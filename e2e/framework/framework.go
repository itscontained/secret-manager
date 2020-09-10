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
	"os"
	"os/exec"
	"time"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// KubectlPath defines the full path of the kubectl binary
	KubectlPath = "/usr/local/bin/kubectl"
	// RunID is a unique ID for this e2e test run
	RunID = uuid.NewUUID()
)

const (
	// Poll how often to poll for conditions
	Poll = 2 * time.Second

	// DefaultTimeout time to wait for operations to complete
	DefaultTimeout = 90 * time.Second
)

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
type Framework struct {
	// Name of the e2e test suite
	BaseName  string
	Namespace string

	// A Kubernetes and Service Catalog client
	KubeClient client.Client
	KubeConfig *restclient.Config
	HelmValues string
}

// NewDefaultFramework makes a new framework and sets up a BeforeEach/AfterEach for
// you (you can write additional before/after each functions).
func NewDefaultFramework(baseName, helmValues string) *Framework {
	defer ginkgo.GinkgoRecover()

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = smv1alpha1.AddToScheme(scheme)

	var kubeConfig *restclient.Config
	var err error
	kcPath := os.Getenv("KUBECONFIG")
	if kcPath != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kcPath)
	} else {
		kubeConfig, err = restclient.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}
	assert.Nil(ginkgo.GinkgoT(), err, "creting kubernetes API client configuration")
	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: scheme})
	assert.Nil(ginkgo.GinkgoT(), err, "creating Kubernetes API client")

	f := &Framework{
		BaseName:   baseName,
		KubeConfig: kubeConfig,
		KubeClient: kubeClient,
		HelmValues: helmValues,
		Namespace:  fmt.Sprintf("e2e-%s-%s", baseName, RunID),
	}
	ginkgo.BeforeEach(f.BeforeEach)
	ginkgo.AfterEach(f.AfterEach)
	return f
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	err := CreateNamespace(f.Namespace, f.KubeClient)
	assert.Nil(ginkgo.GinkgoT(), err, "creating namespace")
	err = f.newSecretManager()
	assert.Nil(ginkgo.GinkgoT(), err, "creating secret-manager")
	_, err = WaitForSMPod(f.Namespace, f.KubeClient)
	assert.Nil(ginkgo.GinkgoT(), err, "waiting for secret-manager to come up")
}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	if ginkgo.CurrentGinkgoTestDescription().Failed {
		logs, err := secretManagerLogs(f.KubeClient, f.Namespace)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		logrus.Println(logs)
	}
	err := f.deleteSecretManager()
	assert.Nil(ginkgo.GinkgoT(), err, "deleting secret-manager %v", f.Namespace)
	err = deleteNamespace(f.Namespace, f.KubeClient)
	assert.Nil(ginkgo.GinkgoT(), err, "deleting namespace %v", f.Namespace)
}

// newSecretManager deploys a secret-manager instance using helm
// you can specify helm values file for a certain configuration
func (f *Framework) newSecretManager() error {
	ginkgo.By("launching secret-manager")
	//nolint:gosec
	cmd := exec.Command("/wait-for-secret-manager.sh", f.Namespace, f.HelmValues, fmt.Sprintf("secret-manager-%s", f.BaseName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error creating secret-manager: %v.\nLogs:\n%v", err, string(out))
	}
	return nil
}

// deletes the secret manager release
func (f *Framework) deleteSecretManager() error {
	ginkgo.By("deleting secret-manager")
	//nolint:gosec
	cmd := exec.Command("helm", "uninstall", fmt.Sprintf("secret-manager-%s", f.BaseName), "-n", f.Namespace)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error deleting secret-manager: %v.\nLogs:\n%v", err, string(out))
	}
	return nil
}

// NewLocalstack deploys a fresh localstack instance into the specified namespace
func (f *Framework) NewLocalstack(namespace string) error {
	ginkgo.By("launching localstack")
	cmd := exec.Command("/wait-for-localstack.sh", namespace, f.HelmValues)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error creating localstack: %v.\nLogs:\n%v", err, string(out))
	}
	return nil
}
