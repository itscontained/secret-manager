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

package controllers

import (
	"context"

	"github.com/go-logr/logr"

	smv1alpha1 "github.com/mcavoyk/secret-manager/pkg/apis/secretmanager/v1alpha1"
	vault "github.com/mcavoyk/secret-manager/pkg/internal/vault"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/tools/record"

	"k8s.io/utils/clock"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ownerKey = ".metadata.controller"
)

// ExternalSecretReconciler reconciles a ExternalSecret object
type ExternalSecretReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Clock    clock.Clock
}

func (r *ExternalSecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("externalsecret", req.NamespacedName)

	extSecret := &smv1alpha1.ExternalSecret{}
	if err := r.Get(ctx, req.NamespacedName, extSecret); err != nil {
		log.Error(err, "unable to fetch ExternalSecret")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: move getStore to function
	var store smv1alpha1.GenericStore
	if extSecret.Kind == smv1alpha1.ClusterSecretStoreKind {
		clusterStore := &smv1alpha1.ClusterSecretStore{}
		ref := types.NamespacedName{
			Name: extSecret.Spec.StoreRef.Name,
		}
		if err := r.Get(ctx, ref, clusterStore); err != nil {
			log.Error(err, "unable to fetch ClusterSecretStore")
			return ctrl.Result{}, err
		}
		store = clusterStore
	} else {
		namespacedStore := &smv1alpha1.SecretStore{}
		ref := types.NamespacedName{
			Namespace: extSecret.Namespace,
			Name:      extSecret.Spec.StoreRef.Name,
		}
		if err := r.Get(ctx, ref, namespacedStore); err != nil {
			log.Error(err, "unable to fetch SecretStore")
			return ctrl.Result{}, err
		}
		store = namespacedStore
	}

	storeSpec := store.GetSpec()
	var storeClient smv1alpha1.StoreClient

	if storeSpec.Vault != nil {
		vaultClient, err := vault.New(ctx, r.Client, store, req.Namespace)
		if err != nil {
			log.Error(err, "unable to setup Vault client")
			return ctrl.Result{}, err
		}
		storeClient = vaultClient
	}

	controllerRef := true
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: extSecret.Name,
			Namespace: extSecret.Namespace,
			Labels: extSecret.Labels,
			Annotations: extSecret.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion: extSecret.APIVersion,
					Kind: extSecret.Kind,
					Name: extSecret.Name,
					UID: extSecret.UID,
					Controller: &controllerRef,
				},
			},
		},
		Data: map[string][]byte{},
	}

	for _, secretRef := range extSecret.Spec.Data {
		secretData, err := storeClient.GetSecret(ctx, secretRef.RemoteRef)
		if err != nil {
			log.Error(err, "failed to fetch secret", "path", secretRef.RemoteRef.Path)
			return ctrl.Result{}, err
		}
		_ = secretData
	}

	_, _ = ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		return nil
	})

	return ctrl.Result{}, nil
}

func (r *ExternalSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &corev1.Secret{}, ownerKey, func(rawObj runtime.Object) []string {
		secret := rawObj.(*corev1.Secret)
		owner := metav1.GetControllerOf(secret)
		if owner == nil {
			return nil
		}

		if owner.APIVersion != smv1alpha1.ExtSecretGroupVersionKind.GroupVersion().String() || owner.Kind != smv1alpha1.ExtSecretKind {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&smv1alpha1.ExternalSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
