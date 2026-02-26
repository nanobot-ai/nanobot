package meta

import (
	"context"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

func TestMetaToolsOnlyExposeUpdateChat(t *testing.T) {
	s := NewServer(nil)

	tools, err := s.tools.List(context.Background(), mcp.Message{}, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("tools.List() failed: %v", err)
	}

	if len(tools.Tools) != 1 {
		t.Fatalf("expected exactly one meta tool, got %d", len(tools.Tools))
	}
	if tools.Tools[0].Name != "update_chat" {
		t.Fatalf("tool name = %q, want %q", tools.Tools[0].Name, "update_chat")
	}
}
