package bifrost

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func toRequest(provider string, req *types.CompletionRequest) (*schemas.BifrostResponsesRequest, error) {
	params := &schemas.ResponsesParameters{}

	if req.SystemPrompt != "" {
		params.Instructions = &req.SystemPrompt
	}

	if req.Temperature != nil {
		if f, err := req.Temperature.Float64(); err == nil {
			params.Temperature = &f
		}
	}

	if req.TopP != nil {
		if f, err := req.TopP.Float64(); err == nil {
			params.TopP = &f
		}
	}

	if req.MaxTokens != 0 {
		params.MaxOutputTokens = &req.MaxTokens
	}

	if req.Truncation != "" {
		params.Truncation = &req.Truncation
	}

	for _, tool := range req.Tools {
		t, err := toTool(tool)
		if err != nil {
			return nil, err
		}
		params.Tools = append(params.Tools, t)
	}

	if req.ToolChoice != "" {
		switch req.ToolChoice {
		case "none", "auto", "required":
			s := req.ToolChoice
			params.ToolChoice = &schemas.ResponsesToolChoice{
				ResponsesToolChoiceStr: &s,
			}
		default:
			params.ToolChoice = &schemas.ResponsesToolChoice{
				ResponsesToolChoiceStruct: &schemas.ResponsesToolChoiceStruct{
					Type: schemas.ResponsesToolChoiceTypeFunction,
					Name: &req.ToolChoice,
				},
			}
		}
	}

	input, err := toInput(req)
	if err != nil {
		return nil, err
	}

	return &schemas.BifrostResponsesRequest{
		Provider:  schemas.ModelProvider(provider),
		Model:     req.Model,
		Input:     input,
		Params:    params,
		Fallbacks: []schemas.Fallback{},
	}, nil
}

func toTool(tool types.ToolUseDefinition) (schemas.ResponsesTool, error) {
	t := schemas.ResponsesTool{
		Type:        schemas.ResponsesToolTypeFunction,
		Name:        &tool.Name,
		Description: &tool.Description,
	}

	if len(tool.Parameters) > 0 {
		var params schemas.ToolFunctionParameters
		if err := json.Unmarshal(tool.Parameters, &params); err != nil {
			return t, fmt.Errorf("failed to unmarshal tool parameters for %q: %w", tool.Name, err)
		}
		t.ResponsesToolFunction = &schemas.ResponsesToolFunction{
			Parameters: &params,
		}
	}

	return t, nil
}

func toInput(req *types.CompletionRequest) ([]schemas.ResponsesMessage, error) {
	var result []schemas.ResponsesMessage

	for _, msg := range req.Input {
		msgs, err := toMessages(msg)
		if err != nil {
			return nil, err
		}
		result = append(result, msgs...)
	}

	return result, nil
}

func toMessages(msg types.Message) ([]schemas.ResponsesMessage, error) {
	var result []schemas.ResponsesMessage

	var contentBlocks []schemas.ResponsesMessageContentBlock
	for _, item := range msg.Items {
		if item.Content != nil {
			if block, ok := toContentBlock(*item.Content); ok {
				if msg.Role == "user" {
					contentBlocks = append(contentBlocks, block)
				} else {
					if m, ok := toAssistantTextMessage(msg.Role, *item.Content); ok {
						result = append(result, m)
					}
				}
			}
		}

		if item.ToolCall != nil {
			result = append(result, toFunctionCallMessage(item))
		}

		if item.ToolCallResult != nil {
			result = append(result, toFunctionCallOutputMessage(item.ToolCallResult))
		}
	}

	if len(contentBlocks) > 0 {
		role := schemas.ResponsesInputMessageRoleUser
		msgType := schemas.ResponsesMessageTypeMessage
		result = append(result, schemas.ResponsesMessage{
			Type: &msgType,
			Role: &role,
			Content: &schemas.ResponsesMessageContent{
				ContentBlocks: contentBlocks,
			},
		})
	}

	return result, nil
}

func toAssistantTextMessage(role string, content mcp.Content) (schemas.ResponsesMessage, bool) {
	if content.Type != "text" {
		return schemas.ResponsesMessage{}, false
	}

	msgType := schemas.ResponsesMessageTypeMessage
	r := schemas.ResponsesMessageRoleType(role)
	return schemas.ResponsesMessage{
		Type: &msgType,
		Role: &r,
		Content: &schemas.ResponsesMessageContent{
			ContentBlocks: []schemas.ResponsesMessageContentBlock{
				{
					Type: schemas.ResponsesOutputMessageContentTypeText,
					Text: &content.Text,
				},
			},
		},
	}, true
}

func toFunctionCallMessage(item types.CompletionItem) schemas.ResponsesMessage {
	msgType := schemas.ResponsesMessageTypeFunctionCall
	return schemas.ResponsesMessage{
		ID:   &item.ID,
		Type: &msgType,
		ResponsesToolMessage: &schemas.ResponsesToolMessage{
			CallID:    &item.ToolCall.CallID,
			Name:      &item.ToolCall.Name,
			Arguments: &item.ToolCall.Arguments,
		},
	}
}

func toFunctionCallOutputMessage(result *types.ToolCallResult) schemas.ResponsesMessage {
	msgType := schemas.ResponsesMessageTypeFunctionCallOutput

	var sb strings.Builder
	for _, c := range result.Output.Content {
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	outputStr := sb.String()
	if outputStr == "" {
		outputStr = "completed"
	}

	return schemas.ResponsesMessage{
		Type: &msgType,
		ResponsesToolMessage: &schemas.ResponsesToolMessage{
			CallID: &result.CallID,
			Output: &schemas.ResponsesToolMessageOutputStruct{
				ResponsesToolCallOutputStr: &outputStr,
			},
		},
	}
}

func toContentBlock(content mcp.Content) (schemas.ResponsesMessageContentBlock, bool) {
	switch content.Type {
	case "text":
		return schemas.ResponsesMessageContentBlock{
			Type: schemas.ResponsesInputMessageContentBlockTypeText,
			Text: &content.Text,
		}, true
	case "image":
		url := content.ToImageURL()
		return schemas.ResponsesMessageContentBlock{
			Type: schemas.ResponsesInputMessageContentBlockTypeImage,
			ResponsesInputMessageContentBlockImage: &schemas.ResponsesInputMessageContentBlockImage{
				ImageURL: &url,
			},
		}, true
	}
	return schemas.ResponsesMessageContentBlock{}, false
}
