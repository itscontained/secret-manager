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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/aws/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/internal/store"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Client = &AWS{}

type AWS struct {
	kubeClient ctrlclient.Client
	store      smv1alpha1.GenericStore

	client *secretsmanager.Client
}

func New(ctx context.Context, kubeClient ctrlclient.Client, store smv1alpha1.GenericStore) (store.Client, error) {
	v := &AWS{
		kubeClient: kubeClient,
		store:      store,
	}

	cfg, err := v.newConfig(ctx)
	if err != nil {
		return nil, err
	}

	v.client = secretsmanager.New(*cfg)
	return v, nil
}

func (a *AWS) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	version := ""
	if ref.Version != nil {
		version = *ref.Version
	}
	data, err := a.readSecret(ctx, *ref.ID, version)
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
		return nil, err
	}
	secretData := make(map[string][]byte)
	err = json.Unmarshal([]byte(*resp.SecretString), &secretData)
	if err != nil {
		return nil, err
	}
	return secretData, nil
}

func (a *AWS) newConfig(ctx context.Context) (*aws.Config, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	spec := *a.store.GetSpec().AWS
	if spec.Region != nil {
		cfg.Region = *spec.Region
	}
	if spec.Auth == nil {
		return &cfg, nil
	}
	scoped := true
	if a.store.GetTypeMeta().String() == "ClusterSecretStore" {
		scoped = false
	}
	if spec.Auth.AccessKeyID == nil || spec.Auth.SecretAccessKey == nil {
		return nil, fmt.Errorf("missing accessKeyID/secretAccessKey in store config")
	}
	aKid, err := a.secretKeyRef(ctx, a.store.GetNamespace(), *spec.Auth.AccessKeyID, scoped)
	if err != nil {
		return nil, err
	}
	sak, err := a.secretKeyRef(ctx, a.store.GetNamespace(), *spec.Auth.SecretAccessKey, scoped)
	if err != nil {
		return nil, err
	}
	nScp := aws.NewStaticCredentialsProvider(aKid, sak, "")
	cfg.Credentials = nScp
	if spec.Auth.Role != nil {
		role, err := a.secretKeyRef(ctx, a.store.GetNamespace(), *spec.Auth.Role, scoped)
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
	err := a.kubeClient.Get(ctx, ref, &secret)
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
