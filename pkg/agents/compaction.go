package agents

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

const (
	compactionSummaryPrefix = "## Compacted Context\n"

	clearedContentPlaceholder = "[content cleared during context compaction]"

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

	runModel := ""
	if run != nil && run.PopulatedRequest != nil {
		runModel = run.PopulatedRequest.Model
	}
	if runModel == "" {
		runModel = req.Model
	}

	splitIdx := len(req.Input) - retain
	older := req.Input[:splitIdx]
	recent := req.Input[splitIdx:]

	if len(older) == 0 {
		return false, nil
	}

	// Render the older messages into a transcript BEFORE clearing their content.
	olderTranscript := renderMessages(older)
	if strings.TrimSpace(olderTranscript) == "" {
		return false, nil
	}

	log.Infof(ctx, "starting compaction: model=%s totalMessages=%d olderMessages=%d retainRecent=%d",
		runModel, len(req.Input), len(older), len(recent))

	// Extract any existing summary from a previous compaction message so the
	// summarizer can build on it rather than starting from scratch.
	existingSummary := extractCompactionSummary(req.Input)

	// Build the summarizer prompt.
	recentTranscript := renderMessages(recent)
	userContent := buildCompactionPrompt(existingSummary, olderTranscript, recentTranscript)

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

	// Step 1: Clear content from older messages in-place. This preserves the
	// message structure (roles, tool call/result pairing) while reclaiming tokens.
	for i := range older {
		older[i] = clearMessageContent(older[i])
	}

	// Step 2: Insert a summary message between the cleared older messages and
	// the recent messages so the model has context about what was cleared.
	summaryMessage := types.Message{
		ID:   "compaction-summary-" + uuid.String(),
		Role: "user",
		Items: []types.CompletionItem{
			{
				ID: uuid.String(),
				Content: &mcp.Content{
					Type: "text",
					Text: "## Compacted Context\nThe conversation history above has been compacted to manage context length. " +
						"Earlier messages have had their content cleared. Here is a summary of what happened:\n\n" +
						summaryText,
				},
			},
		},
	}

	// Rebuild: summary + cleared older + recent.
	// The summary goes first to avoid breaking tool call/result pairing
	// at the older/recent boundary.
	newInput := make([]types.Message, 0, 1+len(older)+len(recent))
	newInput = append(newInput, summaryMessage)
	newInput = append(newInput, older...)
	newInput = append(newInput, recent...)
	req.Input = newInput

	// Propagate changes to the execution state.
	if run.PopulatedRequest != nil {
		run.PopulatedRequest.Input = req.Input
	}
	if prev != nil && prev.PopulatedRequest != nil {
		prev.PopulatedRequest.Input = req.Input
		prev.ToolOutputs = nil
	}

	if summaryResp.Usage != nil {
		run.Usage = types.MergeUsage(run.Usage, *summaryResp.Usage)
	}

	log.Infof(ctx, "compaction complete: model=%s clearedMessages=%d retainedRecent=%d",
		runModel, len(older), len(recent))

	return true, nil
}

// clearMessageContent replaces the content of a message with short placeholders
// while preserving the message structure (role, IDs, tool call/result pairing).
func clearMessageContent(msg types.Message) types.Message {
	newItems := make([]types.CompletionItem, len(msg.Items))
	for i, item := range msg.Items {
		newItems[i] = clearItemContent(item)
	}
	msg.Items = newItems
	return msg
}

func clearItemContent(item types.CompletionItem) types.CompletionItem {
	if item.Content != nil {
		item.Content = &mcp.Content{
			Type: "text",
			Text: clearedContentPlaceholder,
		}
	}
	if item.ToolCallResult != nil {
		item.ToolCallResult = &types.ToolCallResult{
			CallID: item.ToolCallResult.CallID,
			Output: types.CallResult{
				Content: []mcp.Content{
					{Type: "text", Text: clearedContentPlaceholder},
				},
			},
		}
	}
	if item.Reasoning != nil {
		item.Reasoning = nil
	}
	// Keep ToolCall as-is — it's small (just name + args) and needed for pairing.
	return item
}

// extractCompactionSummary finds the most recent compaction summary message
// in the message list. This allows subsequent compactions to build on the
// previous summary rather than starting from scratch.
func extractCompactionSummary(messages []types.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != "user" || len(msg.Items) != 1 {
			continue
		}
		item := msg.Items[0]
		if item.Content == nil {
			continue
		}
		if strings.HasPrefix(item.Content.Text, compactionSummaryPrefix) {
			return strings.TrimSpace(item.Content.Text[len(compactionSummaryPrefix):])
		}
	}
	return ""
}

func buildCompactionPrompt(existingSummary, olderTranscript, recentTranscript string) string {
	var builder strings.Builder

	if existingSummary != "" {
		builder.WriteString("### Previous Compaction Summary\nThis conversation was previously compacted. Here is the existing summary to build upon:\n\n")
		builder.WriteString(existingSummary)
		builder.WriteString("\n\n")
	}

	builder.WriteString("### History To Compact\n")
	builder.WriteString(strings.TrimSpace(olderTranscript))
	builder.WriteString("\n\n")
	builder.WriteString("### Recent Messages (for context only — do not summarize)\n")
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
