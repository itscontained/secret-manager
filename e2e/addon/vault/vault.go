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
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"

	rbacv1 "k8s.io/api/rbac/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	selector           string = "app=vault"
	serviceAddress     string = "http://vault.%s:8200"
	serviceAccountName string = "vault"
	defaultToken       string = "root"
	secretPolicy       string = `
path "%s/*" {
  capabilities = ["read"]
}
`
)

// Vault describes the configuration details for an instance of Vault
// deployed to the test cluster
type Vault struct {
	// Kubectl is the path to kubectl
	Kubectl string

	// Namespace is the namespace to deploy Vault into
	Namespace string

	// KubeClient is a configured Kubernetes clientset for addons to use.
	KubeClient ctrlclient.Client

	// Host is the hostname that can be used to connect to Vault
	Host string

	// BasePath is root of deployment manifests
	BasePath string

	vaultClient *vault.Client
	kvPath      string
	secretRole  string
	kubePath    string
	kubeRole    string
}

func (v *Vault) Setup() error {
	if v.Kubectl == "" {
		return errors.New("kubectl must be set")
	}

	if v.Namespace == "" {
		return errors.New("namespace must be set")
	}
	v.kvPath = "secret"

	err := Run(fmt.Sprintf("%s apply --timeout 2m -f %s -n %s", v.Kubectl, filepath.Join(v.BasePath, "/k8s/vault/vault.yaml"), v.Namespace))
	if err != nil {
		return err
	}
	// create cluster resources because namespace is not known beforehand
	err = v.createClusterResources()
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	err = Run(fmt.Sprintf("%s wait pod --for=condition=Ready --timeout 2m -n %s -l %s", v.Kubectl, v.Namespace, selector))
	if err != nil {
		return err
	}
	v.Host = fmt.Sprintf(serviceAddress, v.Namespace)
	vaultConfig := vault.DefaultConfig()
	vaultConfig.Address = v.Host
	v.vaultClient, err = vault.NewClient(vaultConfig)
	if err != nil {
		return err
	}
	v.vaultClient.SetToken(defaultToken)

	v.secretRole = "kv-reader"
	err = v.vaultClient.Sys().PutPolicy(v.secretRole, fmt.Sprintf(secretPolicy, v.kvPath))
	return err
}

func Run(cmd string) error {
	splitCmd := strings.Split(cmd, " ")
	c := exec.Command(splitCmd[0], splitCmd[1:]...)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing command: %s\nerr: %w", string(out), err)
	}
	return nil
}

func (v *Vault) CreateSecret(name string, secret map[string]string) error {
	vaultPath := fmt.Sprintf("%s/data/%s", v.kvPath, name)
	vaultData := make(map[string]interface{}, 1)
	vaultData["data"] = secret
	_, err := v.vaultClient.Logical().Write(vaultPath, vaultData)
	return err
}

func (v *Vault) createClusterResources() error {
	vaultCRB := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("vault-auth-%s", v.Namespace),
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "system:auth-delegator",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      serviceAccountName,
				Namespace: v.Namespace,
			},
		},
	}

	err := v.KubeClient.Create(context.Background(), vaultCRB)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}
