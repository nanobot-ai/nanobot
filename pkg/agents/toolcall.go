package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func (a *Agents) toolCalls(ctx context.Context, config types.Config, run *types.Execution, opts []types.CompletionOptions) error {
	// Get agent config for truncation settings
	agentName := run.Request.GetAgent()
	agent, ok := config.Agents[agentName]
	if !ok {
		return fmt.Errorf("agent %s not found in config", agentName)
	}

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

		if targetServer.Target.External {
			// Handled externally, so terminate the run waiting for the client
			run.Done = true
			continue
		}

		callOutput, err := a.invoke(ctx, &agent, targetServer, tools.ToolCallInvocation{
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

func (a *Agents) invoke(ctx context.Context, agent *types.Agent, target types.TargetMapping[types.TargetTool], funcCall tools.ToolCallInvocation, opts []types.CompletionOptions) (*types.Message, error) {
	var (
		data map[string]any
	)

	if funcCall.ToolCall.Arguments != "" {
		data = make(map[string]any)
		if err := json.Unmarshal([]byte(funcCall.ToolCall.Arguments), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal function call arguments: %w", err)
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

	// Apply truncation if needed (unless disabled or this is an error)
	if !response.IsError && !agent.DisableToolTruncation {
		maxLines := DefaultMaxLines
		maxBytes := DefaultMaxBytes

		if agent.ToolOutputMaxLines != nil {
			maxLines = *agent.ToolOutputMaxLines
		}
		if agent.ToolOutputMaxBytes != nil {
			maxBytes = *agent.ToolOutputMaxBytes
		}

		truncResult, truncErr := truncateToolOutput(ctx, target.TargetName, response, maxLines, maxBytes)
		if truncErr != nil {
			// Log error but don't fail the tool call
			fmt.Fprintf(os.Stderr, "Warning: failed to truncate tool output: %v\n", truncErr)
		} else if truncResult != nil {
			response.Content = truncResult.Content
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
