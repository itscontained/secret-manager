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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"github.com/imdario/mergo"

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	ctxlog "github.com/itscontained/secret-manager/pkg/log"
	"github.com/itscontained/secret-manager/pkg/store"
	_ "github.com/itscontained/secret-manager/pkg/store/register" // register known store backends
	storeschema "github.com/itscontained/secret-manager/pkg/store/schema"
	"github.com/itscontained/secret-manager/pkg/util/merge"

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
	errTemplateFailed      = "failed to merge secret with template field"
)

// ExternalSecretReconciler reconciles a ExternalSecret object
type ExternalSecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Clock  clock.Clock

	Reader client.Reader
}

func (r *ExternalSecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("externalsecret", req.NamespacedName)
	ctx = ctxlog.IntoContext(ctx, log)

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
		s, err := r.getStore(ctx, extSecret)
		if err != nil {
			return fmt.Errorf("%s: %w", errStoreNotFound, err)
		}

		storeClient, err := storeschema.GetStore(s)
		if err != nil {
			return fmt.Errorf("%s: %w", errStoreSetupFailed, err)
		}

		storeClient, err = storeClient.New(ctx, s, r.Client, req.Namespace)
		if err != nil {
			return fmt.Errorf("%s: %w", errStoreSetupFailed, err)
		}

		err = controllerutil.SetControllerReference(extSecret, &secret.ObjectMeta, r.Scheme)
		if err != nil {
			return fmt.Errorf("failed to set ExternalSecret controller reference: %w", err)
		}

		secret.Labels = extSecret.Labels
		secret.Annotations = extSecret.Annotations
		secret.Data, err = r.getSecret(ctx, storeClient, extSecret)
		if err != nil {
			return fmt.Errorf("%s: %w", errGetSecretDataFailed, err)
		}

		if &extSecret.Spec.Template != nil && extSecret.Spec.Template.Raw != nil {
			err = r.templateSecret(secret, extSecret.Spec.Template)
			if err != nil {
				return fmt.Errorf("%s: %w", errTemplateFailed, err)
			}
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

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Secret{}, ownerKey, func(rawObj runtime.Object) []string {
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

func (r *ExternalSecretReconciler) getSecret(ctx context.Context, storeClient store.Client, extSecret *smv1alpha1.ExternalSecret) (map[string][]byte, error) {
	secretDataMap := make(map[string][]byte)
	for _, remoteRef := range extSecret.Spec.DataFrom {
		secretMap, err := storeClient.GetSecretMap(ctx, remoteRef)
		if err != nil {
			return nil, fmt.Errorf("name %q: %w", remoteRef.Name, err)
		}
		secretDataMap = merge.Merge(secretDataMap, secretMap)
	}

	for _, secretRef := range extSecret.Spec.Data {
		secretData, err := storeClient.GetSecret(ctx, secretRef.RemoteRef)
		if err != nil {
			return nil, fmt.Errorf("name %q: %w", secretRef.RemoteRef.Name, err)
		}
		secretDataMap[secretRef.SecretKey] = secretData
	}

	return secretDataMap, nil
}

func (r *ExternalSecretReconciler) getStore(ctx context.Context, extSecret *smv1alpha1.ExternalSecret) (smv1alpha1.GenericStore, error) {
	r.Log.V(1).Info("getting store configuration")
	var secretStore smv1alpha1.GenericStore
	storeType := "ClusterSecretStore"
	ref := types.NamespacedName{
		Name: extSecret.Spec.StoreRef.Name,
	}
	if extSecret.Spec.StoreRef.Kind == smv1alpha1.ClusterSecretStoreKind {
		secretStore = &smv1alpha1.ClusterSecretStore{}
		r.Log.V(1).Info("using ClusterSecretStore")
	} else {
		secretStore = &smv1alpha1.SecretStore{}
		ref.Namespace = extSecret.Namespace
		storeType = "SecretStore"
		r.Log.V(1).Info("using SecretStore")
	}
	if err := r.Reader.Get(ctx, ref, secretStore); err != nil {
		return nil, fmt.Errorf("%s %q: %w", storeType, ref.Name, err)
	}
	return secretStore, nil
}

func (r *ExternalSecretReconciler) templateSecret(secret *corev1.Secret, template runtime.RawExtension) error {
	templatedSecret := &corev1.Secret{}
	if err := json.Unmarshal(template.Raw, templatedSecret); err != nil {
		return fmt.Errorf("error unmarshalling json: %w", err)
	}

	return mergo.Merge(secret, templatedSecret, mergo.WithOverride)
}
