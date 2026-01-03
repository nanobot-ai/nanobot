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

	// Create progress accumulator to capture partial responses on error
	var accumulator *progressAccumulator
	if opt.ProgressToken != nil {
		accumulator = newProgressAccumulator(opt.ProgressToken)
		ctx = progress.WithInterceptor(ctx, accumulator.captureProgress)
	}

	var response *types.CompletionResponse
	var err error

	if strings.HasPrefix(req.Model, "claude") {
		response, err = c.anthropic.Complete(ctx, req, opts...)
	} else if c.useCompletions {
		response, err = c.completions.Complete(ctx, req, opts...)
	} else {
		response, err = c.responses.Complete(ctx, req, opts...)
	}

	// If there's an error and we have accumulated partial progress, return it
	if err != nil && accumulator != nil {
		if partialResponse := accumulator.getPartialResponse(ctx, err); partialResponse != nil {
			return partialResponse, nil // Return partial response, no error
		}
	}

	return response, err
}
