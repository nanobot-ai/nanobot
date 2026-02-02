package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

const (
	workflowsDir    = "workflows"
	debounceTimeout = 100 * time.Millisecond
)

// subscription holds the session reference and subscribed URIs for that session
type subscription struct {
	session *mcp.Session
	uris    map[string]struct{}
}

type Server struct {
	mu            sync.RWMutex
	watcher       *fsnotify.Watcher
	subscriptions map[string]*subscription // sessionID -> subscription
	sessions      map[string]*mcp.Session  // sessionID -> session (all initialized sessions)
	watcherCtx    context.Context
	watcherCancel context.CancelFunc
	watcherOnce   sync.Once
}

func NewServer() *Server {
	return &Server{
		subscriptions: make(map[string]*subscription),
		sessions:      make(map[string]*mcp.Session),
	}
}

// Close stops the file watcher and cleans up resources
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watcherCancel != nil {
		s.watcherCancel()
	}
	if s.watcher != nil {
		return s.watcher.Close()
	}
	return nil
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
	case "resources/list":
		mcp.Invoke(ctx, msg, s.resourcesList)
	case "resources/read":
		mcp.Invoke(ctx, msg, s.resourcesRead)
	case "resources/subscribe":
		mcp.Invoke(ctx, msg, s.resourcesSubscribe)
	case "resources/unsubscribe":
		mcp.Invoke(ctx, msg, s.resourcesUnsubscribe)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%v", msg.Method))
	}
}

func (s *Server) initialize(ctx context.Context, msg mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	// Track this session for sending list_changed notifications
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	s.mu.Lock()
	s.sessions[sessionID] = msg.Session
	s.mu.Unlock()

	// Clean up session when it ends
	context.AfterFunc(msg.Session.Context(), func() {
		s.mu.Lock()
		delete(s.sessions, sessionID)
		s.mu.Unlock()
	})

	// Start watcher when first session initializes
	if err := s.ensureWatcher(); err != nil {
		log.Errorf(ctx, "failed to start file watcher: %v", err)
	}

	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Resources: &mcp.ResourcesServerCapability{
				Subscribe:   true,
				ListChanged: true,
			},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}

// parseWorkflowDescription extracts description from workflow markdown content.
// Looks for: # Workflow: <name>\n\n<description paragraph>
func parseWorkflowDescription(content string) string {
	lines := strings.Split(content, "\n")

	// Find the "# Workflow:" line
	startIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# Workflow:") {
			startIdx = i + 1
			break
		}
	}

	if startIdx == -1 || startIdx >= len(lines) {
		return ""
	}

	// Skip empty lines after header
	for startIdx < len(lines) && strings.TrimSpace(lines[startIdx]) == "" {
		startIdx++
	}

	if startIdx >= len(lines) {
		return ""
	}

	// Collect description paragraph (until empty line or next header)
	var desc []string
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			break
		}
		desc = append(desc, trimmed)
	}

	return strings.Join(desc, " ")
}

func (s *Server) resourcesList(ctx context.Context, _ mcp.Message, _ mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	workflowsPath := filepath.Join(".", workflowsDir)

	entries, err := os.ReadDir(workflowsPath)
	if err != nil {
		// Directory doesn't exist or can't be read - return empty list
		return &mcp.ListResourcesResult{Resources: []mcp.Resource{}}, nil
	}

	var result []mcp.Resource
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")

		// Read file to extract description
		content, err := os.ReadFile(filepath.Join(workflowsPath, entry.Name()))
		if err != nil {
			// Skip files we can't read
			continue
		}

		description := parseWorkflowDescription(string(content))

		result = append(result, mcp.Resource{
			URI:         fmt.Sprintf("workflow:///%s", name),
			Name:        name,
			Description: description,
			MimeType:    "text/markdown",
		})
	}

	return &mcp.ListResourcesResult{Resources: result}, nil
}

func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	workflowName, err := s.parseWorkflowURI(request.URI)
	if err != nil {
		return nil, err
	}

	workflowPath := filepath.Join(".", workflowsDir, workflowName+".md")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", request.URI)
	}

	contentStr := string(content)
	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{
			{
				URI:      request.URI,
				Name:     workflowName,
				MIMEType: "text/markdown",
				Text:     &contentStr,
			},
		},
	}, nil
}

// parseWorkflowURI extracts the workflow name from a workflow:///name URI
func (s *Server) parseWorkflowURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, "workflow:///") {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid workflow URI format, expected workflow:///name")
	}

	workflowName := strings.TrimPrefix(uri, "workflow:///")
	if workflowName == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("workflow name is required")
	}

	// Remove .md extension if present (we'll add it back when needed)
	workflowName = strings.TrimSuffix(workflowName, ".md")

	return workflowName, nil
}

func (s *Server) resourcesSubscribe(ctx context.Context, msg mcp.Message, request mcp.SubscribeRequest) (*mcp.SubscribeResult, error) {
	workflowName, err := s.parseWorkflowURI(request.URI)
	if err != nil {
		return nil, err
	}

	// Verify the workflow file exists
	workflowPath := filepath.Join(".", workflowsDir, workflowName+".md")
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", request.URI)
	}

	// Ensure the watcher is running
	if err := s.ensureWatcher(); err != nil {
		return nil, mcp.ErrRPCInternal.WithMessage("failed to start file watcher: %v", err)
	}

	sessionID, _ := types.GetSessionAndAccountID(ctx)

	// Add subscription
	s.mu.Lock()
	defer s.mu.Unlock()

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

	sessionID := msg.Session.ID()
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

