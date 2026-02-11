package agents

import (
	"context"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestGuardAfterToolTriggersCompaction(t *testing.T) {
	ctx := context.Background()
	agents := &Agents{}
	config := types.Config{}

	run := &types.Execution{
		PopulatedRequest: &types.CompletionRequest{
			Model: "gpt-5.0",
			Input: []types.Message{makeMessageWithTokens(100000)},
		},
		Response: &types.CompletionResponse{
			Model:  "gpt-5.0",
			Output: makeMessageWithTokens(20000),
		},
		ToolOutputs: map[string]types.ToolOutput{
			"tool": {
				Output: makeMessageWithTokens(40000),
				Done:   true,
			},
		},
	}

	if !agents.guardAfterTool(ctx, config, run) {
		t.Fatalf("expected guard to request compaction")
	}

	runSmall := &types.Execution{
		PopulatedRequest: &types.CompletionRequest{
			Model: "gpt-5.0",
			Input: []types.Message{makeMessageWithTokens(1000)},
		},
		Response: &types.CompletionResponse{
			Model:  "gpt-5.0",
			Output: makeMessageWithTokens(1000),
		},
		ToolOutputs: map[string]types.ToolOutput{
			"tool": {
				Output: makeMessageWithTokens(1000),
				Done:   true,
			},
		},
	}

	if agents.guardAfterTool(ctx, config, runSmall) {
		t.Fatalf("expected guard to allow continuation for small outputs")
	}
}

func makeMessageWithTokens(tokens int) types.Message {
	text := strings.Repeat("a", tokens*4)
	return types.Message{
		Role: "user",
		Items: []types.CompletionItem{{
			Content: &mcp.Content{Type: "text", Text: text},
		}},
	}
}
