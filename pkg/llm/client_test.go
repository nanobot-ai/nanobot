package llm

import (
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestResolveProvider(t *testing.T) {
	cfg := Config{
		DefaultModel:     "openai/gpt-4.1",
		DefaultMiniModel: "anthropic/claude-haiku-4-5",
		Providers: map[string]ProviderConfig{
			"openai":    {Dialect: types.DialectOpenResponses},
			"anthropic": {Dialect: types.DialectAnthropicMessages},
			"azure":     {Dialect: types.DialectOpenResponses},
		},
	}

	tests := []struct {
		name         string
		model        string
		wantModel    string
		wantProvider string
	}{
		// Alias expansion
		{"default alias", "default", "gpt-4.1", "openai"},
		{"empty alias", "", "gpt-4.1", "openai"},
		{"mini alias", "mini", "claude-haiku-4-5", "anthropic"},

		// Explicit provider prefix
		{"openai prefix", "openai/gpt-4o", "gpt-4o", "openai"},
		{"anthropic prefix", "anthropic/claude-3-7-sonnet-latest", "claude-3-7-sonnet-latest", "anthropic"},
		{"azure prefix", "azure/gpt-4o", "gpt-4o", "azure"},
		{"unknown provider prefix", "vertex/gemini-pro", "gemini-pro", "vertex"},

		// Default fallbacks (no prefix)
		{"claude", "claude-haiku-4-5", "claude-haiku-4-5", "anthropic"},
		{"claude prefix", "claude-3-7-sonnet-latest", "claude-3-7-sonnet-latest", "anthropic"},
		{"openai", "gpt-4.1", "gpt-4.1", "openai"},
		{"unknown model", "gemini-pro", "gemini-pro", "openai"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModel, gotProvider := resolveProvider(tt.model, cfg)
			if gotModel != tt.wantModel {
				t.Errorf("model: got %q, want %q", gotModel, tt.wantModel)
			}
			if gotProvider != tt.wantProvider {
				t.Errorf("provider: got %q, want %q", gotProvider, tt.wantProvider)
			}
		})
	}
}
