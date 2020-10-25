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

// ExternalSecretSpec defines the desired state of ExternalSecret
type ExternalSecretSpec struct {

	// StoreRef is a reference to the store backend for this secret.
	// If the 'kind' field is not set, or set to 'SecretStore', a SecretStore resource
	// with the given name in the same namespace as the SecretStore will be used.
	// If the 'kind' field is set to 'ClusterSecretStore', a ClusterSecretStore with the
	// provided name will be used.
	// The 'name' field in this stanza is required at all times.
	StoreRef ObjectReference `json:"storeRef"`

	// Template which will be deep merged into the generated secret.
	// Can be used to set for example annotations or type on the generated secret.
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:Format=any
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Template []byte `json:"template,omitempty"`

	// Data is a list of references to secret values.
	// +optional
	Data []KeyReference `json:"data,omitempty"`

	// DataFrom references a map of secrets to embed within the generated secret.
	// +optional
	DataFrom []RemoteReference `json:"dataFrom,omitempty"`
}

// ObjectReference is a reference to an object with a given name, kind and group.
type ObjectReference struct {
	// Name of the resource being referred to.
	Name string `json:"name"`

	// Kind of the resource being referred to.
	// +optional
	Kind string `json:"kind,omitempty"`

	// Group of the resource being referred to.
	// +optional
	Group string `json:"group,omitempty"`
}

type KeyReference struct {
	// The key in the generated secret to place fetched secret value into.
	SecretKey string `json:"secretKey"`
	// RemoteRef describes the path and other parameters to access the secret for the specific SecretStore
	RemoteRef RemoteReference `json:"remoteRef"`
}

type RemoteReference struct {
	// Name of the key, path, or id in the SecretStore.
	Name string `json:"name"`

	// Property to extract secret value at path in the SecretStore.
	// Can be omitted if not supported by SecretStore or if entire secret should
	// be fetched as in dataFrom reference.
	// +optional
	Property *string `json:"property,omitempty"`

	// Version of the secret to fetch from the SecretStore. Must be a supported parameter
	// by the referenced SecretStore.
	// +optional
	Version *string `json:"version,omitempty"`
}

// ExternalSecretStatus defines the observed state of ExternalSecret
type ExternalSecretStatus struct {
	// List of status conditions to indicate the status of ExternalSecret.
	// Known condition types are `Ready`.
	smmeta.ConditionedStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// ExternalSecret resource can be created which will ensure a Secret is available with the referenced contents of the external store
// +kubebuilder:printcolumn:name="LAST SYNC",type="date",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="STORE",type="string",JSONPath=".spec.storeRef.name",priority=1
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={secretmanager},shortName=es
type ExternalSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExternalSecretSpec   `json:"spec,omitempty"`
	Status ExternalSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExternalSecretList contains a list of ExternalSecret
type ExternalSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalSecret `json:"items"`
}
