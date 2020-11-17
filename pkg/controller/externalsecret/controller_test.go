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

	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	storeint "github.com/itscontained/secret-manager/pkg/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
							Name:     "secret/data/foo",
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
			defer func() {
				By("Deleting the ExternalSecret successfully")
				Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			}()

			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				By("Fetching the ExternalSecret successfully")
				k8sClient.Get(context.Background(), key, fetched)
				By("Checking the Status Condition")
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Unavailable()) &&
					matches(fetchedCond.Message, errStoreNotFound)
			}, timeout, interval).Should(BeTrue())
		})

		It("An ExternalSecret referencing a SecretStore with invalid credentials should be NotReady", func() {
			store := sampleStore.DeepCopy()
			By("Creating the SecretStore successfully")
			Expect(k8sClient.Create(context.Background(), store)).Should(Succeed())
			defer func() {
				By("Deleting the SecretStore successfully")
				Expect(k8sClient.Delete(context.Background(), store)).Should(Succeed())
			}()
			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key",
						RemoteRef: smv1alpha1.RemoteReference{
							Name:     "secret/data/foo",
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

			storeFactory.WithNew(func(context.Context, smv1alpha1.GenericStore,
				client.Client, string) (storeint.Client, error) {
				return nil, fmt.Errorf("artificial test error")
			})

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			defer func() {
				By("Deleting the ExternalSecret successfully")
				Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			}()
			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				By("Fetching the ExternalSecret successfully")
				k8sClient.Get(context.Background(), key, fetched)
				By("Checking the status condition")
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Unavailable()) &&
					matches(fetchedCond.Message, errStoreSetupFailed)
			}, timeout, interval).Should(BeTrue())
		})

		It("An ExternalSecret with a valid SecretStore should generate a Secret", func() {
			store := sampleStore.DeepCopy()
			By("Creating the SecretStore successfully")
			Expect(k8sClient.Create(context.Background(), store)).Should(Succeed())
			defer func() {
				By("Deleting the SecretStore successfully")
				Expect(k8sClient.Delete(context.Background(), store)).Should(Succeed())
			}()
			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key",
						RemoteRef: smv1alpha1.RemoteReference{
							Name:     "secret/data/foo",
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
					Labels: map[string]string{
						"label-key": "label-value",
					},
					Annotations: map[string]string{
						"annotation-key": "annotation-value",
					},
				},
				Spec: spec,
			}

			testSecretData := []byte("this-is-a-secret")
			expectedData := map[string][]byte{
				"key": testSecretData,
			}
			storeFactory.WithGetSecret(testSecretData, nil)
			storeFactory.WithNew(func(context.Context, smv1alpha1.GenericStore,
				client.Client, string) (storeint.Client, error) {
				return storeFactory, nil
			})

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			defer func() {
				By("Deleting the ExternalSecret successfully")
				Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			}()
			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				By("Fetching the ExternalSecret successfully")
				Expect(k8sClient.Get(context.Background(), key, fetched)).Should(Succeed())
				By("Checking the status condition")
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Available())
			}, timeout, interval).Should(BeTrue(), "The ExternalSecret should have a ready condition")

			fetchedSecret := &corev1.Secret{}
			Eventually(func() bool {
				By("Fetching the Secret successfully")
				Expect(k8sClient.Get(context.Background(), key, fetchedSecret)).Should(Succeed())
				return matches(string(fetchedSecret.Data["key"]), string(expectedData["key"]))
			}, timeout, interval).Should(BeTrue(), "The generated secret should be created")

			Expect(len(fetchedSecret.OwnerReferences)).Should(BeIdenticalTo(1),
				"The owner reference of the secret should be set")
			Expect(fetchedSecret.OwnerReferences[0].Kind).Should(BeIdenticalTo(smv1alpha1.ExtSecretKind),
				"The owner kind should be ExternalSecret")
			Expect(fetchedSecret.OwnerReferences[0].Name).Should(BeIdenticalTo(toCreate.Name),
				"The owner name should be the name of the ExternalSecret")

			Expect(fetchedSecret.Labels["label-key"]).Should(BeIdenticalTo(toCreate.Labels["label-key"]),
				"The secret should have labels of the ExternalSecret")
			Expect(fetchedSecret.Annotations["annotation-key"]).Should(BeIdenticalTo(toCreate.Annotations["annotation-key"]),
				"The secret should have annotations of the ExternalSecret")
		})

		It("An ExternalSecret with dataFrom specified should generate secret", func() {
			store := sampleStore.DeepCopy()
			By("Creating the SecretStore successfully")
			Expect(k8sClient.Create(context.Background(), store)).Should(Succeed())
			defer func() {
				By("Deleting the SecretStore successfully")
				Expect(k8sClient.Delete(context.Background(), store)).Should(Succeed())
			}()
			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key2",
						RemoteRef: smv1alpha1.RemoteReference{
							Name:     "secret/data/foo",
							Property: smmeta.String("property"),
						},
					},
				},
				DataFrom: []smv1alpha1.RemoteReference{
					{
						Name:     "secret/data/bar",
						Property: smmeta.String("property"),
					},
					{
						Name:     "secret/data/bar",
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
				"key1": testSecretMap["key1"],
				"key2": testSecretData,
			}
			storeFactory.WithGetSecretMap(testSecretMap, nil)
			storeFactory.WithGetSecret(testSecretData, nil)
			storeFactory.WithNew(func(context.Context, smv1alpha1.GenericStore,
				client.Client, string) (storeint.Client, error) {
				return storeFactory, nil
			})

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			defer func() {
				By("Deleting the ExternalSecret successfully")
				Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			}()
			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				By("Fetching the ExternalSecret successfully")
				Expect(k8sClient.Get(context.Background(), key, fetched)).Should(Succeed())
				By("Checking the Status Condition")
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Available())
			}, timeout, interval).Should(BeTrue(), "The ExternalSecret should have a ready condition")

			fetchedSecret := &corev1.Secret{}
			Eventually(func() bool {
				By("Fetching the Secret successfully")
				Expect(k8sClient.Get(context.Background(), key, fetchedSecret)).Should(Succeed())
				return true
			}, timeout, interval).Should(BeTrue(), "The generated secret should be created")
			defer func() {
				By("Deleting the secret successfully")
				Expect(k8sClient.Delete(context.Background(), fetchedSecret)).Should(Succeed())
			}()

			Expect(fetchedSecret.Data).Should(Equal(expectedMap), "Secret data should match test data")
		})

		It("An ExternalSecret with a template fields should be set", func() {
			store := sampleStore.DeepCopy()
			By("Creating the SecretStore successfully")
			Expect(k8sClient.Create(context.Background(), store)).Should(Succeed())
			defer func() {
				By("Deleting the SecretStore successfully")
				Expect(k8sClient.Delete(context.Background(), store)).Should(Succeed())
			}()
			expectedAnnotations := map[string]string{
				"testKey": "testValue",
			}

			templateObject := map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": expectedAnnotations,
				},
			}
			templateObjectBytes, err := json.Marshal(templateObject)
			Expect(err).Should(BeNil())

			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key1",
						RemoteRef: smv1alpha1.RemoteReference{
							Name:     "secret/data/foo",
							Property: smmeta.String("property"),
						},
					},
				},
				Template: runtime.RawExtension{
					Raw: templateObjectBytes,
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

			testSecretData := []byte("this-is-a-test-secret")
			expectedData := map[string][]byte{
				"key1": testSecretData,
			}
			storeFactory.WithGetSecret(testSecretData, nil)
			storeFactory.WithNew(func(context.Context, smv1alpha1.GenericStore,
				client.Client, string) (storeint.Client, error) {
				return storeFactory, nil
			})

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			defer func() {
				By("Deleting the ExternalSecret successfully")
				Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			}()

			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), key, fetched)).Should(Succeed())
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Available())
			}, timeout, interval).Should(BeTrue(), "The ExternalSecret should have a ready condition")

			Expect(fetched.Spec.Template).Should(Equal(templateObjectBytes), "ExternalSecret should have template set")

			fetchedSecret := &corev1.Secret{}
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), key, fetchedSecret)).Should(Succeed())
				return true
			}, timeout, interval).Should(BeTrue(), "The generated secret should be created")
			defer func() {
				By("Deleting the Secret successfully")
				Expect(k8sClient.Delete(context.Background(), fetchedSecret)).Should(Succeed())
			}()

			Expect(fetchedSecret.Data).Should(Equal(expectedData), "Secret data should match test data")
			Expect(fetchedSecret.Annotations).Should(Equal(expectedAnnotations), "Secret generated should have annotations from ExternalSecret template field")
		})

		It("An ExternalSecret with invalid template field should report NotReady", func() {
			store := sampleStore.DeepCopy()
			By("Creating the SecretStore successfully")
			Expect(k8sClient.Create(context.Background(), store)).Should(Succeed())
			defer func() {
				By("Deleting the SecretStore successfully")
				Expect(k8sClient.Delete(context.Background(), store)).Should(Succeed())
			}()
			templateObject := map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": "invalid type",
				},
			}
			templateObjectBytes, err := json.Marshal(templateObject)
			Expect(err).Should(BeNil())

			spec := smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "key1",
						RemoteRef: smv1alpha1.RemoteReference{
							Name:     "secret/data/foo",
							Property: smmeta.String("property"),
						},
					},
				},
				Template: runtime.RawExtension{
					Raw: templateObjectBytes,
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

			testSecretData := []byte("this-is-a-test-secret")
			storeFactory.WithGetSecret(testSecretData, nil)
			storeFactory.WithNew(func(context.Context, smv1alpha1.GenericStore,
				client.Client, string) (storeint.Client, error) {
				return storeFactory, nil
			})

			By("Creating the ExternalSecret successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			defer func() {
				By("Deleting the ExternalSecret successfully")
				Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())
			}()

			fetched := &smv1alpha1.ExternalSecret{}
			Eventually(func() bool {
				By("Fetching the ExternalSecret successfully")
				Expect(k8sClient.Get(context.Background(), key, fetched)).Should(Succeed())
				By("Checking the Status Condition")
				fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
				return fetchedCond.Matches(smmeta.Unavailable()) &&
					matches(fetchedCond.Message, errTemplateFailed)
			}, timeout, interval).Should(BeTrue(), "The ExternalSecret should have a NotReady condition")
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
