package mcp

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var mcpTracePropagator = propagation.NewCompositeTextMapPropagator(
	propagation.TraceContext{},
	propagation.Baggage{},
)

var mcpTracer = otel.Tracer("nanobot/mcp")

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

func startInboundMessageSpan(ctx context.Context, msg Message, sessionID string) (context.Context, trace.Span) {
	name := "mcp.message"
	if msg.Method != "" {
		name = "mcp." + msg.Method
	}

	attrs := []attribute.KeyValue{
		attribute.String("rpc.system", "mcp"),
	}
	if msg.Method != "" {
		attrs = append(attrs, attribute.String("mcp.method", msg.Method))
	}
	if msg.ID != nil {
		attrs = append(attrs, attribute.String("mcp.request_id", MessageIDString(msg.ID)))
	}
	if sessionID != "" {
		attrs = append(attrs, attribute.String("mcp.session_id", sessionID))
	}

	switch msg.Method {
	case "resources/read":
		if uri := gjson.GetBytes(msg.Params, "uri").String(); uri != "" {
			attrs = append(attrs, attribute.String("mcp.resource.uri", uri))
		}
	case "tools/call", "prompts/get":
		if name := gjson.GetBytes(msg.Params, "name").String(); name != "" {
			attrs = append(attrs, attribute.String("mcp.name", name))
		}
	}

	return mcpTracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attrs...))
}

func recordInboundMessageSpanError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
