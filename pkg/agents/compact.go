package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

const (
	compactionThreshold      = 0.835
	compactionSummaryMetaKey = "ai.nanobot.meta/compaction-summary"
)

const defaultContextWindow = 200_000

// getContextWindowSize returns the context window size for the given model.
// If configOverride is > 0, it is used directly. Otherwise, defaults to 200k.
func getContextWindowSize(configOverride int) int {
	if configOverride > 0 {
		fmt.Printf("[DEBUG compact] getContextWindowSize: using config override %d\n", configOverride)
		return configOverride
	}
	fmt.Printf("[DEBUG compact] getContextWindowSize: using default %d\n", defaultContextWindow)
	return defaultContextWindow
}

// shouldCompact returns true if the estimated token count of the request
// exceeds the compaction threshold of the context window.
func shouldCompact(req types.CompletionRequest, contextWindowSize int) bool {
	if contextWindowSize <= 0 {
		fmt.Printf("[DEBUG compact] shouldCompact: contextWindowSize=%d <= 0, returning false\n", contextWindowSize)
		return false
	}

	estimated := estimateTokens(req.Input, req.SystemPrompt, req.Tools)
	threshold := int(float64(contextWindowSize) * compactionThreshold)
	result := estimated > threshold
	fmt.Printf("[DEBUG compact] shouldCompact: estimatedTokens=%d threshold=%d (%.1f%% of %d) shouldCompact=%v\n",
		estimated, threshold, compactionThreshold*100, contextWindowSize, result)
	if result {
		fmt.Printf("[DEBUG compact] shouldCompact: over threshold by %d tokens (%.1f%% of context window)\n",
			estimated-threshold, float64(estimated)/float64(contextWindowSize)*100)
	}
	return result
}

// IsCompactionSummary checks whether a message is a compaction summary
// by looking for the compaction summary meta key.
func IsCompactionSummary(msg types.Message) bool {
	if len(msg.Items) == 0 {
		return false
	}
	item := msg.Items[0]
	if item.Content == nil || item.Content.Meta == nil {
		return false
	}
	_, ok := item.Content.Meta[compactionSummaryMetaKey]
	return ok
}

// splitHistoryAndNewInput separates the full populated input into history (messages
// from previous turns) and new input (messages from the current request).
// It finds the boundary by matching the first message ID from currentRequestInput.
func splitHistoryAndNewInput(fullInput, currentRequestInput []types.Message) (history, newInput []types.Message) {
	fmt.Printf("[DEBUG compact] splitHistoryAndNewInput: fullInput=%d messages, currentRequestInput=%d messages\n",
		len(fullInput), len(currentRequestInput))

	if len(currentRequestInput) == 0 {
		fmt.Printf("[DEBUG compact] splitHistoryAndNewInput: no currentRequestInput, all %d messages treated as history\n", len(fullInput))
		return fullInput, nil
	}

	firstNewID := currentRequestInput[0].ID
	for i, msg := range fullInput {
		if msg.ID == firstNewID {
			fmt.Printf("[DEBUG compact] splitHistoryAndNewInput: boundary found at index %d (msgID=%q), history=%d newInput=%d\n",
				i, firstNewID, i, len(fullInput)-i)
			return fullInput[:i], fullInput[i:]
		}
	}

	// If we can't find the boundary, treat everything as history
	fmt.Printf("[DEBUG compact] splitHistoryAndNewInput: boundary msgID=%q NOT found in fullInput, treating all %d messages as history\n",
		firstNewID, len(fullInput))
	return fullInput, nil
}

type compactResult struct {
	compactedInput   []types.Message
	archivedMessages []types.Message
}

