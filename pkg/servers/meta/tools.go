package meta

import (
	"context"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func (s *Server) updateChat(ctx context.Context, data struct {
	ID    string `json:"chatId"`
	Title string `json:"title"`
}) (*types.Chat, error) {
	// Ensure rename operations trigger resource notifications for active UI sessions.
	s.ensureManagerEventSubscription(ctx)

	mcpSession := mcp.SessionFromContext(ctx)
	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return nil, err
	}

	chatSession, err := manager.DB.GetByIDByAccountID(ctx, data.ID, accountID)
	if err != nil {
		return nil, err
	}

	if data.Title != "" {
		updated, _, err := manager.UpdateThreadDescription(ctx, data.ID, accountID, data.Title)
		if err != nil {
			return nil, err
		}
		chatSession = updated
	}

	chat := chatFromSession(chatSession, accountID)
	return &chat, nil
}

func (s *Server) getManagerAndAccountID(mcpSession *mcp.Session) (*session.Manager, string, error) {
	var (
		manager   *session.Manager
		accountID string
	)

	if !mcpSession.Get(session.ManagerSessionKey, &manager) || manager == nil ||
		!mcpSession.Get(types.AccountIDSessionKey, &accountID) {
		return nil, "", mcp.ErrRPCInvalidParams.WithMessage("session store or account not found")
	}
	return manager, accountID, nil
}

func chatFromSession(s *session.Session, currentAccountID string) types.Chat {
	availableAgentIDs := availableAgentIDsFromSession(s)
	currentAgentID := currentAgentIDFromSession(s, availableAgentIDs)

	return types.Chat{
		ID:                s.SessionID,
		Title:             s.Description,
		Created:           s.CreatedAt,
		ReadOnly:          s.AccountID != currentAccountID,
		CurrentAgentID:    currentAgentID,
		AvailableAgentIDs: availableAgentIDs,
		WorkflowURIs:      s.WorkflowURIs,
	}
}

func availableAgentIDsFromSession(s *session.Session) []string {
	config := types.Config(s.Config)
	if len(config.Publish.Entrypoint) == 0 {
		return nil
	}

	// Preserve order from publish.entrypoint while dropping blanks and duplicates.
	result := make([]string, 0, len(config.Publish.Entrypoint))
	seen := map[string]struct{}{}
	for _, id := range config.Publish.Entrypoint {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}

func currentAgentIDFromSession(s *session.Session, availableAgentIDs []string) string {
	if s.State.Attributes != nil {
		if current, ok := s.State.Attributes[types.CurrentAgentSessionKey].(string); ok && strings.TrimSpace(current) != "" {
			return current
		}
	}

	if len(availableAgentIDs) > 0 {
		return availableAgentIDs[0]
	}
	return ""
}
