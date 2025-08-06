package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

var setCurrentAgentInputSchema json.RawMessage

func init() {
	schema, err := jsonschema.For[struct {
		Agent string `json:"agent"`
	}]()
	if err != nil {
		panic(fmt.Sprintf("failed to create set_current_agent input schema: %v", err))
	}
	setCurrentAgentInputSchema, err = json.Marshal(schema)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal set_current_agent input schema: %v", err))
	}
}

type setCurrentAgentCall struct {
	s *Server
}

func (c setCurrentAgentCall) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "set_current_agent",
		Description: "Set the current agent the user is chatting with",
		InputSchema: setCurrentAgentInputSchema,
	}
}

func (c setCurrentAgentCall) Invoke(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var (
		result *types.CallResult
		err    error
	)
	if c.s.isAgentPassthrough {
		result, err = c.setRemote(ctx, msg, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to set current agent on remote: %w", err)
		}
	} else {
		agent, _ := payload.Arguments["agent"].(string)
		result, err = c.setLocal(ctx, agent)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat locally: %w", err)
		}
	}

	return &mcp.CallToolResult{
		IsError: result.IsError,
		Content: result.Content,
	}, nil
}

func (c setCurrentAgentCall) setRemote(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*types.CallResult, error) {
	target, err := c.s.data.GetCurrentAgentTargetMapping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current agent target mapping: %w", err)
	}
	return c.s.runtime.Call(ctx, target.MCPServer, "set_current_agent", struct{}{}, tools.CallOptions{
		ProgressToken: msg.ProgressToken(),
		LogData: map[string]any{
			"mcpToolName": payload.Name,
		},
	})
}

func (c setCurrentAgentCall) setLocal(ctx context.Context, agent string) (*types.CallResult, error) {
	if agent != "" {
		if err := c.s.data.SetCurrentAgent(ctx, agent); err != nil {
			return nil, fmt.Errorf("failed to set current agent: %w", err)
		}
	}
	return &types.CallResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: agent,
			},
		},
	}, nil
}
