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
	smmeta "github.com/mcavoyk/secret-manager/pkg/apis/meta/v1"

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

	// The amount of time before the secret-manager will renew the values of
	// of the ExternalSecret. If not set the secret will only be synced on
	// creation of the ExternalSecret.
	// +kubebuilder:validation:Format=duration
	// +optional
	RenewAfter *metav1.Duration `json:"renewAfter,omitempty"`

	// Template which will be deep merged into the generated secret.
	// Can be used to set for example annotations or type on the generated secret.
	// +optional
	Template *JSON `json:"template,omitempty"`

	// Data is a list of references to secrets values
	Data []KeyReference `json:"data"`

	// DataFrom refernces a map of secrets to embed within the generated secret.
	// +optional
	DataFrom *RemoteReference `json:"dataFrom"`
}

// JSON represents any valid JSON value.
type JSON struct {
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:Format=any
	Raw []byte `json:",inline"`
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
	// The key in the generated secret to place fetched value into. If not specified
	// and the SecretStore value has multiple keys, all keys will be placed into secret with
	// the key name from the source. If omitted and only a single value is present in SecretManager
	// value, the data will be placed under the 'secret' key in the generated secret.
	SecretKey *string `json:"secretKey,omitempty"`

	// RemoteRef describes the path and other parameters to access the secret for the specific SecretManager
	RemoteRef RemoteReference `json:"remoteRef"`
}

type RemoteReference struct {
	// Path to the key in the SecretManager
	Path string `json:"path"`

	// Property to extract secret value at path in SecretManager if path specifies multiple
	// secret values. Can be omitted if not supported by SecretManager or if entire secret should
	// be fetched.
	// +optional
	Property *string `json:"property,omitempty"`

	// Version of the secret to fetch from the SecretManager.
	// +optional
	Version *string `json:"version,omitempty"`
}

// ExternalSecretStatus defines the observed state of ExternalSecret
type ExternalSecretStatus struct {
	// List of status conditions to indicate the status of ExternalSecret.
	// Known condition types are `Ready`.
	// +optional
	Conditions smmeta.ConditionedStatus `json:",inline"`

	// RenewalTime is the time at which the secret values will be next
	// renewed.
	// If not set, no upcoming renewal is scheduled.
	// +optional
	RenewalTime *metav1.Time `json:"renewalTime,omitempty"`
}

// +kubebuilder:object:root=true

// ExternalSecret is the Schema for the externalsecrets API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={secretmanager}
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
