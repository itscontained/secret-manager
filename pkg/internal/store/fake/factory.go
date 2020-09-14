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

	"github.com/go-logr/logr"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/internal/store"

	client "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Factory = &Factory{}

type Factory struct {
	NewFn          func(context.Context, client.Client, smv1alpha1.GenericStore, string) (*Factory, error)
	GetSecretFn    func(context.Context, smv1alpha1.RemoteReference) ([]byte, error)
	GetSecretMapFn func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error)
}

func New() *Factory {
	v := &Factory{
		GetSecretFn: func(context.Context, smv1alpha1.RemoteReference) ([]byte, error) {
			return nil, nil
		},
		GetSecretMapFn: func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error) {
			return nil, nil
		},
	}

	v.NewFn = func(context.Context, client.Client, smv1alpha1.GenericStore, string) (*Factory, error) {
		return v, nil
	}

	return v
}

func (v *Factory) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	return v.GetSecretFn(ctx, ref)
}

func (v *Factory) WithGetSecret(secData []byte, err error) *Factory {
	v.GetSecretFn = func(context.Context, smv1alpha1.RemoteReference) ([]byte, error) {
		return secData, err
	}
	return v
}

func (v *Factory) GetSecretMap(ctx context.Context, ref smv1alpha1.RemoteReference) (map[string][]byte, error) {
	return v.GetSecretMapFn(ctx, ref)
}

func (v *Factory) WithGetSecretMap(secData map[string][]byte, err error) *Factory {
	v.GetSecretMapFn = func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error) {
		return secData, err
	}
	return v
}

func (v *Factory) WithNew(f func(context.Context, client.Client, smv1alpha1.GenericStore, string) (*Factory, error)) *Factory {
	v.NewFn = f
	return v
}

func (v *Factory) New(ctx context.Context, store smv1alpha1.GenericStore, kubeClient client.Client, kubeReader client.Reader, namespace string, log logr.Logger) (store.Client, error) {
	_, err := v.NewFn(ctx, kubeClient, store, namespace)
	if err != nil {
		return nil, err
	}

	return v, nil
}
