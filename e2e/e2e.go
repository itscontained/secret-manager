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

package e2e

import (
	"testing"

	"github.com/itscontained/secret-manager/e2e/framework"
	// import tests
	_ "github.com/itscontained/secret-manager/e2e/tests"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"

	"github.com/sirupsen/logrus"

	// required
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// RunE2ETests runs e2e tests using the Ginkgo runner
func RunE2ETests(t *testing.T) {
	logrus.SetOutput(ginkgo.GinkgoWriter)
	logrus.Infof("Starting e2e run %q on Ginkgo node %d", framework.RunID, config.GinkgoConfig.ParallelNode)
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "secret-manager e2e suite")
}
