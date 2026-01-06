package middleware

import (
	"net/http"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware wraps an HTTP handler with OpenTelemetry tracing
func TracingMiddleware(next http.Handler, tp *telemetry.TracerProvider) http.Handler {
	if tp == nil {
		return next
	}

	// Wrap the handler to add custom attributes after otelhttp creates the span
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add MCP-specific attributes to the span created by otelhttp
		AddSpanAttributes(r)
		next.ServeHTTP(w, r)
	})

	// Use otelhttp for automatic HTTP instrumentation
	return otelhttp.NewHandler(wrapped, "http.request",
		otelhttp.WithSpanOptions(
			trace.WithAttributes(
				attribute.String("service.name", "nanobot"),
				attribute.String("http.scheme", "http"),
			),
		),
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			// Custom span name: METHOD /path
			return r.Method + " " + r.URL.Path
		}),
		otelhttp.WithFilter(func(r *http.Request) bool {
			// Don't trace health check endpoints
			return r.URL.Path != "/health" && r.URL.Path != "/healthz"
		}),
	)
}

// AddSpanAttributes adds MCP-specific attributes to the current span
func AddSpanAttributes(r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	if !span.IsRecording() {
		return
	}

	// Add MCP session ID if present
	if sessionID := r.Header.Get(mcp.SessionIDHeader); sessionID != "" {
		span.SetAttributes(attribute.String("mcp.session_id", sessionID))
	}

	// Add user agent
	if ua := r.Header.Get("User-Agent"); ua != "" {
		span.SetAttributes(attribute.String("http.user_agent", ua))
	}

	// Add client IP (respecting X-Forwarded-For)
	clientIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		clientIP = xff
	}
	span.SetAttributes(attribute.String("http.client_ip", clientIP))
}
