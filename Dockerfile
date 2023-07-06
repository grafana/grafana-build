FROM golang:1.20-alpine

ARG DAGGER_VERSION=v0.6.3

WORKDIR /src

ADD . .
RUN apk add --update git docker wget bash
RUN go build -o /src/grafana-build ./cmd

# Install dagger
RUN wget "https://github.com/dagger/dagger/releases/download/${DAGGER_VERSION}/dagger_${DAGGER_VERSION}_linux_amd64.tar.gz"
RUN tar -xvf ./dagger_${DAGGER_VERSION}_linux_amd64.tar.gz && mv dagger /bin && rm dagger_${DAGGER_VERSION}_linux_amd64.tar.gz

ENTRYPOINT ["dagger", "run", "/src/grafana-build"]
