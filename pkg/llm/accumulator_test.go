package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestNewProgressAccumulator(t *testing.T) {
	token := "test-token"
	acc := newProgressAccumulator(token)

	if acc == nil {
		t.Fatal("expected non-nil accumulator")
	}

	if acc.progressToken != token {
		t.Errorf("expected token %v, got %v", token, acc.progressToken)
	}

	if acc.response.InternalMessages == nil {
		t.Error("expected InternalMessages to be initialized")
	}

	if len(acc.response.InternalMessages) != 0 {
		t.Errorf("expected empty InternalMessages, got %d items", len(acc.response.InternalMessages))
	}
}

func TestCaptureProgress_NewMessage(t *testing.T) {
	acc := newProgressAccumulator("token")

	prog := &types.CompletionProgress{
		Model:     "gpt-4",
		Agent:     "test-agent",
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-1",
			Content: &mcp.Content{
				Type: "text",
				Text: "Hello",
			},
		},
	}

	acc.captureProgress(context.Background(), prog)

	if acc.response.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", acc.response.Model)
	}

	if acc.response.Agent != "test-agent" {
		t.Errorf("expected agent test-agent, got %s", acc.response.Agent)
	}

	if len(acc.response.InternalMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(acc.response.InternalMessages))
	}

	msg := acc.response.InternalMessages[0]
	if msg.ID != "msg-1" {
		t.Errorf("expected message ID msg-1, got %s", msg.ID)
	}

	if msg.Role != "assistant" {
		t.Errorf("expected role assistant, got %s", msg.Role)
	}

	if !msg.HasMore {
		t.Error("expected HasMore to be true")
	}

	if len(msg.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(msg.Items))
	}

	if msg.Items[0].Content.Text != "Hello" {
		t.Errorf("expected text 'Hello', got '%s'", msg.Items[0].Content.Text)
	}
}

func TestCaptureProgress_DefaultRole(t *testing.T) {
	acc := newProgressAccumulator("token")

	prog := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "", // Empty role should default to "assistant"
		Item: types.CompletionItem{
			ID: "item-1",
			Content: &mcp.Content{
				Type: "text",
				Text: "Hello",
			},
		},
	}

	acc.captureProgress(context.Background(), prog)

	if len(acc.response.InternalMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(acc.response.InternalMessages))
	}

	if acc.response.InternalMessages[0].Role != "assistant" {
		t.Errorf("expected default role 'assistant', got '%s'", acc.response.InternalMessages[0].Role)
	}
}

func TestCaptureProgress_NewItemInExistingMessage(t *testing.T) {
	acc := newProgressAccumulator("token")

	// First progress - creates message and first item
	prog1 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-1",
			Content: &mcp.Content{
				Type: "text",
				Text: "First",
			},
		},
	}
	acc.captureProgress(context.Background(), prog1)

	// Second progress - adds new item to same message
	prog2 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-2",
			Content: &mcp.Content{
				Type: "text",
				Text: "Second",
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	if len(acc.response.InternalMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(acc.response.InternalMessages))
	}

	msg := acc.response.InternalMessages[0]
	if len(msg.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(msg.Items))
	}

	if msg.Items[0].Content.Text != "First" {
		t.Errorf("expected first item 'First', got '%s'", msg.Items[0].Content.Text)
	}

	if msg.Items[1].Content.Text != "Second" {
		t.Errorf("expected second item 'Second', got '%s'", msg.Items[1].Content.Text)
	}
}

func TestCaptureProgress_PartialTextAccumulation(t *testing.T) {
	acc := newProgressAccumulator("token")

	// First partial text
	prog1 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			Content: &mcp.Content{
				Type: "text",
				Text: "Hello ",
			},
		},
	}
	acc.captureProgress(context.Background(), prog1)

	// Second partial text - should append
	prog2 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			Content: &mcp.Content{
				Type: "text",
				Text: "World",
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	if len(acc.response.InternalMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(acc.response.InternalMessages))
	}

	msg := acc.response.InternalMessages[0]
	if len(msg.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(msg.Items))
	}

	if msg.Items[0].Content.Text != "Hello World" {
		t.Errorf("expected accumulated text 'Hello World', got '%s'", msg.Items[0].Content.Text)
	}
}

func TestCaptureProgress_NonPartialReplacement(t *testing.T) {
	acc := newProgressAccumulator("token")

	// First partial text
	prog1 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			Content: &mcp.Content{
				Type: "text",
				Text: "Partial ",
			},
		},
	}
	acc.captureProgress(context.Background(), prog1)

	// Complete (non-partial) item - should replace
	prog2 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: false,
			Content: &mcp.Content{
				Type: "text",
				Text: "Complete",
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	if len(acc.response.InternalMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(acc.response.InternalMessages))
	}

	msg := acc.response.InternalMessages[0]
	if len(msg.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(msg.Items))
	}

	if msg.Items[0].Content.Text != "Complete" {
		t.Errorf("expected replaced text 'Complete', got '%s'", msg.Items[0].Content.Text)
	}
}

