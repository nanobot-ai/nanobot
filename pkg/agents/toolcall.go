package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/contextguard"
	"github.com/nanobot-ai/nanobot/pkg/log"
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

		if targetServer.Target.External {
			// Handled externally, so terminate the run waiting for the client
			run.Done = true
			continue
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

		if a.guardAfterTool(ctx, config, run) {
			run.PendingCompaction = true
			break
		}
	}

	if len(run.ToolOutputs) == 0 {
		run.Done = true
	}

	return nil
}

func (a *Agents) invoke(ctx context.Context, config types.Config, target types.TargetMapping[types.TargetTool], funcCall tools.ToolCallInvocation, opts []types.CompletionOptions) (*types.Message, error) {
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
	if response != nil {
		response = a.truncateToolResult(funcCall.ToolCall.CallID, response)
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

func (a *Agents) guardAfterTool(ctx context.Context, config types.Config, run *types.Execution) bool {
	if run == nil || run.Response == nil || run.PopulatedRequest == nil {
		return false
	}

	model := run.Response.Model
	if model == "" {
		model = run.PopulatedRequest.Model
	}

	messages := make([]types.Message, 0, len(run.PopulatedRequest.Input)+1+len(run.ToolOutputs))
	messages = append(messages, run.PopulatedRequest.Input...)
	messages = append(messages, run.Response.Output)

	if len(run.ToolOutputs) > 0 {
		for _, callID := range slices.Sorted(maps.Keys(run.ToolOutputs)) {
			toolCall := run.ToolOutputs[callID]
			if toolCall.Done {
				messages = append(messages, toolCall.Output)
			}
		}
	}

	guard := contextguard.NewService(contextguard.Config{WarnThreshold: config.Compaction.EffectiveGuardThreshold()})
	result := guard.Evaluate(contextguard.State{
		Model:        model,
		SystemPrompt: run.PopulatedRequest.SystemPrompt,
		Tools:        run.PopulatedRequest.Tools,
		Messages:     messages,
	})

	switch result.Status {
	case contextguard.StatusNeedsCompaction, contextguard.StatusOverLimit:
		log.Infof(ctx, "context guard triggered after tool: status=%s model=%s inputTokens=%d usable=%d", result.Status, model, result.Totals.InputTokens, result.Limits.Usable)
		return true
	default:
		return false
	}
}
