// Package telemetry provides OpenTelemetry setup for Ginger applications.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds telemetry configuration.
type Config struct {
	ServiceName    string
	ServiceVersion string
	// Exporter selects the trace exporter: "stdout" (default) or "otlp".
	// For OTLP set OTEL_EXPORTER_OTLP_ENDPOINT in the environment.
	Exporter string
}

// Provider wraps the OTel TracerProvider and exposes Shutdown.
type Provider struct {
	tp *sdktrace.TracerProvider
}

// Setup initialises the global OTel tracer provider.
// Call provider.Shutdown(ctx) in your OnStop hook.
func Setup(ctx context.Context, cfg Config) (*Provider, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: resource: %w", err)
	}

	exp, err := newExporter(ctx, cfg.Exporter)
	if err != nil {
		return nil, fmt.Errorf("telemetry: exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return &Provider{tp: tp}, nil
}

// Shutdown flushes pending spans and stops the provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	return p.tp.Shutdown(ctx)
}

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// newExporter builds the span exporter selected by kind.
// "otlp" uses OTLP/HTTP; anything else falls back to stdout.
func newExporter(ctx context.Context, kind string) (sdktrace.SpanExporter, error) {
	if kind == "otlp" {
		return otlptracehttp.New(ctx)
	}
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}
