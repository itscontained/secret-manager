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
	"encoding/base64"

	"github.com/go-logr/logr"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	vault "github.com/itscontained/secret-manager/pkg/internal/vault"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/tools/record"

	"k8s.io/utils/clock"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      extSecret.Name,
			Namespace: extSecret.Namespace,
		},
	}

	result, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		// TODO: move getStore to function
		var store smv1alpha1.GenericStore
		if extSecret.Kind == smv1alpha1.ClusterSecretStoreKind {
			clusterStore := &smv1alpha1.ClusterSecretStore{}
			ref := types.NamespacedName{
				Name: extSecret.Spec.StoreRef.Name,
			}
			if err := r.Get(ctx, ref, clusterStore); err != nil {
				log.Error(err, "unable to fetch ClusterSecretStore")
				return err
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
				return err
			}
			store = namespacedStore
		}

		storeSpec := store.GetSpec()
		var storeClient smv1alpha1.StoreClient

		if storeSpec.Vault != nil {
			vaultClient, err := vault.New(ctx, r.Client, store, req.Namespace)
			if err != nil {
				log.Error(err, "unable to setup Vault client")
				return err
			}
			storeClient = vaultClient
		}

		secret.ObjectMeta.Labels = extSecret.Labels
		secret.ObjectMeta.Annotations = extSecret.Annotations
		secret.Data = map[string][]byte{}
		if err := controllerutil.SetControllerReference(extSecret, &secret.ObjectMeta, r.Scheme); err != nil {
			return err
		}

		// TODO: move fetching data from secretStore to function
		for _, secretRef := range extSecret.Spec.Data {
			secretData, err := storeClient.GetSecret(ctx, secretRef.RemoteRef)
			if err != nil {
				log.Error(err, "failed to fetch secret", "path", secretRef.RemoteRef.Path)
				return err
			}
			dstBytes := make([]byte, base64.StdEncoding.EncodedLen(len(secretData)))
			base64.StdEncoding.Encode(dstBytes, secretData)
			secret.Data[secretRef.SecretKey] = dstBytes
		}

		if extSecret.Spec.DataFrom != nil {
			secretMap, err := storeClient.GetSecretMap(ctx, *extSecret.Spec.DataFrom)
			if err != nil {
				log.Error(err, "failed to fetch secret", "path", extSecret.Spec.DataFrom.Path)
				return err
			}
			for secretKey, secretData := range secretMap {
				dstBytes := make([]byte, base64.StdEncoding.EncodedLen(len(secretData)))
				base64.StdEncoding.Encode(dstBytes, secretData)
				secret.Data[secretKey] = dstBytes
			}
		}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("reconciled ExternalSecret", "operation", result)
	return ctrl.Result{}, nil
}

func (r *ExternalSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Clock == nil {
		r.Clock = clock.RealClock{}
	}

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
