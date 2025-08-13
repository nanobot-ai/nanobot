package agent

import (
	"context"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type getChatCall struct {
	s *Server
}

func (c getChatCall) getChatLocal(ctx context.Context) (*types.ChatData, error) {
	var (
		run          types.Execution
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

	prompts, err := c.s.data.PublishedPromptMappings(ctx, sessiondata.WithAllowMissing())
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

	return &types.ChatData{
		ID:           sessionID,
		CurrentAgent: currentAgent,
		Ext: types.ChatDataExtension{
			Tools:   run.ToolToMCPServer,
			Prompts: prompts,
			Agents:  agents,
		},
		Messages: types.ConsolidateTools(allMessages),
	}, nil
}
