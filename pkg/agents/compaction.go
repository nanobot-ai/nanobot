package agents

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nanobot-ai/nanobot/pkg/contextguard"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

const (
	compactionSummaryPrefix = "## Compacted Context\n"

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
	// Adjust split forward so tool results aren't separated from their calls.
	splitIdx = adjustSplitForToolPairs(req.Input, splitIdx)

	// If the recent window alone would still exceed the context limit,
	// reduce retain until it fits (keeping at least 1 message).
	guard := contextguard.NewService(contextguard.Config{WarnThreshold: config.Compaction.EffectiveGuardThreshold(), UseTiktoken: true})
	for splitIdx < len(req.Input)-1 {
		result := guard.Evaluate(contextguard.State{
			Model:        runModel,
			SystemPrompt: req.SystemPrompt,
			Tools:        req.Tools,
			Messages:     req.Input[splitIdx:],
		})
		if result.Status != contextguard.StatusOverLimit {
			break
		}
		splitIdx++
	}

	older := req.Input[:splitIdx]
	recent := req.Input[splitIdx:]

	if len(older) == 0 {
		return false, nil
	}

	// Render the older messages into a transcript before they are dropped.
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

	// Build the summary as a user→assistant pair so that the summary
	// (assistant role) naturally alternates with a following user message
	// such as a tool_result. The user message satisfies the Anthropic
	// requirement that the first message use the "user" role.
	summaryUser := types.Message{
		ID:   "compaction-summary-" + uuid.String(),
		Role: "user",
		Items: []types.CompletionItem{
			{
				ID: uuid.String(),
				Content: &mcp.Content{
					Type: "text",
					Text: "The earlier portion of this conversation has been compacted. A summary follows.",
				},
			},
		},
	}
	summaryAssistant := types.Message{
		ID:   "compaction-summary-asst-" + uuid.String(),
		Role: "assistant",
		Items: []types.CompletionItem{
			{
				ID: uuid.String(),
				Content: &mcp.Content{
					Type: "text",
					Text: compactionSummaryPrefix +
						"The earlier portion of this conversation has been summarized to manage context length. " +
						"Here is what happened:\n\n" +
						summaryText,
				},
			},
		},
	}

	// Rebuild: summary pair + recent messages.
	newInput := make([]types.Message, 0, 2+len(recent))
	newInput = append(newInput, summaryUser, summaryAssistant)
	// If the first retained message is also assistant role, merge its
	// content into the summary assistant message to maintain alternation.
	if len(recent) > 0 && recent[0].Role == "assistant" {
		newInput[1].Items = append(newInput[1].Items, recent[0].Items...)
		recent = recent[1:]
	}
	newInput = append(newInput, recent...)

	// Strip tool_result items whose corresponding tool_call was in the
	// dropped older messages. adjustSplitForToolPairs handles most cases
	// but cannot remove the very last message, so orphans can survive.
	newInput = stripOrphanedToolItems(newInput)

	req.Input = newInput

	// Propagate changes to the execution state.
	if run.PopulatedRequest != nil {
		run.PopulatedRequest.Input = req.Input
	}
	if prev != nil {
		if prev.PopulatedRequest != nil {
			prev.PopulatedRequest.Input = req.Input
		}
		prev.ToolOutputs = nil
		// Clear the previous response so populateRequest does not re-add
		// assistant messages with orphaned tool_use blocks.
		if prev.Response != nil {
			prev.Response.InternalMessages = nil
			prev.Response.Output = types.Message{}
		}
	}

	if summaryResp.Usage != nil {
		run.Usage = types.MergeUsage(run.Usage, *summaryResp.Usage)
	}

	log.Infof(ctx, "compaction complete: model=%s droppedMessages=%d retainedRecent=%d",
		runModel, len(older), len(recent))

	return true, nil
}

// adjustSplitForToolPairs moves the split index forward so that tool result
// messages in the recent portion are not orphaned from their tool calls.
// If a message at the start of recent has tool results whose corresponding
// tool calls would be in the older (dropped) portion, it moves that message
// into the older set.
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
	// calls not in the recent window, move it into the older (dropped) set.
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

// stripOrphanedToolItems removes ToolCallResult items that reference
// ToolCall IDs not present anywhere in messages, and ToolCall items without
// a matching ToolCallResult. Messages that become empty after stripping are
// removed entirely.
func stripOrphanedToolItems(messages []types.Message) []types.Message {
	callIDs := map[string]bool{}
	resultIDs := map[string]bool{}
	for _, msg := range messages {
		for _, item := range msg.Items {
			if item.ToolCall != nil && item.ToolCall.CallID != "" {
				callIDs[item.ToolCall.CallID] = true
			}
			if item.ToolCallResult != nil && item.ToolCallResult.CallID != "" {
				resultIDs[item.ToolCallResult.CallID] = true
			}
		}
	}

	var result []types.Message
	for _, msg := range messages {
		var kept []types.CompletionItem
		for _, item := range msg.Items {
			if item.ToolCallResult != nil && item.ToolCallResult.CallID != "" {
				if !callIDs[item.ToolCallResult.CallID] {
					continue
				}
			}
			if item.ToolCall != nil && item.ToolCall.CallID != "" {
				if !resultIDs[item.ToolCall.CallID] {
					continue
				}
			}
			kept = append(kept, item)
		}
		if len(kept) > 0 {
			msg.Items = kept
			result = append(result, msg)
		}
	}
	return result
}

// extractCompactionSummary finds the most recent compaction summary message
// in the message list. This allows subsequent compactions to build on the
// previous summary rather than starting from scratch.
func extractCompactionSummary(messages []types.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != "assistant" {
			continue
		}
		for _, item := range msg.Items {
			if item.Content != nil && strings.HasPrefix(item.Content.Text, compactionSummaryPrefix) {
				return strings.TrimSpace(item.Content.Text[len(compactionSummaryPrefix):])
			}
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
