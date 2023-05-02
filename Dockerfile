FROM golang:1.20

WORKDIR /src

ADD . .
RUN go build -o /src/grafana-build ./cmd

ENTRYPOINT ["/src/grafana-build"]