func TestCaptureProgress_ToolCallAccumulation(t *testing.T) {
	acc := newProgressAccumulator("token")

	// First partial tool call - name only
	prog1 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			ToolCall: &types.ToolCall{
				CallID:    "call-1",
				Name:      "test_tool",
				Arguments: "{\"arg",
			},
		},
	}
	acc.captureProgress(context.Background(), prog1)

	// Second partial tool call - append arguments
	prog2 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			ToolCall: &types.ToolCall{
				Arguments: "\": \"value\"}",
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	if len(acc.response.InternalMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(acc.response.InternalMessages))
	}

	msg := acc.response.InternalMessages[0]
	if len(msg.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(msg.Items))
	}

	toolCall := msg.Items[0].ToolCall
	if toolCall == nil {
		t.Fatal("expected tool call to be set")
	}

	if toolCall.Name != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got '%s'", toolCall.Name)
	}

	if toolCall.Arguments != "{\"arg\": \"value\"}" {
		t.Errorf("expected accumulated arguments, got '%s'", toolCall.Arguments)
	}
}

func TestCaptureProgress_ReasoningAccumulation(t *testing.T) {
	acc := newProgressAccumulator("token")

	// First partial reasoning
	prog1 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			Reasoning: &types.Reasoning{
				Summary: []types.SummaryText{
					{Text: "First "},
				},
			},
		},
	}
	acc.captureProgress(context.Background(), prog1)

	// Second partial reasoning - should append
	prog2 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			Reasoning: &types.Reasoning{
				Summary: []types.SummaryText{
					{Text: "Second"},
				},
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	if len(acc.response.InternalMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(acc.response.InternalMessages))
	}

	msg := acc.response.InternalMessages[0]
	if len(msg.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(msg.Items))
	}

	reasoning := msg.Items[0].Reasoning
	if reasoning == nil {
		t.Fatal("expected reasoning to be set")
	}

	// AppendProgress appends summary arrays rather than concatenating text
	if len(reasoning.Summary) != 2 {
		t.Fatalf("expected 2 summary items, got %d", len(reasoning.Summary))
	}

	if reasoning.Summary[0].Text != "First " {
		t.Errorf("expected first summary 'First ', got '%s'", reasoning.Summary[0].Text)
	}

	if reasoning.Summary[1].Text != "Second" {
		t.Errorf("expected second summary 'Second', got '%s'", reasoning.Summary[1].Text)
	}
}

func TestCaptureProgress_NilContentInitialization(t *testing.T) {
	acc := newProgressAccumulator("token")

	// Create item with nil content
	prog1 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			ToolCall: &types.ToolCall{
				CallID: "call-1",
				Name:   "tool",
			},
		},
	}
	acc.captureProgress(context.Background(), prog1)

	// Now send content - should initialize nil content
	prog2 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-1",
			Partial: true,
			Content: &mcp.Content{
				Type: "text",
				Text: "Text",
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	msg := acc.response.InternalMessages[0]
	if msg.Items[0].Content == nil {
		t.Fatal("expected content to be initialized")
	}

	if msg.Items[0].Content.Text != "Text" {
		t.Errorf("expected text 'Text', got '%s'", msg.Items[0].Content.Text)
	}
}

func TestGetPartialResponse_NoMessages(t *testing.T) {
	acc := newProgressAccumulator("token")

	result := acc.getPartialResponse(context.Background(), errors.New("test error"))

	if result != nil {
		t.Error("expected nil response when no messages accumulated")
	}
}

func TestGetPartialResponse_FiltersIncompleteToolCalls(t *testing.T) {
	acc := newProgressAccumulator("token")

	// Add a message with mixed content
	prog := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-1",
			Content: &mcp.Content{
				Type: "text",
				Text: "Some text",
			},
		},
	}
	acc.captureProgress(context.Background(), prog)

	// Add a partial tool call (should be filtered)
	prog2 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-2",
			Partial: true,
			ToolCall: &types.ToolCall{
				CallID: "call-1",
				Name:   "tool",
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	// Add a tool call with HasMore (should be filtered)
	prog3 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-3",
			Partial: true,
			ToolCall: &types.ToolCall{
				CallID: "call-2",
				Name:   "tool2",
			},
		},
	}
	acc.captureProgress(context.Background(), prog3)

	// Add a complete tool call (should be kept)
	prog4 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID:      "item-4",
			Partial: false,
			ToolCall: &types.ToolCall{
				CallID:    "call-3",
				Name:      "tool3",
				Arguments: "{}",
			},
		},
	}
	acc.captureProgress(context.Background(), prog4)

	result := acc.getPartialResponse(context.Background(), errors.New("test error"))

	if result == nil {
		t.Fatal("expected non-nil response")
	}

	// Should have text content + complete tool call + error message = 3 items
	if len(result.Output.Items) != 3 {
		t.Fatalf("expected 3 items (text, complete tool, error), got %d", len(result.Output.Items))
	}

	// First item should be text
	if result.Output.Items[0].Content == nil || result.Output.Items[0].Content.Text != "Some text" {
		t.Error("expected first item to be text content")
	}

	// Second item should be complete tool call
	if result.Output.Items[1].ToolCall == nil || result.Output.Items[1].ToolCall.CallID != "call-3" {
		t.Error("expected second item to be complete tool call")
	}

	// Last item should be error message
	lastItem := result.Output.Items[len(result.Output.Items)-1]
	if lastItem.Content == nil || lastItem.Content.Text != "\n\n[Error: test error]" {
		t.Errorf("expected error message, got '%s'", lastItem.Content.Text)
	}
}

