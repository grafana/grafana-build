package otel

import (
	"context"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func createExporterClient(ctx context.Context) otlptrace.Client {
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	if endpoint == "" {
		log.Print("OTEL_EXPORTER_OTLP_ENDPOINT not set. Disabling tracing.")
		return nil
	}
	if protocol == "grpc" {
		return otlptracegrpc.NewClient()
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return otlptracehttp.NewClient()
	}
	return otlptracegrpc.NewClient()
}

func Setup(ctx context.Context) func(context.Context) error {
	client := createExporterClient(ctx)
	if client == nil {
		return func(ctx context.Context) error {
			return nil
		}
	}
	exporter, err := otlptrace.New(ctx, client)

	if err != nil {
		log.Fatal(err)
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("grafana-build"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	otel.SetTextMapPropagator(propagation.TraceContext{})
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}

func FindParentTrace(ctx context.Context) context.Context {
	traceParent := os.Getenv("TRACEPARENT")
	if traceParent == "" {
		return ctx
	}
	log.Printf("Parent trace found: %s", traceParent)
	carrier := make(propagation.MapCarrier)
	carrier.Set("traceparent", traceParent)
	prop := otel.GetTextMapPropagator()
	ctx = prop.Extract(ctx, carrier)
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, span.SpanContext())
}

// Tracer is a simple wrapper around otel.Tracer in order to abstract that
// package.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func RecordFailed(span trace.Span, err error, msg string) {
	span.RecordError(err)
	span.SetStatus(codes.Error, msg)
}
