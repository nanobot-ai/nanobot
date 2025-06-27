package ollama

import (
	"encoding/json"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func toResponse(resp *Response) (*types.CompletionResponse, error) {
	result := &types.CompletionResponse{
		Model: resp.Model,
	}

	// Handle tool calls
	if len(resp.Message.ToolCalls) > 0 {
		for _, toolCall := range resp.Message.ToolCalls {
			args, _ := json.Marshal(toolCall.Function.Arguments)
			result.Output = append(result.Output, types.CompletionItem{
				ToolCall: &types.ToolCall{
					CallID:    toolCall.ID,
					Name:      toolCall.Function.Name,
					Arguments: string(args),
				},
			})
		}
	} else if resp.Message.Content != "" {
		// Handle regular text response
		result.Output = append(result.Output, types.CompletionItem{
			Message: &mcp.SamplingMessage{
				Role: "assistant",
				Content: mcp.Content{
					Type: "text",
					Text: resp.Message.Content,
				},
			},
		})
	} else {
		// If no content and no tool calls, add an empty message to prevent nil response
		result.Output = append(result.Output, types.CompletionItem{
			Message: &mcp.SamplingMessage{
				Role: "assistant",
				Content: mcp.Content{
					Type: "text",
					Text: "",
				},
			},
		})
	}

	return result, nil
}

func toRequest(req *types.CompletionRequest) (Request, error) {
	result := Request{
		Model:  req.Model,
		Stream: true,
	}

	// Convert options
	if req.Temperature != nil || req.TopP != nil || req.MaxTokens > 0 {
		options := &Options{}
		if req.Temperature != nil {
			if temp, err := req.Temperature.Float64(); err == nil {
				options.Temperature = &temp
			}
		}
		if req.TopP != nil {
			if topP, err := req.TopP.Float64(); err == nil {
				options.TopP = &topP
			}
		}
		if req.MaxTokens > 0 {
			options.NumCtx = &req.MaxTokens
		}
		result.Options = options
	}

	// Convert tools
	for _, tool := range req.Tools {
		var parameters map[string]interface{}
		if len(tool.Parameters) > 0 {
			if err := json.Unmarshal(tool.Parameters, &parameters); err != nil {
				// If unmarshal fails, create empty parameters object
				parameters = make(map[string]interface{})
			}
		} else {
			parameters = make(map[string]interface{})
		}

		result.Tools = append(result.Tools, Tool{
			Type: "function",
			Function: Function{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  parameters,
			},
		})
	}

	// Add system message if present
	if req.SystemPrompt != "" {
		result.Messages = append(result.Messages, Message{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	// Convert input messages
	for _, input := range req.Input {
		if input.Message != nil {
			result.Messages = append(result.Messages, Message{
				Role:    input.Message.Role,
				Content: contentToString(input.Message.Content),
			})
		}
		if input.ToolCall != nil {
			// Assistant message with tool call
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(input.ToolCall.Arguments), &args); err != nil {
				return Request{}, fmt.Errorf("failed to unmarshal tool call arguments: %w", err)
			}

			result.Messages = append(result.Messages, Message{
				Role: "assistant",
				ToolCalls: []ToolCall{
					{
						ID:   input.ToolCall.CallID,
						Type: "function",
						Function: FunctionCall{
							Name:      input.ToolCall.Name,
							Arguments: args,
						},
					},
				},
			})
		}
		if input.ToolCallResult != nil {
			// Tool result message
			result.Messages = append(result.Messages, Message{
				Role:    "tool",
				Content: contentArrayToString(input.ToolCallResult.Output.Content),
			})
		}
	}

	// Ensure we have at least one message
	if len(result.Messages) == 0 {
		result.Messages = append(result.Messages, Message{
			Role:    "user",
			Content: "Hello",
		})
	}

	return result, nil
}

func contentToString(content mcp.Content) string {
	switch content.Type {
	case "text":
		return content.Text
	case "image":
		// Ollama supports images, but we'll need to handle them differently
		// For now, just return a description
		return fmt.Sprintf("[Image: %s]", content.MIMEType)
	default:
		return content.Text
	}
}

func contentArrayToString(contents []mcp.Content) string {
	if len(contents) == 0 {
		return ""
	}
	if len(contents) == 1 {
		return contentToString(contents[0])
	}

	var result string
	for i, content := range contents {
		if i > 0 {
			result += "\n"
		}
		result += contentToString(content)
	}
	return result
}
