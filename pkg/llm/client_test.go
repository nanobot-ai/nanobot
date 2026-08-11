package llm

import (
	"context"
	"slices"
	"testing"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
)

func TestDynamicConfigUsesRegisteredProviderEndpoint(t *testing.T) {
	session := mcp.NewEmptySession(context.Background())
	session.SetEnv(map[string]string{"MINIMAX_API_KEY": "test-key"})
	client := NewClient(Config{
		LLMProviders: map[string]LLMProviderConfig{
			"minimax": {
				Dialect: types.DialectOpenAIChatCompletions,
				APIKey:  "${MINIMAX_API_KEY}",
				BaseURL: "${MINIMAX_BASE_URL}",
				Region:  "${MINIMAX_REGION}",
			},
		},
		ProviderEndpoints: map[string][]ProviderEndpoint{
			"minimax": {
				{Region: "global_en", OpenAIBaseURL: "https://api.minimax.io/v1"},
				{Region: "cn_zh", OpenAIBaseURL: "https://api.minimaxi.com/v1"},
			},
		},
	})

	dynamic := client.dynamicConfig(session.Context())
	if got := dynamic.LLMProviders["minimax"].BaseURL; got != "https://api.minimax.io/v1" {
		t.Fatalf("expected registered MiniMax endpoint, got %q", got)
	}
}

func TestDynamicConfigUsesRegisteredRegionalProviderEndpoint(t *testing.T) {
	session := mcp.NewEmptySession(context.Background())
	session.SetEnv(map[string]string{
		"MINIMAX_API_KEY": "test-key",
		"MINIMAX_REGION":  "cn_zh",
	})
	client := NewClient(Config{
		LLMProviders: map[string]LLMProviderConfig{
			"minimax": {
				Dialect: types.DialectOpenAIChatCompletions,
				APIKey:  "${MINIMAX_API_KEY}",
				BaseURL: "${MINIMAX_BASE_URL}",
				Region:  "${MINIMAX_REGION}",
			},
		},
		ProviderEndpoints: map[string][]ProviderEndpoint{
			"minimax": {
				{Region: "global_en", OpenAIBaseURL: "https://api.minimax.io/v1"},
				{Region: "cn_zh", OpenAIBaseURL: "https://api.minimaxi.com/v1"},
			},
		},
	})

	dynamic := client.dynamicConfig(session.Context())
	provider := dynamic.LLMProviders["minimax"]
	if provider.Region != "cn_zh" {
		t.Fatalf("expected MiniMax region cn_zh, got %q", provider.Region)
	}
	if provider.BaseURL != "https://api.minimaxi.com/v1" {
		t.Fatalf("expected MiniMax China endpoint, got %q", provider.BaseURL)
	}
}

func TestModelConfig(t *testing.T) {
	cacheWrite := 0.375
	client := NewClient(Config{
		Models: map[string]map[string]ModelConfig{
			"minimax": {
				"MiniMax-M2.7": {
					ContextWindow: 204_800,
					Pricing: ModelPricing{
						Input:      0.3,
						Output:     1.2,
						CacheRead:  0.06,
						CacheWrite: &cacheWrite,
					},
					InputModalities: []string{"text"},
					Thinking:        []string{"always_on"},
				},
			},
		},
	})

	model, ok := client.ModelConfig("minimax/MiniMax-M2.7")
	if !ok {
		t.Fatal("expected MiniMax model config")
	}
	if model.ContextWindow != 204_800 {
		t.Fatalf("expected context window 204800, got %d", model.ContextWindow)
	}
	if model.Pricing.Input != 0.3 || model.Pricing.Output != 1.2 || model.Pricing.CacheRead != 0.06 {
		t.Fatalf("unexpected MiniMax model pricing: %+v", model.Pricing)
	}
	if model.Pricing.CacheWrite == nil || *model.Pricing.CacheWrite != 0.375 {
		t.Fatalf("unexpected MiniMax cache write pricing: %v", model.Pricing.CacheWrite)
	}
	if !slices.Equal(model.InputModalities, []string{"text"}) {
		t.Fatalf("unexpected MiniMax input modalities: %v", model.InputModalities)
	}
	if !slices.Equal(model.Thinking, []string{"always_on"}) {
		t.Fatalf("unexpected MiniMax thinking modes: %v", model.Thinking)
	}
}

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

func TestCompleteUnknownDialect(t *testing.T) {
	session := mcp.NewEmptySession(context.Background())
	client := NewClient(Config{
		LLMProviders: map[string]LLMProviderConfig{
			"custom": {Dialect: "UnknownDialect"},
		},
	})

	_, err := client.Complete(session.Context(), types.CompletionRequest{Model: "custom/model"})
	wantErr := `unsupported LLM provider dialect "UnknownDialect"`
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
