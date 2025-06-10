package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/obot-platform/nanobot/pkg/complete"
	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/tools"
	"github.com/obot-platform/nanobot/pkg/types"
)

func (a *Agents) toolCalls(ctx context.Context, run *run, opts []types.CompletionOptions) error {
	for _, output := range run.Response.Output {
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

		callOutput, err := a.invoke(ctx, targetServer, functionCall, opts)
		if err != nil {
			return fmt.Errorf("failed to invoke tool %s on MCP server %s: %w", functionCall.Name, targetServer.MCPServer, err)
		}

		if run.ToolOutputs == nil {
			run.ToolOutputs = make(map[string]toolCall)
		}

		for _, opt := range opts {
			if opt.Progress != nil {
				data, err := json.Marshal(map[string]any{
					"type":     "nanobot/toolcall/output",
					"target":   targetServer,
					"toolCall": functionCall,
					"output":   callOutput,
				})
				if err == nil {
					opt.Progress <- data
				}
			}
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

func (a *Agents) confirm(ctx context.Context, target types.TargetMapping, funcCall *types.ToolCall) error {
	if a.confirmations == nil {
		return nil
	}
	if _, ok := a.config.Agents[target.MCPServer]; ok {
		// Don't require confirmations to talk to another agent
		return nil
	}
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil
	}
	return a.confirmations.Confirm(ctx, session, target, funcCall)
}

func (a *Agents) invoke(ctx context.Context, target types.TargetMapping, funcCall *types.ToolCall, opts []types.CompletionOptions) ([]types.CompletionInput, error) {
	var (
		data map[string]any
	)

	if funcCall.Arguments != "" {
		data = make(map[string]any)
		if err := json.Unmarshal([]byte(funcCall.Arguments), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal function call arguments: %w", err)
		}
	}

	if err := a.confirm(ctx, target, funcCall); err != nil {
		return nil, fmt.Errorf("failed to confirm tool call: %w", err)
	}

	response, err := a.registry.Call(ctx, target.MCPServer, target.TargetName, data, tools.CallOptions{
		ProgressToken: complete.Complete(opts...).ProgressToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke tool %s on mcp server %s: %w", target.TargetName, target.MCPServer, err)
	}

	return []types.CompletionInput{
		{
			ToolCallResult: &types.ToolCallResult{
				CallID:     funcCall.CallID,
				Output:     *response,
				OutputRole: a.config.Flows[target.MCPServer].OutputRole,
			},
		},
	}, nil
}
