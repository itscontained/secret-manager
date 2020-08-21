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

package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=false
// +kubebuilder:object:generate:false
// +k8s:deepcopy-gen:interfaces=nil
// +k8s:deepcopy-gen=nil

// GenericStore is a common interface for interacting with ClusterSecretStore
// or a namespaced SecretStore
type GenericStore interface {
	runtime.Object
	metav1.Object

	GetObjectMeta() *metav1.ObjectMeta
	GetSpec() *SecretStoreSpec
	GetStatus() *SecretStoreStatus
}

// +kubebuilder:object:root:false
// +kubebuilder:object:generate:false
var _ GenericStore = &SecretStore{}

// +kubebuilder:object:root:false
// +kubebuilder:object:generate:false
var _ GenericStore = &ClusterSecretStore{}

func (c *ClusterSecretStore) GetObjectMeta() *metav1.ObjectMeta {
	return &c.ObjectMeta
}
func (c *ClusterSecretStore) GetSpec() *SecretStoreSpec {
	return &c.Spec
}
func (c *ClusterSecretStore) GetStatus() *SecretStoreStatus {
	return &c.Status
}
func (c *ClusterSecretStore) SetSpec(spec SecretStoreSpec) {
	c.Spec = spec
}
func (c *ClusterSecretStore) SetStatus(status SecretStoreStatus) {
	c.Status = status
}
func (c *ClusterSecretStore) Copy() GenericStore {
	return c.DeepCopy()
}

func (c *SecretStore) GetObjectMeta() *metav1.ObjectMeta {
	return &c.ObjectMeta
}
func (c *SecretStore) GetSpec() *SecretStoreSpec {
	return &c.Spec
}
func (c *SecretStore) GetStatus() *SecretStoreStatus {
	return &c.Status
}
func (c *SecretStore) SetSpec(spec SecretStoreSpec) {
	c.Spec = spec
}
func (c *SecretStore) SetStatus(status SecretStoreStatus) {
	c.Status = status
}
func (c *SecretStore) Copy() GenericStore {
	return c.DeepCopy()
}

// +kubebuilder:object:root=false
// +kubebuilder:object:generate:false
// +k8s:deepcopy-gen:interfaces=nil
// +k8s:deepcopy-gen=nil

// StoreClient is a common interface for interacting with SecretStore
// backends
type StoreClient interface {
	GetSecret(ctx context.Context, ref RemoteReference) ([]byte, error)
	GetSecretMap(ctx context.Context, ref RemoteReference) (map[string][]byte, error)
}
