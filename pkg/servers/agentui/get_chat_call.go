package agentui

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type getChatCall struct {
	s *Server
}

func (c getChatCall) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "get_chat",
		Description: "Returns the contents of the current thread",
		InputSchema: mcp.EmptyObjectSchema,
	}
}

func (c getChatCall) Invoke(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var (
		chatData  *types.ChatData
		sessionID = mcp.SessionFromContext(ctx).Parent.ID()
		err       error
	)

	chatData, err = c.getChatRemote(ctx, msg, payload)
	if err != nil {
		chatData = &types.ChatData{}
		return nil, fmt.Errorf("failed to get chat from remote agent: %w", err)
	}
	chatData.ID = sessionID

	chatDataJSON, err := json.Marshal(chatData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat data: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type:              "text",
				Text:              string(chatDataJSON),
				StructuredContent: chatData,
			},
		},
	}, nil
}

func (c getChatCall) getChatRemote(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*types.ChatData, error) {
	currentAgent := c.s.data.CurrentAgent(ctx)
	result, err := c.s.runtime.Call(ctx, currentAgent, "get_chat", struct{}{}, tools.CallOptions{
		ProgressToken: msg.ProgressToken(),
		LogData: map[string]any{
			"mcpToolName": payload.Name,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call get_chat on remote agent %s: %w", currentAgent, err)
	}
	if result.IsError {
		return nil, fmt.Errorf("remote agent %s returned error: %s", currentAgent, result.Content)
	}
	for _, content := range result.Content {
		if content.Type == "text" {
			var chatData types.ChatData
			if err := json.Unmarshal([]byte(content.Text), &chatData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal chat data from remote agent %s: %w", currentAgent, err)
			}
			return &chatData, nil
		}
	}

	return nil, fmt.Errorf("remote agent %s did not return any chat data", currentAgent)
}
