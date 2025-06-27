package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/confirm"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/schema"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type Agents struct {
	config        types.Config
	completer     types.Completer
	registry      *tools.Service
	confirmations *confirm.Service
}

type ToolListOptions struct {
	ToolName string
	Names    []string
}

func New(completer types.Completer, registry *tools.Service, config types.Config) *Agents {
	return &Agents{
		config:        config,
		completer:     completer,
		registry:      registry,
		confirmations: confirm.New(),
	}
}

func (a *Agents) addTools(ctx context.Context, req *types.CompletionRequest, agent *types.Agent) (types.ToolMappings, error) {
	toolMappings, err := a.registry.BuildToolMappings(ctx, slices.Concat(agent.Tools, agent.Agents, agent.Flows))
	if err != nil {
		return nil, fmt.Errorf("failed to build tool mappings: %w", err)
	}

	for _, key := range slices.Sorted(maps.Keys(toolMappings)) {
		toolMapping := toolMappings[key]

		tool := toolMapping.Target.(mcp.Tool)
		req.Tools = append(req.Tools, types.ToolUseDefinition{
			Name:        key,
			Parameters:  schema.ValidateAndFixToolSchema(tool.InputSchema),
			Description: tool.Description,
			Attributes:  agent.ToolExtensions[key],
		})
	}

	return toolMappings, nil
}

func populateToolCallResult(previousRun *run, req *types.CompletionRequest, callID string) {
	if previousRun.ToolOutputs == nil {
		previousRun.ToolOutputs = make(map[string]toolCall)
	}
	var newContent []mcp.Content
	for _, input := range req.Input {
		if input.Message != nil {
			newContent = append(newContent, input.Message.Content)
		}
	}

	previousRun.ToolOutputs[callID] = toolCall{
		Output: []types.CompletionItem{
			{
				ToolCallResult: &types.ToolCallResult{
					CallID: callID,
					Output: types.CallResult{
						Content: newContent,
					},
				},
			},
		},
		Done: true,
	}
	req.InputAsToolResult = nil
	req.Input = nil
}

func (a *Agents) populateRequest(ctx context.Context, run *run, previousRun *run) (types.CompletionRequest, types.ToolMappings, error) {
	req := run.Request

	if previousRun != nil {
		input := previousRun.PopulatedRequest.Input

		for _, output := range previousRun.Response.Output {
			prevInput := output
			if prevInput.ToolCall != nil {
				if _, exists := previousRun.ToolOutputs[prevInput.ToolCall.CallID]; !exists {
					if req.InputAsToolResult != nil && *req.InputAsToolResult {
						populateToolCallResult(previousRun, &req, prevInput.ToolCall.CallID)
					} else {
						continue
					}
				}
			}
			input = append(input, prevInput)
		}

		for _, callID := range slices.Sorted(maps.Keys(previousRun.ToolOutputs)) {
			toolCall := previousRun.ToolOutputs[callID]
			if toolCall.Done {
				input = append(input, toolCall.Output...)
			}
		}

		input = append(input, req.Input...)
		req.Input = input
	}

	agent, ok := a.config.Agents[req.Model]
	if !ok {
		return req, nil, nil
	}

	if req.SystemPrompt != "" {
		var agentInstructions types.DynamicInstructions
		if err := json.Unmarshal([]byte(strings.TrimSpace(req.SystemPrompt)), &agentInstructions); err == nil &&
			agentInstructions.IsPrompt() {
			req.SystemPrompt = ""
			agent.Instructions = agentInstructions
		}
	}

	if req.SystemPrompt == "" && agent.Instructions.IsSet() {
		var err error
		req.SystemPrompt, err = a.registry.GetDynamicInstruction(ctx, agent.Instructions)
		if err != nil {
			return req, nil, err
		}
	}

	if req.TopP == nil && agent.TopP != nil {
		req.TopP = agent.TopP
	}

	if req.Temperature == nil && agent.Temperature != nil {
		req.Temperature = agent.Temperature
	}

	if req.Truncation == "" && agent.Truncation != "" {
		req.Truncation = agent.Truncation
	}

	if req.MaxTokens == 0 && agent.MaxTokens != 0 {
		req.MaxTokens = agent.MaxTokens
	}

	if req.ToolChoice == "" && agent.ToolChoice != "" {
		req.ToolChoice = agent.ToolChoice
	}

	if previousRun != nil {
		// Don't allow tool choice if this is a follow-on request
		req.ToolChoice = ""
	}

	if req.OutputSchema == nil && agent.Output != nil && len(agent.Output.ToSchema()) > 0 {
		req.OutputSchema = &types.OutputSchema{
			Name:        agent.Output.Name,
			Description: agent.Output.Description,
			Schema:      agent.Output.ToSchema(),
			Strict:      agent.Output.Strict,
		}
	}

	if req.OutputSchema != nil && req.OutputSchema.Name == "" {
		req.OutputSchema.Name = "output_schema"
	}

	if req.ThreadName == "" {
		req.ThreadName = agent.ThreadName
	}

	req.Model = agent.Model

	toolMapping, err := a.addTools(ctx, &req, &agent)
	if err != nil {
		return req, nil, fmt.Errorf("failed to add tools: %w", err)
	}

	// Validate and fix tool input schemas
	for i, tool := range req.Tools {
		fixedSchema := schema.ValidateAndFixToolSchema(tool.Parameters)
		req.Tools[i].Parameters = fixedSchema
	}

	return req, toolMapping, nil
}

