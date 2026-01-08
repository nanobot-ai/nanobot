package agents

import (
	"context"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

// mockCompleter is a mock implementation of types.Completer for testing
type mockCompleter struct{}

func (m *mockCompleter) Complete(ctx context.Context, req types.CompletionRequest, opts ...types.CompletionOptions) (*types.CompletionResponse, error) {
	return &types.CompletionResponse{}, nil
}

func TestToolCalls_ToolNotFound(t *testing.T) {
	agents := New(&mockCompleter{}, nil)

	run := &types.Execution{
		Response: &types.CompletionResponse{
			Output: types.Message{
				Items: []types.CompletionItem{
					{
						ToolCall: &types.ToolCall{
							CallID: "call_123",
							Name:   "unknown_tool",
						},
					},
				},
			},
		},
		ToolToMCPServer: map[string]types.TargetMapping[types.TargetTool]{},
	}

	err := agents.toolCalls(context.Background(), types.Config{}, run, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check that error was stored as tool output
	toolOutput, exists := run.ToolOutputs["call_123"]
	if !exists {
		t.Fatal("expected tool output to be stored")
	}

	if !toolOutput.Done {
		t.Error("expected tool output to be marked as Done")
	}

	// Verify error content
	if len(toolOutput.Output.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(toolOutput.Output.Items))
	}

	result := toolOutput.Output.Items[0].ToolCallResult
	if result == nil {
		t.Fatal("expected ToolCallResult to be set")
	}

	if !result.Output.IsError {
		t.Error("expected IsError to be true")
	}

	if len(result.Output.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Output.Content))
	}

	errorText := result.Output.Content[0].Text
	if errorText != "Error: tool unknown_tool not found" {
		t.Errorf("unexpected error text: %s", errorText)
	}
}

func TestToolCalls_MultipleToolsOneNotFound(t *testing.T) {
	agents := New(&mockCompleter{}, nil)

	run := &types.Execution{
		Response: &types.CompletionResponse{
			Output: types.Message{
				Items: []types.CompletionItem{
					{
						ToolCall: &types.ToolCall{
							CallID: "call_1",
							Name:   "unknown_tool1",
						},
					},
					{
						ToolCall: &types.ToolCall{
							CallID: "call_2",
							Name:   "unknown_tool2",
						},
					},
				},
			},
		},
		ToolToMCPServer: map[string]types.TargetMapping[types.TargetTool]{},
	}

	err := agents.toolCalls(context.Background(), types.Config{}, run, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Both tools should have error outputs
	if len(run.ToolOutputs) != 2 {
		t.Fatalf("expected 2 tool outputs, got %d", len(run.ToolOutputs))
	}

	// Verify both are error outputs
	for callID, output := range run.ToolOutputs {
		if !output.Done {
			t.Errorf("expected %s to be Done", callID)
		}
		if !output.Output.Items[0].ToolCallResult.Output.IsError {
			t.Errorf("expected %s to be an error", callID)
		}
	}
}

func TestToolCalls_SkipsAlreadyDone(t *testing.T) {
	agents := New(&mockCompleter{}, nil)

	run := &types.Execution{
		Response: &types.CompletionResponse{
			Output: types.Message{
				Items: []types.CompletionItem{
					{
						ToolCall: &types.ToolCall{
							CallID: "call_done",
							Name:   "test_tool",
						},
					},
				},
			},
		},
		ToolToMCPServer: map[string]types.TargetMapping[types.TargetTool]{
			"test_tool": {
				MCPServer:  "test_server",
				TargetName: "test_tool",
			},
		},
		ToolOutputs: map[string]types.ToolOutput{
			"call_done": {
				Done: true,
				Output: types.Message{
					Role: "user",
					Items: []types.CompletionItem{
						{
							ToolCallResult: &types.ToolCallResult{
								CallID: "call_done",
								Output: types.CallResult{
									Content: []mcp.Content{{Type: "text", Text: "already done"}},
								},
							},
						},
					},
				},
			},
		},
	}

	err := agents.toolCalls(context.Background(), types.Config{}, run, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Output should remain unchanged (only 1 output, the original)
	if len(run.ToolOutputs) != 1 {
		t.Errorf("expected 1 tool output, got %d", len(run.ToolOutputs))
	}

	output := run.ToolOutputs["call_done"]
	if output.Output.Items[0].ToolCallResult.Output.Content[0].Text != "already done" {
		t.Error("expected output to remain unchanged")
	}
}

func TestToolCalls_ExternalToolTerminatesRun(t *testing.T) {
	agents := New(&mockCompleter{}, nil)

	run := &types.Execution{
		Response: &types.CompletionResponse{
			Output: types.Message{
				Items: []types.CompletionItem{
					{
						ToolCall: &types.ToolCall{
							CallID: "call_external",
							Name:   "external_tool",
						},
					},
				},
			},
		},
		ToolToMCPServer: map[string]types.TargetMapping[types.TargetTool]{
			"external_tool": {
				MCPServer:  "external_server",
				TargetName: "external_tool",
				Target: types.TargetTool{
					External: true,
				},
			},
		},
	}

	err := agents.toolCalls(context.Background(), types.Config{}, run, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !run.Done {
		t.Error("expected run to be marked as Done for external tool")
	}

	// External tools should not have outputs created
	if len(run.ToolOutputs) != 0 {
		t.Error("expected no tool outputs for external tool")
	}
}

func TestToolCalls_NoToolCallsMarksDone(t *testing.T) {
	agents := New(&mockCompleter{}, nil)

	run := &types.Execution{
		Response: &types.CompletionResponse{
			Output: types.Message{
				Items: []types.CompletionItem{
					{
						Content: &mcp.Content{Type: "text", Text: "just text, no tool calls"},
					},
				},
			},
		},
		ToolToMCPServer: map[string]types.TargetMapping[types.TargetTool]{},
	}

	err := agents.toolCalls(context.Background(), types.Config{}, run, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !run.Done {
		t.Error("expected run to be marked as Done when there are no tool outputs")
	}
}

func TestCreateErrorToolOutput(t *testing.T) {
	callID := "test_call"
	errorMsg := "test error message"

	output := createErrorToolOutput(callID, errorMsg)

	if !output.Done {
		t.Error("expected Done to be true")
	}

	if output.Output.Role != "user" {
		t.Errorf("expected Role to be 'user', got '%s'", output.Output.Role)
	}

	if len(output.Output.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(output.Output.Items))
	}

	result := output.Output.Items[0].ToolCallResult
	if result == nil {
		t.Fatal("expected ToolCallResult to be set")
	}

	if result.CallID != callID {
		t.Errorf("expected CallID to be '%s', got '%s'", callID, result.CallID)
	}

	if !result.Output.IsError {
		t.Error("expected IsError to be true")
	}

	if len(result.Output.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Output.Content))
	}

	content := result.Output.Content[0]
	if content.Type != "text" {
		t.Errorf("expected content type 'text', got '%s'", content.Type)
	}

	if content.Text != errorMsg {
		t.Errorf("expected error message '%s', got '%s'", errorMsg, content.Text)
	}
}
