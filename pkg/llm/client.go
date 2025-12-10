package llm

import (
	"context"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/llm/anthropic"
	"github.com/nanobot-ai/nanobot/pkg/llm/completions"
	"github.com/nanobot-ai/nanobot/pkg/llm/progress"
	"github.com/nanobot-ai/nanobot/pkg/llm/responses"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

var _ types.Completer = (*Client)(nil)

type Config struct {
	DefaultModel string
	Responses    responses.Config
	Anthropic    anthropic.Config
}

func NewClient(cfg Config) *Client {
	return &Client{
		useCompletions: cfg.Responses.ChatCompletionAPI,
		defaultModel:   cfg.DefaultModel,
		completions: completions.NewClient(completions.Config{
			APIKey:  cfg.Responses.APIKey,
			BaseURL: cfg.Responses.BaseURL,
			Headers: cfg.Responses.Headers,
		}),
		responses: responses.NewClient(cfg.Responses),
		anthropic: anthropic.NewClient(cfg.Anthropic),
	}
}

type Client struct {
	defaultModel   string
	useCompletions bool
	completions    *completions.Client
	responses      *responses.Client
	anthropic      *anthropic.Client
}

func (c Client) Complete(ctx context.Context, req types.CompletionRequest, opts ...types.CompletionOptions) (ret *types.CompletionResponse, _ error) {
	defer func() {
		if ret != nil && ret.Agent == "" {
			ret.Agent = req.Agent
		}
	}()
	if req.Model == "default" || req.Model == "" {
		req.Model = c.defaultModel
	}

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

	if strings.HasPrefix(req.Model, "claude") {
		return c.anthropic.Complete(ctx, req, opts...)
	}
	if c.useCompletions {
		return c.completions.Complete(ctx, req, opts...)
	}
	return c.responses.Complete(ctx, req, opts...)
}
