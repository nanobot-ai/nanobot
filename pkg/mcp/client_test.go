package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestServerUnmarshalJSONNoTools(t *testing.T) {
	var server Server
	if err := json.Unmarshal([]byte(`{"url":"https://example.com/mcp","noTools":true}`), &server); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !server.NoTools {
		t.Fatal("expected noTools true")
	}
}

func TestClientListToolsNoToolsDoesNotNeedSession(t *testing.T) {
	tools, err := (&Client{noTools: true}).ListTools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools.Tools) != 0 {
		t.Fatalf("expected no tools, got %#v", tools.Tools)
	}
}

func TestClientCallNoToolsDoesNotNeedSession(t *testing.T) {
	_, err := (&Client{noTools: true}).Call(context.Background(), "hidden", nil)
	if err == nil || !strings.Contains(err.Error(), "no tools available") {
		t.Fatalf("expected no tools error, got %v", err)
	}
}

func TestClientCallToolOverridesDisableAbsentTool(t *testing.T) {
	_, err := (&Client{toolOverrides: ToolOverrides{"allowed": {}}}).Call(context.Background(), "hidden", nil)
	if err == nil || !strings.Contains(err.Error(), `tool "hidden" not found`) {
		t.Fatalf("expected tool not found error, got %v", err)
	}
}
