package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/obot-platform/nanobot/pkg/mcp/auditlogs"
)

type recordingAuditLogCollector struct {
	mu      sync.Mutex
	entries []auditlogs.MCPAuditLog
}

func (c *recordingAuditLogCollector) CollectMCPAuditEntry(entry auditlogs.MCPAuditLog) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, entry)
}

func (*recordingAuditLogCollector) Close() {}

func (c *recordingAuditLogCollector) entry(callType string) (auditlogs.MCPAuditLog, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, entry := range c.entries {
		if entry.CallType == callType {
			return entry, true
		}
	}
	return auditlogs.MCPAuditLog{}, false
}

func TestHTTPServerNotificationAuditLogHasProcessingTime(t *testing.T) {
	collector := new(recordingAuditLogCollector)
	handler := MessageHandlerFunc(func(ctx context.Context, msg Message) {
		if msg.Method == "initialize" {
			if err := msg.Reply(ctx, InitializeResult{
				ProtocolVersion: "2025-06-18",
				ServerInfo: ServerInfo{
					Name:    "test-server",
					Version: "1.0.0",
				},
			}); err != nil {
				t.Errorf("reply to initialize: %v", err)
			}
		}
	})
	server, err := NewHTTPServer(nil, handler, HTTPServerOptions{
		BaseContext:       t.Context(),
		AuditLogCollector: collector,
	})
	if err != nil {
		t.Fatalf("create HTTP server: %v", err)
	}

	initialize := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`,
	))
	initializeResponse := httptest.NewRecorder()
	server.ServeHTTP(initializeResponse, initialize)
	if initializeResponse.Code != http.StatusOK {
		t.Fatalf("initialize status = %d, want %d; body: %s", initializeResponse.Code, http.StatusOK, initializeResponse.Body.String())
	}

	initialized := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
	))
	initialized.Header.Set("Mcp-Session-Id", initializeResponse.Header().Get("Mcp-Session-Id"))
	initializedResponse := httptest.NewRecorder()
	server.ServeHTTP(initializedResponse, initialized)
	if initializedResponse.Code != http.StatusAccepted {
		t.Fatalf("initialized status = %d, want %d; body: %s", initializedResponse.Code, http.StatusAccepted, initializedResponse.Body.String())
	}

	auditLog, ok := collector.entry("notifications/initialized")
	if !ok {
		t.Fatal("notifications/initialized audit log was not collected")
	}
	if auditLog.ProcessingTimeMs < 1 {
		t.Fatalf("ProcessingTimeMs = %d, want at least 1", auditLog.ProcessingTimeMs)
	}
}
