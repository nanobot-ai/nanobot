package meta

import (
	"context"
	"fmt"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type ChatsData struct {
	Chats []ChatDescription `json:"chats"`
}

type ChatDescription struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Created    time.Time `json:"created"`
	ReadOnly   bool      `json:"readonly,omitempty"`
	Visibility string    `json:"visibility,omitempty"`
}

func (s *Server) listChats(ctx context.Context, _ struct{}) (*ChatsData, error) {
	mcpSession := mcp.SessionFromContext(ctx)
	var (
		store     session.Store
		accountID string
	)

	if !mcpSession.Get(session.StoreSessionKey, &store) || !mcpSession.Get(types.AccountIDSessionKey, &accountID) {
		return &ChatsData{
			Chats: []ChatDescription{},
		}, nil
	}

	session, err := store.Get(ctx, mcpSession.Parent.ID())
	if err != nil {
		return nil, err
	}

	sessions, err := store.FindByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	found := false
	chats := make([]ChatDescription, 0, len(sessions))
	for _, s := range sessions {
		if s.ID == session.ID {
			found = true
		}
		chats = append(chats, ChatDescription{
			ID:         s.SessionID,
			Title:      s.Description,
			Created:    s.CreatedAt,
			Visibility: visibility(s.IsPublic),
		})
	}

	if !found {
		chats = append([]ChatDescription{
			{
				ID:         session.SessionID,
				Title:      session.Description,
				Created:    session.CreatedAt,
				ReadOnly:   session.AccountID != accountID,
				Visibility: visibility(session.IsPublic),
			},
		}, chats...)
	}

	return &ChatsData{
		Chats: chats,
	}, nil
}

func visibility(isPublic bool) string {
	if isPublic {
		return "public"
	}
	return "private"
}

func (s *Server) clone(ctx context.Context, _ struct{}) (string, error) {
	var (
		store     session.Store
		accountID string
	)

	mcpSession := mcp.SessionFromContext(ctx)
	if mcpSession.Parent != nil {
		mcpSession = mcpSession.Parent
	}

	if !mcpSession.Get(session.StoreSessionKey, &store) {
		return "", fmt.Errorf("session store not found")
	}

	if !mcpSession.Get(types.AccountIDSessionKey, &accountID) {
		return "", fmt.Errorf("account ID not found in session")
	}

	stored, err := store.Get(ctx, mcpSession.ID())
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	stored = stored.Clone(accountID)
	if err := store.Create(ctx, stored); err != nil {
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
