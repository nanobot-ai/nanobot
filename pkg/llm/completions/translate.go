package completions

import (
	"encoding/json"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

// toRequest converts a nanobot CompletionRequest to an OpenAI ChatCompletionRequest
func toRequest(req *types.CompletionRequest) (*ChatCompletionRequest, error) {
	openaiReq := &ChatCompletionRequest{
		Model:             req.Model,
		Stream:            true, // Always use streaming for consistency with other backends
		ParallelToolCalls: &[]bool{false}[0], // Disable parallel tool calls for sequential execution
	}

	// Convert messages
	messages := make([]ChatMessage, 0, len(req.Input))
	for _, input := range req.Input {
		if input.Message != nil {
			msg, err := convertMessage(input.Message)
			if err != nil {
				return nil, fmt.Errorf("failed to convert message: %w", err)
			}
			messages = append(messages, msg)
		}

		if input.ToolCall != nil {
			// Convert tool call to assistant message with tool_calls
			args := input.ToolCall.Arguments
			if args == "" {
				args = "{}" // OpenAI requires at least empty JSON object
			}

			toolCall := ToolCall{
				ID:   input.ToolCall.CallID,
				Type: "function",
				Function: FunctionCall{
					Name:      input.ToolCall.Name,
					Arguments: args,
				},
			}

			toolMsg := ChatMessage{
				Role:      "assistant",
				Content:   "", // Content field required even for tool calls
				ToolCalls: []ToolCall{toolCall},
			}
			messages = append(messages, toolMsg)
		}

		if input.ToolCallResult != nil {
			// Add tool call result as a tool message
			toolMsg := ChatMessage{
				Role:       "tool",
				Content:    formatToolCallResult(input.ToolCallResult),
				ToolCallID: input.ToolCallResult.CallID,
			}
			messages = append(messages, toolMsg)
		}
	}
	openaiReq.Messages = messages

	// Convert tools
	if len(req.Tools) > 0 {
		tools := make([]Tool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			openaiTool, err := convertTool(tool)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool: %w", err)
			}
			tools = append(tools, openaiTool)
		}
		openaiReq.Tools = tools
	}

	// Set parameters
	if req.MaxTokens > 0 {
		openaiReq.MaxTokens = &req.MaxTokens
	}
	if req.Temperature != nil {
		temp, err := req.Temperature.Float64()
		if err != nil {
			return nil, fmt.Errorf("invalid temperature: %w", err)
		}
		openaiReq.Temperature = &temp
	}
	if req.TopP != nil {
		topP, err := req.TopP.Float64()
		if err != nil {
			return nil, fmt.Errorf("invalid top_p: %w", err)
		}
		openaiReq.TopP = &topP
	}

	// Convert OutputSchema to ResponseFormat for structured output
	if req.OutputSchema != nil {
		responseFormat := &ResponseFormat{
			Type: "json_schema",
			JSONSchema: &ResponseFormatSchema{
				Name:        req.OutputSchema.Name,
				Description: req.OutputSchema.Description,
				Strict:      req.OutputSchema.Strict,
			},
		}

		// Use the schema if available, otherwise use fields to build schema
		if len(req.OutputSchema.Schema) > 0 {
			var schema interface{}
			if err := json.Unmarshal(req.OutputSchema.Schema, &schema); err != nil {
				return nil, fmt.Errorf("failed to parse output schema: %w", err)
			}
			responseFormat.JSONSchema.Schema = schema
		} else {
			// Use ToSchema() method to build schema from fields
			schemaData := req.OutputSchema.ToSchema()
			if len(schemaData) > 0 {
				var schema interface{}
				if err := json.Unmarshal(schemaData, &schema); err != nil {
					return nil, fmt.Errorf("failed to build output schema: %w", err)
				}
				responseFormat.JSONSchema.Schema = schema
			}
		}

		openaiReq.ResponseFormat = responseFormat
	}

	return openaiReq, nil
}

// convertMessage converts an MCP SamplingMessage to an OpenAI ChatMessage
func convertMessage(msg *mcp.SamplingMessage) (ChatMessage, error) {
	chatMsg := ChatMessage{
		Role: msg.Role,
	}

	// Convert content - note that MCP Content is a single struct, not a slice
	if msg.Content.Text != "" {
		chatMsg.Content = msg.Content.Text
	} else if msg.Content.Type == "image" && msg.Content.Data != "" {
		chatMsg.Content = []ContentPart{{
			Type: "image_url",
			ImageURL: &struct {
				URL    string `json:"url"`
				Detail string `json:"detail,omitempty"`
			}{
				URL: msg.Content.Data,
			},
		}}
	}

	return chatMsg, nil
}

// convertTool converts a nanobot ToolUseDefinition to an OpenAI Tool
func convertTool(tool types.ToolUseDefinition) (Tool, error) {
	return Tool{
		Type: "function",
		Function: Function{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		},
	}, nil
}

// formatToolCallResult formats a tool call result for OpenAI
func formatToolCallResult(result *types.ToolCallResult) string {
	if len(result.Output.Content) == 0 {
		return ""
	}

	// Simple case: single text content
	if len(result.Output.Content) == 1 && result.Output.Content[0].Text != "" {
		return result.Output.Content[0].Text
	}

	// Multiple content parts - format as JSON
	output := make([]map[string]interface{}, 0, len(result.Output.Content))
	for _, content := range result.Output.Content {
		part := map[string]interface{}{}
		if content.Text != "" {
			part["text"] = content.Text
		}
		if content.Type == "image" && content.Data != "" {
			part["image"] = content.Data
		}
		output = append(output, part)
	}

	data, _ := json.Marshal(output)
	return string(data)
}

// toResponse converts an OpenAI ChatCompletionResponse to a nanobot CompletionResponse
func toResponse(resp *ChatCompletionResponse) (*types.CompletionResponse, error) {
	nanobotResp := &types.CompletionResponse{
		Model: resp.Model,
	}

	for _, choice := range resp.Choices {
		// Handle text content
		if content := getTextContent(choice.Message.Content); content != "" {
			nanobotResp.Output = append(nanobotResp.Output, types.CompletionItem{
				Message: &mcp.SamplingMessage{
					Role: choice.Message.Role,
					Content: mcp.Content{
						Type: "text",
						Text: content,
					},
				},
			})
		}

		// Handle tool calls - FORCE SEQUENTIAL EXECUTION by returning only the first tool call
		// This ensures tool calls are processed one at a time, preventing parallel execution
		for _, toolCall := range choice.Message.ToolCalls {
			if toolCall.Function.Name == "" {
				// Skip tool calls with empty names to avoid errors
				continue
			}
			// Only add the first valid tool call to force sequential processing
			nanobotResp.Output = append(nanobotResp.Output, types.CompletionItem{
				ToolCall: &types.ToolCall{
					CallID:    toolCall.ID,
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments,
				},
			})
			// Break after first tool call to ensure sequential execution
			break
		}
	}

	return nanobotResp, nil
}

// getTextContent extracts text content from various content formats
func getTextContent(content interface{}) string {
	switch c := content.(type) {
	case string:
		return c
	case []ContentPart:
		for _, part := range c {
			if part.Type == "text" && part.Text != "" {
				return part.Text
			}
		}
	}
	return ""
}
