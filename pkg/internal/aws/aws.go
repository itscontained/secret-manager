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

package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	vault "github.com/hashicorp/vault/api"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/internal/store"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Client = &AWS{}

type Client interface {
	NewRequest(method, requestPath string) *vault.Request
	RawRequestWithContext(ctx context.Context, r *vault.Request) (*vault.Response, error)
	SetToken(v string)
	Token() string
}

type AWS struct {
	kubeClient ctrlclient.Client
	store      smv1alpha1.GenericStore
	namespace  string

	client *secretsmanager.Client
}

func New(ctx context.Context, kubeclient ctrlclient.Client, store smv1alpha1.GenericStore, namespace string) (store.Client, error) {
	v := &AWS{
		kubeClient: kubeclient,
		namespace:  namespace,
		store:      store,
	}

	cfg, err := v.newConfig()
	if err != nil {
		return nil, err
	}

	v.client = secretsmanager.New(*cfg)
	return v, nil
}

func (a *AWS) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	data, err := a.readSecret(ctx, *ref.ID, *ref.Version)
	if err != nil {
		return nil, err
	}
	property := ""
	if ref.Property != nil {
		property = *ref.Property
	}
	value, exists := data[property]
	if !exists {
		return nil, fmt.Errorf("property %q not found in secret response", property)
	}
	return value, nil
}

func (a *AWS) GetSecretMap(ctx context.Context, ref smv1alpha1.RemoteReference) (map[string][]byte, error) {
	version := ""
	if ref.Version != nil {
		version = *ref.Version
	}

	return a.readSecret(ctx, *ref.ID, version)
}

func (a *AWS) readSecret(ctx context.Context, id, version string) (map[string][]byte, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(id),
	}
	if version != "" {
		input.VersionStage = aws.String(version)
	}
	req := a.client.GetSecretValueRequest(input)
	resp, err := req.Send(ctx)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			case secretsmanager.ErrCodeInvalidParameterException:
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())
			case secretsmanager.ErrCodeInvalidRequestException:
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())
			case secretsmanager.ErrCodeDecryptionFailure:
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())
			case secretsmanager.ErrCodeInternalServiceError:
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil, err
	}

	secretData := resp.String()
	fmt.Println(secretData)
	byteMap := make(map[string][]byte)
	return byteMap, nil
}

func (a *AWS) newConfig() (*aws.Config, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	if a.store.GetSpec().AWS.Region != "" {
		cfg.Region = a.store.GetSpec().AWS.Region
	}

	return &cfg, nil
}
