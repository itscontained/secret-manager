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

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

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

	secret := &corev1.Secret{}
	err := r.Get(ctx, req.NamespacedName, secret)
	_ = err

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