// ensureWatcher starts the file watcher if it hasn't been started yet
func (s *Server) ensureWatcher() error {
	var initErr error
	s.watcherOnce.Do(func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			initErr = err
			return
		}

		workflowsPath := filepath.Join(".", workflowsDir)

		// Ensure the workflows directory exists
		if err := os.MkdirAll(workflowsPath, 0755); err != nil {
			watcher.Close()
			initErr = err
			return
		}

		if err := watcher.Add(workflowsPath); err != nil {
			watcher.Close()
			initErr = err
			return
		}

		s.watcher = watcher
		s.watcherCtx, s.watcherCancel = context.WithCancel(context.Background())

		go s.watchLoop()
	})

	return initErr
}

// watchLoop processes file system events and sends notifications
func (s *Server) watchLoop() {
	// Debounce map: filename -> last event time
	debounce := make(map[string]time.Time)
	debounceMu := sync.Mutex{}

	// Timer for processing debounced events
	ticker := time.NewTicker(debounceTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-s.watcherCtx.Done():
			return

		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			// Only process write, create, remove, and rename events for .md files
			if filepath.Ext(event.Name) != ".md" {
				continue
			}

			if !event.Op.Has(fsnotify.Write | fsnotify.Create | fsnotify.Remove | fsnotify.Rename) {
				continue
			}

			// Record the event for debouncing
			debounceMu.Lock()
			debounce[event.Name] = time.Now()
			isRemove := event.Op.Has(fsnotify.Remove | fsnotify.Rename)
			isCreate := event.Op.Has(fsnotify.Create)
			debounceMu.Unlock()

			// If it's a remove event, handle it immediately (no more writes expected)
			if isRemove {
				s.handleFileChange(event.Name, true, true)
				debounceMu.Lock()
				delete(debounce, event.Name)
				debounceMu.Unlock()
			} else if isCreate {
				// Handle create events immediately
				s.handleFileChange(event.Name, false, true)
				debounceMu.Lock()
				delete(debounce, event.Name)
				debounceMu.Unlock()
			}

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Errorf(s.watcherCtx, "workflow watcher error: %v", err)

		case <-ticker.C:
			// Process debounced events
			now := time.Now()
			debounceMu.Lock()
			for filename, eventTime := range debounce {
				if now.Sub(eventTime) >= debounceTimeout {
					delete(debounce, filename)
					// Process in goroutine to avoid blocking
					// Write events don't change the list
					go s.handleFileChange(filename, false, false)
				}
			}
			debounceMu.Unlock()
		}
	}
}

// handleFileChange sends notifications to all subscribed sessions for the changed file
// and sends list_changed notifications to all sessions if the list changed (create/delete)
func (s *Server) handleFileChange(filename string, isDelete bool, listChanged bool) {
	// Convert filename to workflow URI
	basename := filepath.Base(filename)
	workflowName := strings.TrimSuffix(basename, ".md")
	uri := fmt.Sprintf("workflow:///%s", workflowName)

	// Only send updated notifications for write/delete events, not create
	if !listChanged || isDelete {
		s.mu.RLock()
		// Copy the sessions that need notification
		var (
			sessionsToNotify      []*mcp.Session
			sessionsToUnsubscribe []string
		)
		for sessionID, sub := range s.subscriptions {
			if _, ok := sub.uris[uri]; ok {
				sessionsToNotify = append(sessionsToNotify, sub.session)
				if isDelete {
					sessionsToUnsubscribe = append(sessionsToUnsubscribe, sessionID)
				}
			}
		}
		s.mu.RUnlock()

		// Send notifications
		for _, session := range sessionsToNotify {
			s.sendResourceUpdatedNotification(session, uri)
		}

		// Auto-unsubscribe deleted resources
		if isDelete && len(sessionsToUnsubscribe) > 0 {
			s.mu.Lock()
			for _, sessionID := range sessionsToUnsubscribe {
				if sub, ok := s.subscriptions[sessionID]; ok {
					delete(sub.uris, uri)
					if len(sub.uris) == 0 {
						delete(s.subscriptions, sessionID)
					}
				}
			}
			s.mu.Unlock()
		}
	}

	// Send list_changed notification to all sessions if the list changed
	if listChanged {
		s.sendListChangedNotification()
	}
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
		log.Errorf(s.watcherCtx, "failed to marshal notification params: %v", err)
		return
	}
	notification.Params = paramsBytes

	if err := session.Send(s.watcherCtx, notification); err != nil {
		log.Errorf(s.watcherCtx, "failed to send resource updated notification: %v", err)
	}
}

// sendListChangedNotification sends a notifications/resources/list_changed message to all sessions
func (s *Server) sendListChangedNotification() {
	notification := mcp.Message{
		JSONRPC: "2.0",
		Method:  "notifications/resources/list_changed",
	}

	// Get all sessions
	s.mu.RLock()
	sessions := make([]*mcp.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	s.mu.RUnlock()

	// Send to all sessions
	for _, session := range sessions {
		if err := session.Send(s.watcherCtx, notification); err != nil && !errors.Is(err, mcp.ErrNoReader) {
			log.Errorf(s.watcherCtx, "failed to send list_changed notification: %v", err)
		}
	}
}
