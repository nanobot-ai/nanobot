package llm

import (
	"context"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/llm/anthropic"
	"github.com/nanobot-ai/nanobot/pkg/llm/ollama"
	"github.com/nanobot-ai/nanobot/pkg/llm/responses"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

var _ types.Completer = (*Client)(nil)

type Config struct {
	DefaultModel string
	Responses    responses.Config
	Anthropic    anthropic.Config
	Ollama       ollama.Config
}

func NewClient(cfg Config) *Client {
	return &Client{
		defaultModel: cfg.DefaultModel,
		responses:    responses.NewClient(cfg.Responses),
		anthropic:    anthropic.NewClient(cfg.Anthropic),
		ollama:       ollama.NewClient(cfg.Ollama),
	}
}

type Client struct {
	defaultModel string
	responses    *responses.Client
	anthropic    *anthropic.Client
	ollama       *ollama.Client
}

func (c *Client) handleAssistantRolesFromTools(req types.CompletionRequest) (_ types.CompletionRequest, resp *types.CompletionResponse) {
	if len(req.Input) > 0 {
		if last := req.Input[len(req.Input)-1]; last.ToolCallResult != nil &&
			last.ToolCallResult.OutputRole == "assistant" &&
			len(last.ToolCallResult.Output.Content) > 0 {
			resp = &types.CompletionResponse{
				Model: req.Model,
			}
			for _, content := range last.ToolCallResult.Output.Content {
				resp.Output = append(resp.Output, types.CompletionItem{
					Message: &mcp.SamplingMessage{
						Role:    "assistant",
						Content: content,
					},
				})
			}
			return req, resp
		}
	}

	newInput := make([]types.CompletionItem, 0, len(req.Input))
	for _, input := range req.Input {
		if input.ToolCallResult != nil && input.ToolCallResult.OutputRole == "assistant" &&
			len(input.ToolCallResult.Output.Content) > 0 {
			// elide the tool call result if it is an assistant response
			input = types.CompletionItem{
				ToolCallResult: &types.ToolCallResult{
					CallID: input.ToolCallResult.CallID,
					Output: types.CallResult{
						Content: []mcp.Content{{Text: "complete"}},
					},
				},
			}
		}
		newInput = append(newInput, input)
	}
	req.Input = newInput
	return req, nil
}

func (c Client) Complete(ctx context.Context, req types.CompletionRequest, opts ...types.CompletionOptions) (*types.CompletionResponse, error) {
	if req.Model == "default" || req.Model == "" {
		req.Model = c.defaultModel
	}

	req, resp := c.handleAssistantRolesFromTools(req)
	if resp != nil {
		return resp, nil
	}

	if strings.HasPrefix(req.Model, "claude") {
		return c.anthropic.Complete(ctx, req, opts...)
	}

	// Route to Ollama for common Ollama model patterns
	if strings.Contains(req.Model, "llama") ||
		strings.Contains(req.Model, "mistral") ||
		strings.Contains(req.Model, "gemma") ||
		strings.Contains(req.Model, "phi") ||
		strings.Contains(req.Model, "qwen") ||
		strings.Contains(req.Model, "codellama") ||
		strings.HasPrefix(req.Model, "ollama:") {
		// Remove ollama: prefix if present
		req.Model = strings.TrimPrefix(req.Model, "ollama:")
		return c.ollama.Complete(ctx, req, opts...)
	}

	return c.responses.Complete(ctx, req, opts...)
}
