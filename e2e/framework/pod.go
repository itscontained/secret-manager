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
	"fmt"

	corev1 "k8s.io/api/core/v1"

	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// podRunningReady checks whether pod p's phase is running and it has a ready
// condition of status true.
func podRunningReady(p *corev1.Pod) (bool, error) {
	// Check the phase is running.
	if p.Status.Phase != corev1.PodRunning {
		return false, fmt.Errorf("want pod '%s' on '%s' to be '%v' but was '%v'",
			p.ObjectMeta.Name, p.Spec.NodeName, corev1.PodRunning, p.Status.Phase)
	}
	// Check the ready condition is true.
	if !podutil.IsPodReady(p) {
		return false, fmt.Errorf("pod '%s' on '%s' didn't have condition {%v %v}; conditions: %v",
			p.ObjectMeta.Name, p.Spec.NodeName, corev1.PodReady, corev1.ConditionTrue, p.Status.Conditions)
	}
	return true, nil
}
