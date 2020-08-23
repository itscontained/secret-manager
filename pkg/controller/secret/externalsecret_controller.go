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
	"fmt"
	"time"

	"github.com/go-logr/logr"

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/internal/store"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/utils/clock"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	ownerKey     = ".metadata.controller"
	requeueAfter = time.Second * 30

	errStoreNotFound       = "cannot get store reference"
	errStoreSetupFailed    = "cannot setup store client"
	errGetSecretDataFailed = "cannot get ExternalSecret data from store"
)

// ExternalSecretReconciler reconciles a ExternalSecret object
type ExternalSecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Clock  clock.Clock

	storeFactory store.Factory
}

func (r *ExternalSecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("externalsecret", req.NamespacedName)

	extSecret := &smv1alpha1.ExternalSecret{}
	if err := r.Get(ctx, req.NamespacedName, extSecret); err != nil {
		log.Error(err, "unable to get ExternalSecret")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      extSecret.Name,
			Namespace: extSecret.Namespace,
		},
	}

	result, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		store, err := r.getStore(ctx, extSecret)
		if err != nil {
			return fmt.Errorf("%s: %w", errStoreNotFound, err)
		}

		storeClient, err := r.storeFactory(ctx, store, r.Client, req.Namespace)
		if err != nil {
			return fmt.Errorf("%s: %w", errStoreSetupFailed, err)
		}

		secret.ObjectMeta.Labels = extSecret.Labels
		secret.ObjectMeta.Annotations = extSecret.Annotations
		secret.Data = map[string][]byte{}
		err = controllerutil.SetControllerReference(extSecret, &secret.ObjectMeta, r.Scheme)
		if err != nil {
			return fmt.Errorf("failed to set ExternalSecret controller reference: %w", err)
		}

		secret.Data, err = r.getSecret(ctx, storeClient, extSecret)
		if err != nil {
			return fmt.Errorf("%s: %w", errGetSecretDataFailed, err)
		}

		return nil
	})

	if err != nil {
		log.Error(err, "error while reconciling ExternalSecret")
		extSecret.Status.SetConditions(smmeta.Unavailable().WithMessage(err.Error()))
		_ = r.Status().Update(ctx, extSecret)
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	log.Info("successfully reconcile ExternalSecret", "operation", result)
	extSecret.Status.SetConditions(smmeta.Available())
	_ = r.Status().Update(ctx, extSecret)
	return ctrl.Result{}, nil
}

func (r *ExternalSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Clock == nil {
		r.Clock = clock.RealClock{}
	}

	if r.storeFactory == nil {
		r.storeFactory = store.New
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

func (r *ExternalSecretReconciler) getSecret(ctx context.Context, storeClient smv1alpha1.StoreClient, extSecret *smv1alpha1.ExternalSecret) (map[string][]byte, error) {
	secretDataMap := make(map[string][]byte)
	for _, secretRef := range extSecret.Spec.Data {
		secretData, err := storeClient.GetSecret(ctx, secretRef.RemoteRef)
		if err != nil {
			return nil, fmt.Errorf("path %q: %w", extSecret.Spec.DataFrom.Path, err)
		}
		dstBytes := make([]byte, base64.StdEncoding.EncodedLen(len(secretData)))
		base64.StdEncoding.Encode(dstBytes, secretData)
		secretDataMap[secretRef.SecretKey] = dstBytes
	}

	if extSecret.Spec.DataFrom != nil {
		secretMap, err := storeClient.GetSecretMap(ctx, *extSecret.Spec.DataFrom)
		if err != nil {
			return nil, fmt.Errorf("path %q: %w", extSecret.Spec.DataFrom.Path, err)
		}
		for secretKey, secretData := range secretMap {
			dstBytes := make([]byte, base64.StdEncoding.EncodedLen(len(secretData)))
			base64.StdEncoding.Encode(dstBytes, secretData)
			secretDataMap[secretKey] = dstBytes
		}
	}

	return secretDataMap, nil
}

func (r *ExternalSecretReconciler) getStore(ctx context.Context, extSecret *smv1alpha1.ExternalSecret) (smv1alpha1.GenericStore, error) {
	if extSecret.Kind == smv1alpha1.ClusterSecretStoreKind {
		clusterStore := &smv1alpha1.ClusterSecretStore{}
		ref := types.NamespacedName{
			Name: extSecret.Spec.StoreRef.Name,
		}
		if err := r.Get(ctx, ref, clusterStore); err != nil {
			return nil, fmt.Errorf("ClusterSecretStore %q: %w", ref.Name, err)
		}
		return clusterStore, nil
	}

	namespacedStore := &smv1alpha1.SecretStore{}
	ref := types.NamespacedName{
		Namespace: extSecret.Namespace,
		Name:      extSecret.Spec.StoreRef.Name,
	}
	if err := r.Get(ctx, ref, namespacedStore); err != nil {
		return nil, fmt.Errorf("SecretStore %q: %w", ref.Name, err)
	}
	return namespacedStore, nil
}
