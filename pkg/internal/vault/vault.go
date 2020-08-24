/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vault

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strings"

	vault "github.com/hashicorp/vault/api"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/internal/store"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Client = &Vault{}

type Client interface {
	NewRequest(method, requestPath string) *vault.Request
	RawRequestWithContext(ctx context.Context, r *vault.Request) (*vault.Response, error)
	SetToken(v string)
	Token() string
}

type Vault struct {
	kubeClient ctrlclient.Client
	store      smv1alpha1.GenericStore
	namespace  string

	client Client
}

func New(ctx context.Context, kubeclient ctrlclient.Client, store smv1alpha1.GenericStore, namespace string) (store.Client, error) {
	v := &Vault{
		kubeClient: kubeclient,
		namespace:  namespace,
		store:      store,
	}

	cfg, err := v.newConfig()
	if err != nil {
		return nil, err
	}

	client, err := vault.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing Vault client: %s", err.Error())
	}

	if v.store.GetSpec().Vault.Namespace != nil {
		client.SetNamespace(*v.store.GetSpec().Vault.Namespace)
	}

	if err := v.setToken(ctx, client); err != nil {
		return nil, err
	}

	v.client = client

	return v, nil
}

func (v *Vault) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	version := ""
	if ref.Version != nil {
		version = *ref.Version
	}

	data, err := v.readSecret(ctx, ref.Path, version)
	if err != nil {
		return nil, err
	}
	property := ""
	if ref.Property != nil {
		property = *ref.Property
	}
	value, exists := data[property]
	if !exists {
		return nil, fmt.Errorf("property %q not found in secret response", property)
	}
	return value, nil
}

func (v *Vault) GetSecretMap(ctx context.Context, ref smv1alpha1.RemoteReference) (map[string][]byte, error) {
	version := ""
	if ref.Version != nil {
		version = *ref.Version
	}

	return v.readSecret(ctx, ref.Path, version)
}

func (v *Vault) readSecret(ctx context.Context, path, version string) (map[string][]byte, error) {
	storeSpec := v.store.GetSpec()
	kvPath := storeSpec.Vault.Path
	if !strings.HasSuffix(kvPath, "/data") {
		kvPath = fmt.Sprintf("%s/data", kvPath)
	}

	req := v.client.NewRequest(http.MethodGet, fmt.Sprintf("/v1/%s/%s", kvPath, path))
	if version != "" {
		req.Params.Set("version", version)
	}

	resp, err := v.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	vaultSecret, err := vault.ParseSecret(resp.Body)
	if err != nil {
		return nil, err
	}

	kvVersion := smv1alpha1.DefaultVaultKVEngineVersion
	if storeSpec.Vault.Version != nil {
		kvVersion = *storeSpec.Vault.Version
	}

	secretData := vaultSecret.Data
	if kvVersion == smv1alpha1.DefaultVaultKVEngineVersion {
		dataInt, ok := vaultSecret.Data["data"]
		if !ok {
			return nil, fmt.Errorf("unexpected secret data response")
		}
		secretData, ok = dataInt.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected secret data format")
		}
	}

	byteMap := make(map[string][]byte, len(secretData))
	for k, v := range secretData {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected secret type")
		}
		byteMap[k] = []byte(str)
	}

	return byteMap, nil
}

