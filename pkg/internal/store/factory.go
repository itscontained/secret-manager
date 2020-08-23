package store

import (
	"context"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is a common interface for interacting with SecretStore
// backends
type Client interface {
	GetSecret(ctx context.Context, ref smv1alpha1.RemoteReference) ([]byte, error)
	GetSecretMap(ctx context.Context, ref smv1alpha1.RemoteReference) (map[string][]byte, error)
}

// Factory returns a StoreClient
type Factory interface {
	New(ctx context.Context, store smv1alpha1.GenericStore, kubeClient client.Client, namespace string) (Client, error)
}
