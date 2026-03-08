package anthropic

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestToRequestDropsResourceLinkContent(t *testing.T) {
	req := types.CompletionRequest{
		Model: "claude-opus-4-6",
		Input: []types.Message{
			{
				Role: "user",
				Items: []types.CompletionItem{
					{
						Content: &mcp.Content{
							Type: "text",
							Text: "what's in this file",
						},
					},
					{
						Content: &mcp.Content{
							Type:     "resource_link",
							Name:     "screenshot.png",
							URI:      "file:///screenshot.png",
							MIMEType: "image/png",
						},
					},
				},
			},
			{
				Role: "user",
				Items: []types.CompletionItem{
					{
						Content: &mcp.Content{
							Type: "text",
							Text: "The user has attached the following file \"screenshot.png\".",
						},
					},
				},
			},
		},
	}

	anthropicReq, err := toRequest(&req)
	if err != nil {
		t.Fatalf("toRequest failed: %v", err)
	}

	if len(anthropicReq.Messages) != 2 {
		t.Fatalf("expected 2 messages after dropping resource link, got %d", len(anthropicReq.Messages))
	}

	for i, msg := range anthropicReq.Messages {
		if msg.Content == nil {
			t.Fatalf("message %d has nil content", i)
		}
		if len(msg.Content) == 0 {
			t.Fatalf("message %d has empty content", i)
		}
	}

	data, err := json.Marshal(anthropicReq)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	if strings.Contains(string(data), `"content":null`) {
		t.Fatalf("request still contains null content: %s", data)
	}
}

func TestToRequestSupportsToolSearchAndDeferredTools(t *testing.T) {
	req := types.CompletionRequest{
		Model: "claude-sonnet-4-6",
		Tools: []types.ToolUseDefinition{
			{
				Name:        "gmail_search",
				Parameters:  json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`),
				Description: "Search Gmail messages",
				Attributes: map[string]any{
					"defer_loading": true,
				},
			},
			{
				Name: "anthropic_tool_search",
				Attributes: map[string]any{
					"type": "tool_search_tool_bm25_20251119",
				},
			},
		},
	}

	anthropicReq, err := toRequest(&req)
	if err != nil {
		t.Fatalf("toRequest failed: %v", err)
	}

	data, err := json.Marshal(anthropicReq)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	payload := string(data)
	if !strings.Contains(payload, `"defer_loading":true`) {
		t.Fatalf("expected defer_loading flag in payload: %s", payload)
	}
	if !strings.Contains(payload, `"type":"tool_search_tool_bm25_20251119"`) {
		t.Fatalf("expected tool search tool type in payload: %s", payload)
	}
}

func TestToResponseIgnoresToolSearchBlocks(t *testing.T) {
	text := "Done."
	resp, err := toResponse(&Response{
		ID:    "msg_123",
		Model: "claude-sonnet-4-6",
		Content: []Content{
			{
				Type:       "server_tool_use",
				Name:       "anthropic_tool_search",
				RawContent: json.RawMessage(`{"query":"gmail unread messages"}`),
			},
			{
				Type:       "tool_search_tool_result",
				RawContent: json.RawMessage(`{"results":[{"tool":"gmail_search"}]}`),
			},
			{
				Type:  "tool_use",
				ID:    "toolu_123",
				Name:  "gmail_search",
				Input: map[string]any{"query": "unread"},
			},
			{
				Type: "text",
				Text: &text,
			},
		},
	}, time.Now())
	if err != nil {
		t.Fatalf("toResponse failed: %v", err)
	}

	if len(resp.Output.Items) != 2 {
		t.Fatalf("expected only tool_use + text items, got %d", len(resp.Output.Items))
	}
	if resp.Output.Items[0].ToolCall == nil || resp.Output.Items[0].ToolCall.Name != "gmail_search" {
		t.Fatalf("expected first item to be the discovered tool call, got %+v", resp.Output.Items[0])
	}
	if resp.Output.Items[1].Content == nil || resp.Output.Items[1].Content.Text != text {
		t.Fatalf("expected final text item, got %+v", resp.Output.Items[1])
	}
}
