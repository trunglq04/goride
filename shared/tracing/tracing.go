package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string
	Environment    string
	JaegerEndpoint string
}

func InitTracer(cfg Config) (func(context.Context) error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Initialize the otlptracegrpc exporter
	traceExporter, err := newExporter(ctx, cfg.JaegerEndpoint) // choose OTLP tho Jaeger for export reports
	if err != nil {
		return nil, err
	}

	// 2. Set service details (required for Jaeger filtering) and  create the TracerProvider
	traceProvider, err := newTraceProvider(ctx, cfg, traceExporter)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(traceProvider)

	// Propagator
	otel.SetTextMapPropagator(newPropagator())

	return func(ctx context.Context) error {
		err := traceProvider.Shutdown(ctx)
		return err
	}, nil
}

func GetTracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}

func newExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	return otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // W3C Trace Context header (Trace-Id, Span-Id, Trace-State)
		propagation.Baggage{},      // W3C Baggage header
	)
}

func newTraceProvider(ctx context.Context, cfg Config, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironmentNameKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return traceProvider, nil
}