const previousRunKey = "thread"

func (a *Agents) Complete(ctx context.Context, req types.CompletionRequest, opts ...types.CompletionOptions) (_ *types.CompletionResponse, err error) {
	var (
		previousRunKey = previousRunKey
		session        = mcp.SessionFromContext(ctx)
		isChat         = session != nil
		previousRun    *run
		currentRun     = &run{}
	)

	if req.ThreadName != "" {
		previousRunKey = fmt.Sprintf("%s/%s", previousRunKey, req.ThreadName)
	}

	if isChat && a.config.Agents[req.Model].Chat != nil && !*a.config.Agents[req.Model].Chat {
		isChat = false
	}

	if ch := complete.Complete(opts...).Chat; ch != nil {
		isChat = *ch
	}

	if isChat && req.InputAsToolResult == nil {
		req.InputAsToolResult = &isChat
	}

	// Save the original request to the run status
	currentRun.Request = req

	if isChat {
		var fallBack *run
		if lookup := (run{}); session.Get(previousRunKey, &lookup) {
			fallBack = &lookup
			previousRun = &lookup
		}

		if req.NewThread && previousRun != nil {
			session.Set(previousRunKey+"/"+time.Now().Format(time.RFC3339), previousRun)
			session.Set(previousRunKey, nil)
		}

		defer func() {
			if err != nil && fallBack != nil {
				session.Set(previousRunKey, fallBack)
			}
		}()
	}

	for {
		if err := a.run(ctx, currentRun, previousRun, opts); err != nil {
			return nil, err
		}

		if isChat {
			session.Set(previousRunKey, currentRun)
		}

		if err := a.toolCalls(ctx, currentRun, opts); err != nil {
			return nil, err
		}

		if isChat {
			for _, toolOutput := range currentRun.ToolOutputs {
				for _, output := range toolOutput.Output {
					if output.ToolCallResult != nil && output.ToolCallResult.Output.ChatResponse {
						return &types.CompletionResponse{
							Output: []types.CompletionItem{
								{
									ToolCallResult: output.ToolCallResult,
								},
							},
						}, nil
					}
				}
			}
		}

		if currentRun.Done {
			if isChat {
				currentRun.Response.ChatResponse = true
				session.Set(previousRunKey, currentRun)
			}
			return currentRun.Response, nil
		}

		previousRun = currentRun
		currentRun = &run{
			Request: req.Reset(),
		}
	}
}

func (a *Agents) run(ctx context.Context, run *run, prev *run, opts []types.CompletionOptions) error {
	completionRequest, toolMapping, err := a.populateRequest(ctx, run, prev)
	if err != nil {
		return err
	}

	// Save the populated request to the run status
	run.PopulatedRequest = &completionRequest
	run.ToolToMCPServer = toolMapping

	resp, err := a.completer.Complete(ctx, completionRequest, opts...)
	if err != nil {
		return err
	}

	run.Response = resp
	return nil
}
