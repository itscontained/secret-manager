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

// Configures an store to sync secrets using AWS SecretManager
type AWSStore struct {
	// Region configures the region to send requests to.
	// +optional
	Region *string `json:"region"`
	// Auth configures how secret-manager authenticates with AWS.
	// +optional
	AuthSecretRef *AWSAuth `json:"authSecretRef"`
}

// Configuration used to authenticate with AWS.
// Any of `AccessKeyID`, `SecretAccessKey` or `Role` can be specified. If not set we fall-back to using env vars, shared
// credentials file or AWS Instance metadata
type AWSAuth struct {
	// The AccessKeyID is used for authentication. If not set we fall-back to using env vars, shared credentials file
	// or AWS Instance metadata
	// see: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials
	// +optional
	AccessKeyID *smmeta.SecretKeySelector `json:"accessKeyID,omitempty"`
	// The SecretAccessKey is used for authentication. If not set we fall-back to using env vars, shared credentials file
	// or AWS Instance metadata
	// see: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials
	// +optional
	SecretAccessKey *smmeta.SecretKeySelector `json:"secretAccessKey,omitempty"`
	// Role is a Role ARN which the SecretManager provider will assume using either the explicit credentials
	// AccessKeyID/SecretAccessKey or the inferred credentials from environment variables, shared credentials
	// file or AWS Instance metadata
	// +optional
	Role *smmeta.SecretKeySelector `json:"role,omitempty"`
}
