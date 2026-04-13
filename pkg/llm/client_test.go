package llm

import (
	"context"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestResolveProvider(t *testing.T) {
	cfg := Config{
		DefaultModel:     "openai/gpt-4.1",
		DefaultMiniModel: "anthropic/claude-haiku-4-5",
		LLMProviders: map[string]LLMProviderConfig{
			"openai":    {Dialect: types.DialectOpenAIResponses},
			"anthropic": {Dialect: types.DialectAnthropicMessages},
			"azure":     {Dialect: types.DialectOpenAIResponses},
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

func TestResolveProviderMalformedModel(t *testing.T) {
	cfg := Config{
		LLMProviders: map[string]LLMProviderConfig{
			"openai": {Dialect: types.DialectOpenAIResponses},
		},
	}

	tests := []struct {
		name         string
		model        string
		wantModel    string
		wantProvider string
	}{
		{"extra slash", "openai/model/extra", "model/extra", "openai"},
		{"single slash", "openai/gpt-4o", "gpt-4o", "openai"},
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

func TestCompleteUnknownProvider(t *testing.T) {
	session := mcp.NewEmptySession(context.Background())

	client := NewClient(Config{
		DefaultModel: "vertex/gemini-pro",
		LLMProviders: map[string]LLMProviderConfig{
			"openai": {Dialect: types.DialectOpenAIResponses},
		},
	})

	ctx := session.Context()
	_, err := client.Complete(ctx, types.CompletionRequest{Model: "vertex/gemini-pro"})
	wantErr := `unknown LLM provider "vertex": not defined in llmProviders config`
	if err == nil || err.Error() != wantErr {
		t.Errorf("got error %v, want %q", err, wantErr)
	}
}

func TestDynamicConfigProviderResolution(t *testing.T) {
	session := mcp.NewEmptySession(context.Background())
	session.SetEnv(map[string]string{
		"MY_API_KEY":    "sk-test-12345",
		"MY_BASE_URL":   "https://api.example.com/v1",
		"MY_AUTH_TOKEN": "bearer-token",
	})

	client := NewClient(Config{
		LLMProviders: map[string]LLMProviderConfig{
			"literal": {
				Dialect: types.DialectOpenAIResponses,
				APIKey:  "sk-literal-key",
				BaseURL: "https://literal.example.com/v1",
				Headers: map[string]string{"X-Custom": "value"},
			},
			"from-env": {
				Dialect: types.DialectAnthropicMessages,
				APIKey:  "${MY_API_KEY}",
				BaseURL: "${MY_BASE_URL}",
				Headers: map[string]string{"Authorization": "Bearer ${MY_AUTH_TOKEN}"},
			},
		},
	})

	dynamic := client.dynamicConfig(session.Context())

	literal := dynamic.LLMProviders["literal"]
	if literal.APIKey != "sk-literal-key" {
		t.Errorf("literal APIKey: got %q, want %q", literal.APIKey, "sk-literal-key")
	}
	if literal.BaseURL != "https://literal.example.com/v1" {
		t.Errorf("literal BaseURL: got %q, want %q", literal.BaseURL, "https://literal.example.com/v1")
	}
	if literal.Headers["X-Custom"] != "value" {
		t.Errorf("literal Headers[X-Custom]: got %q, want %q", literal.Headers["X-Custom"], "value")
	}

	fromEnv := dynamic.LLMProviders["from-env"]
	if fromEnv.APIKey != "sk-test-12345" {
		t.Errorf("from-env APIKey: got %q, want %q", fromEnv.APIKey, "sk-test-12345")
	}
	if fromEnv.BaseURL != "https://api.example.com/v1" {
		t.Errorf("from-env BaseURL: got %q, want %q", fromEnv.BaseURL, "https://api.example.com/v1")
	}
	if fromEnv.Headers["Authorization"] != "Bearer bearer-token" {
		t.Errorf("from-env Headers[Authorization]: got %q, want %q", fromEnv.Headers["Authorization"], "Bearer bearer-token")
	}
}
