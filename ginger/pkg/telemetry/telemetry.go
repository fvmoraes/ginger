// Package telemetry provides OpenTelemetry setup for Ginger applications.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
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
	// Exporter: "stdout" | "otlp" (extend as needed)
	Exporter string
}

// Provider wraps the OTel TracerProvider with a shutdown function.
type Provider struct {
	tp *sdktrace.TracerProvider
}

// Setup initializes the global OTel tracer provider.
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

// Shutdown flushes and stops the tracer provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	return p.tp.Shutdown(ctx)
}

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func newExporter(ctx context.Context, kind string) (sdktrace.SpanExporter, error) {
	// Default to stdout; swap for OTLP in production.
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}
