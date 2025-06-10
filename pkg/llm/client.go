package llm

import (
	"context"

	"github.com/obot-platform/nanobot/pkg/llm/anthropic"
	"github.com/obot-platform/nanobot/pkg/llm/responses"
	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
)

var _ types.Completer = (*Client)(nil)

type Config struct {
	DefaultModel string
	Responses    responses.Config
	Anthropic    anthropic.Config
}

func NewClient(cfg Config, config types.Config) *Client {
	return &Client{
		defaultModel: cfg.DefaultModel,
		responses:    responses.NewClient(cfg.Responses, config),
		anthropic:    anthropic.NewClient(cfg.Anthropic, config),
	}
}

type Client struct {
	defaultModel string
	responses    *responses.Client
	anthropic    *anthropic.Client
}

func (c Client) Complete(ctx context.Context, req types.CompletionRequest, opts ...types.CompletionOptions) (*types.CompletionResponse, error) {
	if req.Model == "default" || req.Model == "" {
		req.Model = c.defaultModel
	}
	if len(req.Input) > 0 {
		if last := req.Input[len(req.Input)-1]; last.ToolCallResult != nil &&
			last.ToolCallResult.OutputRole == "assistant" &&
			len(last.ToolCallResult.Output.Content) > 0 {
			resp := &types.CompletionResponse{
				Model: req.Model,
			}
			for _, content := range last.ToolCallResult.Output.Content {
				resp.Output = append(resp.Output, types.CompletionOutput{
					Message: &mcp.SamplingMessage{
						Role:    "assistant",
						Content: content,
					},
				})
			}
			return resp, nil
		}
	}
	newInput := make([]types.CompletionInput, 0, len(req.Input))
	for _, input := range req.Input {
		if input.ToolCallResult != nil && input.ToolCallResult.OutputRole == "assistant" &&
			len(input.ToolCallResult.Output.Content) > 0 {
			// elide the tool call result if it is an assistant response
			input = types.CompletionInput{
				ToolCallResult: &types.ToolCallResult{
					CallID: input.ToolCallResult.CallID,
					Output: mcp.CallToolResult{
						Content: []mcp.Content{{Text: "complete"}},
					},
				},
			}
		}
		newInput = append(newInput, input)
	}
	req.Input = newInput
	return c.responses.Complete(ctx, req, opts...)
}
