package mcp

import (
	"context"
	"errors"
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

// a lazy-auth server (e.g. BigQuery) returns 200 on initialize with no 401
// challenge. When the server is flagged OAuthRequired and no token is stored, initialize
// must still surface AuthRequiredErr so the standard OAuth flow runs.
func TestHTTPClientOAuthRequiredTriggersOnClean200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(SessionIDHeader, "sess-1")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`))
	}))
	defer ts.Close()

	initMsg, err := NewMessage("initialize", InitializeRequest{ProtocolVersion: "2025-06-18"})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}

	// OAuthRequired=true, no stored token (Start not called -> oauthTokenLoaded stays false):
	// a clean 200 must be turned into an auth-required error.
	c, err := newHTTPClient("test", Server{BaseURL: ts.URL, OAuthRequired: true}, HTTPClientOptions{}, nil, nil, false)
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	err = c.initialize(context.Background(), *initMsg)
	if _, ok := errors.AsType[AuthRequiredErr](err); !ok {
		t.Fatalf("OAuthRequired=true on clean 200: want AuthRequiredErr, got %v", err)
	}
	// The guard fields (oauthTokenLoaded / oauthStarted) are exercised via the e2e path;
	// here we keep to the synthetic-trigger assertion which returns before any wire I/O.
}
