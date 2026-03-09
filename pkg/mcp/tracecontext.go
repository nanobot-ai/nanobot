package mcp

import (
	"context"
	"crypto/rand"
	"fmt"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var mcpTracePropagator = propagation.NewCompositeTextMapPropagator(
	propagation.TraceContext{},
	propagation.Baggage{},
)

func ensureOutboundTraceContext(ctx context.Context) (context.Context, error) {
	if trace.SpanContextFromContext(ctx).IsValid() {
		return ctx, nil
	}

	var traceID trace.TraceID
	if _, err := rand.Read(traceID[:]); err != nil {
		return nil, fmt.Errorf("failed to generate trace ID: %w", err)
	}

	var spanID trace.SpanID
	if _, err := rand.Read(spanID[:]); err != nil {
		return nil, fmt.Errorf("failed to generate span ID: %w", err)
	}

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     false,
	})

	return trace.ContextWithSpanContext(ctx, spanContext), nil
}
