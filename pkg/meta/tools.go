package meta

import (
	"context"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func (s *Server) setAgent(ctx context.Context, args struct {
	Agent string `json:"agent"`
}) (string, error) {
	if args.Agent != "" {
		if _, ok := s.data.Agents(ctx)[args.Agent]; !ok {
			return "", mcp.ErrRPCInvalidParams.WithMessage("agent %q does not exist", args.Agent)
		}
	}
	s.data.SetCurrentAgent(ctx, args.Agent)
	return s.data.CurrentAgent(ctx), nil
}

type ChatsData struct {
	Chats []ChatDescription `json:"chats"`
}

type ChatDescription struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
}

func (s *Server) listChats(ctx context.Context, _ struct{}) (*ChatsData, error) {
	mcpSession := mcp.SessionFromContext(ctx)
	var store session.Store

	if !mcpSession.Get(session.StoreSessionKey, &store) {
		return &ChatsData{
			Chats: []ChatDescription{},
		}, nil
	}

	session, err := store.Get(ctx, mcpSession.Parent.ID())
	if err != nil {
		return nil, err
	}

	sessions, err := store.FindByAccount(ctx, session.AccountID)
	if err != nil {
		return nil, err
	}

	chats := make([]ChatDescription, 0, len(sessions))
	for _, s := range sessions {
		var description string
		chats = append(chats, ChatDescription{
			ID:          s.SessionID,
			Description: description,
			Created:     s.CreatedAt,
		})
	}

	return &ChatsData{
		Chats: chats,
	}, nil
}

func (s *Server) getChat(ctx context.Context, _ struct{}) (*ChatData, error) {
	var run types.Execution
	session := mcp.SessionFromContext(ctx)
	session.Get(types.PreviousExecutionKey, &run)

	var (
		allMessages       []types.Message
		processedMessages []types.Message
	)
	if run.PopulatedRequest != nil {
		allMessages = run.PopulatedRequest.Input
	}
	if run.Response != nil {
		allMessages = append(allMessages, run.Response.Output)
	}

	tools := map[string]struct {
		msgIndex  int
		itemIndex int
	}{}
	for _, msg := range allMessages {
		var processedItems []types.CompletionItem
		for _, output := range msg.Items {
			if output.ToolCallResult != nil && output.ToolCall == nil {
				if i, ok := tools[output.ToolCallResult.CallID]; ok {
					targetMsg := msg
					if len(processedMessages) > i.msgIndex {
						targetMsg = processedMessages[i.msgIndex]
					}
					targetMsg.Items[i.itemIndex].ToolCallResult = output.ToolCallResult
					continue
				}
			} else if output.ToolCall != nil && output.ToolCallResult == nil {
				x := tools[output.ToolCall.CallID]
				x.msgIndex = len(processedMessages)
				x.itemIndex = len(processedItems)
				tools[output.ToolCall.CallID] = x
			}
			processedItems = append(processedItems, output)
		}
		msg.Items = processedItems
		if len(msg.Items) > 0 {
			processedMessages = append(processedMessages, msg)
		}
	}

	sessionID := session.ID()
	if session.Parent != nil {
		sessionID = session.Parent.ID()
	}

	prompts, err := s.data.PromptMappings(ctx, sessiondata.WithAllowMissing())
	if err != nil {
		return nil, err
	}

	return &ChatData{
		ID:           sessionID,
		Messages:     processedMessages,
		Tools:        run.ToolToMCPServer,
		Prompts:      prompts,
		CurrentAgent: s.data.CurrentAgent(ctx),
		Agents:       s.data.Agents(ctx),
	}, nil
}

type ChatData struct {
	ID           string               `json:"id"`
	Messages     []types.Message      `json:"messages"`
	Tools        types.ToolMappings   `json:"tools"`
	Prompts      types.PromptMappings `json:"prompts"`
	CurrentAgent string               `json:"currentAgent"`
	Agents       types.Agents         `json:"agents,omitempty"`
}
