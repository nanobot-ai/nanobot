package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/obot-platform/nanobot/pkg/mcp"
)

// fakeMCPServer is a minimal Streamable-HTTP MCP server that completes the
// initialize handshake and records DELETE (session termination) requests.
func fakeMCPServer(deletes *int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			atomic.AddInt32(deletes, 1)
			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			var msg struct {
				ID     json.RawMessage `json:"id"`
				Method string          `json:"method"`
			}
			_ = json.Unmarshal(body, &msg)
			if msg.Method == "initialize" {
				w.Header().Set("Mcp-Session-Id", "test-session-1")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":` + string(msg.ID) +
					`,"result":{"protocolVersion":"2025-06-18","capabilities":{},` +
					`"serverInfo":{"name":"fake","version":"1"}}}`))
				return
			}
			// notifications/initialized and anything else
			w.WriteHeader(http.StatusAccepted)
		case http.MethodGet:
			// SSE event stream is unused in this test
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

// Compile-time assert the factory is closable so session teardown reaps it.
var _ interface{ Close(bool) } = (*clientFactory)(nil)

// TestClientFactoryCloseDeletesUpstreamSession proves that closing a
// clientFactory terminates the upstream MCP session (sends DELETE), preventing
// the per-reconnect session leak observed against downstream MCP servers.
func TestClientFactoryCloseDeletesUpstreamSession(t *testing.T) {
	var deletes int32
	srv := fakeMCPServer(&deletes)
	defer srv.Close()

	ctx := context.Background()
	client, err := mcp.NewClient(ctx, "grocy", mcp.Server{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	f := newClientFactory(func(*mcp.SessionState) (*mcp.Client, error) {
		return client, nil
	})
	if _, err := f.get("envhash"); err != nil {
		t.Fatalf("factory.get failed: %v", err)
	}

	if got := atomic.LoadInt32(&deletes); got != 0 {
		t.Fatalf("expected no DELETE before close, got %d", got)
	}

	f.Close(false)

	if got := atomic.LoadInt32(&deletes); got != 1 {
		t.Fatalf("expected exactly one upstream DELETE after factory close, got %d", got)
	}
}
