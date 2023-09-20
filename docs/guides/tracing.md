# Tracing with OpenTelemetry

grafana-build also supports OpenTelemetry for tracing using either `http/protobuf` or `grpc` as protocols.

```
# Enable gRPC exporter (explicit via protocol)
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=https://tempo-somewhere.grafana.net:443
export OTEL_EXPORTER_OTLP_HEADERS=Authorization=...

# Enable gRPC exporter (implicit via no protocol in endpoint)
export OTEL_EXPORTER_OTLP_ENDPOINT=tempo-somewhere.grafana.net:443
export OTEL_EXPORTER_OTLP_HEADERS=Authorization=...

# Enable HTTP exporter (explicit via protocol)
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
export OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-gateway-somewhere.grafana.net/otlp
export OTEL_EXPORTER_OTLP_HEADERS=Authorization=...

# Enable HTTP exporter (implicit via protocol in endpoint)
export OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-gateway-somewhere.grafana.net/otlp
export OTEL_EXPORTER_OTLP_HEADERS=Authorization=...

```
