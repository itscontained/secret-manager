module github.com/itscontained/secret-manager

go 1.14

require (
	github.com/aws/aws-sdk-go-v2 v0.24.0
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/hashicorp/vault/api v1.0.4
	github.com/imdario/mergo v0.3.11
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	google.golang.org/api v0.33.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/klog/v2 v2.3.0
	k8s.io/utils v0.0.0-20201015054608-420da100c033
	oss.indeed.com/go/go-groups v1.1.3
	sigs.k8s.io/controller-runtime v0.6.3
)
