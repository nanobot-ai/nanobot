package agentui

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
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

func (c chatCall) Invoke(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	description := c.s.describeSession(ctx, payload.Arguments)
	currentAgent := c.s.data.CurrentAgent(ctx)

	c.s.data.CurrentAgent(ctx)
	client, err := c.s.runtime.GetClient(ctx, currentAgent)
	if err != nil {
		return nil, err
	}

	result, err := client.Call(ctx, types.AgentTool, payload.Arguments, mcp.CallOption{
		ProgressToken: msg.ProgressToken(),
		Meta:          payload.Meta,
	})
	if err != nil {
		return nil, err
	}

	mcpResult := mcp.CallToolResult{
		IsError: result.IsError,
		Content: result.Content,
	}

	if description != nil {
		<-description
	}

	err = msg.Reply(ctx, mcpResult)
	return &mcpResult, err
}
