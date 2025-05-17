package agents

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/openai"
	"github.com/obot-platform/nanobot/pkg/tools"
	"github.com/obot-platform/nanobot/pkg/types"
)

type Agents struct {
	config    types.Config
	completer types.Completer
	registry  *tools.Registry
}

type ToolListOptions struct {
	ToolName string
	Names    []string
}

func New(completer types.Completer, registry *tools.Registry, config types.Config) *Agents {
	return &Agents{
		config:    config,
		completer: completer,
		registry:  registry,
	}
}

func (a *Agents) addTools(ctx context.Context, req *types.Request, agent *types.Agent) (types.ToolMappings, error) {
	toolMappings, err := a.registry.BuildToolMappings(ctx, agent.Tools)
	if err != nil {
		return nil, fmt.Errorf("failed to build tool mappings: %w", err)
	}

	for _, key := range slices.Sorted(maps.Keys(toolMappings)) {
		toolMapping := toolMappings[key]

		customTool := &types.CustomTool{
			Name:        key,
			Parameters:  toolMapping.Tool.InputSchema,
			Description: toolMapping.Tool.Description,
			Attributes:  agent.ToolExtensions[key],
		}

		req.Tools = append(req.Tools, types.Tool{
			CustomTool: customTool,
		})
	}

	return toolMappings, nil
}

func (a *Agents) populateRequest(ctx context.Context, run *run, previousRun *run) (types.Request, types.ToolMappings, error) {
	req := run.Request

	req.Store = &[]bool{false}[0]

	if previousRun != nil {
		input := previousRun.PopulatedRequest.Input.GetItems()

		for _, output := range previousRun.Response.Output {
			if output.Reasoning != nil {
				// Skip reasoning output
				continue
			}
			input = append(input, output.ToInput())
		}

		for _, callID := range slices.Sorted(maps.Keys(previousRun.ToolOutputs)) {
			toolCall := previousRun.ToolOutputs[callID]
			if toolCall.Done {
				input = append(input, toolCall.Output...)
			}
		}

		input = append(input, req.Input.GetItems()...)
		req.Input = types.Input{
			Items: input,
		}
	}

	agent, ok := a.config.Agents[req.Model]
	if !ok {
		return req, nil, nil
	}

	if req.Instructions == nil && agent.Instructions != "" {
		req.Instructions = &agent.Instructions
	}

	if req.TopP == nil && agent.TopP != nil {
		req.TopP = agent.TopP
	}

	if req.Temperature == nil && agent.Temperature != nil {
		req.Temperature = agent.Temperature
	}

	if req.Truncation == nil && agent.Truncation != nil {
		req.Truncation = agent.Truncation
	}

	if req.MaxOutputTokens == nil && agent.MaxTokens != 0 {
		req.MaxOutputTokens = &agent.MaxTokens
	}

	if req.ToolChoice == nil && agent.ToolChoice != "" {
		switch agent.ToolChoice {
		case "none", "auto", "required":
			req.ToolChoice = &types.ToolChoice{
				Mode: agent.ToolChoice,
			}
		case "file_search", "web_search_preview", "computer_use_preview":
			req.ToolChoice = &types.ToolChoice{
				HostedTool: &types.HostedTool{
					Type: agent.ToolChoice,
				},
			}
		default:
			req.ToolChoice = &types.ToolChoice{
				FunctionTool: &types.FunctionTool{
					Name: agent.ToolChoice,
				},
			}
		}
	}

	if previousRun != nil {
		// Don't allow tool choice if this is a follow on request
		req.ToolChoice = nil
	}

	if req.Text == nil && agent.Output != nil && len(agent.Output.Schema) > 0 {
		name := agent.Output.Name
		if name == "" {
			name = "output_schema"
		}
		req.Text = &types.TextFormatting{
			Format: types.Format{
				JSONSchema: &types.JSONSchema{
					Name:        name,
					Description: agent.Output.Description,
					Schema:      agent.Output.Schema,
					Strict:      agent.Output.Strict,
				},
			},
		}
	}

	req.Model = agent.Model

	toolMapping, err := a.addTools(ctx, &req, &agent)
	if err != nil {
		return req, nil, fmt.Errorf("failed to add tools: %w", err)
	}

	return req, toolMapping, nil
}

const previousRunKey = "previous_run"

func (a *Agents) Complete(ctx context.Context, req types.Request, opts ...types.CompletionOptions) (*types.Response, error) {
	var (
		opt            = openai.CompleteCompletionOptions(opts...)
		previousRunKey = previousRunKey + "/" + req.Model
		session        = mcp.SessionFromContext(ctx)
		stateful       = session != nil && (opt.Continue || a.config.Agents[req.Model].Stateful)
		previousRun    *run
		currentRun     = &run{
			Request: req,
		}
	)

	if stateful {
		previousRun, _ = session.Get(previousRunKey).(*run)
	}

	for {
		if err := a.run(ctx, currentRun, previousRun, opts); err != nil {
			return nil, err
		}

		if err := a.toolCalls(ctx, currentRun); err != nil {
			return nil, err
		}

		if currentRun.Done {
			if stateful {
				session.Set(previousRunKey, currentRun)
			}
			return currentRun.Response, nil
		}

		previousRun = currentRun
		currentRun = &run{
			Request: types.Request{
				Model: currentRun.Request.Model,
			},
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
