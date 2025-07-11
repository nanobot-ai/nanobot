package anthropic

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func toResponse(resp *Response, created time.Time) (*types.CompletionResponse, error) {
	result := &types.CompletionResponse{
		Model: resp.Model,
		Output: types.Message{
			ID:      resp.ID,
			Created: &created,
			Role:    "assistant",
		},
	}

	for contentIndex, content := range resp.Content {
		if content.Type == "tool_use" {
			args, _ := json.Marshal(content.Input)
			result.Output.Items = append(result.Output.Items, types.CompletionItem{
				ID: fmt.Sprintf("%s-%d", resp.ID, contentIndex),
				ToolCall: &types.ToolCall{
					CallID:    content.ID,
					Name:      content.Name,
					Arguments: string(args),
				},
			})
		} else if content.Type == "text" && content.Text != nil {
			result.Output.Items = append(result.Output.Items, types.CompletionItem{
				ID: fmt.Sprintf("%s-%d", resp.ID, contentIndex),
				Content: &mcp.Content{
					Type: "text",
					Text: *content.Text,
				},
			})
		} else if content.Type == "image" {
			result.Output.Items = append(result.Output.Items, types.CompletionItem{
				ID: fmt.Sprintf("%s-%d", resp.ID, contentIndex),
				Content: &mcp.Content{
					Type:     "image",
					MIMEType: content.Source.MediaType,
					Data:     content.Source.Data,
				},
			})
		}
	}

	return result, nil
}

func toRequest(req *types.CompletionRequest) (Request, error) {
	// TODO: handle output schema

	if req.MaxTokens == 0 {
		req.MaxTokens = 64_000
	}

	result := Request{
		Model:       req.Model,
		System:      req.SystemPrompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Metadata:    req.Metadata,
	}

	for _, tool := range req.Tools {
		result.Tools = append(result.Tools, CustomTool{
			Name:        tool.Name,
			InputSchema: tool.Parameters,
			Description: tool.Description,
			Attributes:  tool.Attributes,
		})
	}

	if req.ToolChoice != "" {
		switch req.ToolChoice {
		case "auto":
			result.ToolChoice = &ToolChoice{
				Type: "auto",
			}
		case "none":
			result.ToolChoice = &ToolChoice{
				Type: "none",
			}
		default:
			result.ToolChoice = &ToolChoice{
				Type: "tool",
				Name: req.ToolChoice,
			}
		}
	}

	for _, msg := range req.Input {
		for _, input := range msg.Items {
			if input.Content != nil {
				result.Messages = append(result.Messages, Message{
					Content: contentToContent([]mcp.Content{*input.Content}),
					Role:    msg.Role,
				})
			}
			if input.ToolCall != nil {
				args := map[string]any{}
				if err := json.Unmarshal([]byte(input.ToolCall.Arguments), &args); err != nil {
					return Request{}, fmt.Errorf("failed to unmarshal tool call arguments: %w", err)
				}
				result.Messages = append(result.Messages, Message{
					Content: []Content{
						{
							Type:  "tool_use",
							ID:    input.ToolCall.CallID,
							Input: args,
							Name:  input.ToolCall.Name,
						},
					},
					Role: "assistant",
				})
			}
			if input.ToolCallResult != nil {
				result.Messages = append(result.Messages, Message{
					Content: []Content{
						{
							Type:      "tool_result",
							ToolUseID: input.ToolCallResult.CallID,
							Content:   contentToContent(input.ToolCallResult.Output.Content),
							IsError:   input.ToolCallResult.Output.IsError,
						},
					},
					Role: "user",
				})
			}
		}
	}

	return result, nil
}

func contentToContent(content []mcp.Content) (result []Content) {
	for _, item := range content {
		if item.Type == "text" || item.Type == "" {
			result = append(result, Content{
				Type: "text",
				Text: &item.Text,
			})
		} else if item.Type == "image" {
			result = append(result, Content{
				Type: "image",
				Source: ImageSource{
					MediaType: item.MIMEType,
					Data:      item.Data,
				},
			})
		}
	}
	return
}
