# Build the manager binary
ARG BUILDPLATFORM=linux/amd64
ARG GO_VERSION=1.15.3
FROM --platform=$BUILDPLATFORM golang:$GO_VERSION-buster as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd cmd/
COPY pkg pkg/
COPY Makefile Makefile

# Copy the dirs for make
COPY build build/
COPY .git .git/

# Build
ARG ARCHS
RUN make build-multiarch

FROM alpine:3.12

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /
LABEL maintainer="DirtyCajunRice,mcavoyk" \
  org.opencontainers.image.created=$BUILD_DATE \
  org.opencontainers.image.url="https://github.com/itscontained/secret-manager" \
  org.opencontainers.image.source="https://github.com/itscontained/secret-manager" \
  org.opencontainers.image.version=$VERSION \
  org.opencontainers.image.revision=$VCS_REF \
  org.opencontainers.image.vendor="itscontained" \
  org.opencontainers.image.title="secret-manager" \
  org.opencontainers.image.description="Secret Manager is a set of Kubernetes CRDs and controllers which define a common method of interacting with External SecretStores." \
  org.opencontainers.image.licenses="APACHE"
COPY --from=builder "/workspace/bin/manager-$TARGETOS-$TARGETARCH" /manager

# Run as UID for nobody since k8s pod securityContext runAsNonRoot can't resolve the user ID:
# https://github.com/kubernetes/kubernetes/issues/40958
USER 65534

ENTRYPOINT ["/manager"]
