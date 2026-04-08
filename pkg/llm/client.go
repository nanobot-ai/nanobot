package llm

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/envvar"
	"github.com/nanobot-ai/nanobot/pkg/llm/anthropic"
	"github.com/nanobot-ai/nanobot/pkg/llm/completions"
	"github.com/nanobot-ai/nanobot/pkg/llm/progress"
	"github.com/nanobot-ai/nanobot/pkg/llm/responses"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

var _ types.Completer = (*Client)(nil)

// LLMProviderConfig holds the configuration for a named LLM provider.
type LLMProviderConfig struct {
	Dialect types.Dialect
	APIKey  string // supports ${VAR} syntax
	BaseURL string // supports ${VAR} syntax
	Headers map[string]string
}

type Config struct {
	DefaultModel, DefaultMiniModel string
	LLMProviders                   map[string]LLMProviderConfig
}

func NewClient(cfg Config) *Client {
	return &Client{
		defaultModel:     cfg.DefaultModel,
		defaultMiniModel: cfg.DefaultMiniModel,
		cfg:              cfg,
	}
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
	case types.DialectOpenAIResponses, types.DialectOpenResponses:
		// DialectOpenAIResponses and DialectOpenResponses are intentionally distinct specs that currently
		// share the same client implementation but may diverge
		fallthrough
	default:
		return responses.NewClient(responses.Config{
			APIKey:  providerCfg.APIKey,
			BaseURL: providerCfg.BaseURL,
			Headers: providerCfg.Headers,
		}).Complete(ctx, req, opts...)
	}
}

func (c Client) dynamicConfig(ctx context.Context) Config {
	cfg := Config{
		DefaultModel:     c.defaultModel,
		DefaultMiniModel: c.defaultMiniModel,
		LLMProviders:     map[string]LLMProviderConfig{},
	}

	// Start with built-in/static provider refs (env var names)
	for name, p := range c.cfg.LLMProviders {
		cfg.LLMProviders[name] = LLMProviderConfig{
			Dialect: p.Dialect,
			APIKey:  p.APIKey,
			BaseURL: p.BaseURL,
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
		cfg.LLMProviders[name] = LLMProviderConfig{
			Dialect: p.Dialect,
			APIKey:  p.APIKey,
			BaseURL: p.BaseURL,
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
		cfg.LLMProviders[name] = LLMProviderConfig{
			Dialect: p.Dialect,
			APIKey:  envvar.ReplaceString(env, p.APIKey),
			BaseURL: envvar.ReplaceString(env, p.BaseURL),
			Headers: envvar.ReplaceMap(env, p.Headers),
		}
	}

	return cfg
}
