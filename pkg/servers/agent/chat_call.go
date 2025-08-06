package agent

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

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

func (c chatCall) Invoke(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target, err := c.s.data.GetCurrentAgentTargetMapping(ctx)
	if err != nil {
		return nil, err
	}

	description := c.s.describeSession(ctx, payload.Arguments)

	result, err := c.s.runtime.Call(ctx, target.MCPServer, target.TargetName, payload.Arguments, tools.CallOptions{
		ProgressToken: msg.ProgressToken(),
		LogData: map[string]any{
			"mcpToolName": payload.Name,
		},
	})
	if err != nil {
		return nil, err
	}

	if !c.s.isAgentPassthrough && result.ChatResponse && result.Agent != "" {
		c.s.data.SetCurrentAgent(ctx, result.Agent)
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
