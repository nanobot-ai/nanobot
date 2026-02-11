package agents

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

const (
	transcriptContentLimit = 4_000
)

func (a *Agents) runCompaction(ctx context.Context, config types.Config, prev *types.Execution, run *types.Execution, req *types.CompletionRequest) (bool, error) {
	if len(req.Input) == 0 {
		return false, nil
	}

	agentName := config.Compaction.AgentName()
	agent, ok := config.Agents[agentName]
	if !ok {
		return false, fmt.Errorf("compaction agent %q not defined", agentName)
	}

	retain := config.Compaction.RecentCount()
	if retain < 1 {
		retain = 1
	}

	if retain >= len(req.Input) {
		retain = len(req.Input) - 1
	}

	if retain < 1 {
		return false, nil
	}

	totalMessages := len(req.Input)
	runModel := ""
	if run != nil && run.PopulatedRequest != nil {
		runModel = run.PopulatedRequest.Model
	}
	if runModel == "" {
		runModel = req.Model
	}

	log.Infof(ctx, "starting compaction: model=%s totalMessages=%d retainRecent=%d", runModel, totalMessages, retain)

	splitIdx := len(req.Input) - retain
	// Adjust split forward to avoid orphaning tool results whose tool calls
	// would end up in the summarized older portion.
	splitIdx = adjustSplitForToolPairs(req.Input, splitIdx)

	older := req.Input[:splitIdx]
	recent := req.Input[splitIdx:]

	if len(older) == 0 {
		return false, nil
	}

	olderTranscript := renderMessages(older)
	if strings.TrimSpace(olderTranscript) == "" {
		return false, nil
	}

	recentTranscript := renderMessages(recent)

	systemPrompt := ""
	if agent.Instructions.IsSet() {
		prompt, err := a.registry.GetDynamicInstruction(ctx, agent.Instructions)
		if err != nil {
			return false, fmt.Errorf("failed to resolve compaction instructions: %w", err)
		}
		systemPrompt = strings.TrimSpace(prompt)
	} else if req.SystemPrompt != "" {
		systemPrompt = strings.TrimSpace(req.SystemPrompt)
	}

	userContent := buildCompactionPrompt(olderTranscript, recentTranscript)

	summaryReq := types.CompletionRequest{
		Agent:        agentName,
		Model:        agent.Model,
		SystemPrompt: systemPrompt,
		MaxTokens:    agent.MaxTokens,
		Temperature:  agent.Temperature,
		TopP:         agent.TopP,
		Input: []types.Message{
			{
				Role: "user",
				Items: []types.CompletionItem{
					{
						ID: uuid.String(),
						Content: &mcp.Content{
							Type: "text",
							Text: userContent,
						},
					},
				},
			},
		},
	}

	summaryResp, err := a.completer.Complete(ctx, summaryReq)
	if err != nil {
		return false, fmt.Errorf("failed to run compaction agent: %w", err)
	}

	summaryText := collectSummaryText(summaryResp)
	if summaryText == "" {
		summaryText = "(compaction summary unavailable)"
	}

	now := time.Now()
	summaryMessage := types.Message{
		ID:      "summary-" + uuid.String(),
		Created: &now,
		Role:    "user",
		Items: []types.CompletionItem{
			{
				ID: "summary-item-" + uuid.String(),
				Content: &mcp.Content{
					Type: "text",
					Text: "[Previous conversation summary]\n\n" + summaryText,
				},
			},
		},
	}

	newHistory := append([]types.Message{summaryMessage}, recent...)

	req.Input = newHistory
	if run.PopulatedRequest != nil {
		run.PopulatedRequest.Input = newHistory
	}

	if prev != nil && prev.PopulatedRequest != nil {
		prev.PopulatedRequest.Input = newHistory
		prev.ToolOutputs = nil
	}

	if summaryResp.Usage != nil {
		run.Usage = types.MergeUsage(run.Usage, *summaryResp.Usage)
	}

	log.Infof(ctx, "compaction complete: model=%s summarizedMessages=%d retainedRecent=%d", runModel, len(older), len(recent))

	return true, nil
}

func buildCompactionPrompt(olderTranscript, recentTranscript string) string {
	var builder strings.Builder
	builder.WriteString("### History To Compact\n")
	builder.WriteString(strings.TrimSpace(olderTranscript))
	builder.WriteString("\n\n")
	builder.WriteString("### Recent Messages (retain verbatim)\n")
	builder.WriteString(strings.TrimSpace(recentTranscript))
	return builder.String()
}

