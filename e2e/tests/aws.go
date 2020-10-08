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

	"github.com/itscontained/secret-manager/e2e/framework"
	smmeta "github.com/itscontained/secret-manager/pkg/apis/meta/v1"
	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"

	// use dot imports
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = ginkgo.Describe("[aws]", func() {
	f := framework.NewDefaultFramework("aws", "default")

	ginkgo.BeforeEach(func() {
		err := f.NewLocalstack(f.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err, "creating localstack")
	})

	ginkgo.It("should sync secrets", func() {
		// create AWS SM Secret
		err := framework.CreateAWSSecretsManagerSecret(f.Namespace, "my-aws-secret", `{"username":"bob", "password":"abc123xyz456"}`)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		key := types.NamespacedName{
			Name:      "aws-sm-secret",
			Namespace: f.Namespace,
		}

		// create store
		store := &smv1alpha1.SecretStore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "localstack-sm",
				Namespace: f.Namespace,
			},
			Spec: smv1alpha1.SecretStoreSpec{
				AWS: &smv1alpha1.AWSStore{
					Region: smmeta.String("us-east-1"),
				},
			},
		}
		ginkgo.By("Creating the SecretStore successfully")
		gomega.Expect(f.KubeClient.Create(context.Background(), store)).Should(gomega.Succeed())

		// create ES
		ginkgo.By("Creating the ExternalSecret successfully")
		gomega.Expect(f.KubeClient.Create(context.Background(), &smv1alpha1.ExternalSecret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "aws-sm-secret",
				Namespace: f.Namespace,
				Labels: map[string]string{
					"aws": "yep",
				},
				Annotations: map[string]string{
					"aws": "ok",
				},
			},
			Spec: smv1alpha1.ExternalSecretSpec{
				StoreRef: smv1alpha1.ObjectReference{
					Name: store.Name,
					Kind: smv1alpha1.SecretStoreKind,
				},
				Data: []smv1alpha1.KeyReference{
					{
						SecretKey: "username-from-aws",
						RemoteRef: smv1alpha1.RemoteReference{
							Name:     smmeta.String("my-aws-secret"),
							Property: smmeta.String("username"),
							Version:  smmeta.String("AWSCURRENT"),
						},
					},
					{
						SecretKey: "password-from-aws",
						RemoteRef: smv1alpha1.RemoteReference{
							Name:     smmeta.String("my-aws-secret"),
							Property: smmeta.String("password"),
							Version:  smmeta.String("AWSCURRENT"),
						},
					},
				},
			},
		})).Should(gomega.Succeed())

		// wait for secret to appear
		fetched := &smv1alpha1.ExternalSecret{}
		gomega.Eventually(func() bool {
			ginkgo.By("Fetching the ExternalSecret successfully")
			gomega.Expect(f.KubeClient.Get(context.Background(), key, fetched)).Should(gomega.Succeed())
			ginkgo.By("Checking the status condition")
			fetchedCond := fetched.Status.GetCondition(smmeta.TypeReady)
			return fetchedCond.Matches(smmeta.Available())
		}, framework.DefaultTimeout, framework.Poll).Should(gomega.BeTrue(), "The ExternalSecret should have a ready condition")

		fetchedSecret := &corev1.Secret{}
		gomega.Eventually(func() map[string][]byte {
			ginkgo.By("Fetching the Secret successfully")
			gomega.Expect(f.KubeClient.Get(context.Background(), key, fetchedSecret)).Should(gomega.Succeed())
			return fetchedSecret.Data
		}, framework.DefaultTimeout, framework.Poll).Should(gomega.Equal(map[string][]byte{
			"username-from-aws": []byte(b64enc("bob")),
			"password-from-aws": []byte(b64enc("abc123xyz456")),
		}), "The generated secret should be created")
	})
})
