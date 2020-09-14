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

package gcp

import (
	"context"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/internal/store"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Client = &GCP{}

type GCP struct {
	kubeClient ctrlclient.Client
	store      smv1alpha1.GenericStore

	client *secretmanager.Client
}

func New(ctx context.Context, kubeClient ctrlclient.Client, store smv1alpha1.GenericStore) (store.Client, error) {
	g := &GCP{
		kubeClient: kubeClient,
		store:      store,
	}
	err := g.newClient(ctx)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *GCP) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	version := "latest"
	if ref.Version != nil {
		version = *ref.Version
	}
	data, err := g.readSecret(ctx, *ref.Name, version)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", data)
	return data["data"], nil
}

func (g *GCP) GetSecretMap(ctx context.Context, ref smv1alpha1.RemoteReference) (map[string][]byte, error) {
	version := "latest"
	if ref.Version != nil {
		version = *ref.Version
	}
	return g.readSecret(ctx, *ref.Name, version)
}

func (g *GCP) readSecret(ctx context.Context, id, version string) (map[string][]byte, error) {
	projectId := g.store.GetSpec().GCP.AuthSecretRef.ProjectID
	name := id
	if !strings.HasPrefix(id, "projects/") && projectId != nil {
		name = fmt.Sprintf("projects/%s/secrets/%s/versions/%s", *projectId, id, version)
	}
	req := &secretmanagerpb.AccessSecretVersionRequest{Name: name}
	resp, err := g.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	data := string(resp.Payload.Data)
	secretData := make(map[string][]byte)
	secretData["data"] = []byte(data)
	return secretData, nil
}

// secret name test-secret
func (g *GCP) newClient(ctx context.Context) error {
	var err error
	var clientOption option.ClientOption
	spec := *g.store.GetSpec().GCP
	if spec.AuthSecretRef == nil {
		g.client, err = secretmanager.NewClient(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	if spec.AuthSecretRef.JSON != nil && spec.AuthSecretRef.File != nil {
		return fmt.Errorf("multiple authentication methods configured")
	}
	if spec.AuthSecretRef.File != nil {
		clientOption = option.WithCredentialsFile(*spec.AuthSecretRef.File)
	}
	scoped := true
	fmt.Print(g.store.GetTypeMeta().String())
	if g.store.GetTypeMeta().String() == "ClusterSecretStore" {
		scoped = false
	}
	if spec.AuthSecretRef.JSON != nil {
		data, err := g.secretKeyRef(ctx, g.store.GetNamespace(), *spec.AuthSecretRef.JSON, scoped)
		if err != nil {
			return err
		}
		clientOption = option.WithCredentialsJSON([]byte(data))
	}
	g.client, err = secretmanager.NewClient(ctx, clientOption)
	if err != nil {
		return err
	}
	return nil
}

func (g *GCP) secretKeyRef(ctx context.Context, namespace string, secretRef smmeta.SecretKeySelector, scoped bool) (string, error) {
	var secret corev1.Secret
	ref := types.NamespacedName{
		Namespace: namespace,
		Name:      secretRef.Name,
	}
	if !scoped && secretRef.Namespace != nil {
		ref.Namespace = *secretRef.Namespace
	}
	err := g.kubeClient.Get(ctx, ref, &secret)
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