func collectSummaryText(resp *types.CompletionResponse) string {
	if resp == nil {
		return ""
	}

	var parts []string
	for _, item := range resp.Output.Items {
		if item.Content != nil && item.Content.Text != "" {
			parts = append(parts, strings.TrimSpace(item.Content.Text))
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func renderMessages(messages []types.Message) string {
	var builder strings.Builder
	for _, msg := range messages {
		builder.WriteString("[")
		builder.WriteString(msg.Role)
		builder.WriteString("] ")
		builder.WriteString(renderItems(msg.Items))
		builder.WriteString("\n")
	}
	return builder.String()
}

func renderItems(items []types.CompletionItem) string {
	var parts []string
	for _, item := range items {
		switch {
		case item.Content != nil:
			parts = append(parts, truncateText(flattenContent(*item.Content)))
		case item.ToolCall != nil:
			parts = append(parts, fmt.Sprintf("tool call %s args=%s", item.ToolCall.Name, truncateText(item.ToolCall.Arguments)))
		case item.ToolCallResult != nil:
			parts = append(parts, truncateText(flattenCallResult(item.ToolCallResult.Output)))
		case item.Reasoning != nil:
			parts = append(parts, truncateText(flattenReasoning(item.Reasoning)))
		}
	}
	return strings.Join(parts, " ")
}

func flattenCallResult(result types.CallResult) string {
	var parts []string
	for _, content := range result.Content {
		parts = append(parts, flattenContent(content))
	}
	if result.IsError {
		parts = append(parts, "(tool reported an error)")
	}
	return strings.Join(parts, " ")
}

func flattenReasoning(reasoning *types.Reasoning) string {
	if reasoning == nil {
		return ""
	}
	var parts []string
	for _, summary := range reasoning.Summary {
		if summary.Text != "" {
			parts = append(parts, summary.Text)
		}
	}
	return strings.Join(parts, " ")
}

func flattenContent(content mcp.Content) string {
	var parts []string
	if content.Text != "" {
		parts = append(parts, content.Text)
	}
	if content.Resource != nil {
		if content.Resource.Text != "" {
			parts = append(parts, content.Resource.Text)
		}
		if content.Resource.URI != "" {
			parts = append(parts, fmt.Sprintf("(resource: %s)", content.Resource.URI))
		}
	}
	if len(content.StructuredContent) > 0 {
		parts = append(parts, "[structured content]")
	}
	for _, child := range content.Content {
		parts = append(parts, flattenContent(child))
	}
	if content.ToolUseID != "" {
		parts = append(parts, fmt.Sprintf("(tool-use-id %s)", content.ToolUseID))
	}
	return strings.Join(parts, " ")
}

// adjustSplitForToolPairs moves the split index forward so that tool result
// messages are not separated from their corresponding tool call messages.
// It collects all tool call IDs present in messages[splitIdx:] and, if any
// message at the start of that range has tool results referencing calls not
// in the retained set, it moves the boundary forward past those messages.
func adjustSplitForToolPairs(messages []types.Message, splitIdx int) int {
	if splitIdx >= len(messages) {
		return splitIdx
	}

	// Collect tool call IDs present in the recent window.
	toolCallIDs := map[string]bool{}
	for _, msg := range messages[splitIdx:] {
		for _, item := range msg.Items {
			if item.ToolCall != nil && item.ToolCall.CallID != "" {
				toolCallIDs[item.ToolCall.CallID] = true
			}
		}
	}

	// Walk forward from splitIdx: if a message has tool results referencing
	// calls not in the recent window, move it into the older (summarized) set.
	for splitIdx < len(messages)-1 {
		msg := messages[splitIdx]
		hasOrphan := false
		for _, item := range msg.Items {
			if item.ToolCallResult != nil && item.ToolCallResult.CallID != "" {
				if !toolCallIDs[item.ToolCallResult.CallID] {
					hasOrphan = true
					break
				}
			}
		}
		if !hasOrphan {
			break
		}
		splitIdx++
	}

	return splitIdx
}

func truncateText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return text
	}
	if utf8.RuneCountInString(text) <= transcriptContentLimit {
		return text
	}
	runes := []rune(text)
	return string(runes[:transcriptContentLimit]) + " …"
}
