package modelcaps

import "testing"

func TestNormalizeVariants(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		context int
		reserve int
	}{
		{"plain claude", "claude-opus-4-6", 200_000, 5_000},
		{"slash prefixed", "anthropic/claude-opus-4-6", 200_000, 5_000},
		{"colon prefixed", "anthropic:claude-opus-4-6", 200_000, 5_000},
		{"gpt standard", "gpt-5.0", 272_000, 5_000},
		{"gpt prefixed", "openai/gpt-5.0", 272_000, 5_000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ContextWindow(tc.model); got != tc.context {
				t.Fatalf("context window mismatch: got %d want %d", got, tc.context)
			}
			if got := ReservedOutput(tc.model); got != tc.reserve {
				t.Fatalf("reserve mismatch: got %d want %d", got, tc.reserve)
			}
		})
	}
}
