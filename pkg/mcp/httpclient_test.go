package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

func TestHTTPClientInitializationAcceptsNoContentNotification(t *testing.T) {
	var initialized atomic.Bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var msg Message
		if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch msg.Method {
		case "initialize":
			result, err := json.Marshal(InitializeResult{
				ProtocolVersion: "2025-06-18",
				ServerInfo: ServerInfo{
					Name:    "test",
					Version: "1.0.0",
				},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set(SessionIDHeader, "test-session")
			_ = json.NewEncoder(w).Encode(Message{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Result:  result,
			})
		case "notifications/initialized":
			initialized.Store(true)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected method", http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	client, err := NewClient(t.Context(), "test", Server{BaseURL: ts.URL})
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}
	defer client.Close(false)

	if !initialized.Load() {
		t.Fatal("expected notifications/initialized")
	}
}

func TestHTTPClientNoContentMessageHandling(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		wantErr bool
	}{
		{
			name: "notification",
			message: Message{
				JSONRPC: "2.0",
				Method:  "notifications/test",
			},
		},
		{
			name: "response",
			message: Message{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`{}`),
			},
		},
		{
			name: "request",
			message: Message{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "ping",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			defer ts.Close()

			client, err := newHTTPClient("test", Server{BaseURL: ts.URL}, HTTPClientOptions{}, &SessionState{ID: "test-session"}, nil, false)
			if err != nil {
				t.Fatalf("newHTTPClient failed: %v", err)
			}
			if err := client.Start(t.Context(), func(context.Context, Message) {}); err != nil {
				t.Fatalf("failed to start client: %v", err)
			}
			defer client.Close(false)

			err = client.Send(t.Context(), tt.message)
			if tt.wantErr && err == nil {
				t.Fatal("expected request with no response to fail")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected message with no response to succeed: %v", err)
			}
		})
	}
}

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

func TestHTTPClientBlockingOptions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	blocked, err := newHTTPClient("test", Server{BaseURL: ts.URL}, HTTPClientOptions{
		BlockLoopback: true,
	}, nil, nil, false)
	if err != nil {
		t.Fatalf("newHTTPClient failed: %v", err)
	}

	_, err = blocked.oauthHandler.metadataClient.Get(ts.URL)
	if err == nil {
		t.Fatal("expected loopback metadata request to be blocked")
	}
	if !strings.Contains(err.Error(), "blocked loopback IP") {
		t.Fatalf("expected loopback block error, got %v", err)
	}

	_, err = blocked.httpClient.Get(ts.URL)
	if err == nil {
		t.Fatal("expected loopback MCP request to be blocked")
	}
	if !strings.Contains(err.Error(), "blocked loopback IP") {
		t.Fatalf("expected loopback block error, got %v", err)
	}

	serverURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	allowListed, err := newHTTPClient("test", Server{BaseURL: ts.URL}, HTTPClientOptions{
		BlockLoopback: true,
		AllowedHosts:  []string{serverURL.Host},
	}, nil, nil, false)
	if err != nil {
		t.Fatalf("newHTTPClient failed: %v", err)
	}

	resp, err := allowListed.oauthHandler.metadataClient.Get(ts.URL)
	if err != nil {
		t.Fatalf("expected allow-listed loopback metadata request to be allowed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	resp, err = allowListed.httpClient.Get(ts.URL)
	if err != nil {
		t.Fatalf("expected allow-listed loopback MCP request to be allowed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	allowed, err := newHTTPClient("test", Server{BaseURL: ts.URL}, HTTPClientOptions{}, nil, nil, false)
	if err != nil {
		t.Fatalf("newHTTPClient failed: %v", err)
	}

	resp, err = allowed.oauthHandler.metadataClient.Get(ts.URL)
	if err != nil {
		t.Fatalf("expected loopback metadata request to be allowed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	resp, err = allowed.httpClient.Get(ts.URL)
	if err != nil {
		t.Fatalf("expected loopback MCP request to be allowed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}
