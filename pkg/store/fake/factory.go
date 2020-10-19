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
	"github.com/itscontained/secret-manager/pkg/store"
	"github.com/itscontained/secret-manager/pkg/store/schema"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Client = &Client{}

type Client struct {
	NewFn func(context.Context, smv1alpha1.GenericStore, client.Client,
		string) error
	GetSecretFn    func(context.Context, smv1alpha1.RemoteReference) ([]byte, error)
	GetSecretMapFn func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error)
}

func New() *Client {
	v := &Client{
		GetSecretFn: func(context.Context, smv1alpha1.RemoteReference) ([]byte, error) {
			return nil, nil
		},
		GetSecretMapFn: func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error) {
			return nil, nil
		},
	}

	v.NewFn = func(context.Context, smv1alpha1.GenericStore, client.Client, string) error {
		return nil
	}

	return v
}

func (v *Client) RegisterAs(storeSpec *smv1alpha1.SecretStoreSpec) {
	schema.ForceRegister(v, storeSpec)
}

func (v *Client) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	return v.GetSecretFn(ctx, ref)
}

func (v *Client) WithGetSecret(secData []byte, err error) *Client {
	v.GetSecretFn = func(context.Context, smv1alpha1.RemoteReference) ([]byte, error) {
		return secData, err
	}
	return v
}

func (v *Client) GetSecretMap(ctx context.Context, ref smv1alpha1.RemoteReference) (map[string][]byte, error) {
	return v.GetSecretMapFn(ctx, ref)
}

func (v *Client) WithGetSecretMap(secData map[string][]byte, err error) *Client {
	v.GetSecretMapFn = func(context.Context, smv1alpha1.RemoteReference) (map[string][]byte, error) {
		return secData, err
	}
	return v
}

func (v *Client) WithNew(f func(context.Context, smv1alpha1.GenericStore, client.Client,
	string) error) *Client {
	v.NewFn = f
	return v
}

func (v *Client) New(ctx context.Context, store smv1alpha1.GenericStore, kube client.Client, namespace string) error {
	err := v.NewFn(ctx, store, kube, namespace)
	if err != nil {
		return err
	}

	return nil
}
