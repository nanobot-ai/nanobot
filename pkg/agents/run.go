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
	"github.com/nanobot-ai/nanobot/pkg/llm/progress"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/schema"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

type Agents struct {
	completer types.Completer
	registry  *tools.Service
}

type ToolListOptions struct {
	ToolName string
	Names    []string
}

func New(completer types.Completer, registry *tools.Service) *Agents {
	return &Agents{
		completer: completer,
		registry:  registry,
	}
}

func (a *Agents) addTools(ctx context.Context, req *types.CompletionRequest, agent *types.Agent) (types.ToolMappings, error) {
	toolMappings, err := a.registry.BuildToolMappings(ctx, slices.Concat(agent.Tools, agent.Agents, agent.Flows, agent.MCPServers))
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
	req.Reasoning = agent.Reasoning

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

func (a *Agents) replacePrompt(ctx context.Context, agentConfig types.Agent, items []types.CompletionItem) (result []types.CompletionItem, messages []mcp.PromptMessage, err error) {
	if len(items) != 1 || items[0].Content == nil || items[0].Content.Type != "text" {
		return items, nil, nil
	}

	var uiAction types.UIAction
	if err := json.Unmarshal([]byte(items[0].Content.Text), &uiAction); err != nil {
		// Ignore invalid JSON
		return items, nil, nil
	}

	if uiAction.Prompt != nil && uiAction.Prompt.Prompt != "" {
		newItem := items[0]
		newItem.Content = &mcp.Content{
			Type: "text",
			Text: uiAction.Prompt.Prompt,
		}
		return []types.CompletionItem{newItem}, nil, nil
	} else if uiAction.Prompt != nil && len(uiAction.Prompt.RenderedMessages) > 0 {
		return items, uiAction.Prompt.RenderedMessages, nil
	} else if uiAction.Prompt != nil && uiAction.Prompt.PromptName != "" {
		prompts, err := sessiondata.NewData(a.registry).BuildPromptMappings(ctx, slices.Concat(agentConfig.MCPServers, agentConfig.Prompts)...)
		if err != nil {
			return items, nil, nil
		}
		prompt, ok := prompts[uiAction.Prompt.PromptName]
		if !ok {
			return nil, nil, fmt.Errorf("prompt %q not found", uiAction.Prompt.PromptName)
		}

		messages, err := a.registry.GetPrompt(ctx, prompt.MCPServer, prompt.TargetName, uiAction.Prompt.Params)
		if err != nil {
			return items, nil, err
		}

		uiAction.Prompt.RenderedMessages = messages.Messages
		newText, err := json.Marshal(uiAction)
		if err != nil {
			return items, nil, fmt.Errorf("failed to marshal UIAction: %w", err)
		}

		// purposefully modify the original message so it gets saved, so we don't call on each turn
		items[0].Content.Text = string(newText)

		return items, messages.Messages, err
	}

	return items, nil, nil
}

func (a *Agents) handleUIAction(ctx context.Context, config types.Config, req types.CompletionRequest, opts []types.CompletionOptions) (types.CompletionRequest, *types.CompletionResponse, error) {
	var (
		uiAction    types.UIAction
		err         error
		lastMessage types.Message
		opt         = complete.Complete(opts...)
	)

	if len(req.Input) == 0 {
		return req, nil, nil
	}

	// copy
	newReq := req
	newReq.Input = nil
	for _, msg := range req.Input {
		if msg.Role != "user" {
			newReq.Input = append(newReq.Input, msg)
			continue
		}
		newItems, promptMessages, err := a.replacePrompt(ctx, config.Agents[req.Agent], msg.Items)
		if err != nil {
			return req, nil, fmt.Errorf("failed to replace prompt: %w", err)
		}
		if len(promptMessages) > 0 {
			for i, promptMessage := range promptMessages {
				newMessage := types.Message{
					ID:      fmt.Sprintf("%s-prompt-%d", msg.ID, i),
					Created: msg.Created,
					Role:    promptMessage.Role,
					Items: []types.CompletionItem{
						{
							ID:      fmt.Sprintf("%s-prompt-item-%d", msg.ID, i),
							Content: &promptMessage.Content,
						},
					},
				}
				if newMessage.Role == "" {
					newMessage.Role = "user"
				}
				// Reset id's for first element to match original message id
				if i == 0 {
					newMessage.ID = msg.ID
					if len(msg.Items) == 1 {
						newMessage.Items[0].ID = msg.Items[0].ID
					}
				}

				newReq.Input = append(newReq.Input, newMessage)
			}
		} else {
			msg.Items = newItems
			newReq.Input = append(newReq.Input, msg)
		}

	}

	lastMessage = newReq.Input[len(req.Input)-1]
	if lastMessage.Role != "user" {
		return newReq, nil, nil
	}

	if len(lastMessage.Items) != 1 || lastMessage.Items[0].Content == nil || lastMessage.Items[0].Content.Type != "text" {
		return newReq, nil, nil
	}

	if err := json.Unmarshal([]byte(lastMessage.Items[0].Content.Text), &uiAction); err != nil {
		// Ignore invalid JSON
		return newReq, nil, nil
	}

	if uiAction.Tool != nil && uiAction.Tool.ToolName != "" {
		args := []byte("{}")
		if len(uiAction.Tool.Params) > 0 {
			args, err = json.Marshal(uiAction.Tool.Params)
			if err != nil {
				return newReq, nil, err
			}
		}
		resp := &types.CompletionResponse{
			Output: types.Message{
				ID:   uuid.String(),
				Role: "assistant",
				Items: []types.CompletionItem{
					{
						ID: "fc_" + uuid.String(),
						ToolCall: &types.ToolCall{
							Arguments: string(args),
							CallID:    uuid.String(),
							Name:      uiAction.Tool.ToolName,
						},
					},
				},
			},
		}

		if opt.ProgressToken != nil {
			progress.Send(ctx, &types.CompletionProgress{
				MessageID: resp.Output.ID,
				Role:      "assistant",
				Item:      resp.Output.Items[0],
			}, opt.ProgressToken)
		}

		return newReq, resp, nil
	}

	return newReq, nil, nil
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

	for session != nil && session.Parent != nil {
		session = session.Parent
	}

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

			finalResponse := *currentRun.Response

			if startID != "" && currentRun.PopulatedRequest != nil {
				i := slices.IndexFunc(currentRun.PopulatedRequest.Input, func(msg types.Message) bool {
					return msg.ID == startID
				})
				if i >= 0 {
					finalResponse.InternalMessages = types.ConsolidateTools(currentRun.PopulatedRequest.Input[i:])
				}
			}

			return &finalResponse, nil
		}

		previousRun = currentRun
		currentRun = &types.Execution{
			Request: req.Reset(),
		}
	}
}

