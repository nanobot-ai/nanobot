package mcp

import (
	"encoding/json"
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
