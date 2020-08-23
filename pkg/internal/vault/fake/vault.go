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

package fake

import (
	"context"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Vault struct {
	NewFn          func(context.Context, ctrlclient.Client, smv1alpha1.GenericStore, string) (*Vault, error)
	GetSecretFn    func(context.Context, smv1alpha1.RemoteReference) ([]byte, error)
	GetSecretMapFn func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error)
}

func New() *Vault {
	v := &Vault{
		GetSecretFn: func(context.Context, smv1alpha1.RemoteReference) ([]byte, error) {
			return nil, nil
		},
		GetSecretMapFn: func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error) {
			return nil, nil
		},
	}

	v.NewFn = func(context.Context, ctrlclient.Client, smv1alpha1.GenericStore, string) (*Vault, error) {
		return v, nil
	}

	return v
}

func (v *Vault) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	return v.GetSecretFn(ctx, ref)
}

func (v *Vault) WithGetSecret(secData []byte, err error) *Vault {
	v.GetSecretFn = func(context.Context, smv1alpha1.RemoteReference) ([]byte, error) {
		return secData, err
	}
	return v
}

func (v *Vault) GetSecretMap(ctx context.Context, ref smv1alpha1.RemoteReference) (map[string][]byte, error) {
	return v.GetSecretMapFn(ctx, ref)
}

func (v *Vault) WithGetSecretMap(secData map[string][]byte, err error) *Vault {
	v.GetSecretMapFn = func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error) {
		return secData, err
	}
	return v
}

func (v *Vault) WithNew(f func(context.Context, ctrlclient.Client, smv1alpha1.GenericStore, string) (*Vault, error)) *Vault {
	v.NewFn = f
	return v
}

func (v *Vault) New(ctx context.Context, kubeClient ctrlclient.Client, store smv1alpha1.GenericStore, ns string) (*Vault, error) {
	_, err := v.NewFn(ctx, kubeClient, store, ns)
	if err != nil {
		return nil, err
	}

	return v, nil
}
