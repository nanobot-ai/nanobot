package llm

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/envvar"
	"github.com/nanobot-ai/nanobot/pkg/llm/anthropic"
	"github.com/nanobot-ai/nanobot/pkg/llm/progress"
	"github.com/nanobot-ai/nanobot/pkg/llm/responses"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

var _ types.Completer = (*Client)(nil)

// ProviderConfig holds the configuration for a named LLM provider.
// APIKey and BaseURL are environment variable names whose values are resolved
// at request time from the session environment.
type ProviderConfig struct {
	Dialect types.Dialect
	APIKey  string // env var name
	BaseURL string // env var name
	Headers map[string]string
}

type Config struct {
	DefaultModel, DefaultMiniModel string
	Providers                      map[string]ProviderConfig
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

	providerCfg := dynamic.Providers[provider]
	switch providerCfg.Dialect {
	case types.DialectAnthropicMessages:
		return anthropic.NewClient(anthropic.Config{
			APIKey:  providerCfg.APIKey,
			BaseURL: providerCfg.BaseURL,
			Headers: providerCfg.Headers,
		}).Complete(ctx, req, opts...)
	default: // DialectOpenResponses or ""
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
		Providers:        map[string]ProviderConfig{},
	}

	// Start with built-in/static provider refs (env var names)
	for name, p := range c.cfg.Providers {
		cfg.Providers[name] = ProviderConfig{
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
	for name, p := range typesConfig.Providers {
		cfg.Providers[name] = ProviderConfig{
			Dialect: p.Dialect,
			APIKey:  p.APIKey,
			BaseURL: p.BaseURL,
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
	for name, p := range cfg.Providers {
		cfg.Providers[name] = ProviderConfig{
			Dialect: p.Dialect,
			APIKey:  envvar.ReplaceString(env, p.APIKey),
			BaseURL: envvar.ReplaceString(env, p.BaseURL),
			Headers: envvar.ReplaceMap(env, p.Headers),
		}
	}

	return cfg
}

func parseHeaderEnv(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	result := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &result); err == nil {
		return result
	}

	for part := range strings.SplitSeq(raw, ",") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k != "" {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
