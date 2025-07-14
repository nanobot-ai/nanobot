package responses

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func send(ctx context.Context, progress *types.CompletionProgress, progressToken any) {
	if progressToken == "" || progressToken == nil {
		return
	}
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return
	}

	_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
		ProgressToken: progressToken,
		Meta: map[string]any{
			types.CompletionProgressMetaKey: progress,
		},
	})
}

func progressResponse(ctx context.Context, agentName, modelName string, resp *http.Response, progressToken any) (response Response, seen bool, err error) {
	lines := bufio.NewScanner(resp.Body)
	defer resp.Body.Close()

	progress := types.CompletionProgress{
		Agent: agentName,
		Model: modelName,
	}

	for lines.Scan() {
		line := lines.Text()

		header, body, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(header) {
		case "data":
			var event Progress
			body = strings.TrimSpace(body)
			data := []byte(body)
			if err := json.Unmarshal(data, &event); err != nil {
				log.Errorf(ctx, "failed to decode event: %v: %s", err, body)
				continue
			}

			switch event.Type {
			case "response.created":
				progress.Model = event.Response.Model
				progress.MessageID = event.Response.ID
			case "response.output_item.added":
				if event.Item.Type == "function_call" {
					progress.Item = types.CompletionItem{
						Partial: true,
						HasMore: true,
						ID:      event.Item.ID,
						ToolCall: &types.ToolCall{
							CallID: event.Item.CallID,
							Name:   event.Item.Name,
						},
					}
				} else if event.Item.Type == "message" {
					progress.Item = types.CompletionItem{
						Partial: true,
						HasMore: true,
						ID:      event.Item.ID,
						Content: &mcp.Content{
							Type: "text",
						},
					}
				}
			case "response.function_call_arguments.delta":
				progress.Item.ToolCall.Arguments = event.Delta
				send(ctx, &progress, progressToken)
			case "response.output_item.done":
				if progress.Item.ID != "" {
					progress.Item = types.CompletionItem{
						Partial: true,
						ID:      progress.Item.ID,
					}
					send(ctx, &progress, progressToken)
				}
				progress.Item = types.CompletionItem{}
			case "response.output_text.delta":
				if progress.Item.Content != nil {
					progress.Item.Content.Text = event.Delta
					send(ctx, &progress, progressToken)
				}
			}

			if event.Type == "response.completed" || event.Type == "response.failed" || event.Type == "response.incomplete" {
				log.Messages(ctx, "responses-api", false, data)
				response = event.Response
				seen = true
			}
		}
	}

	err = lines.Err()
	return
}
