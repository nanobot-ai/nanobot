package responses

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func toResponse(req *types.CompletionRequest, resp *Response) (*types.CompletionResponse, error) {
	var created *time.Time
	if i, err := resp.CreatedAt.Int64(); err == nil {
		t := time.Unix(i, 0)
		created = &t
	}
	result := &types.CompletionResponse{
		Model: resp.Model,
		Output: types.Message{
			ID:      resp.ID,
			Created: created,
			Role:    "assistant",
		},
	}

	for _, output := range resp.Output {
		if output.ComputerCall != nil {
			for _, tool := range req.Tools {
				if tool.Attributes["type"] == "computer_use_preview" {
					args, _ := json.Marshal(output.ComputerCall.Action)
					result.Output.Items = append(result.Output.Items, types.CompletionItem{
						ID: output.ComputerCall.ID,
						ToolCall: &types.ToolCall{
							Name:      tool.Name,
							Arguments: string(args),
							CallID:    output.ComputerCall.CallID,
						},
					})
					break
				}
			}
		} else if output.FunctionCall != nil {
			result.Output.Items = append(result.Output.Items, types.CompletionItem{
				ID: output.FunctionCall.ID,
				ToolCall: &types.ToolCall{
					Name:      output.FunctionCall.Name,
					Arguments: output.FunctionCall.Arguments,
					CallID:    output.FunctionCall.CallID,
				},
			})
		} else if output.Message != nil {
			result.Output.Items = append(result.Output.Items, toSamplingMessageFromOutputMessage(output.Message)...)
			result.Output.Role = output.Message.Role
		} else if output.Reasoning != nil && output.Reasoning.EncryptedContent != nil {
			result.Output.Items = append(result.Output.Items, types.CompletionItem{
				ID: output.Reasoning.ID,
				Reasoning: &types.Reasoning{
					EncryptedContent: *output.Reasoning.EncryptedContent,
					Summary:          output.Reasoning.GetSummary(),
				},
			})
		}
	}

	return result, nil
}

func toSamplingMessageFromOutputMessage(output *Message) (result []types.CompletionItem) {
	for _, content := range output.Content {
		if content.OutputText != nil {
			result = append(result, types.CompletionItem{
				ID: output.ID,
				Content: &mcp.Content{
					Type: "text",
					Text: content.OutputText.Text,
				},
			})
		} else if content.Refusal != nil {
			result = append(result, types.CompletionItem{
				ID: output.ID,
				Content: &mcp.Content{
					Type: "text",
					Text: "REFUSAL: " + content.Refusal.Refusal,
				},
			})
		}
	}
	return
}

