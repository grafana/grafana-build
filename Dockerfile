FROM golang:1.21-alpine

ARG DAGGER_VERSION=v0.8.4

WORKDIR /src
RUN apk add --no-cache git docker wget bash

# Install dagger
RUN wget "https://github.com/dagger/dagger/releases/download/${DAGGER_VERSION}/dagger_${DAGGER_VERSION}_linux_amd64.tar.gz" && \
    tar -xvf ./dagger_${DAGGER_VERSION}_linux_amd64.tar.gz && mv dagger /bin && rm dagger_${DAGGER_VERSION}_linux_amd64.tar.gz

ADD . .
RUN go build -o /src/grafana-build ./cmd

ENTRYPOINT ["dagger", "run", "/src/grafana-build"]
