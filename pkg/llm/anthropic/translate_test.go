package anthropic

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
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

// TestToRequestOmitsSamplingControls ensures deprecated temperature and top_p fields are omitted.
func TestToRequestOmitsSamplingControls(t *testing.T) {
	temperature := json.Number("0.7")
	topP := json.Number("0.9")

	req, err := toRequest(&types.CompletionRequest{
		Model:       "claude-opus-4-8",
		Temperature: &temperature,
		TopP:        &topP,
	})
	if err != nil {
		t.Fatalf("toRequest failed: %v", err)
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}
	if _, ok := raw["temperature"]; ok {
		t.Fatalf("request includes temperature: %s", data)
	}
	if _, ok := raw["top_p"]; ok {
		t.Fatalf("request includes top_p: %s", data)
	}
}
