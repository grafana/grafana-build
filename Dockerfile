FROM golang:1.20-alpine

WORKDIR /src

ADD . .
RUN apk add --update git docker
RUN go build -o /src/grafana-build ./cmd

ENTRYPOINT ["/src/grafana-build"]
