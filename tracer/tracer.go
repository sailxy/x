package tracer

import (
	"context"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type StdoutConfig struct {
	ServiceName string
}

func InitStdoutTracer(c StdoutConfig) error {
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(io.Discard))
	if err != nil {
		return err
	}
	tp := newTracerProvider(exporter, c.ServiceName)
	otel.SetTracerProvider(tp)
	return nil
}

type HTTPConfig struct {
	ServiceName string
	Endpoint    string
}

func InitHTTPTracer(c HTTPConfig) error {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(c.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return err
	}
	tp := newTracerProvider(exporter, c.ServiceName)
	otel.SetTracerProvider(tp)
	return nil
}

func newTracerProvider(exporter trace.SpanExporter, serviceName string) *trace.TracerProvider {
	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
}