func toRequest(completion *types.CompletionRequest) (req Request, _ error) {
	req = Request{
		Model: completion.Model,
		Store: &[]bool{false}[0],
	}

	if reasoningPrefix.MatchString(req.Model) {
		req.Include = append(req.Include, "reasoning.encrypted_content")
		req.Reasoning = &ResponseReasoning{}
		if completion.Reasoning != nil && completion.Reasoning.Summary != "" {
			req.Reasoning.Summary = &completion.Reasoning.Summary
		} else {
			req.Reasoning.Summary = &[]string{"auto"}[0]
		}
		if completion.Reasoning != nil && completion.Reasoning.Effort != "" {
			req.Reasoning.Effort = &completion.Reasoning.Effort
		} else {
			req.Reasoning.Effort = &[]string{"medium"}[0]
		}
	}

	if completion.Truncation != "" {
		req.Truncation = &completion.Truncation
	}

	if completion.Temperature != nil {
		req.Temperature = completion.Temperature
	}

	if completion.TopP != nil {
		req.TopP = completion.TopP
	}

	if len(completion.Metadata) > 0 {
		req.Metadata = map[string]string{}
		for k, v := range completion.Metadata {
			req.Metadata[k] = fmt.Sprint(v)
		}
	}

	if completion.SystemPrompt != "" {
		req.Instructions = &completion.SystemPrompt
	}

	if completion.MaxTokens != 0 {
		req.MaxOutputTokens = &completion.MaxTokens
	}

	if completion.ToolChoice != "" {
		switch completion.ToolChoice {
		case "none", "auto", "required":
			req.ToolChoice = &ToolChoice{
				Mode: completion.ToolChoice,
			}
		case "file_search", "web_search_preview", "computer_use_preview":
			req.ToolChoice = &ToolChoice{
				HostedTool: &HostedTool{
					Type: completion.ToolChoice,
				},
			}
		default:
			req.ToolChoice = &ToolChoice{
				FunctionTool: &FunctionTool{
					Name: completion.ToolChoice,
				},
			}
		}
	}

	if completion.OutputSchema != nil {
		req.Text = &TextFormatting{
			Format: Format{
				JSONSchema: &JSONSchema{
					Name:        completion.OutputSchema.Name,
					Description: completion.OutputSchema.Description,
					Schema:      completion.OutputSchema.ToSchema(),
					Strict:      completion.OutputSchema.Strict,
				},
			},
		}
		if req.Text.Format.Name == "" {
			req.Text.Format.Name = "output-schema"
		}
	}

	for _, tool := range completion.Tools {
		req.Tools = append(req.Tools, Tool{
			CustomTool: &CustomTool{
				Name:        tool.Name,
				Parameters:  tool.Parameters,
				Description: tool.Description,
				Attributes:  tool.Attributes,
			},
		})
	}

	for _, msg := range completion.Input {
		for _, input := range msg.Items {
			if input.Content != nil {
				inputItem, ok := messageToInputItem(msg.Role, *input.Content)
				if ok {
					req.Input.Items = append(req.Input.Items, inputItem)
				}
			}
			if input.ToolCall != nil {
				inputItem, err := toolCallToInputItem(completion, input.ID, input.ToolCall)
				if err != nil {
					return req, err
				}
				req.Input.Items = append(req.Input.Items, inputItem)
			}
			if input.ToolCallResult != nil {
				req.Input.Items = append(req.Input.Items, toolCallResultToInputItems(completion, input.ToolCallResult)...)
			}
			if input.Reasoning != nil && input.Reasoning.EncryptedContent != "" {
				// summary must not be nil
				summary := make([]SummaryText, 0)
				for _, s := range input.Reasoning.Summary {
					summary = append(summary, SummaryText{
						Text: s.Text,
					})
				}

				req.Input.Items = append(req.Input.Items, InputItem{
					Item: &Item{
						Reasoning: &Reasoning{
							ID:               input.ID,
							EncryptedContent: &input.Reasoning.EncryptedContent,
							Summary:          summary,
						},
					},
				})
			}
		}
	}

	return req, nil
}

func isComputerUse(completion *types.CompletionRequest, name string) bool {
	for _, toolDef := range completion.Tools {
		if toolDef.Name == name && toolDef.Attributes["type"] == "computer_use_preview" {
			return true
		}
	}
	return false
}

func getToolCall(completion *types.CompletionRequest, callID string) types.ToolCall {
	for _, msg := range completion.Input {
		for _, input := range msg.Items {
			if input.ToolCall != nil && input.ToolCall.CallID == callID {
				return *input.ToolCall
			}
		}
	}
	return types.ToolCall{}
}

func contentToInputItem(content mcp.Content) (InputItemContent, bool) {
	switch content.Type {
	case "text":
		return InputItemContent{
			InputText: &InputText{
				Text: content.Text,
			},
		}, true
	case "image":
		url := content.ToImageURL()
		return InputItemContent{
			InputImage: &InputImage{
				ImageURL: &url,
			},
		}, true
	case "audio":
		return InputItemContent{
			InputFile: &InputFile{
				FileData: &content.Data,
			},
		}, true
	case "resources":
		if content.Resource != nil {
			return InputItemContent{
				InputFile: toInputFile(content.Resource),
			}, true
		}
	}
	return InputItemContent{}, false
}

