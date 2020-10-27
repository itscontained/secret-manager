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

package tests

import (
	"context"

	vaultaddon "github.com/itscontained/secret-manager/e2e/addon/vault"
	"github.com/itscontained/secret-manager/e2e/framework"
	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("[vault]", func() {
	f := framework.NewDefaultFramework("vault", "secret-manager")
	// TODO: make framework addon setup more generic
	vaultDeploy := vaultaddon.Vault{
		Kubectl:    "kubectl",
		Namespace:  f.Namespace,
		KubeClient: f.KubeClient,
	}

	BeforeEach(func() {
		err := vaultDeploy.Setup()
		assert.Nil(GinkgoT(), err, "creating vault")
	})

	It("should sync secrets", func() {
		// create secret
		secretName := "teamA/example-service"
		secret := map[string]string{
			"new-secret": "test-value123",
		}
		err := vaultDeploy.CreateSecret(secretName, secret)
		Expect(err).ToNot(HaveOccurred())

		key := types.NamespacedName{
			Name:      "vault-secret1",
			Namespace: f.Namespace,
		}

		storeSpec, err := vaultDeploy.CreateSecretStoreVaultKubeAuth()
		By("Setup Vault Kubernetes auth backend successfully")
		Expect(err).Should(Succeed())
		// create store
		store := &smv1alpha1.SecretStore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vault",
				Namespace: f.Namespace,
			},
			Spec: *storeSpec,
		}
		By("Creating the Vault SecretStore successfully")
		Expect(f.KubeClient.Create(context.Background(), store)).Should(Succeed())

		// create ES
		By("Creating the ExternalSecret successfully")
		Expect(f.KubeClient.Create(context.Background(), &smv1alpha1.ExternalSecret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: f.Namespace,
			},
			Spec: smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "user-a",
						RemoteRef: smv1alpha1.RemoteReference{
							Name:     secretName,
							Property: smmeta.String("new-secret"),
						},
					},
				},
			},
		})).Should(Succeed())

		// wait for secret to appear
		fetched := &smv1alpha1.ExternalSecret{}
		Eventually(func() bool {
			By("Fetching the ExternalSecret successfully")
			Expect(f.KubeClient.Get(context.Background(), key, fetched)).Should(Succeed())
			By("Checking the status condition")
			fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
			return fetchedCond.Matches(smmeta.Available())
		}, framework.DefaultTimeout, framework.Poll).Should(BeTrue(), "The ExternalSecret should have a ready condition")

		fetchedSecret := &corev1.Secret{}
		Eventually(func() map[string][]byte {
			By("Fetching the Secret successfully")
			Expect(f.KubeClient.Get(context.Background(), key, fetchedSecret)).Should(Succeed())
			return fetchedSecret.Data
		}, framework.DefaultTimeout, framework.Poll).Should(Equal(map[string][]byte{
			"user-a": []byte("test-value123"),
		}), "The generated secret should be created")
	})
})
