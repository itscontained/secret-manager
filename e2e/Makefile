IMG=secret-manager-e2e:test

start-kind: ## Start kind cluster
	kind create cluster \
	  --name secret-manager \
	  --config kind.yaml \
	  --retain \
	  --image "kindest/node:v1.19.1"

test: e2e-image ## Run e2e tests against current kube context
	$(MAKE) -C ../ docker-build IMG=secret-manager:test
	kind load docker-image --name="secret-manager" secret-manager:test
	kind load docker-image --name="secret-manager" secret-manager-e2e:test
	./run.sh

e2e-bin:
	CGO_ENABLED=0 ginkgo build .

e2e-image: e2e-bin
	-rm -rf ./k8s/deploy
	mkdir -p k8s
	$(MAKE) -C ../ crds-to-chart
	cp -r ../deploy ./k8s
	docker build -t ${IMG} .

help: ## displays this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_\/-]+:.*?## / {printf "\033[34m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | \
		sort | \
		grep -v '#'