func (a *Agents) runBefore(ctx context.Context, config types.Config, req types.CompletionRequest) (types.CompletionRequest, *types.CompletionResponse, error) {
	return req, nil, nil
}

func (a *Agents) runAfter(ctx context.Context, config types.Config, req types.CompletionRequest, resp *types.CompletionResponse) (*types.CompletionResponse, error) {
	return resp, nil
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

	run.ToolToMCPServer = allToolMappings

	completionRequest, resp, err := a.runBefore(ctx, config, completionRequest)
	if err != nil {
		return fmt.Errorf("failed to run before agent: %w", err)
	} else if resp != nil {
		run.PopulatedRequest = &completionRequest
		run.Response = resp
		return nil
	}

	run.PopulatedRequest = &completionRequest

	modifiedRequest, resp, err := a.handleUIAction(ctx, config, completionRequest, opts)
	if err != nil {
		return fmt.Errorf("failed to handle UI action: %w", err)
	} else if resp != nil {
		run.Response = resp
		return nil
	}

	resp, err = a.completer.Complete(ctx, modifiedRequest, opts...)
	if err != nil {
		return err
	}

	resp, err = a.runAfter(ctx, config, completionRequest, resp)
	if err != nil {
		return fmt.Errorf("failed to run after agent: %w", err)
	}

	run.Response = resp
	return nil
}