func fcOutputText(callID, text string) *InputItem {
	return &InputItem{
		Item: &Item{
			FunctionCallOutput: &FunctionCallOutput{
				CallID: callID,
				Output: text,
			},
		},
	}
}

func fcOutputImage(callID string, imageURL string) *InputItem {
	return &InputItem{
		Item: &Item{
			ComputerCallOutput: &ComputerCallOutput{
				CallID: callID,
				Output: ComputerScreenshot{
					ImageURL: imageURL,
				},
			},
		},
	}
}

func toolCallResultToInputItems(completion *types.CompletionRequest, toolCallResult *types.ToolCallResult) (result []InputItem) {
	var (
		isComputerUseCall = isComputerUse(completion, getToolCall(completion, toolCallResult.CallID).Name)
		outputType        = "text"
		fcOutput          *InputItem
	)

	if isComputerUseCall {
		outputType = "image"
	}

	for _, output := range toolCallResult.Output.Content {
		if fcOutput == nil && outputType == output.Type {
			if output.Type == "text" {
				fcOutput = fcOutputText(toolCallResult.CallID, output.Text)
			} else {
				fcOutput = fcOutputImage(toolCallResult.CallID, output.ToImageURL())
			}
			result = append(result, *fcOutput)
			continue
		}

		inputItemContent, ok := contentToInputItem(output)
		if !ok {
			continue
		}

		result = append(result, InputItem{
			Item: &Item{
				InputMessage: &InputMessage{
					Content: InputContent{
						InputItemContent: []InputItemContent{
							inputItemContent,
						},
					},
					Role: "user",
				},
			},
		})
	}

	if fcOutput == nil {
		// This can happen if the MCP server returns an empty response or only an image
		result = append(result, InputItem{
			Item: &Item{
				FunctionCallOutput: &FunctionCallOutput{
					CallID: toolCallResult.CallID,
					Output: "completed",
				},
			},
		})
	}

	return result
}

func toolCallToInputItem(completion *types.CompletionRequest, id string, toolCall *types.ToolCall) (InputItem, error) {
	if isComputerUse(completion, toolCall.Name) {
		var args ComputerCallAction
		if toolCall.Arguments != "" {
			if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
				return InputItem{}, fmt.Errorf("failed to unmarshal function call arguments for computer call: %w", err)
			}
		}
		return InputItem{
			Item: &Item{
				ComputerCall: &ComputerCall{
					ID:     id,
					CallID: toolCall.CallID,
					Action: args,
				},
			},
		}, nil
	}

	return InputItem{
		Item: &Item{
			FunctionCall: &FunctionCall{
				Arguments: toolCall.Arguments,
				CallID:    toolCall.CallID,
				Name:      toolCall.Name,
				ID:        id,
			},
		},
	}, nil
}

func messageToInputItem(role string, content mcp.Content) (InputItem, bool) {
	if role == "assistant" && content.Type == "text" {
		return InputItem{
			Item: &Item{
				Message: &Message{
					Content: []MessageContent{
						{
							OutputText: &OutputText{
								Text: content.Text,
							},
						},
					},
					Role: role,
				},
			},
		}, true
	}

	inputItemContent, ok := contentToInputItem(content)
	if !ok {
		return InputItem{}, false
	}

	return InputItem{
		Item: &Item{
			InputMessage: &InputMessage{
				Content: InputContent{
					InputItemContent: []InputItemContent{
						inputItemContent,
					},
				},
				Role: role,
			},
		},
	}, true
}

func toInputFile(file *mcp.EmbeddedResource) *InputFile {
	if file.Text != "" {
		fileData := base64.StdEncoding.EncodeToString([]byte(file.Text))
		return &InputFile{
			FileData: &fileData,
			Filename: file.URI,
		}
	}
	if file.Blob != "" {
		return &InputFile{
			FileData: &file.Blob,
			Filename: file.URI,
		}
	}
	return &InputFile{}
}