// compact performs conversation compaction by summarizing history messages
// into a condensed summary, allowing the conversation to continue within
// the context window limits.
//
// On re-compaction, only the messages since the previous summary are summarized
// (with the previous summary included as context). This keeps the summarization
// input bounded rather than growing with the full conversation.
func (a *Agents) compact(ctx context.Context, req types.CompletionRequest, currentRequestInput []types.Message, previousCompacted []types.Message) (*compactResult, error) {
	fmt.Printf("[DEBUG compact] ========== COMPACTION STARTED ==========\n")
	fmt.Printf("[DEBUG compact] model=%q totalInputMessages=%d currentRequestInputMessages=%d previousCompactedMessages=%d\n",
		req.Model, len(req.Input), len(currentRequestInput), len(previousCompacted))
	fmt.Printf("[DEBUG compact] systemPrompt length=%d, tools=%d\n", len(req.SystemPrompt), len(req.Tools))

	history, newInput := splitHistoryAndNewInput(req.Input, currentRequestInput)

	fmt.Printf("[DEBUG compact] after split: history=%d messages, newInput=%d messages\n", len(history), len(newInput))
	for i, msg := range history {
		itemSummary := fmt.Sprintf("role=%s items=%d", msg.Role, len(msg.Items))
		if IsCompactionSummary(msg) {
			itemSummary += " [COMPACTION_SUMMARY]"
		}
		fmt.Printf("[DEBUG compact]   history[%d]: id=%q %s\n", i, msg.ID, itemSummary)
	}
	for i, msg := range newInput {
		fmt.Printf("[DEBUG compact]   newInput[%d]: id=%q role=%s items=%d\n", i, msg.ID, msg.Role, len(msg.Items))
	}

	// Split history into: messages before/including the previous summary, and messages after it.
	// We only need to summarize the messages after the previous summary, using the summary as context.
	var previousSummaryText string
	var sinceLastSummary []types.Message
	lastSummaryIdx := -1
	for i, msg := range history {
		if IsCompactionSummary(msg) {
			lastSummaryIdx = i
			if len(msg.Items) > 0 && msg.Items[0].Content != nil {
				previousSummaryText = msg.Items[0].Content.Text
			}
		}
	}
	if lastSummaryIdx >= 0 {
		sinceLastSummary = history[lastSummaryIdx+1:]
		fmt.Printf("[DEBUG compact] found previous summary at history[%d], summarizing %d messages since then\n",
			lastSummaryIdx, len(sinceLastSummary))
		fmt.Printf("[DEBUG compact] previous summary text length=%d\n", len(previousSummaryText))
	} else {
		sinceLastSummary = history
		fmt.Printf("[DEBUG compact] no previous summary found, summarizing all %d history messages\n", len(sinceLastSummary))
	}

	fmt.Printf("[DEBUG compact] history messages for archival: %d\n", len(history))

	// Build summarization transcript from only the messages since the last summary
	transcript := buildTranscript(sinceLastSummary)
	fmt.Printf("[DEBUG compact] built transcript: %d bytes from %d messages\n", len(transcript), len(sinceLastSummary))

	var summaryPrompt string
	if previousSummaryText != "" {
		fmt.Printf("[DEBUG compact] using RE-COMPACTION prompt (previous summary + new messages)\n")
		summaryPrompt = fmt.Sprintf(`You are a conversation summarizer. Below is a previous summary of an earlier portion of the conversation, followed by the new messages since that summary.

Please provide an updated summary that incorporates both the previous summary and the new messages. Capture:
1. All key topics discussed
2. Important decisions made
3. Relevant context and details that would be needed to continue the conversation
4. Any pending questions or tasks

Be thorough but concise. This summary will replace the conversation history to allow the conversation to continue.
Be clear about whether any task is incomplete, and if so, which parts are finished and which parts are not.
If the user asked for something that has already been completed, state this clearly to avoid the next agent repeating the work.

--- PREVIOUS SUMMARY ---
%s
--- END PREVIOUS SUMMARY ---

--- NEW MESSAGES ---
%s
--- END NEW MESSAGES ---

Provide your updated summary now:`, previousSummaryText, transcript)
	} else {
		fmt.Printf("[DEBUG compact] using INITIAL compaction prompt (full transcript)\n")
		summaryPrompt = fmt.Sprintf(`You are a conversation summarizer. Below is a conversation transcript between a user and an assistant.
Please provide a detailed summary that captures:
1. All key topics discussed
2. Important decisions made
3. Relevant context and details that would be needed to continue the conversation
4. Any pending questions or tasks

Be thorough but concise. This summary will replace the conversation history to allow the conversation to continue.
Be clear about whether any task is incomplete, and if so, which parts are finished and which parts are not.
If the user asked for something that has already been completed, state this clearly to avoid the next agent repeating the work.

--- CONVERSATION TRANSCRIPT ---
%s
--- END TRANSCRIPT ---

Provide your summary now:`, transcript)
	}

	fmt.Printf("[DEBUG compact] summary prompt length=%d bytes\n", len(summaryPrompt))
	fmt.Printf("[DEBUG compact] sending summarization request to LLM (model=%q)...\n", req.Model)

	summaryReq := types.CompletionRequest{
		Model: req.Model,
		Input: []types.Message{
			{
				ID:   uuid.String(),
				Role: "user",
				Items: []types.CompletionItem{
					{
						ID: uuid.String(),
						Content: &mcp.Content{
							Type: "text",
							Text: summaryPrompt,
						},
					},
				},
			},
		},
	}

	resp, err := a.completer.Complete(ctx, summaryReq)
	if err != nil {
		fmt.Printf("[DEBUG compact] ERROR: summarization LLM call failed: %v\n", err)
		return nil, fmt.Errorf("compaction summarization failed: %w", err)
	}

	// Extract summary text from response
	summaryText := extractTextFromResponse(resp)
	fmt.Printf("[DEBUG compact] summarization response: text length=%d\n", len(summaryText))
	if summaryText == "" {
		fmt.Printf("[DEBUG compact] ERROR: compaction produced empty summary\n")
		return nil, fmt.Errorf("compaction produced empty summary")
	}

	// Create summary message with compaction metadata
	now := time.Now()
	summaryMessage := types.Message{
		ID:      "compaction-summary-" + uuid.String(),
		Created: &now,
		Role:    "user",
		Items: []types.CompletionItem{
			{
				ID: uuid.String(),
				Content: &mcp.Content{
					Type: "text",
					Text: fmt.Sprintf("The conversation became too long and was compacted. The following is a summary. Please do not acknowledge the summary in your response, but continue where you left off with the conversation.\n\n[Conversation Summary]\n%s", summaryText),
					Meta: map[string]any{
						compactionSummaryMetaKey: true,
					},
				},
			},
		},
	}

	// Build the compacted input: summary + new user messages
	compactedInput := []types.Message{summaryMessage}
	compactedInput = append(compactedInput, newInput...)

	// Build archived messages: previous compacted + all history from this compaction (including old summaries)
	archivedMessages := make([]types.Message, 0, len(previousCompacted)+len(history))
	archivedMessages = append(archivedMessages, previousCompacted...)
	archivedMessages = append(archivedMessages, history...)

	fmt.Printf("[DEBUG compact] compaction complete: compactedInput=%d messages, archivedMessages=%d messages\n",
		len(compactedInput), len(archivedMessages))
	fmt.Printf("[DEBUG compact] summary message id=%q, summary text length=%d\n", summaryMessage.ID, len(summaryText))
	fmt.Printf("[DEBUG compact] ========== COMPACTION FINISHED ==========\n")

	return &compactResult{
		compactedInput:   compactedInput,
		archivedMessages: archivedMessages,
	}, nil
}

