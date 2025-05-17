package agents

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
)

func (a *Agents) toolCalls(ctx context.Context, run *run) error {
	for _, output := range run.Response.Output {
		var computerUseCall bool
		functionCall := output.FunctionCall
		if functionCall == nil && output.ComputerCall != nil {
			for _, tool := range run.PopulatedRequest.Tools {
				if tool.CustomTool != nil && tool.CustomTool.Attributes["type"] == "computer_use_preview" {
					args, _ := json.Marshal(output.ComputerCall.Action)
					functionCall = &types.FunctionCall{
						Name:      tool.CustomTool.Name,
						Arguments: string(args),
						CallID:    output.ComputerCall.CallID,
					}
					computerUseCall = true
					break
				}
			}
		}

		if functionCall == nil {
			continue
		}
		if run.ToolOutputs[functionCall.CallID].Done {
			continue
		}

		targetServer, ok := run.ToolToMCPServer[functionCall.Name]
		if !ok {
			return fmt.Errorf("can not map tool %s to a MCP server", functionCall.Name)
		}

		callOutput, err := a.Invoke(ctx, targetServer, functionCall)
		if err != nil {
			return fmt.Errorf("failed to invoke tool %s on MCP server %s: %w", functionCall.Name, targetServer, err)
		}

		if computerUseCall {
			for _, item := range callOutput {
				if item.Item == nil || item.Item.InputMessage == nil {
					continue
				}
				for _, content := range item.Item.InputMessage.Content.InputItemContent {
					if content.InputImage != nil && content.InputImage.ImageURL != nil {
						callOutput = []types.InputItem{
							{
								Item: &types.Item{
									ComputerCallOutput: &types.ComputerCallOutput{
										CallID: functionCall.CallID,
										Output: types.ComputerScreenshot{
											ImageURL: *content.InputImage.ImageURL,
										},
									},
								},
							},
						}
					}
				}
			}
		}

		if run.ToolOutputs == nil {
			run.ToolOutputs = make(map[string]toolCall)
		}

		run.ToolOutputs[functionCall.CallID] = toolCall{
			Output: callOutput,
			Done:   true,
		}
	}

	if len(run.ToolOutputs) == 0 {
		run.Done = true
	}

	return nil
}

func (a *Agents) Invoke(ctx context.Context, target types.ToolMapping, funcCall *types.FunctionCall) ([]types.InputItem, error) {
	var (
		data map[string]any
	)

	if funcCall.Arguments != "" {
		data = make(map[string]any)
		if err := json.Unmarshal([]byte(funcCall.Arguments), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal function call arguments: %w", err)
		}
	}

	response, err := a.registry.Call(ctx, target.Server, target.ToolName, data)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke tool %s on mcp server %s: %w", target.Tool, target.Server, err)
	}

	var (
		funcResponseSet  bool
		result           []types.InputItem
		inputItemContent []types.InputItemContent
	)

	for _, content := range response.Content {
		switch content.Type {
		case "text":
			if funcResponseSet {
				inputItemContent = append(inputItemContent, types.InputItemContent{
					InputText: &types.InputText{
						Text: content.Text,
					},
				})
			} else {
				result = append(result, types.InputItem{
					Item: &types.Item{
						FunctionCallOutput: &types.FunctionCallOutput{
							CallID: funcCall.CallID,
							Output: content.Text,
						},
					},
				})
			}
			funcResponseSet = true
		case "image":
			inputItemContent = append(inputItemContent, types.InputItemContent{
				InputImage: ToInputImage(content),
			})
		case "resources":
			if content.Resource != nil {
				inputItemContent = append(inputItemContent, types.InputItemContent{
					InputFile: ToInputFile(content.Resource),
				})
			}
		}
	}

	if len(inputItemContent) > 0 {
		result = append(result, types.InputItem{
			Item: &types.Item{
				InputMessage: &types.InputMessage{
					Content: types.InputContent{
						InputItemContent: inputItemContent,
					},
					Role: "user",
				},
			},
		})
	}

	return result, nil
}

func ToInputFile(file *mcp.Resource) *types.InputFile {
	if file.Text != "" {
		fileData := base64.StdEncoding.EncodeToString([]byte(file.Text))
		return &types.InputFile{
			FileData: &fileData,
			Filename: file.URI,
		}
	}
	if file.Blob != "" {
		return &types.InputFile{
			FileData: &file.Blob,
			Filename: file.URI,
		}
	}
	return &types.InputFile{}
}

func ToInputImage(img mcp.Content) *types.InputImage {
	data := "data:"
	if img.MIMEType != "" {
		data += img.MIMEType + ";"
	}
	data += "base64," + img.Data
	return &types.InputImage{
		ImageURL: &data,
	}
}
