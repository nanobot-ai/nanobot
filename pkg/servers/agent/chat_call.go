package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const progressSessionKey = "progress"

type chatCall struct {
	s *Server
}

func (c chatCall) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        types.AgentTool,
		Description: types.AgentToolDescription,
		InputSchema: types.ChatInputSchema,
	}
}

func closeProgress(ctx context.Context, session *mcp.Session, err error) {
	var response types.CompletionResponse
	session.Get(progressSessionKey, &response)
	response.HasMore = false
	if err != nil {
		response.Error = err.Error()
	}
	if len(response.InternalMessages) > 0 {
		response.Output = response.InternalMessages[len(response.InternalMessages)-1]
		response.InternalMessages = response.InternalMessages[:len(response.InternalMessages)-1]
		// Sort items to ensure consistent display order: reasoning, content, tools
		sortCompletionItems(&response.Output)
	}
	response.ProgressToken = nil
	session.Set(progressSessionKey, &response)

	_ = session.SendPayload(ctx, "notifications/resources/updated", map[string]any{
		"uri": types.ProgressURI,
	})
}

// sortCompletionItems sorts items by type priority: reasoning first, then content, then tools
func sortCompletionItems(msg *types.Message) {
	if len(msg.Items) <= 1 {
		return
	}

	// Define priority order
	itemPriority := func(item types.CompletionItem) int {
		if item.Reasoning != nil {
			return 0 // Reasoning first
		} else if item.Content != nil {
			return 1 // Content second
		} else if item.ToolCall != nil || item.ToolCallResult != nil {
			return 2 // Tools last
		}
		return 99 // Unknown types at the end
	}

	// Sort using a stable sort to preserve relative order within same priority
	for i := 0; i < len(msg.Items)-1; i++ {
		for j := i + 1; j < len(msg.Items); j++ {
			if itemPriority(msg.Items[i]) > itemPriority(msg.Items[j]) {
				msg.Items[i], msg.Items[j] = msg.Items[j], msg.Items[i]
			}
		}
	}
}

func appendProgress(ctx context.Context, session *mcp.Session, progressMessage *mcp.Message) (*mcp.Message, error) {
	if progressMessage.Method != "notifications/progress" {
		return progressMessage, nil
	}

	var event progressPayload

	if err := json.Unmarshal(progressMessage.Params, &event); err != nil {
		return progressMessage, nil
	}
	if event.Meta.Progress == nil || event.Meta.Progress.MessageID == "" {
		return progressMessage, nil
	}

	// Get the current response from the session
	var response types.CompletionResponse
	session.Get(progressSessionKey, &response)

	defer func() {
		_ = session.SendPayload(ctx, "notifications/resources/updated", map[string]any{
			"uri": types.ProgressURI,
		})
	}()

	// Handle tool call results specially - we need to find the matching tool call
	// and add the result to it directly
	progressItem := event.Meta.Progress.Item
	if progressItem.ToolCallResult != nil {
		for msgIndex, msg := range response.InternalMessages {
			for itemIndex, item := range msg.Items {
				if item.ToolCall != nil && item.ToolCall.CallID == progressItem.ToolCallResult.CallID {
					// Modify in place for tool call results to match existing behavior
					response.InternalMessages[msgIndex].Items[itemIndex].ToolCallResult = progressItem.ToolCallResult
					response.HasMore = true
					session.Set(progressSessionKey, response)
					return nil, nil
				}
			}
		}
	}

	// Use types.AppendProgress for all other cases
	updatedResponse := types.AppendProgress(response, *event.Meta.Progress)
	updatedResponse.HasMore = true

	session.Set(progressSessionKey, updatedResponse)
	return nil, nil
}

func (c chatCall) Invoke(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	description := c.s.describeSession(ctx, payload.Arguments)
	if description != nil {
		defer func() {
			<-description
		}()
	}

	if attachments, _ := payload.Arguments["attachments"].([]any); len(attachments) > 0 {
		var err error
		payload.Arguments["attachments"], err = c.inlineAttachments(ctx, attachments)
		if err != nil {
			return nil, err
		}
	}

	async := msg.Meta()[types.AsyncMetaKey]
	if (async == "true" || async == true) && msg.ProgressToken() != nil {
		nctx := types.NanobotContext(ctx)
		session := mcp.SessionFromContext(ctx)
		mcp.SessionFromContext(ctx).Go(types.WithNanobotContext(session.Context(), nctx), func(ctx context.Context) {
			_, _ = c.chatInvoke(ctx, msg, payload)
		})
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Text: fmt.Sprintf("Chat request has been sent to the agent. You can track the progress of the response in the resource %s",
						types.ProgressURI),
				},
				{
					Type:     "resource_link",
					URI:      types.ProgressURI,
					MIMEType: types.ToolResultMimeType,
				},
			},
		}, nil
	}

	return c.chatInvoke(ctx, msg, payload)
}

func (c chatCall) chatInvoke(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (_ *mcp.CallToolResult, retErr error) {
	session := mcp.SessionFromContext(ctx).Parent

	defer func() {
		closeProgress(ctx, session, retErr)
	}()
	defer session.AddFilter(func(ctx context.Context, msg *mcp.Message) (*mcp.Message, error) {
		return appendProgress(ctx, session, msg)
	})()

	session.Set(progressSessionKey, &types.CompletionResponse{
		ProgressToken: msg.ProgressToken(),
	})

	result, err := c.s.runtime.Call(ctx, c.s.agentName, c.s.agentName, payload.Arguments, tools.CallOptions{
		ProgressToken: msg.ProgressToken(),
		LogData: map[string]any{
			"mcpToolName": payload.Name,
		},
	})
	if err != nil {
		return nil, err
	}

	mcpResult := mcp.CallToolResult{
		StructuredContent: result.StructuredContent,
		IsError:           result.IsError,
		Content:           result.Content,
	}

	err = msg.Reply(ctx, mcpResult)
	return &mcpResult, err
}

func GetMessages(ctx context.Context) ([]types.Message, error) {
	var (
		run         types.Execution
		allMessages []types.Message
	)

	session := mcp.SessionFromContext(ctx)
	session.Get(types.PreviousExecutionKey, &run)

	if run.PopulatedRequest != nil {
		allMessages = run.PopulatedRequest.Input
	}
	if run.Response != nil {
		allMessages = append(allMessages, run.Response.Output)
	}

	return types.ConsolidateTools(allMessages), nil
}

type progressPayload struct {
	Meta progressPayloadMeta `json:"_meta"`
}

type progressPayloadMeta struct {
	Progress *types.CompletionProgress `json:"ai.nanobot.progress/completion"`
}
