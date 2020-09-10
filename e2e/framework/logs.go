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
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func secretManagerLogs(c client.Client, namespace string) (string, error) {
	pod, err := WaitForSMPod(namespace, c)
	if err != nil {
		return "", err
	}
	if isRunning, err := podRunningReady(pod); err == nil && isRunning {
		return Logs(pod)
	}
	return "", fmt.Errorf("no secret-manager pod is running (logs)")
}

// WaitForSMPod returns the ingress controller running pod
func WaitForSMPod(ns string, c client.Client) (*corev1.Pod, error) {
	var pod *corev1.Pod
	var l corev1.PodList
	err := wait.Poll(Poll, DefaultTimeout, func() (bool, error) {
		ls, _ := labels.Parse("app=secret-manager")
		err := c.List(context.Background(), &l, &client.ListOptions{
			LabelSelector: ls,
		})
		if err != nil {
			return false, nil
		}
		for i := range l.Items {
			if strings.HasPrefix(l.Items[i].GetName(), "secret-manager") {
				isRunning, err := podRunningReady(&l.Items[i])
				if err != nil {
					continue
				}
				if isRunning {
					pod = &l.Items[i]
					return true, nil
				}
			}
		}
		return false, nil
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return nil, fmt.Errorf("timeout waiting at least one secret-manager pod running in namespace %v", ns)
		}
		return nil, err
	}
	return pod, nil
}

// Logs returns the log entries of a given Pod.
func Logs(pod *corev1.Pod) (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	if len(pod.Spec.Containers) != 1 {
		return "", fmt.Errorf("could not determine which container to use")
	}
	//nolint:gosec
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v logs --namespace %s %s", KubectlPath, pod.Namespace, pod.Name))
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not execute '%s %s': %v", cmd.Path, cmd.Args, err)
	}

	if execErr.Len() > 0 {
		return "", fmt.Errorf("stderr: %v", execErr.String())
	}

	return execOut.String(), nil
}
