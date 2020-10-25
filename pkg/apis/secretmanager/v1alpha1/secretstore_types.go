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
	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretStoreSpec defines the authentication methods used
type SecretStoreSpec struct {
	// Vault configures this store to sync secrets using a HashiCorp Vault
	// KV backend.
	// +optional
	Vault *VaultStore `json:"vault,omitempty"`
	// AWS configures this store to sync secrets using AWS SecretManager
	// +optional
	AWS *AWSStore `json:"aws,omitempty"`
	// GCP configures this store to sync secrets using GCP Secret Manager
	// +optional
	GCP *GCPStore `json:"gcp,omitempty"`
}

type SecretStoreStatus struct {
	// List of status conditions to indicate the status of SecretStore.
	// Known condition types are `Ready`.
	// +optional
	Conditions smmeta.ConditionedStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// SecretStore represents a secure external location for storing secrets, which can be referenced as part of `storeRef` fields
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={secretmanager},shortName=ss
type SecretStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SecretStoreSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// SecretStoreList contains a list of SecretStore
type SecretStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalSecret `json:"items"`
}

// +kubebuilder:object:root=true

// ClusterSecretStore represents a secure external location for storing secrets, which can be referenced as part of `storeRef` fields
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={secretmanager},shortName=css
type ClusterSecretStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SecretStoreSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterSecretStoreList contains a list of ClusterSecretStore
type ClusterSecretStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalSecret `json:"items"`
}
