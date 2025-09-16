package sampling

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

var ErrNoMatchingModel = fmt.Errorf("no matching model found")

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
		return nil, ErrNoMatchingModel
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

	var currentRole string
	for _, content := range req.Messages {
		role := content.Role
		if role == "" {
			role = "user"
		}

		var id string
		if role != currentRole {
			if opt.ProgressToken != nil {
				id = fmt.Sprint(opt.ProgressToken)
			}
			if id != "" && len(request.Input) > 0 {
				id = fmt.Sprintf("%s-%d", id, len(request.Input))
			}
			if id == "" {
				id = uuid.String()
			}
			now := time.Now()
			request.Input = append(request.Input, types.Message{
				ID:      id,
				Created: &now,
				Role:    role,
			})
			currentRole = role
		}

		inputIndex := len(request.Input) - 1
		itemsLength := len(request.Input[inputIndex].Items)
		request.Input[inputIndex].Items = append(request.Input[inputIndex].Items, types.CompletionItem{
			ID:      fmt.Sprintf("%s_%d", id, itemsLength),
			Content: &content.Content,
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

	if _, ok := config.Agents[request.Model]; ok {
		resp.Agent = request.Model
	}

	result = &types.CallResult{
		Model:        resp.Model,
		ChatResponse: resp.ChatResponse,
	}

	if _, ok := config.Agents[request.Model]; ok {
		result.Agent = request.Model
	}

	return CompletionResponseToCallResult(resp, false)
}

func CompletionResponseToCallResult(resp *types.CompletionResponse, includeMessages bool) (*types.CallResult, error) {
	result := &types.CallResult{
		Model:        resp.Model,
		ChatResponse: resp.ChatResponse,
		IsError:      resp.Error != "",
	}

	for _, output := range resp.Output.Items {
		if output.ToolCallResult != nil {
			cp := output.ToolCallResult.Output
			result = &cp
			break
		}
		if output.Content == nil {
			continue
		}
		result.Content = append(result.Content, *output.Content)
	}

	if resp.Error != "" {
		result.IsError = true
		result.Content = append(result.Content, mcp.Content{
			Type: "text",
			Text: resp.Error,
		})
	}

	if len(result.Content) == 0 {
		result.Content = append(result.Content, mcp.Content{
			Type: "text",
			Text: "[NO CONTENT]",
		})
	}

	if includeMessages {
		outputMessages := append(resp.InternalMessages, resp.Output)
		if resp.Error != "" {
			outputMessages = append(outputMessages, types.Message{
				ID:   fmt.Sprintf("%x", sha256.Sum256([]byte(resp.Error)))[:12],
				Role: "assistant",
				Items: []types.CompletionItem{
					{
						Content: &mcp.Content{
							Type: "resource",
							Resource: &mcp.EmbeddedResource{
								MIMEType: types.ErrorMimeType,
								Text:     resp.Error,
							},
						},
					},
				},
			})
		}

		for _, msg := range outputMessages {
			if len(msg.Items) == 0 {
				// Empty message, typically happens in there is an error
				continue
			}
			textData, err := json.Marshal(msg)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal message: %w", err)
			}

			result.Content = append(result.Content, mcp.Content{
				Type: "resource",
				Resource: &mcp.EmbeddedResource{
					URI:      fmt.Sprintf(types.MessageURI, msg.ID),
					MIMEType: types.MessageMimeType,
					Text:     string(textData),
				},
			})
		}
	} else {
		for _, msg := range append(resp.InternalMessages, resp.Output) {
			for _, item := range msg.Items {
				if item.ToolCallResult != nil {
					for _, content := range item.ToolCallResult.Output.Content {
						if strings.HasPrefix(content.MIMEType, "ui://") {
							result.Content = append(result.Content, content)
						}
					}
				}
			}
		}
	}

	return result, nil
}
