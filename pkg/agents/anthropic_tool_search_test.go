package agents

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestSupportsAnthropicToolSearch(t *testing.T) {
	tests := map[string]bool{
		"claude-sonnet-4-6": true,
		"claude-opus-4-6":   true,
		"claude-haiku-4-5":  false,
		"gpt-5.2":           false,
	}

	for model, expected := range tests {
		if got := supportsAnthropicToolSearch(model); got != expected {
			t.Fatalf("supportsAnthropicToolSearch(%q) = %v, want %v", model, got, expected)
		}
	}
}

func TestShouldDeferAnthropicTool(t *testing.T) {
	tests := []struct {
		name    string
		mapping types.TargetMapping[types.TargetTool]
		want    bool
	}{
		{
			name: "obot imported server",
			mapping: types.TargetMapping[types.TargetTool]{
				MCPServer: "gmail",
			},
			want: true,
		},
		{
			name: "nanobot built in server",
			mapping: types.TargetMapping[types.TargetTool]{
				MCPServer: "nanobot.system",
			},
			want: false,
		},
		{
			name: "mcp search server",
			mapping: types.TargetMapping[types.TargetTool]{
				MCPServer: "mcp-server-search",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		if got := shouldDeferAnthropicTool(tt.mapping); got != tt.want {
			t.Fatalf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestEnsureAnthropicToolSearchTool(t *testing.T) {
	req := &types.CompletionRequest{
		Model: "claude-sonnet-4-6",
		Tools: []types.ToolUseDefinition{
			{
				Name: "gmail_search",
				Attributes: map[string]any{
					"defer_loading": true,
				},
			},
		},
	}

	ensureAnthropicToolSearchTool(req)
	ensureAnthropicToolSearchTool(req)

	found := 0
	for _, tool := range req.Tools {
		if tool.Name != anthropicToolSearchToolName {
			continue
		}
		found++
		if got := tool.Attributes["type"]; got != anthropicToolSearchToolType {
			t.Fatalf("tool search type = %v, want %q", got, anthropicToolSearchToolType)
		}
		if got := tool.Attributes["tool_search_type"]; got != "bm25_search" {
			t.Fatalf("tool search mode = %v, want %q", got, "bm25_search")
		}
	}

	if found != 1 {
		t.Fatalf("expected exactly one tool search tool, got %d", found)
	}
}

func TestAddToolsDefersExternalToolsForAnthropicSearch(t *testing.T) {
	agents := New(nil, tools.NewToolsService())
	req := &types.CompletionRequest{
		Model: "claude-sonnet-4-6",
	}

	_, err := agents.addTools(
		types.WithConfig(context.Background(), types.Config{}),
		req,
		&types.Agent{},
		[]types.CompletionOptions{{
			Tools: []mcp.Tool{
				{
					Name:        "search_code",
					Description: "Search code in a repository.",
					InputSchema: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`),
				},
			},
		}},
	)
	if err != nil {
		t.Fatalf("addTools returned error: %v", err)
	}

	var (
		foundDeferredTool bool
		foundSearchTool   bool
	)
	for _, tool := range req.Tools {
		switch tool.Name {
		case "search_code":
			foundDeferredTool = true
			if got := tool.Attributes["defer_loading"]; got != true {
				t.Fatalf("search_code defer_loading = %v, want true", got)
			}
		case anthropicToolSearchToolName:
			foundSearchTool = true
		}
	}

	if !foundDeferredTool {
		t.Fatal("expected external tool to be present in request")
	}
	if !foundSearchTool {
		t.Fatal("expected anthropic tool search tool to be added")
	}
}
