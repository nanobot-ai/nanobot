package llm

import (
	"context"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/llm/anthropic"
	"github.com/nanobot-ai/nanobot/pkg/llm/progress"
	"github.com/nanobot-ai/nanobot/pkg/llm/responses"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

var _ types.Completer = (*Client)(nil)

type Config struct {
	DefaultModel string
	Responses    responses.Config
	Anthropic    anthropic.Config
}

func NewClient(cfg Config) *Client {
	return &Client{
		defaultModel: cfg.DefaultModel,
		responses:    responses.NewClient(cfg.Responses),
		anthropic:    anthropic.NewClient(cfg.Anthropic),
	}
}

type Client struct {
	defaultModel string
	responses    *responses.Client
	anthropic    *anthropic.Client
}

func (c *Client) handleAssistantRolesFromTools(req types.CompletionRequest) (_ types.CompletionRequest, resp *types.CompletionResponse) {
	if len(req.Input) > 0 {
		lastMsg := req.Input[len(req.Input)-1]
		if len(lastMsg.Items) > 0 {
			if last := lastMsg.Items[len(lastMsg.Items)-1]; last.ToolCallResult != nil &&
				last.ToolCallResult.OutputRole == "assistant" &&
				len(last.ToolCallResult.Output.Content) > 0 {
				resp = &types.CompletionResponse{
					Model: req.Model,
					Output: types.Message{
						ID:   uuid.String(),
						Role: "assistant",
					},
				}
				for _, content := range last.ToolCallResult.Output.Content {
					resp.Output.Items = append(resp.Output.Items, types.CompletionItem{
						ID:      uuid.String(),
						Content: &content,
					})
				}
				return req, resp
			}
		}
	}

	newMsgs := make([]types.Message, 0, len(req.Input))
	for _, msg := range req.Input {
		newItems := make([]types.CompletionItem, 0, len(msg.Items))
		for _, input := range msg.Items {
			if input.ToolCallResult != nil && input.ToolCallResult.OutputRole == "assistant" &&
				len(input.ToolCallResult.Output.Content) > 0 {
				// elide the tool call result if it is an assistant response
				input = types.CompletionItem{
					ID: input.ID,
					ToolCallResult: &types.ToolCallResult{
						CallID: input.ToolCallResult.CallID,
						Output: types.CallResult{
							Content: []mcp.Content{{Text: "completed"}},
						},
					},
				}
			}
			newItems = append(newItems, input)
		}
		newMsg := msg
		newMsg.Items = newItems
		newMsgs = append(newMsgs, newMsg)
	}
	req.Input = newMsgs
	return req, nil
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

	req, resp := c.handleAssistantRolesFromTools(req)
	if resp != nil {
		return resp, nil
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
	return c.responses.Complete(ctx, req, opts...)
}
