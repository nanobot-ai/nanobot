package agentui

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
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

	result, err = c.setRemote(ctx, msg, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to set current agent on remote: %w", err)
	}

	return &mcp.CallToolResult{
		StructuredContent: result.StructuredContent,
		IsError:           result.IsError,
		Content:           result.Content,
	}, nil
}

func (c setCurrentAgentCall) setRemote(ctx context.Context, _ mcp.Message, payload mcp.CallToolRequest) (*types.CallResult, error) {
	agentName, _ := payload.Arguments["agent"].(string)
	if err := c.s.data.SetCurrentAgent(ctx, agentName); err != nil {
		return nil, err
	}
	return &types.CallResult{
		IsError:           false,
		StructuredContent: agentName,
		Content: []mcp.Content{
			{
				Text: fmt.Sprintf("Current agent has been set to %s", agentName),
			},
		},
	}, nil
}