// buildTranscript creates a text representation of conversation messages
// suitable for summarization by an LLM.
func buildTranscript(messages []types.Message) string {
	fmt.Printf("[DEBUG compact] buildTranscript: processing %d messages\n", len(messages))
	var sb strings.Builder

	contentItems := 0
	toolCalls := 0
	toolResults := 0
	truncatedResults := 0

	for _, msg := range messages {
		role := msg.Role
		if role == "" {
			role = "unknown"
		}

		for _, item := range msg.Items {
			if item.Content != nil && item.Content.Text != "" {
				fmt.Fprintf(&sb, "[%s]: %s\n", role, item.Content.Text)
				contentItems++
			}
			if item.ToolCall != nil {
				fmt.Fprintf(&sb, "[%s] (tool call: %s): %s\n", role, item.ToolCall.Name, item.ToolCall.Arguments)
				toolCalls++
			}
			if item.ToolCallResult != nil {
				toolResults++
				for _, c := range item.ToolCallResult.Output.Content {
					if c.Text != "" {
						// Truncate very long tool results in the transcript
						text := c.Text
						if len(text) > 5000 {
							text = text[:5000] + "... [truncated]"
							truncatedResults++
						}
						fmt.Fprintf(&sb, "[tool result]: %s\n", text)
					}
				}
			}
		}
	}

	fmt.Printf("[DEBUG compact] buildTranscript: contentItems=%d toolCalls=%d toolResults=%d truncatedResults=%d transcriptBytes=%d\n",
		contentItems, toolCalls, toolResults, truncatedResults, sb.Len())

	return sb.String()
}

// extractTextFromResponse extracts the text content from a completion response.
func extractTextFromResponse(resp *types.CompletionResponse) string {
	if resp == nil {
		fmt.Printf("[DEBUG compact] extractTextFromResponse: response is nil\n")
		return ""
	}

	var texts []string
	for i, item := range resp.Output.Items {
		if item.Content != nil && item.Content.Text != "" {
			texts = append(texts, item.Content.Text)
			fmt.Printf("[DEBUG compact] extractTextFromResponse: item[%d] text length=%d\n", i, len(item.Content.Text))
		}
	}
	result := strings.Join(texts, "\n")
	fmt.Printf("[DEBUG compact] extractTextFromResponse: extracted %d text parts, total length=%d\n", len(texts), len(result))
	return result
}
