package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
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
		chatData  *ChatData
		sessionID = mcp.SessionFromContext(ctx).Parent.ID()
		err       error
	)
	if c.s.isAgentPassthrough {
		chatData, err = c.getChatRemote(ctx, msg, payload)
		if err != nil {
			chatData = &ChatData{}
			//return nil, fmt.Errorf("failed to get chat from remote agent: %w", err)
		}
		chatData.ID = sessionID
	} else {
		chatData, err = c.getChatLocal(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat locally: %w", err)
		}
	}

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

func (c getChatCall) getChatRemote(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*ChatData, error) {
	target, err := c.s.data.GetCurrentAgentTargetMapping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current agent target mapping: %w", err)
	}
	result, err := c.s.runtime.Call(ctx, target.MCPServer, "get_chat", struct{}{}, tools.CallOptions{
		ProgressToken: msg.ProgressToken(),
		LogData: map[string]any{
			"mcpToolName": payload.Name,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call get_chat on remote agent %s: %w", target.MCPServer, err)
	}
	if result.IsError {
		return nil, fmt.Errorf("remote agent %s returned error: %s", target.MCPServer, result.Content)
	}
	for _, content := range result.Content {
		if content.Type == "text" {
			var chatData ChatData
			if err := json.Unmarshal([]byte(content.Text), &chatData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal chat data from remote agent %s: %w", target.MCPServer, err)
			}
			return &chatData, nil
		}
	}

	return nil, fmt.Errorf("remote agent %s did not return any chat data", target.MCPServer)
}

func (c getChatCall) getChatLocal(ctx context.Context) (*ChatData, error) {
	var (
		run          types.Execution
		agentConfig  *types.CustomAgent
		allMessages  []types.Message
		agents       types.Agents
		currentAgent string
	)

	session := mcp.SessionFromContext(ctx)
	session.Get(types.PreviousExecutionKey, &run)

	if run.PopulatedRequest != nil {
		allMessages = run.PopulatedRequest.Input
	}
	if run.Response != nil {
		allMessages = append(allMessages, run.Response.Output)
	}

	sessionID := session.ID()
	if session.Parent != nil {
		sessionID = session.Parent.ID()
	}

	if sessionAgentConfig := (types.CustomAgent{}); session.Get(types.CustomAgentConfigSessionKey, &sessionAgentConfig) {
		agentConfig = &sessionAgentConfig
	}

	prompts, err := c.s.data.PromptMappings(ctx, sessiondata.WithAllowMissing())
	if err != nil {
		return nil, err
	}

	if c.s.multiAgent {
		currentAgent = c.s.data.CurrentAgent(ctx)
		agents, err = c.s.data.Agents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get agents: %w", err)
		}

	}

	return &ChatData{
		ID:           sessionID,
		CurrentAgent: currentAgent,
		Ext: ChatDataExtension{
			CustomAgent: agentConfig,
			Tools:       run.ToolToMCPServer,
			Prompts:     prompts,
			Agents:      agents,
		},
		Messages: types.ConsolidateTools(allMessages),
	}, nil
}
