
# Image URL to use all building/pushing image targets
IMG ?= mcavoyk/secret-manager
CRD_OPTIONS ?= "crd:crdVersions=v1beta1;v1,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

fmt: lint/install # ensure consistent code style
	go run oss.indeed.com/go/go-groups -w .
	golangci-lint run --fix > /dev/null 2>&1 || true

lint: lint/install ## run golangci-lint
	golangci-lint run
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "\033[0;33mdetected fmt problems: run \`\033[0;32mmake fmt\033[0m\`"; \
		exit 1; \
	fi

lint/install:
	@lint_version=$$(golangci-lint --version || ""); \
	if echo $$lint_version | grep -v "1.30.0"; then \
		echo "Installing golangci-lint"; \
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOBIN) v1.30.0; \
	fi

# Run tests
test:
	go test ./... -coverprofile cover.out

# Build manager binary
build: generate
	go build -o bin/manager ./cmd/controller/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./pkg/apis/..." output:crd:artifacts:config=deploy/charts/secret-manager/config/crds

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="build/boilerplate.go.txt" paths="./pkg/apis/..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
