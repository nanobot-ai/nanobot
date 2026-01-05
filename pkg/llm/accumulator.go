package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

// progressAccumulator captures progress messages and assembles them into a partial response
type progressAccumulator struct {
	response      types.CompletionResponse
	progressToken any
}

// newProgressAccumulator creates a new progress accumulator
func newProgressAccumulator(progressToken any) *progressAccumulator {
	return &progressAccumulator{
		response: types.CompletionResponse{
			InternalMessages: []types.Message{},
		},
		progressToken: progressToken,
	}
}

// captureProgress intercepts and accumulates a progress message
func (pa *progressAccumulator) captureProgress(_ context.Context, prog *types.CompletionProgress) {
	// Use types.AppendProgress to handle all the merging logic
	pa.response = types.AppendProgress(pa.response, *prog)

	// Set HasMore flag on the current message if it exists
	if len(pa.response.InternalMessages) > 0 {
		for i := range pa.response.InternalMessages {
			if pa.response.InternalMessages[i].ID == prog.MessageID {
				pa.response.InternalMessages[i].HasMore = true
				break
			}
		}
	}
}

// getPartialResponse returns the accumulated response, filtering out incomplete tool calls and adding an error message
func (pa *progressAccumulator) getPartialResponse(ctx context.Context, err error) *types.CompletionResponse {
	if len(pa.response.InternalMessages) == 0 {
		// No partial response accumulated, return nil
		return nil
	}

	// Take the last message as the output
	if len(pa.response.InternalMessages) > 0 {
		lastMsg := pa.response.InternalMessages[len(pa.response.InternalMessages)-1]

		// Filter out incomplete tool calls
		filteredItems := make([]types.CompletionItem, 0, len(lastMsg.Items))
		for _, item := range lastMsg.Items {
			// Keep text content and reasoning
			if item.Content != nil || item.Reasoning != nil {
				// Remove Partial and HasMore flags for final response
				item.Partial = false
				filteredItems = append(filteredItems, item)
			}
			// Skip incomplete tool calls (Partial or HasMore)
			if item.ToolCall != nil && item.Partial {
				continue
			}
			// Keep complete tool calls (though this is rare for error cases)
			if item.ToolCall != nil && !item.Partial {
				filteredItems = append(filteredItems, item)
			}
		}

		// Add error message as text content
		errorItem := types.CompletionItem{
			ID: "error_" + time.Now().Format(time.RFC3339Nano),
			Content: &mcp.Content{
				Type: "text",
				Text: fmt.Sprintf("\n\n[Error: %v]", err),
			},
		}
		filteredItems = append(filteredItems, errorItem)

		// Send progress for the error message if we have progress tracking
		if pa.progressToken != nil {
			// Use the original progress.Send directly without going through interceptor
			// to avoid recursion
			if session := mcp.SessionFromContext(ctx); session != nil {
				_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
					ProgressToken: pa.progressToken,
					Meta: map[string]any{
						types.CompletionProgressMetaKey: &types.CompletionProgress{
							Model:     pa.response.Model,
							Agent:     pa.response.Agent,
							MessageID: lastMsg.ID,
							Role:      lastMsg.Role,
							Item:      errorItem,
						},
					},
				})
			}
		}

		// Update the output message
		lastMsg.Items = filteredItems
		lastMsg.HasMore = false
		pa.response.Output = lastMsg

		// Remove the last message from InternalMessages since it's now in Output
		if len(pa.response.InternalMessages) > 1 {
			pa.response.InternalMessages = pa.response.InternalMessages[:len(pa.response.InternalMessages)-1]
		} else {
			pa.response.InternalMessages = nil
		}
	}

	pa.response.Error = err.Error()
	return &pa.response
}
