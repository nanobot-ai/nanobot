package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider wraps the OTEL tracer provider for lifecycle management
type TracerProvider struct {
	provider *sdktrace.TracerProvider
}

// Tracer returns a tracer with the given name
func (tp *TracerProvider) Tracer(name string) trace.Tracer {
	if tp == nil || tp.provider == nil {
		return otel.Tracer(name)
	}
	return tp.provider.Tracer(name)
}

// Shutdown shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp == nil || tp.provider == nil {
		return nil
	}
	return tp.provider.Shutdown(ctx)
}

// InitTracer initializes OpenTelemetry tracing with OTLP exporter
// Protocol can be "http", "grpc", or "http/protobuf" (defaults to http)
func InitTracer(ctx context.Context, serviceName, endpoint, protocol string) (*TracerProvider, error) {
	if endpoint == "" {
		// If no endpoint is configured, return nil (tracing disabled)
		return nil, nil
	}

	// Normalize protocol
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	if protocol == "" {
		protocol = "http"
	}

	// Create appropriate exporter based on protocol
	var exporter sdktrace.SpanExporter
	var err error

	switch protocol {
	case "grpc":
		exporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(), // Use insecure for now, can be made configurable
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP gRPC exporter: %w", err)
		}

	case "http", "http/protobuf":
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(), // Use insecure for now, can be made configurable
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP HTTP exporter: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported OTEL protocol: %s (supported: http, grpc, http/protobuf)", protocol)
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set as global tracer provider
	otel.SetTracerProvider(tp)

	return &TracerProvider{provider: tp}, nil
}

// ShutdownWithTimeout shuts down the tracer provider with a timeout
func (tp *TracerProvider) ShutdownWithTimeout(timeout time.Duration) error {
	if tp == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return tp.Shutdown(ctx)
}
