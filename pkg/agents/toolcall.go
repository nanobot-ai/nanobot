package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

// createErrorToolOutput creates a ToolOutput with error content for a failed tool call
func createErrorToolOutput(callID, errorMessage string) types.ToolOutput {
	return types.ToolOutput{
		Output: types.Message{
			Role: "user",
			Items: []types.CompletionItem{
				{
					ToolCallResult: &types.ToolCallResult{
						CallID: callID,
						Output: types.CallResult{
							Content: []mcp.Content{
								{
									Type: "text",
									Text: errorMessage,
								},
							},
							IsError: true,
						},
					},
				},
			},
		},
		Done: true,
	}
}

func (a *Agents) toolCalls(ctx context.Context, config types.Config, run *types.Execution, opts []types.CompletionOptions) error {
	for _, output := range run.Response.Output.Items {
		functionCall := output.ToolCall

		if functionCall == nil {
			continue
		}

		if run.ToolOutputs[functionCall.CallID].Done {
			continue
		}

		targetServer, ok := run.ToolToMCPServer[functionCall.Name]
		if !ok {
			// Store error as tool output instead of returning
			if run.ToolOutputs == nil {
				run.ToolOutputs = make(map[string]types.ToolOutput)
			}
			run.ToolOutputs[functionCall.CallID] = createErrorToolOutput(
				functionCall.CallID,
				fmt.Sprintf("Error: tool %s not found", functionCall.Name),
			)
			continue
		}

		if targetServer.Target.External {
			// Handled externally, so terminate the run waiting for the client
			run.Done = true
			continue
		}

		callOutput, err := a.invoke(ctx, targetServer, tools.ToolCallInvocation{
			MessageID: run.Response.Output.ID,
			ItemID:    output.ID,
			ToolCall:  *functionCall,
		}, opts)
		if err != nil {
			// Store error as tool output instead of returning
			if run.ToolOutputs == nil {
				run.ToolOutputs = make(map[string]types.ToolOutput)
			}
			run.ToolOutputs[functionCall.CallID] = createErrorToolOutput(
				functionCall.CallID,
				fmt.Sprintf("Error calling %s: %v", functionCall.Name, err),
			)
			continue
		}

		if run.ToolOutputs == nil {
			run.ToolOutputs = make(map[string]types.ToolOutput)
		}

		run.ToolOutputs[functionCall.CallID] = types.ToolOutput{
			Output: *callOutput,
			Done:   true,
		}
	}

	if len(run.ToolOutputs) == 0 {
		run.Done = true
	}

	return nil
}

func (a *Agents) invoke(ctx context.Context, target types.TargetMapping[types.TargetTool], funcCall tools.ToolCallInvocation, opts []types.CompletionOptions) (*types.Message, error) {
	var (
		data     map[string]any
		response *types.CallResult
	)

	if funcCall.ToolCall.Arguments != "" {
		data = make(map[string]any)
		if err := json.Unmarshal([]byte(funcCall.ToolCall.Arguments), &data); err != nil {
			// Return error as content instead of returning an error
			response = &types.CallResult{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: fmt.Sprintf("Error unmarshalling arguments for %s: %v", target.TargetName, err),
					},
				},
				IsError: true,
			}
			return &types.Message{
				Role: "user",
				Items: []types.CompletionItem{
					{
						ToolCallResult: &types.ToolCallResult{
							CallID: funcCall.ToolCall.CallID,
							Output: *response,
						},
					},
				},
			}, nil
		}
	}

	response, err := a.registry.Call(ctx, target.MCPServer, target.TargetName, data, tools.CallOptions{
		ProgressToken:      complete.Complete(opts...).ProgressToken,
		ToolCallInvocation: &funcCall,
	})
	if err != nil {
		response = &types.CallResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Error calling %s: %v", target.TargetName, err),
				},
			},
			IsError: true,
		}
	}
	return &types.Message{
		Role: "user",
		Items: []types.CompletionItem{
			{
				ToolCallResult: &types.ToolCallResult{
					CallID: funcCall.ToolCall.CallID,
					Output: *response,
				},
			},
		},
	}, nil
}
