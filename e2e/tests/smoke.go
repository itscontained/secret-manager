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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/itscontained/secret-manager/e2e/framework"

	// use dot imports
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("[smoke]", func() {
	f := framework.NewDefaultFramework("smoke", "default")

	ginkgo.It("should expose metrics", func() {
		po, err := framework.WaitForSMPod(f.Namespace, f.KubeClient)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		ginkgo.By("check pod annotations and labels")
		gomega.Expect(po.ObjectMeta.Annotations["test"]).To(gomega.Equal("foobar"))
		gomega.Expect(po.ObjectMeta.Labels["app"]).To(gomega.Equal("secret-manager"))

		ginkgo.By("scrape custom metrics port")
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:8711/metrics", po.Status.PodIP), nil)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		res, err := http.DefaultClient.Do(req)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		defer res.Body.Close()
		gomega.Expect(res.StatusCode).To(gomega.Equal(http.StatusOK))
		bdy, err := ioutil.ReadAll(res.Body)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(string(bdy)).To(gomega.ContainSubstring("workqueue_work_duration_seconds_bucket{name=\"externalsecret\",le=\"1\"}"))
	})
})
