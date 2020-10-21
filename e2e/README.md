# e2e testing

* tests runs from a pod inside the cluster because we need to speak to localstack (and possibly more secret backends)
* e2e tests cover installation via helm and secret-manager integration into secret backend APIs

## Running e2e tests
### Prerequisites

* docker
* kind `v.0.9.0` or greater
* ginkgo
* kubectl

## Executing tests
Running e2e tests locally:

```shell
$ make start-kind
$ make test
```
