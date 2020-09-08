MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := all

IMG ?= itscontained/secret-manager

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null || echo "canary")
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

BINARY_VERSION ?= ${GIT_TAG}

LDFLAGS += -X github.com/itscontained/secret-manager/pkg/util.version=${BINARY_VERSION}
LDFLAGS += -X github.com/itscontained/secret-manager/pkg/util.gitCommit=${GIT_COMMIT}
LDFLAGS += -X github.com/itscontained/secret-manager/pkg/util.gitState=${GIT_DIRTY}

all: docker-build

fmt: lint/check ## ensure consistent code style
	go run oss.indeed.com/go/go-groups -w .
	golangci-lint run --fix > /dev/null 2>&1 || true

lint/check:
	@if ! golangci-lint --version > /dev/null 2>&1; then \
		echo -e "\033[0;33mgolangci-lint is not installed: run \`\033[0;32mmake lint-install\033[0m\033[0;33m\` or install it from https://golangci-lint.run\033[0m"; \
		exit 1; \
	fi

lint-install: ## installs golangci-lint to the go bin dir
	@if ! golangci-lint --version > /dev/null 2>&1; then \
		echo "Installing golangci-lint"; \
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOBIN) v1.30.0; \
	fi

lint: lint/check ## run golangci-lint
	golangci-lint run
	@if [ -n "$$(go run oss.indeed.com/go/go-groups -d .)" ]; then \
		echo -e "\033[0;33mdetected fmt problems: run \`\033[0;32mmake fmt\033[0m\033[0;33m\`\033[0m"; \
		exit 1; \
	fi

test: lint ## Run tests
	go test -v -race ./... -coverprofile cover.out

build: generate ## Build manager binary
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o bin/manager ./cmd/controller/main.go

manifests: controller-gen ## Generate CRD manifests
	$(CONTROLLER_GEN) "crd:crdVersions=v1" paths="./pkg/apis/..." output:crd:artifacts:config=deploy/crds
	$(CONTROLLER_GEN) "crd:crdVersions=v1beta1,preserveUnknownFields=false" paths="./pkg/apis/..." output:crd:artifacts:config=deploy/crds/legacy

generate: controller-gen ## Generate CRD code
	$(CONTROLLER_GEN) object:headerFile="build/boilerplate.go.txt" paths="./pkg/apis/..."

docker-build: manifests generate test build ## Build the docker image
	docker build . -t ${IMG}

docker-build-only: manifests build ## Build the docker image
	docker build . -t ${IMG}

docker-push: ## Push the docker image
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
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

help: ## displays this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_\/-]+:.*?## / {printf "\033[34m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | \
		sort | \
		grep -v '#'
