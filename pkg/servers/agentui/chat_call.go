package agentui

import (
	"context"
	"fmt"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/agent"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type chatCall struct {
	s *Server
}

func (c chatCall) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        types.AgentTool + "_ui",
		Description: types.AgentToolDescription,
		InputSchema: types.ChatInputSchema,
	}
}

func (c chatCall) inlineAttachments(ctx context.Context, attachments []any) ([]any, error) {
	newAttachments := make([]any, 0, len(attachments))

	messages, err := agent.GetMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

attachmentsLoop:
	for i, attachment := range attachments {
		newAttachments = append(newAttachments, attachment)
		data, ok := attachment.(map[string]any)
		if !ok {
			continue
		}

		uri, ok := data["url"].(string)
		if !ok {
			continue
		}

		if strings.HasPrefix(uri, "data:") || strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
			continue
		}

		for mi := len(messages) - 1; mi >= 0; mi-- {
			for j := len(messages[mi].Items) - 1; j >= 0; j-- {
				item := messages[mi].Items[j]
				if item.ToolCallResult != nil {
					for _, content := range item.ToolCallResult.Output.Content {
						if content.Resource != nil && content.Resource.URI == uri {
							// Drop the attachment from the list
							newAttachments = newAttachments[:i]
							newAttachments = append(newAttachments, map[string]any{
								"url": content.Resource.ToDataURI(),
							})
							continue attachmentsLoop
						}
					}
				}
			}
		}

		clientName := c.s.data.CurrentAgent(ctx)
		if strings.HasPrefix(uri, "nanobot://") {
			clientName = "nanobot.resources"
		}

		client, err := c.s.runtime.GetClient(ctx, clientName)
		if err != nil {
			return nil, err
		}

		// Drop the attachment from the list
		newAttachments = newAttachments[:i]

		resource, err := client.ReadResource(ctx, uri)
		if err != nil {
			return nil, err
		}

		for _, content := range resource.Contents {
			dataURI := content.ToDataURI()
			newAttachments = append(newAttachments, map[string]any{
				"url": dataURI,
			})
		}
	}

	return newAttachments, nil
}

func (c chatCall) Invoke(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	description := c.s.describeSession(ctx, payload.Arguments)
	currentAgent := c.s.data.CurrentAgent(ctx)

	c.s.data.CurrentAgent(ctx)
	client, err := c.s.runtime.GetClient(ctx, currentAgent)
	if err != nil {
		return nil, err
	}

	if attachments, _ := payload.Arguments["attachments"].([]any); len(attachments) > 0 {
		payload.Arguments["attachments"], err = c.inlineAttachments(ctx, attachments)
		if err != nil {
			return nil, err
		}
	}

	result, err := client.Call(ctx, types.AgentTool, payload.Arguments, mcp.CallOption{
		ProgressToken: msg.ProgressToken(),
		Meta:          payload.Meta,
	})
	if err != nil {
		return nil, err
	}

	mcpResult := mcp.CallToolResult{
		StructuredContent: result.StructuredContent,
		IsError:           result.IsError,
		Content:           result.Content,
	}

	if description != nil {
		<-description
	}

	err = msg.Reply(ctx, mcpResult)
	return &mcpResult, err
}
