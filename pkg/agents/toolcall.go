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
			return fmt.Errorf("can not map tool %s to a MCP server", functionCall.Name)
		}

		callOutput, err := a.invoke(ctx, config, targetServer, tools.ToolCallInvocation{
			MessageID: run.Response.Output.ID,
			ItemID:    output.ID,
			ToolCall:  *functionCall,
		}, opts)
		if err != nil {
			return fmt.Errorf("failed to invoke tool %s on MCP server %s: %w", functionCall.Name, targetServer.MCPServer, err)
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

func (a *Agents) confirm(ctx context.Context, config types.Config, target types.TargetMapping[mcp.Tool], funcCall *types.ToolCall) (*types.CallResult, error) {
	if _, ok := config.Agents[target.MCPServer]; ok {
		// Don't require confirmations to talk to another agent
		return nil, nil
	}
	if _, ok := config.Flows[target.MCPServer]; ok {
		// Don't require confirmations to talk to a flow
		return nil, nil
	}
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil, nil
	}
	return a.confirmations.Confirm(ctx, session, target, funcCall)
}

func (a *Agents) invoke(ctx context.Context, config types.Config, target types.TargetMapping[mcp.Tool], funcCall tools.ToolCallInvocation, opts []types.CompletionOptions) (*types.Message, error) {
	var (
		data map[string]any
	)

	if funcCall.ToolCall.Arguments != "" {
		data = make(map[string]any)
		if err := json.Unmarshal([]byte(funcCall.ToolCall.Arguments), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal function call arguments: %w", err)
		}
	}

	response, err := a.confirm(ctx, config, target, &funcCall.ToolCall)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm tool call: %w", err)
	}

	if response == nil {
		response, err = a.registry.Call(ctx, target.MCPServer, target.TargetName, data, tools.CallOptions{
			ProgressToken:      complete.Complete(opts...).ProgressToken,
			ToolCallInvocation: &funcCall,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to invoke tool %s on mcp server %s: %w", target.TargetName, target.MCPServer, err)
		}
	}
	return &types.Message{
		Role: "user",
		Items: []types.CompletionItem{
			{
				ToolCallResult: &types.ToolCallResult{
					CallID:     funcCall.ToolCall.CallID,
					Output:     *response,
					OutputRole: config.Flows[target.MCPServer].OutputRole,
				},
			},
		},
	}, nil
}
