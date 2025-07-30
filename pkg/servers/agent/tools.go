package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func (s *Server) setAgent(ctx context.Context, args struct {
	Agent string `json:"agent"`
}) (string, error) {
	return args.Agent, s.data.SetCurrentAgent(ctx, args.Agent)
}

func (s *Server) getChat(ctx context.Context, _ struct{}) (*ChatData, error) {
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

	prompts, err := s.data.PromptMappings(ctx, sessiondata.WithAllowMissing())
	if err != nil {
		return nil, err
	}

	if s.multiAgent {
		currentAgent = s.data.CurrentAgent(ctx)
		agents, err = s.data.Agents(ctx)
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

type ChatData struct {
	ID           string            `json:"id"`
	Ext          ChatDataExtension `json:"ai.nanobot/ext,omitzero"`
	CurrentAgent string            `json:"currentAgent,omitempty"`
	Messages     []types.Message   `json:"messages"`
}

func (c ChatData) MarshalJSON() ([]byte, error) {
	if c.Messages == nil {
		c.Messages = []types.Message{}
	}
	// We want to omit the empty fields in the extension
	type Alias ChatData
	return json.Marshal(Alias(c))
}

type ChatDataExtension struct {
	CustomAgent *types.CustomAgent   `json:"customAgent,omitempty"`
	Tools       types.ToolMappings   `json:"tools,omitempty"`
	Prompts     types.PromptMappings `json:"prompts,omitempty"`
	Agents      types.Agents         `json:"agents,omitempty"`
}
