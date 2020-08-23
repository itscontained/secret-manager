package store

import (
	"context"
	"fmt"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	vault "github.com/itscontained/secret-manager/pkg/internal/vault"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Factory func(ctx context.Context, store smv1alpha1.GenericStore, kubeClient client.Client, namespace string) (smv1alpha1.StoreClient, error)

func New(ctx context.Context, store smv1alpha1.GenericStore, kubeClient client.Client, namespace string) (smv1alpha1.StoreClient, error) {
	storeSpec := store.GetSpec()
	if storeSpec.Vault != nil {
		vaultClient, err := vault.New(ctx, kubeClient, store, namespace)
		if err != nil {
			return nil, fmt.Errorf("unable to setup Vault client: %w", err)
		}
		return vaultClient, nil
	}

	return nil, fmt.Errorf("SecretStore %q does not have a valid client", store.GetName())
}
