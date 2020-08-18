package vault

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strings"

	vault "github.com/hashicorp/vault/api"

	smv1alpha1 "github.com/mcavoyk/secret-manager/pkg/apis/secretmanager/v1alpha1"

	corelisters "k8s.io/client-go/listers/core/v1"
)

type Interface interface {
	GetSecret(ctx context.Context, path, key, version string) (string, error)
}

type Client interface {
	NewRequest(method, requestPath string) *vault.Request
	RawRequest(r *vault.Request) (*vault.Response, error)
	SetToken(v string)
	Token() string
}

type Vault struct {
	secretsLister corelisters.SecretLister
	store         smv1alpha1.GenericStore
	namespace     string

	client Client
}

func New(namespace string, secretsLister corelisters.SecretLister, store smv1alpha1.GenericStore) (Interface, error) {
	v := &Vault{
		secretsLister: secretsLister,
		namespace:     namespace,
		store:         store,
	}

	cfg, err := v.newConfig()
	if err != nil {
		return nil, err
	}

	client, err := vault.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing Vault client: %s", err.Error())
	}

	if err := v.setToken(client); err != nil {
		return nil, err
	}

	v.client = client

	return v, nil
}

func (v *Vault) GetSecret(ctx context.Context, path, key, version string) (string, error) {
	// TODO: implement
	return "", nil
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

func (v *Vault) setToken(client Client) error {
	tokenRef := v.store.GetSpec().Vault.Auth.TokenSecretRef
	if tokenRef != nil {
		token, err := v.secretKeyRef(v.namespace, tokenRef.Name, tokenRef.Key)
		if err != nil {
			return err
		}
		client.SetToken(token)

		return nil
	}

	appRole := v.store.GetSpec().Vault.Auth.AppRole
	if appRole != nil {
		token, err := v.requestTokenWithAppRoleRef(client, appRole)
		if err != nil {
			return err
		}
		client.SetToken(token)

		return nil
	}

	kubernetesAuth := v.store.GetSpec().Vault.Auth.Kubernetes
	if kubernetesAuth != nil {
		token, err := v.requestTokenWithKubernetesAuth(client, kubernetesAuth)
		if err != nil {
			return fmt.Errorf("error reading Kubernetes service account token from %s: %s", kubernetesAuth.SecretRef.Name, err.Error())
		}
		client.SetToken(token)
		return nil
	}

	return fmt.Errorf("error initializing Vault client: tokenSecretRef, appRoleSecretRef, or Kubernetes auth role not set")
}

func (v *Vault) secretKeyRef(namespace, name, key string) (string, error) {
	secret, err := v.secretsLister.Secrets(namespace).Get(name)
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

func (v *Vault) requestTokenWithAppRoleRef(client Client, appRole *smv1alpha1.VaultAppRole) (string, error) {
	roleID := strings.TrimSpace(appRole.RoleID)

	secretID, err := v.secretKeyRef(v.namespace, appRole.SecretRef.Name, appRole.SecretRef.Key)
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

	resp, err := client.RawRequest(request)
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

func (v *Vault) requestTokenWithKubernetesAuth(client Client, kubernetesAuth *smv1alpha1.VaultKubernetesAuth) (string, error) {
	secret, err := v.secretsLister.Secrets(v.namespace).Get(kubernetesAuth.SecretRef.Name)
	if err != nil {
		return "", err
	}

	key := kubernetesAuth.SecretRef.Key
	keyBytes, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("no data for %q in secret '%s/%s'", key, v.namespace, kubernetesAuth.SecretRef.Name)
	}

	jwt := string(keyBytes)

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

	resp, err := client.RawRequest(request)
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
