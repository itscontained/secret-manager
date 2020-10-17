MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash
.SHELLFLAGS := -euo pipefail -c

IMG_TAG=test
IMG=itscontained/secret-manager-e2e:test

start-kind: ## Start kind cluster
	kind create cluster \
	  --name secret-manager \
	  --config kind.yaml \
	  --retain \
	  --image "kindest/node:v1.19.1"

test: e2e-image ## Run e2e tests against current kube context
	$(MAKE) -C ../ docker-build IMG_TAG=$(IMG_TAG) DOCKER_BUILD_FLAGS="--build-arg ARCHS=amd64"
	kind load docker-image --name="secret-manager" itscontained/secret-manager:$(IMG_TAG)
	kind load docker-image --name="secret-manager" $(IMG)
	./run.sh

e2e-bin:
	CGO_ENABLED=0 ginkgo build .

e2e-image: e2e-bin
	-rm -rf ./k8s/deploy
	mkdir -p k8s
	$(MAKE) -C ../ crds-to-chart
	cp -r ../deploy ./k8s
	docker build -t $(IMG) .

help: ## displays this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_\/-]+:.*?## / {printf "\033[34m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | \
		sort | \
		grep -v '#'
