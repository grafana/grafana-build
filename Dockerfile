FROM ghcr.io/equinix-labs/otel-cli:v0.4.5@sha256:982493a80842650a6a83aab9de84f0db627a5304e1b7ec37c347c4d8c7509565 as otel-cli

FROM alpine:3.20@sha256:a4f4213abb84c497377b8544c81b3564f313746700372ec4fe84653e4fb03805 AS dagger

# TODO: pull the binary from registry.dagger.io/cli:v0.9.8 (or similar) when
# https://github.com/dagger/dagger/issues/6887 is resolved
ARG DAGGER_VERSION=v0.13.3
ADD https://github.com/dagger/dagger/releases/download/${DAGGER_VERSION}/dagger_${DAGGER_VERSION}_linux_amd64.tar.gz /tmp
RUN tar zxf /tmp/dagger_${DAGGER_VERSION}_linux_amd64.tar.gz -C /tmp
RUN mv /tmp/dagger /bin/dagger

FROM golang:1.23-alpine@sha256:383395b794dffa5b53012a212365d40c8e37109a626ca30d6151c8348d380b5f

ARG DAGGER_VERSION=v0.13.3

WORKDIR /src
RUN apk add --no-cache git wget bash jq
RUN apk add --no-cache docker --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community

COPY --from=otel-cli /otel-cli /usr/bin/otel-cli
COPY --from=dagger /bin/dagger /bin/dagger

ADD . .
RUN go build -o /src/grafana-build ./cmd

ENTRYPOINT ["dagger", "run", "/src/grafana-build"]
