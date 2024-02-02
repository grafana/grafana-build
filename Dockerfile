FROM golang:1.21-alpine

ARG DAGGER_VERSION=v0.9.8

WORKDIR /src
RUN apk add --no-cache git wget bash jq
RUN apk add --no-cache docker --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community

# Install dagger & otel-cli
RUN wget "https://github.com/dagger/dagger/releases/download/${DAGGER_VERSION}/dagger_${DAGGER_VERSION}_linux_amd64.tar.gz" && \
    tar -xvf ./dagger_${DAGGER_VERSION}_linux_amd64.tar.gz && mv dagger /bin && rm dagger_${DAGGER_VERSION}_linux_amd64.tar.gz && \
    wget -O /tmp/otel-cli.apk https://github.com/equinix-labs/otel-cli/releases/download/v0.4.0/otel-cli_0.4.0_linux_amd64.apk && \
    apk add --allow-untrusted /tmp/otel-cli.apk

ADD . .
RUN go build -o /src/grafana-build ./cmd

ENTRYPOINT ["dagger", "run", "/src/grafana-build"]
