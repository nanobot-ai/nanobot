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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	// Create span for LLM completion
	tracer := otel.Tracer("llm.client")
	ctx, span := tracer.Start(ctx, "llm.complete",
		trace.WithAttributes(
			attribute.String("llm.model", req.Model),
			attribute.Int("llm.input_messages", len(req.Input)),
			attribute.Int("llm.tools_count", len(req.Tools)),
		),
	)
	defer span.End()

	// Record additional request parameters
	if req.MaxTokens > 0 {
		span.SetAttributes(attribute.Int("llm.max_tokens", req.MaxTokens))
	}
	if req.Temperature != nil {
		if temp, err := req.Temperature.Float64(); err == nil {
			span.SetAttributes(attribute.Float64("llm.temperature", temp))
		}
	}
	if req.Agent != "" {
		span.SetAttributes(attribute.String("llm.agent", req.Agent))
	}

	defer func() {
		if ret != nil && ret.Agent == "" {
			ret.Agent = req.Agent
		}
	}()
	if req.Model == "default" || req.Model == "" {
		req.Model = c.defaultModel
		span.SetAttributes(attribute.String("llm.model", req.Model))
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

	// Determine provider and record in span
	var provider string
	if strings.HasPrefix(req.Model, "claude") {
		provider = "anthropic"
		response, err = c.anthropic.Complete(ctx, req, opts...)
	} else if c.useCompletions {
		provider = "openai-completions"
		response, err = c.completions.Complete(ctx, req, opts...)
	} else {
		provider = "openai-responses"
		response, err = c.responses.Complete(ctx, req, opts...)
	}
	span.SetAttributes(attribute.String("llm.provider", provider))

	// If there's an error and we have accumulated partial progress, return it
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "completion failed")
		if accumulator != nil {
			if partialResponse := accumulator.getPartialResponse(ctx, err); partialResponse != nil {
				span.SetStatus(codes.Ok, "returned partial response")
				// Record partial response metrics
				if len(partialResponse.InternalMessages) > 0 {
					span.SetAttributes(attribute.Int("llm.output_messages", len(partialResponse.InternalMessages)))
				}
				return partialResponse, nil // Return partial response, no error
			}
		}
		return response, err
	}

	// Record successful completion metrics
	if response != nil {
		if len(response.InternalMessages) > 0 {
			span.SetAttributes(attribute.Int("llm.output_messages", len(response.InternalMessages)))
		}
	}

	span.SetStatus(codes.Ok, "completion succeeded")
	return response, err
}
