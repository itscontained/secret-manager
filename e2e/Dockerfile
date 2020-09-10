FROM golang:1.14 as BASE

ENV KUBECTL_VERSION="v1.19.0"
ENV HELM_VERSION="v3.3.1"

RUN go get -u github.com/onsi/ginkgo/ginkgo
RUN wget -q https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl && \
    wget -q https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm && \
    chmod +x /usr/local/bin/helm

FROM alpine:3.11
RUN apk add -U --no-cache \
    ca-certificates \
    bash \
    curl \
    tzdata \
    libc6-compat \
    openssl

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/
COPY --from=BASE /usr/local/bin/kubectl /usr/local/bin/
COPY --from=BASE /usr/local/bin/helm /usr/local/bin/

COPY entrypoint.sh                  /entrypoint.sh
COPY e2e.test                       /e2e.test
COPY wait-for-secret-manager.sh     /wait-for-secret-manager.sh
COPY wait-for-localstack.sh         /wait-for-localstack.sh
COPY k8s                            /k8s
COPY localstack.deployment.yaml     /localstack.deployment.yaml

CMD [ "/entrypoint.sh" ]
