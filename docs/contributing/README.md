# Contributing

Below are a list of resources for contributing and developing secret-manager components:

## Welcome

Hello new or existing contributor! Welcome to our contributors documentation, this is the place where we document all our processes as well as how to compile, run and test secret-manager and all it’s components. This document sometimes can get out of date due to changes in the code and/or tooling. PRs to this part are also very welcome if you happen to find an out of date part.

## Building Secret-Manager

Secret-Manager uses Makefile to install dependencies and build secret-manager. Running `make help` lists developer facing Make targets with descriptions.

## Building the Go binaries (for local OS)

With Golang installed run `make build` to install any missing dependencies and build `secret-manager for the local OS.

## Building images

Docker can build the secret-manager image with `make docker-build`.

Building the Go binaries (for local OS)

## Testing the project

Secret-manager has two types of testing:

* Run `make test` to run any unit and integration tests. This requires [kubebuilder](https://book.kubebuilder.io/quick-start.html#installation) and its other utilities accessible in the local `$PATH`
* End-to-end testings with against a local kubernetes cluster with [kind](https://kind.sigs.k8s.io/). See [end-to-end testing](./e2e.md) for more information
