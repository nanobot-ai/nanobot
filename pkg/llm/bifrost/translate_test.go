package bifrost

import (
	"encoding/json"
	"testing"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestToResponse(t *testing.T) {
	var resp schemas.BifrostResponsesResponse
	if err := json.Unmarshal([]byte(inputExample), &resp); err != nil {
		t.Fatalf("failed to unmarshal example: %v", err)
	}

	got, err := toResponse(&resp)
	if err != nil {
		t.Fatalf("toResponse failed: %v", err)
	}

	if got.Model != "us.anthropic.claude-sonnet-4-6" {
		t.Errorf("expected model %q, got %q", "us.anthropic.claude-sonnet-4-6", got.Model)
	}
	if got.Output.Role != "assistant" {
		t.Errorf("expected role %q, got %q", "assistant", got.Output.Role)
	}
	if len(got.Output.Items) != 1 {
		t.Fatalf("expected 1 output item, got %d", len(got.Output.Items))
	}
	item := got.Output.Items[0]
	if item.Content == nil {
		t.Fatal("expected content, got nil")
	}
	if item.Content.Type != "text" {
		t.Errorf("expected content type %q, got %q", "text", item.Content.Type)
	}
	if item.Content.Text != "Hello there, friend!" {
		t.Errorf("expected text %q, got %q", "Hello there, friend!", item.Content.Text)
	}
}

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

var inputExample = `{
  "id": "7e33cc07-9967-4b2c-b57c-e692030d5e95",
  "object": "",
  "created_at": 1775777817,
  "completed_at": null,
  "error": null,
  "incomplete_details": null,
  "instructions": null,
  "max_output_tokens": null,
  "max_tool_calls": null,
  "model": "us.anthropic.claude-sonnet-4-6",
  "output": [
    {
      "id": "msg_1775777817605391000",
      "type": "message",
      "status": "completed",
      "role": "assistant",
      "content": [
        {
          "type": "output_text",
          "text": "Hello there, friend!",
          "annotations": [],
          "logprobs": []
        }
      ]
    }
  ],
  "previous_response_id": null,
  "prompt_cache_key": null,
  "reasoning": null,
  "safety_identifier": null,
  "service_tier": null,
  "tools": null,
  "usage": {
  "input_tokens": 15,
    "input_tokens_details": null,
    "output_tokens": 8,
    "output_tokens_details": null,
    "total_tokens": 23
  },
  "extra_fields": {
    "request_type": "responses",
    "provider": "bedrock",
    "latency": 4803,
    "chunk_index": 0,
    "provider_response_headers": {
      "X-Amzn-Requestid": "acea774e-69fc-45a9-b4cc-81e3ec09e595"
    }
  }
}`
