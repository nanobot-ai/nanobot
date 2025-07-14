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
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

type Agents struct {
	completer     types.Completer
	registry      *tools.Service
	confirmations *confirm.Service
}

type ToolListOptions struct {
	ToolName string
	Names    []string
}

func New(completer types.Completer, registry *tools.Service) *Agents {
	return &Agents{
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

		tool := toolMapping.Target
		req.Tools = append(req.Tools, types.ToolUseDefinition{
			Name:        key,
			Parameters:  schema.ValidateAndFixToolSchema(tool.InputSchema),
			Description: tool.Description,
			Attributes:  agent.ToolExtensions[toolMapping.Target.Name],
		})
	}

	return toolMappings, nil
}

func populateToolCallResult(previousRun *types.Execution, req *types.CompletionRequest, callID string) {
	if previousRun.ToolOutputs == nil {
		previousRun.ToolOutputs = make(map[string]types.ToolOutput)
	}
	var newContent []mcp.Content
	for _, input := range req.Input {
		for _, item := range input.Items {
			if item.Content != nil {
				newContent = append(newContent, *item.Content)
			}
		}
	}

	previousRun.ToolOutputs[callID] = types.ToolOutput{
		Output: types.Message{
			Role: "user",
			Items: []types.CompletionItem{
				{
					ToolCallResult: &types.ToolCallResult{
						CallID: callID,
						Output: types.CallResult{
							Content: newContent,
						},
					},
				},
			},
		},
		Done: true,
	}
	req.InputAsToolResult = nil
	req.Input = nil
}

func (a *Agents) populateRequest(ctx context.Context, config types.Config, run *types.Execution, previousRun *types.Execution) (types.CompletionRequest, types.ToolMappings, error) {
	req := run.Request

	if previousRun != nil {
		input := previousRun.PopulatedRequest.Input

		for _, outputMessage := range append(previousRun.Response.InternalMessages, previousRun.Response.Output) {
			newItems := make([]types.CompletionItem, 0, len(outputMessage.Items))
			for _, output := range outputMessage.Items {
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
				newItems = append(newItems, prevInput)
			}

			if len(newItems) > 0 {
				outputMessage.Items = newItems
				input = append(input, outputMessage)
			}
		}

		for _, callID := range slices.Sorted(maps.Keys(previousRun.ToolOutputs)) {
			toolCall := previousRun.ToolOutputs[callID]
			if toolCall.Done {
				input = append(input, toolCall.Output)
			}
		}

		input = append(input, req.Input...)
		req.Input = input
	}

	agentName := req.Agent
	if agentName == "" {
		agentName = req.Model
	}

	agent, ok := config.Agents[agentName]
	if !ok {
		return req, nil, nil
	}

	req.Agent = agentName

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

func (a *Agents) Complete(ctx context.Context, req types.CompletionRequest, opts ...types.CompletionOptions) (_ *types.CompletionResponse, err error) {
	var (
		previousExecutionKey = types.PreviousExecutionKey
		session              = mcp.SessionFromContext(ctx)
		isChat               = session != nil
		previousRun          *types.Execution
		currentRun           = &types.Execution{}
		config               = types.ConfigFromContext(ctx)
		startID              = ""
	)

	if len(req.Input) > 0 {
		startID = req.Input[0].ID
		if startID == "" {
			startID = uuid.String()
			req.Input[0].ID = startID
		}
	}

	if req.ThreadName != "" {
		previousExecutionKey = fmt.Sprintf("%s/%s", previousExecutionKey, req.ThreadName)
	}

	if isChat && config.Agents[req.Model].Chat != nil && !*config.Agents[req.Model].Chat {
		isChat = false
	}

	if ch := complete.Complete(opts...).Chat; ch != nil {
		isChat = *ch
	}

	if isChat && req.InputAsToolResult == nil {
		req.InputAsToolResult = &isChat
	}

	// Save the original request to the Execution status
	currentRun.Request = req

	if isChat {
		var fallBack *types.Execution
		if lookup := (types.Execution{}); session.Get(previousExecutionKey, &lookup) {
			fallBack = &lookup
			previousRun = &lookup
		}

		if req.NewThread && previousRun != nil {
			session.Set(previousExecutionKey+"/"+time.Now().Format(time.RFC3339), previousRun)
			session.Set(previousExecutionKey, nil)
		}

		defer func() {
			if err != nil && fallBack != nil {
				session.Set(previousExecutionKey, fallBack)
			}
		}()
	}

	for {
		if err := a.run(ctx, config, currentRun, previousRun, opts); err != nil {
			return nil, err
		}

		if isChat {
			session.Set(previousExecutionKey, currentRun)
		}

		if err := a.toolCalls(ctx, config, currentRun, opts); err != nil {
			return nil, err
		}

		if isChat {
			for _, toolOutput := range currentRun.ToolOutputs {
				for _, output := range toolOutput.Output.Items {
					if output.ToolCallResult != nil && output.ToolCallResult.Output.ChatResponse {
						return &types.CompletionResponse{
							Output: types.Message{
								Items: []types.CompletionItem{
									{
										ToolCallResult: output.ToolCallResult,
									},
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
				session.Set(previousExecutionKey, currentRun)
			}

			finalResponse := currentRun.Response

			if startID != "" && currentRun.PopulatedRequest != nil {
				i := slices.IndexFunc(currentRun.PopulatedRequest.Input, func(msg types.Message) bool {
					return msg.ID == startID
				})
				if i >= 0 {
					finalResponse.InternalMessages = currentRun.PopulatedRequest.Input[i:]
				}
			}

			return finalResponse, nil
		}

		previousRun = currentRun
		currentRun = &types.Execution{
			Request: req.Reset(),
		}
	}
}

func (a *Agents) run(ctx context.Context, config types.Config, run *types.Execution, prev *types.Execution, opts []types.CompletionOptions) error {
	completionRequest, toolMapping, err := a.populateRequest(ctx, config, run, prev)
	if err != nil {
		return err
	}

	// Don't forget about old tools that might not be in use anymore. If the old name mapped to a
	// different tool we will have a problem but, oh well?
	allToolMappings := types.ToolMappings{}
	if prev != nil {
		maps.Copy(allToolMappings, prev.ToolToMCPServer)
	}
	maps.Copy(allToolMappings, toolMapping)

	// Save the populated request to the Execution status
	run.PopulatedRequest = &completionRequest
	run.ToolToMCPServer = allToolMappings

	resp, err := a.completer.Complete(ctx, completionRequest, opts...)
	if err != nil {
		return err
	}

	run.Response = resp
	return nil
}
