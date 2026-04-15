package bifrost

import (
	"testing"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)


func TestToInput_ToolCallResultInUserMessage(t *testing.T) {
	callID := "call_abc123"
	req := &types.CompletionRequest{
		Input: []types.Message{
			{
				Role: "user",
				Items: []types.CompletionItem{
					{Content: &mcp.Content{Type: "text", Text: "use the tool"}},
				},
			},
			{
				Role: "assistant",
				Items: []types.CompletionItem{
					{
						ID: "item_1",
						ToolCall: &types.ToolCall{
							Name:      "my_tool",
							CallID:    callID,
							Arguments: `{"arg":"val"}`,
						},
					},
				},
			},
			{
				Role: "user",
				Items: []types.CompletionItem{
					{
						ToolCallResult: &types.ToolCallResult{
							CallID: callID,
							Output: types.CallResult{
								Content: []mcp.Content{{Type: "text", Text: "tool output"}},
							},
						},
					},
				},
			},
		},
	}

	msgs, err := toInput(req)
	if err != nil {
		t.Fatalf("toInput failed: %v", err)
	}

	// Expect: user message, function_call, function_call_output
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	// First: regular user message
	if msgs[0].Type == nil || *msgs[0].Type != schemas.ResponsesMessageTypeMessage {
		t.Errorf("msg[0]: expected type %q", schemas.ResponsesMessageTypeMessage)
	}

	// Second: function_call from assistant
	if msgs[1].Type == nil || *msgs[1].Type != schemas.ResponsesMessageTypeFunctionCall {
		t.Errorf("msg[1]: expected type %q", schemas.ResponsesMessageTypeFunctionCall)
	}

	// Third: function_call_output — not dropped
	if msgs[2].Type == nil || *msgs[2].Type != schemas.ResponsesMessageTypeFunctionCallOutput {
		t.Errorf("msg[2]: expected type %q, got %v", schemas.ResponsesMessageTypeFunctionCallOutput, msgs[2].Type)
	}
	if msgs[2].ResponsesToolMessage == nil || msgs[2].ResponsesToolMessage.CallID == nil || *msgs[2].ResponsesToolMessage.CallID != callID {
		t.Errorf("msg[2]: expected call_id %q", callID)
	}
	if msgs[2].ResponsesToolMessage.Output == nil || msgs[2].ResponsesToolMessage.Output.ResponsesToolCallOutputStr == nil ||
		*msgs[2].ResponsesToolMessage.Output.ResponsesToolCallOutputStr != "tool output" {
		t.Errorf("msg[2]: expected output %q", "tool output")
	}
}
