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

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateNamespace creates a new namespace in the cluster
func CreateNamespace(name string, c client.Client) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	var err error
	err = wait.Poll(Poll, DefaultTimeout, func() (bool, error) {
		err = c.Create(context.Background(), ns)
		if err != nil {
			log.Errorf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	return nil
}

// deleteNamespace deletes a namespace and all the objects inside
func deleteNamespace(namespace string, c client.Client) error {
	grace := int64(0)
	pb := metav1.DeletePropagationBackground
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	return c.Delete(context.Background(), &ns, &client.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &pb,
	})
}
