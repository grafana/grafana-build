package main

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func setupOTEL(ctx context.Context) func(context.Context) error {
	client := otlptracehttp.NewClient()
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

func findParentTrace(ctx context.Context) context.Context {
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
