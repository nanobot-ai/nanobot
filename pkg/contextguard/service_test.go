package contextguard

import (
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestEvaluateAnthropicLimits(t *testing.T) {
	svc := NewService(Config{})
	state := func(tokens int) State {
		return State{
			Model: "claude-opus-4-8",
			Messages: []types.Message{
				makeTextMessage(tokens),
			},
		}
	}

	tests := []struct {
		name   string
		state  State
		status Status
	}{
		{"ok", state(1000), StatusOK},
		{"needs", state(176000), StatusNeedsCompaction},
		{"over", state(195000), StatusOverLimit},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.Evaluate(tc.state)
			if result.Status != tc.status {
				t.Fatalf("expected status %s, got %s", tc.status, result.Status)
			}
		})
	}
}

func TestEvaluateGPT5Limits(t *testing.T) {
	svc := NewService(Config{})
	state := func(tokens int) State {
		return State{
			Model: "gpt-5.0",
			Messages: []types.Message{
				makeTextMessage(tokens),
			},
		}
	}

	tests := []struct {
		name   string
		state  State
		status Status
	}{
		{"ok", state(100000), StatusOK},
		{"needs", state(241000), StatusNeedsCompaction},
		{"over", state(267000), StatusOverLimit},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.Evaluate(tc.state)
			if result.Status != tc.status {
				t.Fatalf("expected status %s, got %s", tc.status, result.Status)
			}
		})
	}
}

func makeTextMessage(tokens int) types.Message {
	chars := strings.Repeat("a", tokens*4)
	return types.Message{
		Role: "user",
		Items: []types.CompletionItem{
			{
				Content: &mcp.Content{
					Type: "text",
					Text: chars,
				},
			},
		},
	}
}
