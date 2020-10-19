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
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	ctxlog "github.com/itscontained/secret-manager/pkg/log"
	"github.com/itscontained/secret-manager/pkg/store"
	"github.com/itscontained/secret-manager/pkg/store/schema"

	"google.golang.org/api/option"
	"google.golang.org/api/secretmanager/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ store.Client = &GCP{}

type GCP struct {
	kube   ctrlclient.Client
	store  smv1alpha1.GenericStore
	log    logr.Logger
	client *secretmanager.Service
}

func init() {
	schema.Register(&GCP{}, &smv1alpha1.SecretStoreSpec{
		GCP: &smv1alpha1.GCPStore{},
	})
}

func (g *GCP) New(ctx context.Context, store smv1alpha1.GenericStore, kube ctrlclient.Client, namespace string) (store.Client, error) {
	log := ctxlog.FromContext(ctx)
	g = &GCP{
		kube:  kube,
		store: store,
		log:   log,
	}
	err := g.newClient(ctx)
	if err != nil {
		log.Error(err, "could not create new gcp client")
		return nil, err
	}
	return g, nil
}

func (g *GCP) GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error) {
	version := "latest"
	if ref.Version != nil {
		version = *ref.Version
	}
	data, err := g.readSecret(ctx, ref.Name, version)
	if err != nil {
		return nil, err
	}
	return data[ref.Name], nil
}

func (g *GCP) GetSecretMap(ctx context.Context, ref smv1alpha1.RemoteReference) (map[string][]byte, error) {
	version := "latest"
	if ref.Version != nil {
		version = *ref.Version
	}
	return g.readSecret(ctx, ref.Name, version)
}

func (g *GCP) readSecret(ctx context.Context, id, version string) (map[string][]byte, error) {
	projectID := g.store.GetSpec().GCP.ProjectID
	name := id
	if !strings.HasPrefix(id, "projects/") && projectID != nil {
		name = fmt.Sprintf("projects/%s/secrets/%s/versions/%s", *projectID, id, version)
	}
	resp, err := g.client.Projects.Secrets.Versions.Access(name).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	data, err := base64.URLEncoding.DecodeString(resp.Payload.Data)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{id: data}, nil
}

func (g *GCP) newClient(ctx context.Context) error {
	g.log.V(1).Info("creating new gcp api client")
	var err error
	var clientOption option.ClientOption
	spec := g.store.GetSpec().GCP
	if spec.AuthSecretRef == nil {
		g.log.V(1).Info("no authentication defined. using environment variables")
		g.client, err = secretmanager.NewService(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	// TODO: Validating Webhook Candidate
	if spec.AuthSecretRef.JSON != nil && spec.AuthSecretRef.FilePath != nil {
		return fmt.Errorf("multiple authentication methods configured")
	}
	if spec.AuthSecretRef.FilePath != nil {
		g.log.V(1).Info("file authentication defined. using %s", *spec.AuthSecretRef.FilePath)
		clientOption = option.WithCredentialsFile(*spec.AuthSecretRef.FilePath)
	}
	scoped := true
	if g.store.GetTypeMeta().Kind == smv1alpha1.ClusterSecretStoreKind {
		g.log.V(1).Info("removing namespace scope restriction")
		scoped = false
	}
	if spec.AuthSecretRef.JSON != nil {
		g.log.V(1).Info("JSON authentication defined")
		namespace := g.store.GetNamespace()
		if !scoped {
			if spec.AuthSecretRef.JSON.Namespace == nil {
				return fmt.Errorf("authsecretref namespace required when cluster-scoped")
			}
			namespace = *spec.AuthSecretRef.JSON.Namespace
		}
		data, e := g.secretKeyRef(ctx, namespace, *spec.AuthSecretRef.JSON)
		if e != nil {
			return err
		}
		clientOption = option.WithCredentialsJSON([]byte(data))
	}
	g.client, err = secretmanager.NewService(ctx, clientOption)
	if err != nil {
		return err
	}
	return nil
}

func (g *GCP) secretKeyRef(ctx context.Context, namespace string, secretRef smmeta.SecretKeySelector) (string, error) {
	g.log.V(1).Info("retrieving kubernetes secret", "name", secretRef.Name)
	var secret corev1.Secret
	ref := types.NamespacedName{
		Namespace: namespace,
		Name:      secretRef.Name,
	}
	err := g.kube.Get(ctx, ref, &secret)
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
