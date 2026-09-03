package llm

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/obot-platform/nanobot/pkg/complete"
	"github.com/obot-platform/nanobot/pkg/envvar"
	"github.com/obot-platform/nanobot/pkg/llm/anthropic"
	"github.com/obot-platform/nanobot/pkg/llm/completions"
	"github.com/obot-platform/nanobot/pkg/llm/progress"
	"github.com/obot-platform/nanobot/pkg/llm/responses"
	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
	"github.com/obot-platform/nanobot/pkg/uuid"
)

var _ types.Completer = (*Client)(nil)

// LLMProviderConfig holds the configuration for a named LLM provider.
type LLMProviderConfig struct {
	Dialect types.Dialect
	APIKey  string // supports ${VAR} syntax
	BaseURL string // supports ${VAR} syntax
	Region  string // supports ${VAR} syntax
	Headers map[string]string
}

type ProviderEndpoint struct {
	Region           string
	OpenAIBaseURL    string
	AnthropicBaseURL string
	DocsRoot         string
}

// ModelPricing contains USD prices per million tokens.
type ModelPricing struct {
	Input      float64
	Output     float64
	CacheRead  float64
	CacheWrite *float64
}

type ModelConfig struct {
	ContextWindow   int
	Pricing         ModelPricing
	InputModalities []string
	Thinking        []string
}

type Config struct {
	DefaultModel, DefaultMiniModel string
	LLMProviders                   map[string]LLMProviderConfig
	ProviderEndpoints              map[string][]ProviderEndpoint
	Models                         map[string]map[string]ModelConfig
}

func NewClient(cfg Config) *Client {
	return &Client{
		defaultModel:     cfg.DefaultModel,
		defaultMiniModel: cfg.DefaultMiniModel,
		cfg:              cfg,
	}
}

func cloneProviderEndpoints(endpoints map[string][]ProviderEndpoint) map[string][]ProviderEndpoint {
	result := make(map[string][]ProviderEndpoint, len(endpoints))
	for provider, providerEndpoints := range endpoints {
		result[provider] = slices.Clone(providerEndpoints)
	}
	return result
}

func cloneModels(models map[string]map[string]ModelConfig) map[string]map[string]ModelConfig {
	result := make(map[string]map[string]ModelConfig, len(models))
	for provider, providerModels := range models {
		result[provider] = make(map[string]ModelConfig, len(providerModels))
		for model, modelConfig := range providerModels {
			modelConfig.InputModalities = slices.Clone(modelConfig.InputModalities)
			modelConfig.Thinking = slices.Clone(modelConfig.Thinking)
			if modelConfig.Pricing.CacheWrite != nil {
				cacheWrite := *modelConfig.Pricing.CacheWrite
				modelConfig.Pricing.CacheWrite = &cacheWrite
			}
			result[provider][model] = modelConfig
		}
	}
	return result
}

type Client struct {
	defaultModel     string
	defaultMiniModel string
	cfg              Config
}

// resolveProvider resolves the model alias and provider name for a request.
// Model aliases ("default", "mini", "") are expanded to their configured values.
// The provider is parsed from the model string using the "{provider}/{model}" format.
// If no provider prefix is present, it falls back to a heuristic
// (claude prefix → anthropic, else openai).
func resolveProvider(model string, cfg Config) (string, string) {
	switch model {
	case "default", "":
		model = cfg.DefaultModel
	case "mini":
		model = cfg.DefaultMiniModel
	}
	if provider, m, ok := strings.Cut(model, "/"); ok {
		return m, provider
	}
	if strings.HasPrefix(model, "claude") {
		return model, "anthropic"
	}
	return model, "openai"
}

func (c Client) ModelConfig(model string) (ModelConfig, bool) {
	model, provider := resolveProvider(model, c.cfg)
	modelConfig, ok := c.cfg.Models[provider][model]
	if !ok {
		return ModelConfig{}, false
	}
	modelConfig.InputModalities = slices.Clone(modelConfig.InputModalities)
	modelConfig.Thinking = slices.Clone(modelConfig.Thinking)
	if modelConfig.Pricing.CacheWrite != nil {
		cacheWrite := *modelConfig.Pricing.CacheWrite
		modelConfig.Pricing.CacheWrite = &cacheWrite
	}
	return modelConfig, true
}

func (c Client) ContextWindow(model string) int {
	modelConfig, ok := c.ModelConfig(model)
	if !ok {
		return 0
	}
	return modelConfig.ContextWindow
}

func endpointBaseURL(endpoint ProviderEndpoint, dialect types.Dialect) string {
	if dialect == types.DialectAnthropicMessages {
		return endpoint.AnthropicBaseURL
	}
	return endpoint.OpenAIBaseURL
}

func defaultBaseURL(endpoints []ProviderEndpoint, dialect types.Dialect, region string) string {
	if region != "" {
		for _, endpoint := range endpoints {
			if endpoint.Region == region {
				return endpointBaseURL(endpoint, dialect)
			}
		}
	}
	for _, endpoint := range endpoints {
		if baseURL := endpointBaseURL(endpoint, dialect); baseURL != "" {
			return baseURL
		}
	}
	return ""
}

func isUnsetEnvReference(env map[string]string, value string) bool {
	if len(value) < 4 || !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return false
	}
	name := value[2 : len(value)-1]
	if name == "" || strings.ContainsAny(name, " \t\n{}") {
		return false
	}
	_, ok := env[name]
	return !ok
}

