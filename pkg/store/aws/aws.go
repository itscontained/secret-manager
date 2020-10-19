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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/aws/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/go-logr/logr"

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	ctxlog "github.com/itscontained/secret-manager/pkg/log"
	"github.com/itscontained/secret-manager/pkg/store"
	"github.com/itscontained/secret-manager/pkg/store/schema"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Client = &AWS{}

const (
	AWSSecretsmanagerEndpoint = "AWS_SECRETSMANAGER_ENDPOINT"
	AWSSTSEndpoint            = "AWS_STS_ENDPOINT"
)

type AWS struct {
	kube      ctrlclient.Client
	store     smv1alpha1.GenericStore
	log       logr.Logger
	client    *secretsmanager.Client
	namespace string
}

func init() {
	schema.Register(&AWS{}, &smv1alpha1.SecretStoreSpec{
		AWS: &smv1alpha1.AWSStore{},
	})
}

func (a *AWS) New(ctx context.Context, store smv1alpha1.GenericStore, kube ctrlclient.Client, namespace string) error {
	log := ctxlog.FromContext(ctx)
	a = &AWS{
		kube:      kube,
		store:     store,
		log:       log,
		namespace: namespace,
	}

	cfg, err := a.newConfig(ctx)
	if err != nil {
		return err
	}

	a.client = secretsmanager.New(*cfg)
	return nil
}

func (a *AWS) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	version := ""
	if ref.Version != nil {
		version = *ref.Version
	}
	data, err := a.readSecret(ctx, ref.Name, version)
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
	return a.readSecret(ctx, ref.Name, version)
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
		return nil, fmt.Errorf("error getting secret value: %w", err)
	}
	smData := make(map[string]string)
	err = json.Unmarshal([]byte(*resp.SecretString), &smData)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal secret value: %w", err)
	}
	secretData := make(map[string][]byte)
	for k, v := range smData {
		secretData[k] = []byte(v)
	}
	return secretData, nil
}

func (a *AWS) newConfig(ctx context.Context) (*aws.Config, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	cfg.EndpointResolver = &EndpointResolver{}
	spec := *a.store.GetSpec().AWS
	if spec.Region != nil {
		cfg.Region = *spec.Region
	}
	if spec.AuthSecretRef == nil {
		return &cfg, nil
	}
	scoped := true
	if a.store.GetTypeMeta().String() == "ClusterSecretStore" {
		scoped = false
	}
	if spec.AuthSecretRef.AccessKeyID == nil || spec.AuthSecretRef.SecretAccessKey == nil {
		return nil, fmt.Errorf("missing accessKeyID/secretAccessKey in store config")
	}
	aKid, err := a.secretKeyRef(ctx, a.store.GetNamespace(), *spec.AuthSecretRef.AccessKeyID, scoped)
	if err != nil {
		return nil, err
	}
	sak, err := a.secretKeyRef(ctx, a.store.GetNamespace(), *spec.AuthSecretRef.SecretAccessKey, scoped)
	if err != nil {
		return nil, err
	}
	nScp := aws.NewStaticCredentialsProvider(aKid, sak, "secret-manager")
	cfg.Credentials = nScp
	if spec.AuthSecretRef.Role != nil {
		role, err := a.secretKeyRef(ctx, a.store.GetNamespace(), *spec.AuthSecretRef.Role, scoped)
		if err != nil {
			return nil, err
		}
		stsClient := sts.New(cfg)
		stsCp := stscreds.NewAssumeRoleProvider(stsClient, role)
		cfg.Credentials = stsCp
	}
	return &cfg, nil
}

func (a *AWS) secretKeyRef(ctx context.Context, namespace string, secretRef smmeta.SecretKeySelector, scoped bool) (string, error) {
	var secret corev1.Secret
	ref := types.NamespacedName{
		Namespace: namespace,
		Name:      secretRef.Name,
	}
	if !scoped && secretRef.Namespace != nil {
		ref.Namespace = *secretRef.Namespace
	}
	err := a.kube.Get(ctx, ref, &secret)
	if err != nil {
		return "", err
	}
	keyBytes, ok := secret.Data[secretRef.Key]
	if !ok {
		return "", fmt.Errorf("no data for %q in secret '%s/%s'", secretRef.Key, secretRef.Name, namespace)
	}
	value := strings.TrimSpace(string(keyBytes))
	return value, nil
}

// EndpointResolver resolves custom endpoints for aws services
type EndpointResolver struct {
	res endpoints.Resolver
}

// ResolveEndpoint resolves custom endpoints if provided
func (r *EndpointResolver) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
	if ep := os.Getenv(AWSSecretsmanagerEndpoint); ep != "" {
		if service == "secretsmanager" {
			return aws.Endpoint{
				URL: ep,
			}, nil
		}
	}
	if ep := os.Getenv(AWSSTSEndpoint); ep != "" {
		if service == "sts" {
			return aws.Endpoint{
				URL: ep,
			}, nil
		}
	}
	return r.res.ResolveEndpoint(service, region)
}