func TestGetPartialResponse_KeepsReasoningAndContent(t *testing.T) {
	acc := newProgressAccumulator("token")

	// Add reasoning
	prog1 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-1",
			Reasoning: &types.Reasoning{
				Summary: []types.SummaryText{
					{Text: "Thinking..."},
				},
			},
		},
	}
	acc.captureProgress(context.Background(), prog1)

	// Add content
	prog2 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-2",
			Content: &mcp.Content{
				Type: "text",
				Text: "Response text",
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	result := acc.getPartialResponse(context.Background(), errors.New("test error"))

	if result == nil {
		t.Fatal("expected non-nil response")
	}

	// Should have reasoning + content + error = 3 items
	if len(result.Output.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result.Output.Items))
	}

	// Check reasoning
	if result.Output.Items[0].Reasoning == nil {
		t.Error("expected reasoning to be preserved")
	}

	// Check content
	if result.Output.Items[1].Content == nil || result.Output.Items[1].Content.Text != "Response text" {
		t.Error("expected content to be preserved")
	}

	// Verify Partial flag is cleared
	if result.Output.Items[0].Partial {
		t.Error("expected Partial to be cleared for reasoning")
	}
	if result.Output.Items[1].Partial {
		t.Error("expected Partial to be cleared for content")
	}
}

func TestGetPartialResponse_SetsErrorField(t *testing.T) {
	acc := newProgressAccumulator("token")

	prog := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-1",
			Content: &mcp.Content{
				Type: "text",
				Text: "Text",
			},
		},
	}
	acc.captureProgress(context.Background(), prog)

	testErr := errors.New("connection failed")
	result := acc.getPartialResponse(context.Background(), testErr)

	if result == nil {
		t.Fatal("expected non-nil response")
	}

	if result.Error != "connection failed" {
		t.Errorf("expected Error field to be 'connection failed', got '%s'", result.Error)
	}
}

func TestGetPartialResponse_MovesLastMessageToOutput(t *testing.T) {
	acc := newProgressAccumulator("token")

	// Create two messages
	prog1 := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "user",
		Item: types.CompletionItem{
			ID: "item-1",
			Content: &mcp.Content{
				Type: "text",
				Text: "First message",
			},
		},
	}
	acc.captureProgress(context.Background(), prog1)

	prog2 := &types.CompletionProgress{
		MessageID: "msg-2",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-2",
			Content: &mcp.Content{
				Type: "text",
				Text: "Second message",
			},
		},
	}
	acc.captureProgress(context.Background(), prog2)

	result := acc.getPartialResponse(context.Background(), errors.New("test error"))

	if result == nil {
		t.Fatal("expected non-nil response")
	}

	// Last message should be in Output
	if result.Output.ID != "msg-2" {
		t.Errorf("expected output message ID 'msg-2', got '%s'", result.Output.ID)
	}

	if result.Output.Role != "assistant" {
		t.Errorf("expected output role 'assistant', got '%s'", result.Output.Role)
	}

	if result.Output.HasMore {
		t.Error("expected HasMore to be false in output")
	}

	// Only first message should remain in InternalMessages
	if len(result.InternalMessages) != 1 {
		t.Fatalf("expected 1 internal message, got %d", len(result.InternalMessages))
	}

	if result.InternalMessages[0].ID != "msg-1" {
		t.Errorf("expected internal message ID 'msg-1', got '%s'", result.InternalMessages[0].ID)
	}
}

func TestGetPartialResponse_SingleMessageMovesToOutput(t *testing.T) {
	acc := newProgressAccumulator("token")

	prog := &types.CompletionProgress{
		MessageID: "msg-1",
		Role:      "assistant",
		Item: types.CompletionItem{
			ID: "item-1",
			Content: &mcp.Content{
				Type: "text",
				Text: "Only message",
			},
		},
	}
	acc.captureProgress(context.Background(), prog)

	result := acc.getPartialResponse(context.Background(), errors.New("test error"))

	if result == nil {
		t.Fatal("expected non-nil response")
	}

	// Single message should be in Output
	if result.Output.ID != "msg-1" {
		t.Errorf("expected output message ID 'msg-1', got '%s'", result.Output.ID)
	}

	// InternalMessages should be empty
	if result.InternalMessages != nil && len(result.InternalMessages) != 0 {
		t.Errorf("expected empty InternalMessages, got %d", len(result.InternalMessages))
	}
}
