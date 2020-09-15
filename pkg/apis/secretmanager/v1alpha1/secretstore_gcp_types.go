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

import smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"

// Configures an store to sync secrets using GCP Secret Manager
type GCPStore struct {
	// ProjectID is a convenience string to allow the shortening of secret paths.
	// When set, the prefix projects/<ProjectID> can be removed from the name
	ProjectID *string `json:"projectID,omitempty"`
	// Auth configures how secret-manager authenticates with GCP Secret Manager.
	// +optional
	AuthSecretRef *GCPAuth `json:"authSecretRef,omitempty"`
}

// Configuration used to authenticate with GCP.
// Either of `JSON` or `File` can be specified. If not set we fall-back to using
// `GOOGLE_APPLICATION_CREDENTIALS` or the default service account of the compute engine
// see: https://cloud.google.com/docs/authentication/production
type GCPAuth struct {
	// The JSON secret key selector is used for authentication. If not set we fall-back to using
	// `GOOGLE_APPLICATION_CREDENTIALS` or the default service account of the compute engine
	// see: https://cloud.google.com/docs/authentication/production
	// +optional
	JSON *smmeta.SecretKeySelector `json:"json,omitempty"`
	// The File string is used for authentication using a gcp credentials json file. If not set we fall-back to using
	// `GOOGLE_APPLICATION_CREDENTIALS` or the default service account of the compute engine
	// see: https://cloud.google.com/docs/authentication/production
	// +optional
	File *string `json:"file,omitempty"`
}