func (c Client) Complete(ctx context.Context, req types.CompletionRequest, opts ...types.CompletionOptions) (ret *types.CompletionResponse, err error) {
	defer func() {
		if errors.Is(err, context.Canceled) {
			if cancelErr, ok := errors.AsType[*mcp.RequestCancelledError](context.Cause(mcp.UserContext(ctx))); ok && cancelErr != nil {
				err = nil
				ret = &types.CompletionResponse{
					Output: types.Message{
						ID:   uuid.String(),
						Role: "assistant",
						Items: []types.CompletionItem{
							{
								Content: &mcp.Content{
									Type: "text",
									Text: strings.ToUpper(cancelErr.Error()),
								},
							},
						},
					},
				}

				if opt := complete.Complete(opts...); opt.ProgressToken != nil {
					progress.Send(ctx, &types.CompletionProgress{
						MessageID: ret.Output.ID,
						Role:      "assistant",
						Item:      ret.Output.Items[0],
					}, opt.ProgressToken)
				}
			}
		}
		if ret != nil && ret.Agent == "" {
			ret.Agent = req.Agent
		}
	}()

	dynamic := c.dynamicConfig(ctx)

	var provider string
	req.Model, provider = resolveProvider(req.Model, dynamic)

	opt := complete.Complete(opts...)
	if opt.ProgressToken != nil && len(req.Input) > 0 {
		lastMsg := req.Input[len(req.Input)-1]
		if lastMsg.ID != "" && lastMsg.Role == "user" {
			for _, item := range lastMsg.Items {
				progress.Send(ctx, &types.CompletionProgress{
					Model:     req.Model,
					MessageID: lastMsg.ID,
					Role:      lastMsg.Role,
					Item:      item,
				}, opt.ProgressToken)
			}
		}
	}

	providerCfg, ok := dynamic.LLMProviders[provider]
	if !ok {
		return nil, fmt.Errorf("unknown LLM provider %q: not defined in llmProviders config", provider)
	}
	switch providerCfg.Dialect {
	case types.DialectAnthropicMessages:
		return anthropic.NewClient(anthropic.Config{
			APIKey:  providerCfg.APIKey,
			BaseURL: providerCfg.BaseURL,
			Headers: providerCfg.Headers,
		}).Complete(ctx, req, opts...)
	case types.DialectOpenAIChatCompletions:
		return completions.NewClient(completions.Config{
			APIKey:  providerCfg.APIKey,
			BaseURL: providerCfg.BaseURL,
			Headers: providerCfg.Headers,
		}).Complete(ctx, req, opts...)
	case "", types.DialectOpenAIResponses, types.DialectOpenResponses:
		// DialectOpenAIResponses and DialectOpenResponses are intentionally distinct specs that currently
		// share the same client implementation but may diverge
		return responses.NewClient(responses.Config{
			APIKey:  providerCfg.APIKey,
			BaseURL: providerCfg.BaseURL,
			Headers: providerCfg.Headers,
		}).Complete(ctx, req, opts...)
	default:
		return nil, fmt.Errorf("unsupported LLM provider dialect %q", providerCfg.Dialect)
	}
}

func (c Client) dynamicConfig(ctx context.Context) Config {
	cfg := Config{
		DefaultModel:      c.defaultModel,
		DefaultMiniModel:  c.defaultMiniModel,
		LLMProviders:      map[string]LLMProviderConfig{},
		ProviderEndpoints: cloneProviderEndpoints(c.cfg.ProviderEndpoints),
		Models:            cloneModels(c.cfg.Models),
	}

	// Start with built-in/static provider refs (env var names)
	for name, p := range c.cfg.LLMProviders {
		cfg.LLMProviders[name] = LLMProviderConfig{
			Dialect: p.Dialect,
			APIKey:  p.APIKey,
			BaseURL: p.BaseURL,
			Region:  p.Region,
			Headers: maps.Clone(p.Headers),
		}
	}

	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return cfg
	}

	env := session.GetEnvMap()

	// Overlay providers defined in the YAML config for this session
	typesConfig := types.ConfigFromContext(ctx)
	for name, p := range typesConfig.LLMProviders {
		builtIn := cfg.LLMProviders[name]
		cfg.LLMProviders[name] = LLMProviderConfig{
			Dialect: p.Dialect,
			APIKey:  p.APIKey,
			BaseURL: p.BaseURL,
			Region:  builtIn.Region,
			Headers: maps.Clone(p.Headers),
		}
	}

	// Override shared settings from env
	if v := strings.TrimSpace(env["NANOBOT_DEFAULT_MODEL"]); v != "" {
		cfg.DefaultModel = v
	}
	if v := strings.TrimSpace(env["NANOBOT_DEFAULT_MINI_MODEL"]); v != "" {
		cfg.DefaultMiniModel = v
	}

	// Resolve ${VAR} references in provider config using the session env
	for name, p := range cfg.LLMProviders {
		region := p.Region
		if isUnsetEnvReference(env, region) {
			region = ""
		} else {
			region = envvar.ReplaceString(env, region)
		}
		baseURL := p.BaseURL
		if len(cfg.ProviderEndpoints[name]) == 0 || !isUnsetEnvReference(env, baseURL) {
			baseURL = envvar.ReplaceString(env, baseURL)
		} else {
			baseURL = ""
		}
		if baseURL == "" {
			baseURL = defaultBaseURL(cfg.ProviderEndpoints[name], p.Dialect, region)
		}
		cfg.LLMProviders[name] = LLMProviderConfig{
			Dialect: p.Dialect,
			APIKey:  envvar.ReplaceString(env, p.APIKey),
			BaseURL: baseURL,
			Region:  region,
			Headers: envvar.ReplaceMap(env, p.Headers),
		}
	}

	return cfg
}
