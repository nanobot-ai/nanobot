package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClientPassthroughHeaders(t *testing.T) {
	incoming := httptest.NewRequest(http.MethodPost, "http://nanobot.example/mcp", nil)
	incoming.Header.Set("X-Passthrough", "from-request")
	incoming.Header.Add("X-Passthrough", "from-request-1")
	incoming.Header.Set("X-Static", "from-request")
	incoming.Header.Set("X-Not-Allowed", "from-request")

	client, err := newHTTPClient("test", Server{
		BaseURL: "http://mcp.example/mcp",
		Headers: map[string]string{
			"X-Static": "from-config",
		},
		PassthroughHeaders: []string{"X-Passthrough", "X-Static", "X-Not-Present"},
	}, HTTPClientOptions{}, nil, map[string]string{
		"X-Static": "from-config",
	}, false)
	if err != nil {
		t.Fatalf("newHTTPClient failed: %v", err)
	}

	outgoing, err := client.newRequest(WithRequest(context.Background(), incoming), http.MethodPost, nil)
	if err != nil {
		t.Fatalf("newRequest failed: %v", err)
	}

	passthrough := outgoing.Header.Values("X-Passthrough")
	if len(passthrough) != 2 || passthrough[0] != "from-request" || passthrough[1] != "from-request-1" {
		t.Fatalf("X-Passthrough = %v, want %q", passthrough, []string{"from-request", "from-request-1"})
	}
	if got := outgoing.Header.Get("X-Static"); got != "from-config" {
		t.Fatalf("X-Static = %q, want static config value", got)
	}
	if got := outgoing.Header.Get("X-Not-Allowed"); got != "" {
		t.Fatalf("X-Not-Allowed = %q, want empty", got)
	}
}
