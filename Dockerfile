FROM ghcr.io/equinix-labs/otel-cli:v0.4.5 as otel-cli

FROM alpine:3.20 AS dagger

# TODO: pull the binary from registry.dagger.io/cli:v0.9.8 (or similar) when
# https://github.com/dagger/dagger/issues/6887 is resolved
ARG DAGGER_VERSION=v0.13.3
ADD https://github.com/dagger/dagger/releases/download/${DAGGER_VERSION}/dagger_${DAGGER_VERSION}_linux_amd64.tar.gz /tmp
RUN tar zxf /tmp/dagger_${DAGGER_VERSION}_linux_amd64.tar.gz -C /tmp
RUN mv /tmp/dagger /bin/dagger

FROM golang:1.23-alpine

ARG DAGGER_VERSION=v0.13.3

WORKDIR /src
RUN apk add --no-cache git wget bash jq
RUN apk add --no-cache docker --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community

COPY --from=otel-cli /otel-cli /usr/bin/otel-cli
COPY --from=dagger /bin/dagger /bin/dagger

ADD . .
RUN go build -o /src/grafana-build ./cmd

ENTRYPOINT ["dagger", "run", "/src/grafana-build"]
