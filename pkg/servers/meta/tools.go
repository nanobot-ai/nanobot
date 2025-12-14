package meta

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

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

func (s *Server) listAgents(ctx context.Context, _ struct{}) (*types.AgentList, error) {
	agents, err := s.data.Agents(ctx)
	if err != nil {
		return nil, err
	}
	return &types.AgentList{
		Agents: agents,
	}, nil
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

func chatFromSession(session *session.Session, currentAccountID string) types.Chat {
	return types.Chat{
		ID:       session.SessionID,
		Title:    session.Description,
		Created:  session.CreatedAt,
		ReadOnly: session.AccountID != currentAccountID,
	}
}
