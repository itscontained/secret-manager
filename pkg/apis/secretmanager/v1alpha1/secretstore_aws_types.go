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
)

// Configures an store to sync secrets using AWS SecretManager
type AWSStore struct {
	// Auth configures how secret-manager authenticates with AWS.
	// +optional
	Auth *AWSAuth `json:"auth"`
	// Region configures the region to send requests to.
	// +optional
	Region *string `json:"region"`
}

// Configuration used to authenticate with AWS.
// Only one of `Credentials`, or `SecretRef` may be specified.
type AWSAuth struct {
	// Credentials authenticates with AWS using an AccessKeyID, and SecretAccessKey
	// +optional
	Credentials *AWSCredentials `json:"credentials,omitempty"`
	// SecretRef authenticates with AWS using an AccessKeyID, and SecretAccessKey stored in a Secret
	// +optional
	SecretRef *AWSSecretRef `json:"secretRef,omitempty"`
}

type AWSSecretRef struct {
	// AccessKeyID Secret key for AWS Access key ID
	// +optional
	AccessKeyID *smmeta.SecretKeySelector `json:"accessKeyID,omitempty"`
	// SecretAccessKey Secret key for WS Secret Access Key
	// +optional
	SecretAccessKey *smmeta.SecretKeySelector `json:"secretAccessKey,omitempty"`
}

// AWSCredentials authenticates with AWS using an AccessKeyID, and SecretAccessKey
type AWSCredentials struct {
	// AWS Access key ID
	// +optional
	AccessKeyID *string `json:"accessKeyID,omitempty"`
	// AWS Secret Access Key
	// +optional
	SecretAccessKey *string `json:"secretAccessKey,omitempty"`
}
