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

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	fakestore "github.com/itscontained/secret-manager/pkg/internal/store/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ExternalSecrets Controller", func() {

	const timeout = time.Second * 10
	const interval = time.Second * 1

	var secretType = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
	}

	BeforeEach(func() {
		// Add any begin steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("ExternalSecrets", func() {
		It("An ExternalSecret referencing a SecretStore which does not exist should be NotReady", func() {
			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: "does-not-exist",
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key",
						RemoteRef: smv1alpha1.RemoteReference{
							Path:     "secret/data/foo",
							Property: smmeta.String("key"),
						},
					},
				},
			}

			key := types.NamespacedName{
				Name:      secretType.Name,
				Namespace: secretType.Namespace,
			}

			toCreate := &smv1alpha1.ExternalSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, fetched)
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Unavailable()) &&
					matches(fetchedCond.Message, errStoreNotFound)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the ExternalSecret successfully")
			Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
		})

		It("An ExternalSecret referencing a SecretStore with invalid credentials should be NotReady", func() {
			store := sampleStore.DeepCopy()
			Expect(k8sClient.Create(context.Background(), store)).Should(Succeed())

			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key",
						RemoteRef: smv1alpha1.RemoteReference{
							Path:     "secret/data/foo",
							Property: smmeta.String("key"),
						},
					},
				},
			}

			key := types.NamespacedName{
				Name:      secretType.Name,
				Namespace: secretType.Namespace,
			}

			toCreate := &smv1alpha1.ExternalSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			storeFactory.WithNew(func(context.Context, client.Client, smv1alpha1.GenericStore, string) (*fakestore.Factory, error) {
				return nil, fmt.Errorf("fail to setup store client")
			})

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, fetched)
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Unavailable()) &&
					matches(fetchedCond.Message, errStoreSetupFailed)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the ExternalSecret successfully")
			Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			By("Deleting the SecretStore successfully")
			Expect(k8sClient.Delete(context.Background(), store)).Should(Succeed())
		})

		It("An ExternalSecret with a valid SecretStore should generate a Secret", func() {
			store := sampleStore.DeepCopy()
			Expect(k8sClient.Create(context.Background(), store)).Should(Succeed())

			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key",
						RemoteRef: smv1alpha1.RemoteReference{
							Path:     "secret/data/foo",
							Property: smmeta.String("key"),
						},
					},
				},
			}

			key := types.NamespacedName{
				Name:      secretType.Name,
				Namespace: secretType.Namespace,
			}

			toCreate := &smv1alpha1.ExternalSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			testSecretData := []byte("this-is-a-secret")
			expectedData := map[string][]byte{
				"key": []byte(base64.RawStdEncoding.EncodeToString(testSecretData)),
			}
			storeFactory.WithGetSecret(testSecretData, nil)
			storeFactory.WithNew(func(context.Context, client.Client, smv1alpha1.GenericStore, string) (*fakestore.Factory, error) {
				return storeFactory, nil
			})

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), key, fetched)).Should(Succeed())
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Available())
			}, timeout, interval).Should(BeTrue(), "The ExternalSecret should have a ready condition")

			fetchedSecret := &corev1.Secret{}
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), key, fetchedSecret)).Should(Succeed())
				return true
			}, timeout, interval).Should(BeTrue(), "The generated secret should be created")

			Expect(len(fetchedSecret.OwnerReferences)).Should(BeIdenticalTo(1),
				"The owner reference of the secret should be set")
			Expect(fetchedSecret.OwnerReferences[0].Kind).Should(BeIdenticalTo(smv1alpha1.ExtSecretKind),
				"The owner kind should be ExternalSecret")
			Expect(fetchedSecret.OwnerReferences[0].Name).Should(BeIdenticalTo(toCreate.Name),
				"The owner name should be the name of the ExternalSecret")

			Expect(fetchedSecret.Data).Should(Equal(expectedData), "Secret data should match test data")

			By("Deleting the ExternalSecret successfully")
			Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			By("Deleting the SecretStore successfully")
			Expect(k8sClient.Delete(context.Background(), store)).Should(Succeed())
		})

		It("An ExternalSecret with dataFrom specified should generate secret", func() {
			store := sampleStore.DeepCopy()
			Expect(k8sClient.Create(context.Background(), store)).Should(Succeed())

			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key2",
						RemoteRef: smv1alpha1.RemoteReference{
							Path:     "secret/data/foo",
							Property: smmeta.String("property"),
						},
					},
				},
				DataFrom: []smv1alpha1.RemoteReference{
					{
						Path:     "secret/data/bar",
						Property: smmeta.String("property"),
					},
					{
						Path:     "secret/data/bar",
						Property: smmeta.String("property"),
					},
				},
			}

			key := types.NamespacedName{
				Name:      secretType.Name,
				Namespace: secretType.Namespace,
			}

			toCreate := &smv1alpha1.ExternalSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			testSecretMap := map[string][]byte{
				"key1": []byte("value1"),
				"key2": []byte("value2"),
			}
			testSecretData := []byte("value3")
			expectedMap := map[string][]byte{
				"key1": []byte(base64.RawStdEncoding.EncodeToString(testSecretMap["key1"])),
				"key2": []byte(base64.RawStdEncoding.EncodeToString(testSecretData)),
			}
			storeFactory.WithGetSecretMap(testSecretMap, nil)
			storeFactory.WithGetSecret(testSecretData, nil)
			storeFactory.WithNew(func(context.Context, client.Client, smv1alpha1.GenericStore, string) (*fakestore.Factory, error) {
				return storeFactory, nil
			})

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), key, fetched)).Should(Succeed())
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Available())
			}, timeout, interval).Should(BeTrue(), "The ExternalSecret should have a ready condition")

			fetchedSecret := &corev1.Secret{}
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), key, fetchedSecret)).Should(Succeed())
				return true
			}, timeout, interval).Should(BeTrue(), "The generated secret should be created")

			Expect(fetchedSecret.Data).Should(Equal(expectedMap), "Secret data should match test data")

			By("Deleting the ExternalSecret successfully")
			Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			By("Deleting the SecretStore successfully")
			Expect(k8sClient.Delete(context.Background(), store)).Should(Succeed())
		})
	})
})

// arbitrary SecretStore to use when injecting factory
var sampleStore = &smv1alpha1.SecretStore{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "vault",
		Namespace: "default",
	},
	Spec: smv1alpha1.SecretStoreSpec{
		Vault: &smv1alpha1.VaultStore{
			Server: "http://localhost:12345",
		},
	},
}
