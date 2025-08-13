package meta

import (
	"context"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

func (s *Server) deleteChat(ctx context.Context, data struct {
	ID string `json:"chatId"`
}) (*types.Chat, error) {
	mcpSession := mcp.SessionFromContext(ctx)
	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return nil, err
	}

	chatSession, err := manager.DB.GetByIDByAccountID(ctx, data.ID, accountID)
	if err != nil {
		return nil, err
	}

	if err := manager.DB.Delete(ctx, data.ID); err != nil {
		return nil, err
	}

	chat := chatFromSession(chatSession, accountID)
	return &chat, nil
}

func (s *Server) createChat(ctx context.Context, _ struct{}) (*types.Chat, error) {
	mcpSession := mcp.SessionFromContext(ctx)
	var (
		manager   session.Manager
		accountID string
	)

	if !mcpSession.Get(session.ManagerSessionKey, &manager) || !mcpSession.Get(types.AccountIDSessionKey, &accountID) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("session store or account not found")
	}

	var newSession = session.Session{
		Type:      "thread",
		SessionID: uuid.String(),
		AccountID: accountID,
	}
	if err := manager.DB.Create(ctx, &newSession); err != nil {
		return nil, err
	}

	return &types.Chat{
		ID:         newSession.SessionID,
		Title:      newSession.Description,
		Created:    newSession.CreatedAt,
		ReadOnly:   newSession.AccountID != accountID,
		Visibility: visibility(newSession.IsPublic),
	}, nil
}

func (s *Server) updateChat(ctx context.Context, data struct {
	ID    string `json:"chatId"`
	Title string `json:"title"`
}) (*types.Chat, error) {
	mcpSession := mcp.SessionFromContext(ctx)
	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return nil, err
	}

	chatSession, err := manager.DB.GetByIDByAccountID(ctx, data.ID, accountID)
	if err != nil {
		return nil, err
	}

	if data.Title != "" && chatSession.Description != data.Title {
		session, err := manager.DB.Get(ctx, data.ID)
		if err != nil {
			return nil, err
		}

		session.Description = data.Title
		if err := manager.DB.Update(ctx, session); err != nil {
			return nil, err
		}
		chatSession.Description = data.Title
	}

	chat := chatFromSession(chatSession, accountID)
	return &chat, nil
}

func (s *Server) getManagerAndAccountID(mcpSession *mcp.Session) (*session.Manager, string, error) {
	var (
		manager   session.Manager
		accountID string
	)

	if !mcpSession.Get(session.ManagerSessionKey, &manager) || !mcpSession.Get(types.AccountIDSessionKey, &accountID) {
		return nil, "", mcp.ErrRPCInvalidParams.WithMessage("session store or account not found")
	}
	return &manager, accountID, nil
}

func (s *Server) listChats(ctx context.Context, _ struct{}) (*types.ChatList, error) {
	mcpSession := mcp.SessionFromContext(ctx)

	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return nil, err
	}

	sessions, err := manager.DB.FindByAccount(ctx, "thread", accountID)
	if err != nil {
		return nil, err
	}

	chats := make([]types.Chat, 0, len(sessions))
	for _, s := range sessions {
		chats = append(chats, chatFromSession(&s, accountID))
	}

	return &types.ChatList{
		Chats: chats,
	}, nil
}

func (s *Server) getConfig(ctx context.Context, _ struct{}) (ret types.ProjectConfig, _ error) {
	session := mcp.SessionFromContext(ctx)
	session.Get("project", &ret)

	agents, err := s.data.Agents(ctx)
	if err != nil {
		return
	}

	ret.DefaultAgent = ""
	ret.Agents = make([]types.AgentDisplay, 0, len(agents))

	for id, agent := range agents {
		ret.Agents = append(ret.Agents, types.AgentDisplay{
			ID:          id,
			Name:        agent.Name,
			ShortName:   agent.ShortName,
			Description: agent.Description,
		})
	}

	ret.DefaultAgent = s.data.CurrentAgent(ctx)
	return
}

func (s *Server) updateConfig(ctx context.Context, cfg types.ProjectConfig) (types.ProjectConfig, error) {
	session := mcp.SessionFromContext(ctx)
	session = session.Parent
	session.Set("project", &cfg)
	return cfg, nil
}

func chatFromSession(session *session.Session, currentAccountID string) types.Chat {
	return types.Chat{
		ID:         session.SessionID,
		Title:      session.Description,
		Created:    session.CreatedAt,
		ReadOnly:   session.AccountID != currentAccountID,
		Visibility: visibility(session.IsPublic),
	}
}

func visibility(isPublic bool) string {
	if isPublic {
		return "public"
	}
	return "private"
}

func (s *Server) clone(ctx context.Context, _ struct{}) (string, error) {
	var (
		manager   session.Manager
		accountID string
	)

	mcpSession := mcp.SessionFromContext(ctx)
	if mcpSession.Parent != nil {
		mcpSession = mcpSession.Parent
	}

	if !mcpSession.Get(session.ManagerSessionKey, &manager) {
		return "", fmt.Errorf("session store not found")
	}

	if !mcpSession.Get(types.AccountIDSessionKey, &accountID) {
		return "", fmt.Errorf("account ID not found in session")
	}

	stored, err := manager.DB.Get(ctx, mcpSession.ID())
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	stored = stored.Clone(accountID)
	if err := manager.DB.Create(ctx, stored); err != nil {
		return "", fmt.Errorf("failed to create cloned session: %w", err)
	}

	return stored.SessionID, nil
}

func (s *Server) setVisibility(ctx context.Context, args struct {
	Visibility string `json:"visibility"`
}) (string, error) {
	if args.Visibility != "" && args.Visibility != "public" && args.Visibility != "private" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid visibility %q, must be \"public\" or \"private\"", args.Visibility)
	}

	mcpSession := mcp.SessionFromContext(ctx)
	if mcpSession.Parent != nil {
		mcpSession = mcpSession.Parent
	}

	isPublic := args.Visibility == "public"
	mcpSession.Set(types.PublicSessionKey, &isPublic)
	return "Thread visibility set to " + args.Visibility, nil
}
