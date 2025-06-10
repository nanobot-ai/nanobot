package responses

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
)

func toResponse(req *types.CompletionRequest, resp *Response) (*types.CompletionResponse, error) {
	result := &types.CompletionResponse{
		Model: resp.Model,
	}

	for _, output := range resp.Output {
		if output.ComputerCall != nil {
			for _, tool := range req.Tools {
				if tool.Attributes["type"] == "computer_use_preview" {
					args, _ := json.Marshal(output.ComputerCall.Action)
					result.Output = append(result.Output, types.CompletionOutput{
						ToolCall: &types.ToolCall{
							Name:      tool.Name,
							Arguments: string(args),
							CallID:    output.ComputerCall.CallID,
							ID:        output.ComputerCall.ID,
						},
					})
					break
				}
			}
		} else if output.FunctionCall != nil {
			result.Output = append(result.Output, types.CompletionOutput{
				ToolCall: &types.ToolCall{
					Name:      output.FunctionCall.Name,
					Arguments: output.FunctionCall.Arguments,
					CallID:    output.FunctionCall.CallID,
					ID:        output.FunctionCall.ID,
				},
			})
		} else if output.Message != nil {
			result.Output = append(result.Output, toSamplingMessageFromOutputMessage(output.Message)...)
		} else if output.Reasoning != nil && output.Reasoning.EncryptedContent != nil {
			result.Output = append(result.Output, types.CompletionOutput{
				Reasoning: &types.Reasoning{
					ID:               output.Reasoning.ID,
					EncryptedContent: *output.Reasoning.EncryptedContent,
					Summary:          output.Reasoning.GetSummary(),
				},
			})
		}
	}

	return result, nil
}

func toSamplingMessageFromOutputMessage(output *Message) (result []types.CompletionOutput) {
	for _, content := range output.Content {
		if content.OutputText != nil {
			result = append(result, types.CompletionOutput{
				Message: &mcp.SamplingMessage{
					Role: output.Role,
					Content: mcp.Content{
						Type: "text",
						Text: content.OutputText.Text,
					},
				},
			})
		} else if content.Refusal != nil {
			result = append(result, types.CompletionOutput{
				Message: &mcp.SamplingMessage{
					Role: output.Role,
					Content: mcp.Content{
						Type: "text",
						Text: content.Refusal.Refusal,
					},
				},
			})
		}
	}
	return
}

func toRequest(completion *types.CompletionRequest) (req Request, _ error) {
	req = Request{
		Model: completion.Model,
	}

	if reasoningPrefix.MatchString(req.Model) {
		req.Include = append(req.Include, "reasoning.encrypted_content")
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

	for _, input := range completion.Input {
		if input.Message != nil {
			inputItem, ok := messageToInputItem(input.Message)
			if ok {
				req.Input.Items = append(req.Input.Items, inputItem)
			}
		}
		if input.ToolCall != nil {
			inputItem, err := toolCallToInputItem(completion, input.ToolCall)
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
						ID:               input.Reasoning.ID,
						EncryptedContent: &input.Reasoning.EncryptedContent,
						Summary:          summary,
					},
				},
			})
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
	for _, input := range completion.Input {
		if input.ToolCall != nil && input.ToolCall.CallID == callID {
			return *input.ToolCall
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

func toolCallToInputItem(completion *types.CompletionRequest, toolCall *types.ToolCall) (InputItem, error) {
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
					ID:     toolCall.ID,
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
				ID:        toolCall.ID,
			},
		},
	}, nil
}

func messageToInputItem(msg *mcp.SamplingMessage) (InputItem, bool) {
	if msg.Role == "assistant" && msg.Content.Type == "text" {
		return InputItem{
			Item: &Item{
				Message: &Message{
					Content: []MessageContent{
						{
							OutputText: &OutputText{
								Text: msg.Content.Text,
							},
						},
					},
					Role: msg.Role,
				},
			},
		}, true
	}

	inputItemContent, ok := contentToInputItem(msg.Content)
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
				Role: msg.Role,
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
