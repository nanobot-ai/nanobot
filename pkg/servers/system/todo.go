package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"activeForm"`
}

// TodoWrite tool
type TodoWriteParams struct {
	Todos []TodoItem `json:"todos"`
}

func (s *Server) resourcesList(ctx context.Context, _ mcp.Message, _ mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	// Always return one resource (todo:///list) even if the file doesn't exist
	return &mcp.ListResourcesResult{
		Resources: []mcp.Resource{
			{
				URI:         "todo:///list",
				Name:        "Todo List",
				Description: "The current todo list for tracking tasks",
				MimeType:    "application/json",
			},
		},
	}, nil
}

func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if request.URI != "todo:///list" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid todo URI, expected todo:///list")
	}

	// Get session ID
	sessionID, _ := types.GetSessionAndAccountID(ctx)

	// Read from .nanobot/<sessionId>/status/todo.json
	todoPath := filepath.Join(".nanobot", sessionID, "status", "todo.json")

	// Check if file exists
	var contentStr string
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		// Return empty list if file doesn't exist
		contentStr = "[]"
	} else {
		// Read file
		data, err := os.ReadFile(todoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read todo file: %w", err)
		}
		contentStr = string(data)
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{
			{
				URI:      request.URI,
				Name:     "Todo List",
				MIMEType: "application/json",
				Text:     &contentStr,
			},
		},
	}, nil
}

func (s *Server) resourcesSubscribe(ctx context.Context, msg mcp.Message, request mcp.SubscribeRequest) (*mcp.SubscribeResult, error) {
	if request.URI != "todo:///list" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid todo URI, expected todo:///list")
	}

	// Add subscription
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID, _ := types.GetSessionAndAccountID(ctx)
	sub, ok := s.subscriptions[sessionID]
	if !ok {
		sub = &subscription{
			session: msg.Session,
			uris:    make(map[string]struct{}),
		}
		s.subscriptions[sessionID] = sub

		context.AfterFunc(msg.Session.Context(), func() {
			// Clean up subscriptions when session ends
			s.mu.Lock()
			delete(s.subscriptions, sessionID)
			s.mu.Unlock()
		})
	}
	sub.uris[request.URI] = struct{}{}

	return &mcp.SubscribeResult{}, nil
}

func (s *Server) resourcesUnsubscribe(ctx context.Context, msg mcp.Message, request mcp.UnsubscribeRequest) (*mcp.UnsubscribeResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID, _ := types.GetSessionAndAccountID(ctx)
	sub, ok := s.subscriptions[sessionID]
	if !ok {
		// No subscriptions for this session, nothing to do
		return &mcp.UnsubscribeResult{}, nil
	}

	delete(sub.uris, request.URI)

	// Clean up empty subscription entries
	if len(sub.uris) == 0 {
		delete(s.subscriptions, sessionID)
	}

	return &mcp.UnsubscribeResult{}, nil
}

// sendResourceUpdatedNotification sends a notifications/resources/updated message
func (s *Server) sendResourceUpdatedNotification(session *mcp.Session, uri string) {
	notification := mcp.Message{
		JSONRPC: "2.0",
		Method:  "notifications/resources/updated",
	}

	// Create the params with the URI
	params := struct {
		URI string `json:"uri"`
	}{
		URI: uri,
	}

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		log.Errorf(context.Background(), "failed to marshal notification params: %v", err)
		return
	}
	notification.Params = paramsBytes

	if err := session.Send(context.Background(), notification); err != nil {
		log.Errorf(context.Background(), "failed to send resource updated notification: %v", err)
	}
}

func (s *Server) todoWrite(ctx context.Context, params TodoWriteParams) (string, error) {
	// Validate only one in_progress task
	var inProgressCount int
	for _, todo := range params.Todos {
		if todo.Status == "in_progress" {
			inProgressCount++
		}
	}

	if inProgressCount > 1 {
		return "", mcp.ErrRPCInvalidParams.WithMessage("only one task can be in_progress at a time")
	}

	// Get session ID
	sessionID, _ := types.GetSessionAndAccountID(ctx)

	// Write to .nanobot/<sessionId>/status/todo.json
	todoPath := filepath.Join(".nanobot", sessionID, "status", "todo.json")

	// Create directories
	if err := os.MkdirAll(filepath.Dir(todoPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create todo directory: %w", err)
	}

	// Marshal JSON
	todoJSON, err := json.MarshalIndent(params.Todos, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal todos: %w", err)
	}

	// Write file
	if err := os.WriteFile(todoPath, todoJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write todo file: %w", err)
	}

	// Send resource updated notification to subscribed sessions
	s.mu.RLock()
	sub := s.subscriptions[sessionID]
	s.mu.RUnlock()

	if sub != nil {
		go s.sendResourceUpdatedNotification(sub.session, "todo:///list")
	}

	return fmt.Sprintf("Todo list updated:\n\n%s", string(todoJSON)), nil
}