func (v *Vault) newConfig() (*vault.Config, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = v.store.GetSpec().Vault.Server

	certs := v.store.GetSpec().Vault.CABundle
	if len(certs) == 0 {
		return cfg, nil
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(certs)
	if !ok {
		return nil, fmt.Errorf("error loading Vault CA bundle")
	}

	cfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.RootCAs = caCertPool

	return cfg, nil
}

func (v *Vault) setToken(ctx context.Context, client Client) error {
	tokenRef := v.store.GetSpec().Vault.Auth.TokenSecretRef
	if tokenRef != nil {
		token, err := v.secretKeyRef(ctx, v.namespace, tokenRef.Name, tokenRef.Key)
		if err != nil {
			return err
		}
		client.SetToken(token)

		return nil
	}

	appRole := v.store.GetSpec().Vault.Auth.AppRole
	if appRole != nil {
		token, err := v.requestTokenWithAppRoleRef(ctx, client, appRole)
		if err != nil {
			return err
		}
		client.SetToken(token)

		return nil
	}

	kubernetesAuth := v.store.GetSpec().Vault.Auth.Kubernetes
	if kubernetesAuth != nil {
		token, err := v.requestTokenWithKubernetesAuth(ctx, client, kubernetesAuth)
		if err != nil {
			return fmt.Errorf("error reading Kubernetes service account token from %s: %s", kubernetesAuth.SecretRef.Name, err.Error())
		}
		client.SetToken(token)
		return nil
	}

	return fmt.Errorf("error initializing Vault client: tokenSecretRef, appRoleSecretRef, or Kubernetes auth role not set")
}

func (v *Vault) secretKeyRef(ctx context.Context, namespace, name, key string) (string, error) {
	secret := &corev1.Secret{}
	ref := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := v.kubeClient.Get(ctx, ref, secret)
	if err != nil {
		return "", err
	}

	keyBytes, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("no data for %q in secret '%s/%s'", key, name, namespace)
	}

	value := string(keyBytes)
	valueStr := strings.TrimSpace(value)

	return valueStr, nil
}

func (v *Vault) requestTokenWithAppRoleRef(ctx context.Context, client Client, appRole *smv1alpha1.VaultAppRole) (string, error) {
	roleID := strings.TrimSpace(appRole.RoleID)

	secretID, err := v.secretKeyRef(ctx, v.namespace, appRole.SecretRef.Name, appRole.SecretRef.Key)
	if err != nil {
		return "", err
	}

	parameters := map[string]string{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	authPath := appRole.Path
	if authPath == "" {
		authPath = smv1alpha1.DefaultVaultAppRoleAuthMountPath
	}

	url := strings.Join([]string{"/v1", "auth", authPath, "login"}, "/")
	request := client.NewRequest("POST", url)

	err = request.SetJSONBody(parameters)
	if err != nil {
		return "", fmt.Errorf("error encoding Vault parameters: %s", err.Error())
	}

	resp, err := client.RawRequestWithContext(ctx, request)
	if err != nil {
		return "", fmt.Errorf("error logging in to Vault server: %s", err.Error())
	}

	defer resp.Body.Close()

	vaultResult := vault.Secret{}
	if err = resp.DecodeJSON(&vaultResult); err != nil {
		return "", fmt.Errorf("unable to decode JSON payload: %s", err.Error())
	}

	token, err := vaultResult.TokenID()
	if err != nil {
		return "", fmt.Errorf("unable to read token: %s", err.Error())
	}

	if token == "" {
		return "", errors.New("no token returned")
	}

	return token, nil
}

func (v *Vault) requestTokenWithKubernetesAuth(ctx context.Context, client Client, kubernetesAuth *smv1alpha1.VaultKubernetesAuth) (string, error) {
	key := kubernetesAuth.SecretRef.Key
	jwt, err := v.secretKeyRef(ctx, v.namespace, kubernetesAuth.SecretRef.Name, key)
	if err != nil {
		return "", err
	}

	parameters := map[string]string{
		"role": kubernetesAuth.Role,
		"jwt":  jwt,
	}

	authPath := kubernetesAuth.Path
	if authPath == "" {
		authPath = smv1alpha1.DefaultVaultKubernetesAuthMountPath
	}

	url := strings.Join([]string{"/v1", "auth", authPath, "login"}, "/")
	request := client.NewRequest("POST", url)

	err = request.SetJSONBody(parameters)
	if err != nil {
		return "", fmt.Errorf("error encoding Vault parameters: %s", err.Error())
	}

	resp, err := client.RawRequestWithContext(ctx, request)
	if err != nil {
		return "", fmt.Errorf("error calling Vault server: %s", err.Error())
	}

	defer resp.Body.Close()
	vaultResult := vault.Secret{}
	err = resp.DecodeJSON(&vaultResult)
	if err != nil {
		return "", fmt.Errorf("unable to decode JSON payload: %s", err.Error())
	}

	token, err := vaultResult.TokenID()
	if err != nil {
		return "", fmt.Errorf("unable to read token: %s", err.Error())
	}

	return token, nil
}
