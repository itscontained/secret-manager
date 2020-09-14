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
	"errors"
	"fmt"

	"github.com/go-logr/logr"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/internal/aws"
	"github.com/itscontained/secret-manager/pkg/internal/gcp"
	"github.com/itscontained/secret-manager/pkg/internal/store"
	"github.com/itscontained/secret-manager/pkg/internal/vault"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Factory = &Default{}

type Default struct{}

func (f *Default) New(ctx context.Context, genericStore smv1alpha1.GenericStore, kubeClient client.Client, _ client.Reader, namespace string, log logr.Logger) (store.Client, error) {
	var err error
	var storeClient store.Client
	if genericStore.GetSpec().Vault != nil {
		storeClient, err = vault.New(ctx, kubeClient, genericStore, namespace, log)
	} else if genericStore.GetSpec().AWS != nil {
		storeClient, err = aws.New(ctx, kubeClient, genericStore, log)
	} else if genericStore.GetSpec().GCP != nil {
		storeClient, err = gcp.New(ctx, kubeClient, genericStore, log)
	} else {
		return nil, fmt.Errorf("SecretStore %q does not have a valid client", genericStore.GetName())
	}
	if err != nil {
		return nil, errors.New("unable to setup SecretStore client")
	}
	return storeClient, nil
}
