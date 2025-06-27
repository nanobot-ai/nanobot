package sampling

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type Sampler struct {
	completer types.Completer
}

func NewSampler(completer types.Completer) *Sampler {
	return &Sampler{
		completer: completer,
	}
}

type scored struct {
	score float64
	model string
}

func (s *Sampler) sortModels(config types.Config, preferences mcp.ModelPreferences) []string {
	var scoredModels []scored

	for _, modelKey := range slices.Sorted(maps.Keys(config.Agents)) {
		model := config.Agents[modelKey]
		cost := model.Cost
		if preferences.CostPriority != nil {
			cost *= *preferences.CostPriority
		}
		speed := model.Speed
		if preferences.SpeedPriority != nil {
			speed *= *preferences.SpeedPriority
		}
		intelligence := model.Intelligence
		if preferences.IntelligencePriority != nil {
			intelligence *= *preferences.IntelligencePriority
		}
		scoredModels = append(scoredModels, scored{
			score: cost + speed + intelligence,
			model: modelKey,
		})
	}

	sort.Slice(scoredModels, func(i, j int) bool {
		return scoredModels[i].score > scoredModels[j].score
	})

	models := make([]string, len(scoredModels))
	for i, scoredModel := range scoredModels {
		models[i] = scoredModel.model
	}
	return models
}

func (s *Sampler) getMatchingModel(config types.Config, req *mcp.CreateMessageRequest) (string, bool) {
	// Agent by name
	for _, model := range req.ModelPreferences.Hints {
		if _, ok := config.Agents[model.Name]; ok {
			return model.Name, true
		}
	}

	// Model by alias
	for _, model := range req.ModelPreferences.Hints {
		for _, modelKey := range slices.Sorted(maps.Keys(config.Agents)) {
			if slices.Contains(config.Agents[modelKey].Aliases, model.Name) {
				return modelKey, true
			}
		}
	}

	models := s.sortModels(config, req.ModelPreferences)
	if len(models) == 0 {
		return "", false
	}

	return models[0], true
}

type SamplerOptions struct {
	ProgressToken any
	Continue      bool
	AgentOverride types.AgentCall
}

func (s SamplerOptions) Merge(other SamplerOptions) (result SamplerOptions) {
	result.ProgressToken = complete.Last(s.ProgressToken, other.ProgressToken)
	result.Continue = complete.Last(s.Continue, other.Continue)
	result.AgentOverride = complete.Merge(s.AgentOverride, other.AgentOverride)
	return
}

func (s *Sampler) Sample(ctx context.Context, req mcp.CreateMessageRequest, opts ...SamplerOptions) (result *types.CallResult, _ error) {
	opt := complete.Complete(opts...)
	config := types.ConfigFromContext(ctx)

	model, ok := s.getMatchingModel(config, &req)
	if !ok {
		return nil, fmt.Errorf("no matching model found")
	}

	request := types.CompletionRequest{
		Model:             model,
		ToolChoice:        opt.AgentOverride.ToolChoice,
		OutputSchema:      opt.AgentOverride.Output,
		Temperature:       opt.AgentOverride.Temperature,
		TopP:              opt.AgentOverride.TopP,
		NewThread:         opt.AgentOverride.NewThread != nil && *opt.AgentOverride.NewThread,
		InputAsToolResult: opt.AgentOverride.InputAsToolResult,
	}

	if req.MaxTokens != 0 {
		request.MaxTokens = req.MaxTokens
	}
	if req.SystemPrompt != "" {
		request.SystemPrompt = req.SystemPrompt
	}
	if req.Temperature != nil {
		request.Temperature = req.Temperature
	}

	for _, content := range req.Messages {
		request.Input = append(request.Input, types.CompletionItem{
			Message: &content,
		})
	}

	completeOptions := types.CompletionOptions{
		Chat:          opt.AgentOverride.Chat,
		ProgressToken: opt.ProgressToken,
	}

	resp, err := s.completer.Complete(ctx, request, completeOptions)
	if err != nil {
		return nil, err
	}

	result = &types.CallResult{
		Model:        resp.Model,
		ChatResponse: resp.ChatResponse,
	}

	if _, ok := config.Agents[request.Model]; ok {
		result.Agent = request.Model
	}

	for _, output := range resp.Output {
		if output.ToolCallResult != nil {
			return &output.ToolCallResult.Output, nil
		}
		if output.Message == nil {
			continue
		}
		result.Content = append(result.Content, output.Message.Content)
	}

	if len(result.Content) == 0 {
		result.Content = append(result.Content, mcp.Content{
			Type: "text",
			Text: "[NO CONTENT]",
		})
	}

	return result, nil
}
