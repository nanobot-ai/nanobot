package agents

import (
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestEstimateTokens_BasicMessages(t *testing.T) {
	messages := []types.Message{
		{
			Role: "user",
			Items: []types.CompletionItem{
				{Content: &mcp.Content{Type: "text", Text: "Hello, how are you?"}},
			},
		},
		{
			Role: "assistant",
			Items: []types.CompletionItem{
				{Content: &mcp.Content{Type: "text", Text: "I'm doing well, thank you!"}},
			},
		},
	}

	tokens := estimateTokens(messages, "", nil)
	if tokens <= 0 {
		t.Errorf("expected positive token count, got %d", tokens)
	}
	// Two short messages should be a small number of tokens
	if tokens > 100 {
		t.Errorf("expected < 100 tokens for short messages, got %d", tokens)
	}
}

func TestEstimateTokens_WithSystemPrompt(t *testing.T) {
	tokensWithout := estimateTokens(nil, "", nil)
	tokensWith := estimateTokens(nil, "You are a helpful assistant.", nil)
	if tokensWith <= tokensWithout {
		t.Errorf("expected more tokens with system prompt: without=%d, with=%d", tokensWithout, tokensWith)
	}
}

func TestEstimateTokens_WithTools(t *testing.T) {
	tools := []types.ToolUseDefinition{
		{
			Name:        "search",
			Description: "Search the web for information",
			Parameters:  []byte(`{"type":"object","properties":{"query":{"type":"string"}}}`),
		},
	}

	tokensWithout := estimateTokens(nil, "", nil)
	tokensWith := estimateTokens(nil, "", tools)
	if tokensWith <= tokensWithout {
		t.Errorf("expected more tokens with tools: without=%d, with=%d", tokensWithout, tokensWith)
	}
}

func TestEstimateTokens_LargeInput(t *testing.T) {
	longText := strings.Repeat("This is a test sentence. ", 1000)
	messages := []types.Message{
		{
			Role: "user",
			Items: []types.CompletionItem{
				{Content: &mcp.Content{Type: "text", Text: longText}},
			},
		},
	}

	tokens := estimateTokens(messages, "", nil)
	// A large input should produce a significant number of tokens
	if tokens < 1000 {
		t.Errorf("expected > 1000 tokens for large input, got %d", tokens)
	}
}

func TestEstimateTokens_WithToolCalls(t *testing.T) {
	messages := []types.Message{
		{
			Role: "assistant",
			Items: []types.CompletionItem{
				{
					ToolCall: &types.ToolCall{
						Name:      "search",
						Arguments: `{"query": "test"}`,
						CallID:    "call-1",
					},
				},
			},
		},
	}

	tokens := estimateTokens(messages, "", nil)
	if tokens <= 0 {
		t.Errorf("expected positive token count for tool calls, got %d", tokens)
	}
}

func TestEstimateTokens_WithToolResults(t *testing.T) {
	messages := []types.Message{
		{
			Role: "user",
			Items: []types.CompletionItem{
				{
					ToolCallResult: &types.ToolCallResult{
						CallID: "call-1",
						Output: types.CallResult{
							Content: []mcp.Content{
								{Type: "text", Text: "Search results here"},
							},
						},
					},
				},
			},
		},
	}

	tokens := estimateTokens(messages, "", nil)
	if tokens <= 0 {
		t.Errorf("expected positive token count for tool results, got %d", tokens)
	}
}

func TestCountTokens_EmptyString(t *testing.T) {
	tokens := countTokens("")
	if tokens != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", tokens)
	}
}

func TestCountTokens_KnownString(t *testing.T) {
	// "hello world" should produce a small number of tokens
	tokens := countTokens("hello world")
	if tokens < 1 || tokens > 5 {
		t.Errorf("expected 1-5 tokens for 'hello world', got %d", tokens)
	}
}
