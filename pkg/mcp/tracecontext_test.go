package mcp

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestStartOutboundSpanExportsRealOperationSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = provider.Shutdown(context.Background())
	})

	ctx, span := startOutboundSpan(context.Background(), "mcp.tools.call",
		attribute.String("mcp.server.name", "obot"),
		attribute.String("mcp.tool.name", "search"),
	)
	if span == nil {
		t.Fatal("expected span")
	}
	finishOutboundSpan(span, nil)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 exported span, got %d", len(spans))
	}
	if got := spans[0].Name; got != "mcp.tools.call" {
		t.Fatalf("expected span name %q, got %q", "mcp.tools.call", got)
	}
	if got := trace.SpanContextFromContext(ctx); !got.IsValid() {
		t.Fatal("expected span context on returned context")
	}
	var serverName string
	for _, attr := range spans[0].Attributes {
		if attr.Key == "mcp.server.name" {
			serverName = attr.Value.AsString()
			break
		}
	}
	if serverName != "obot" {
		t.Fatalf("expected mcp.server.name %q, got %q", "obot", serverName)
	}
}

func TestFinishOutboundSpanRecordsErrors(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = provider.Shutdown(context.Background())
	})

	_, span := startOutboundSpan(context.Background(), "mcp.tools.call")
	finishOutboundSpan(span, context.Canceled)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 exported span, got %d", len(spans))
	}
	if len(spans[0].Events) == 0 {
		t.Fatal("expected recorded error event")
	}
}
