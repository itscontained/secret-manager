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

package base

import (
	"context"
	"fmt"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	store "github.com/itscontained/secret-manager/pkg/internal/store"
	vault "github.com/itscontained/secret-manager/pkg/internal/vault"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Factory = &Default{}

type Default struct{}

func (f *Default) New(ctx context.Context, store smv1alpha1.GenericStore, kubeClient client.Client, namespace string) (store.Client, error) {
	storeSpec := store.GetSpec()
	if storeSpec.Vault != nil {
		vaultClient, err := vault.New(ctx, kubeClient, store, namespace)
		if err != nil {
			return nil, fmt.Errorf("unable to setup Vault client: %w", err)
		}
		return vaultClient, nil
	}

	return nil, fmt.Errorf("SecretStore %q does not have a valid client", store.GetName())
}
