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

package framework

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/onsi/ginkgo"
)

const (
	localstackDeploy = "/k8s/aws/deploy.sh"
)

// CreateAWSSecretsManagerSecret creates a sm secret with the given value
func CreateAWSSecretsManagerSecret(namespace, name, secret string) error {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return err
	}
	cfg.Region = "us-east-1"
	cfg.Credentials = aws.NewStaticCredentialsProvider("foobar", "foobar", "secret-manager")
	cfg.EndpointResolver = &localResolver{namespace: namespace}
	sm := secretsmanager.New(cfg)
	req := sm.CreateSecretRequest(&secretsmanager.CreateSecretInput{
		Name:         aws.String(name),
		SecretString: aws.String(secret),
	})
	_, err = req.Send(context.Background())
	return err
}

// localResolver resolves endpoints to e2e localstack
type localResolver struct {
	endpoints.Resolver
	namespace string
}

// ResolveEndpoint resolves custom endpoints if provided
func (r *localResolver) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
	return aws.Endpoint{
		URL: fmt.Sprintf("http://localstack.%s", r.namespace),
	}, nil
}

// NewLocalstack deploys a fresh localstack instance into the specified namespace
func (f *Framework) NewLocalstack(namespace string) error {
	ginkgo.By("launching localstack")
	cmd := exec.Command(localstackDeploy, namespace, f.HelmValues)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error creating localstack: %v.\nLogs:\n%v", err, string(out))
	}
	return nil
}
