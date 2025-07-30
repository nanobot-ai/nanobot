package agentbuilder

import (
	"context"
	"encoding/json"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools mcp.ServerTools
	store *Store
}

func NewServer(store *Store) *Server {
	s := &Server{
		store: store,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("create_custom_agent", "Create a custom agent from current thread", s.createAgent),
		mcp.NewServerTool("list_custom_agents", "List available custom agents", s.listAgents),
		mcp.NewServerTool("delete_custom_agent", "Delete a custom agent", s.deleteAgent),
		mcp.NewServerTool("update_custom_agent", "Update a custom agent", s.updateAgent),
	)

	return s
}

type CustomAgents struct {
	CustomAgents []types.CustomAgentMeta `json:"customAgents"`
}

func (s *Server) listAgents(ctx context.Context, _ struct{}) (result CustomAgents, _ error) {
	var (
		accountID string
	)
	session := mcp.SessionFromContext(ctx)
	if !session.Get(types.AccountIDSessionKey, &accountID) {
		return CustomAgents{}, mcp.ErrRPCInvalidParams.WithMessage("missing account ID in session")
	}

	agents, err := s.store.FindByAccountID(ctx, accountID)
	if err != nil {
		return CustomAgents{}, err
	}

	for _, agent := range agents {
		result.CustomAgents = append(result.CustomAgents, types.CustomAgentMeta{
			ID:          agent.UUID,
			Name:        agent.Name,
			Description: agent.Description,
		})
	}

	return result, nil
}

func (s *Server) deleteAgent(ctx context.Context, params types.CustomAgentMeta) (result types.CustomAgentMeta, _ error) {
	var (
		accountID string
	)
	session := mcp.SessionFromContext(ctx)
	if !session.Get(types.AccountIDSessionKey, &accountID) {
		return result, mcp.ErrRPCInvalidParams.WithMessage("missing account ID in session")
	}

	agent, err := s.store.GetByUUIDAndAccountID(ctx, params.ID, accountID)
	if err != nil {
		return result, err
	}

	if agent.AccountID != accountID {
		return result, mcp.ErrRPCInvalidParams.WithMessage("agent does not belong to account")
	}

	return types.CustomAgentMeta{
		ID:          agent.UUID,
		Name:        agent.Name,
		Description: agent.Description,
	}, s.store.Delete(ctx, agent.ID)
}

func (s *Server) updateAgent(ctx context.Context, params types.CustomAgent) (result types.CustomAgent, _ error) {
	var (
		session       = mcp.SessionFromContext(ctx)
		existingAgent types.CustomAgent
	)
	session = session.Parent
	session.Get(types.CustomAgentConfigSessionKey, &existingAgent)
	params.ID = existingAgent.ID
	if params.ID != "" {
		session.Set(types.CustomAgentConfigSessionKey, &params)
		session.Set(types.CustomAgentModifiedSessionKey, true)
	}
	return params, nil
}

func (s *Server) createAgent(ctx context.Context, params types.CustomAgent) (result *types.CustomAgent, _ error) {
	var (
		session   = mcp.SessionFromContext(ctx)
		accountID string
		newID     = uuid.String()
	)

	session = session.Parent
	session.Get(types.AccountIDSessionKey, &accountID)

	toWrite := params
	// zero out the custom agent meta
	toWrite.CustomAgentMeta = types.CustomAgentMeta{}
	configData, err := json.Marshal(toWrite)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("failed to marshal config: " + err.Error())
	}

	err = s.store.Create(ctx, &Agent{
		UUID:        newID,
		SessionID:   session.ID(),
		AccountID:   accountID,
		Config:      string(configData),
		Name:        params.Name,
		Description: params.Description,
		IsPublic:    params.IsPublic,
	})
	if err != nil {
		return nil, err
	}

	params.ID = newID
	return &params, nil
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
	case "tools/list":
		mcp.Invoke(ctx, msg, s.tools.List)
	case "tools/call":
		mcp.Invoke(ctx, msg, s.tools.Call)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage(msg.Method))
	}
}

func (s *Server) initialize(_ context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools:     &mcp.ToolsServerCapability{},
			Resources: &mcp.ResourcesServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}
