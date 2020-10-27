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
	"time"

	vault "github.com/hashicorp/vault/api"

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (v *Vault) CreateSecretStoreVaultKubeAuth() (*smv1alpha1.SecretStoreSpec, error) {
	randomSuffix := uuid.NewUUID()
	v.kubePath = fmt.Sprintf("kubernetes-%s", randomSuffix)
	v.kubeRole = fmt.Sprintf("role-%s", randomSuffix)

	// Create Kubernetes auth backend
	err := v.vaultClient.Sys().EnableAuthWithOptions(v.kubePath, &vault.EnableAuthOptions{
		Type: "kubernetes",
	})
	if err != nil {
		return nil, err
	}

	// Get token reviewer jwt for vault
	tokenReviewerAccount := &corev1.ServiceAccount{}
	v.KubeClient.Get(context.Background(), types.NamespacedName{
		Namespace: v.Namespace,
		Name:      serviceAccountName,
	}, tokenReviewerAccount)

	if len(tokenReviewerAccount.Secrets) == 0 {
		return nil, errors.New("vault serviceaccount has no associated secret")
	}
	tokenSecret := &corev1.Secret{}
	v.KubeClient.Get(context.Background(), types.NamespacedName{
		Namespace: v.Namespace,
		Name:      tokenReviewerAccount.Secrets[0].Name,
	}, tokenSecret)

	kubeReq := v.vaultClient.NewRequest("POST", fmt.Sprintf("/v1/auth/%s/config", v.kubePath))
	kubeData := map[string]interface{}{
		"kubernetes_host":    "https://kubernetes.default",
		"token_reviewer_jwt": string(tokenSecret.Data["token"]),
	}
	kubeReq.SetJSONBody(kubeData)
	_, err = v.vaultClient.RawRequest(kubeReq)
	if err != nil {
		return nil, err
	}

	// Create vault role for Kubernetes auth
	svcAccountName := fmt.Sprintf("kv-reader-%s", randomSuffix)
	req := v.vaultClient.NewRequest("POST", fmt.Sprintf("/v1/auth/%s/role/%s", v.kubePath, v.kubeRole))
	roleData := map[string]interface{}{
		"bound_service_account_names":      []string{svcAccountName},
		"bound_service_account_namespaces": []string{v.Namespace},
		"token_policies":                   []string{v.secretRole},
	}
	req.SetJSONBody(roleData)
	_, err = v.vaultClient.RawRequest(req)
	if err != nil {
		return nil, err
	}

	err = v.KubeClient.Create(context.Background(), &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcAccountName,
			Namespace: v.Namespace,
		},
	})
	if err != nil {
		return nil, err
	}

	svcAccountCreated := &corev1.ServiceAccount{}
	svcAccountKey := types.NamespacedName{
		Name:      svcAccountName,
		Namespace: v.Namespace,
	}

	err = try(3, time.Second*2, func() error {
		err = v.KubeClient.Get(context.Background(), svcAccountKey, svcAccountCreated)
		if err != nil {
			return err
		}

		if len(svcAccountCreated.Secrets) < 1 {
			return errors.New("created service account has no token secret")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &smv1alpha1.SecretStoreSpec{
		Vault: &smv1alpha1.VaultStore{
			Server: v.Host,
			Path:   v.kvPath,
			Auth: smv1alpha1.VaultAuth{
				Kubernetes: &smv1alpha1.VaultKubernetesAuth{
					Path: v.kubePath,
					Role: v.kubeRole,
					SecretRef: &smmeta.SecretKeySelector{
						LocalObjectReference: smmeta.LocalObjectReference{
							Name: svcAccountCreated.Secrets[0].Name,
						},
						Key: "token",
					},
				},
			},
		},
	}, nil
}

// try attempts a function for n retries, with pollPeriod waiting in-between
// when the function returns no error, nil is return. Error is returned
// after retries with error wrapped.
func try(retries int, pollPeriod time.Duration, f func() error) error {
	attempt := 0
	var err error
	for attempt < retries {
		err = f()
		if err == nil {
			return nil
		}
		time.Sleep(pollPeriod)
		attempt++
	}

	return fmt.Errorf("retry attempts failed: %w", err)
}
