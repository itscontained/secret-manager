module github.com/itscontained/secret-manager

go 1.14

require (
	cloud.google.com/go v0.65.0
	github.com/aws/aws-sdk-go-v2 v0.24.0
	github.com/go-logr/logr v0.2.1-0.20200730175230-ee2de8da5be6
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/hashicorp/vault/api v1.0.4
	github.com/imdario/mergo v0.3.11
	github.com/jmespath/go-jmespath v0.3.0 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.15.0 // indirect
	google.golang.org/api v0.30.0
	google.golang.org/genproto v0.0.0-20200911024640-645f7a48b24f
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/klog/v2 v2.3.0
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451
	oss.indeed.com/go/go-groups v1.1.2
	sigs.k8s.io/controller-runtime v0.6.2
)
