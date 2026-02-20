package workflows

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nanobot-ai/nanobot/pkg/fswatch"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"gopkg.in/yaml.v3"
)

const (
	workflowsDir = "workflows"
)

type workflowMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	CreatedAt   string `yaml:"createdAt"`
}

// parseWorkflowFrontmatter extracts YAML frontmatter from workflow content.
// If no frontmatter is found (no opening ---), returns zero-value metadata with a nil error.
func parseWorkflowFrontmatter(content string) (workflowMeta, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return workflowMeta{}, nil
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return workflowMeta{}, fmt.Errorf("frontmatter missing closing delimiter")
	}

	frontmatterYAML := strings.Join(lines[1:endIdx], "\n")
	var meta workflowMeta
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &meta); err != nil {
		return workflowMeta{}, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return meta, nil
}

type Server struct {
	tools          mcp.ServerTools
	watcher        *fswatch.Watcher
	subscriptions  *fswatch.SubscriptionManager
	watcherOnce    sync.Once
	watcherInitErr error
}

func NewServer() *Server {
	s := &Server{
		subscriptions: fswatch.NewSubscriptionManager(context.Background()),
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("recordWorkflowRun", "Record that a workflow was executed in the current chat session", s.recordWorkflowRun),
		mcp.NewServerTool("deleteWorkflow", "Delete a workflow by its URI", s.deleteWorkflow),
	)

	return s
}

// Close stops the file watcher and cleans up resources
func (s *Server) Close() error {
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
	case "notifications/cancelled":
		mcp.HandleCancelled(ctx, msg)
	case "tools/list":
		mcp.Invoke(ctx, msg, s.tools.List)
	case "tools/call":
		mcp.Invoke(ctx, msg, s.tools.Call)
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
	s.subscriptions.AddSession(sessionID, msg.Session)

	// Start watcher when first session initializes
	if err := s.ensureWatcher(); err != nil {
		log.Errorf(ctx, "failed to start file watcher: %v", err)
	}

	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools: &mcp.ToolsServerCapability{},
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

func (s *Server) recordWorkflowRun(ctx context.Context, data struct {
	URI string `json:"uri"`
}) (*map[string]string, error) {
	mcpSession := mcp.SessionFromContext(ctx).Root()

	var uris []string
	mcpSession.Get(types.WorkflowURIsSessionKey, &uris)

	// Deduplicate: only append if URI is not already recorded
	var found bool
	for _, u := range uris {
		if u == data.URI {
			found = true
			break
		}
	}
	if !found {
		uris = append(uris, data.URI)
	}

	mcpSession.Set(types.WorkflowURIsSessionKey, uris)

	return &map[string]string{"uri": data.URI}, nil
}

func (s *Server) deleteWorkflow(ctx context.Context, data struct {
	URI string `json:"uri"`
}) (*struct{}, error) {
	workflowName, err := s.parseWorkflowURI(data.URI)
	if err != nil {
		return nil, err
	}

	workflowPath := filepath.Join(".", workflowsDir, workflowName+".md")
	if err := os.Remove(workflowPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to delete workflow: %w", err)
	}

	return &struct{}{}, nil
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

func (s *Server) resourcesList(ctx context.Context, msg mcp.Message, _ mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
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

		// Read file to extract description and metadata
		contentBytes, err := os.ReadFile(filepath.Join(workflowsPath, entry.Name()))
		if err != nil {
			// Skip files we can't read
			continue
		}

		meta, err := parseWorkflowFrontmatter(string(contentBytes))
		if err != nil {
			log.Debugf(ctx, "failed to parse frontmatter for workflow %s: %v", entry.Name(), err)
		}

		resourceMeta := make(map[string]any)
		if meta.Name != "" {
			resourceMeta["name"] = meta.Name
		}
		if meta.CreatedAt != "" {
			resourceMeta["createdAt"] = meta.CreatedAt
		}

		res := mcp.Resource{
			URI:         fmt.Sprintf("workflow:///%s", name),
			Name:        name,
			Description: meta.Description,
			MimeType:    "text/markdown",
		}
		if len(resourceMeta) > 0 {
			res.Meta = resourceMeta
		}

		result = append(result, res)
	}

	return &mcp.ListResourcesResult{Resources: result}, nil
}

func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	workflowName, err := s.parseWorkflowURI(request.URI)
	if err != nil {
		return nil, err
	}

	workflowPath := filepath.Join(".", workflowsDir, workflowName+".md")
	contentBytes, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", request.URI)
	}

	content := string(contentBytes)
	meta, err := parseWorkflowFrontmatter(content)
	if err != nil {
		log.Debugf(ctx, "failed to parse frontmatter for workflow %s: %v", workflowName, err)
	}

	resourceMeta := make(map[string]any)
	if meta.Name != "" {
		resourceMeta["name"] = meta.Name
	}
	if meta.CreatedAt != "" {
		resourceMeta["createdAt"] = meta.CreatedAt
	}

	rc := mcp.ResourceContent{
		URI:      request.URI,
		Name:     workflowName,
		MIMEType: "text/markdown",
		Text:     &content,
	}
	if len(resourceMeta) > 0 {
		rc.Meta = resourceMeta
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{rc},
	}, nil
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

	sessionID, _ := types.GetSessionAndAccountID(ctx)
	s.subscriptions.Subscribe(sessionID, msg.Session, request.URI)
	return &mcp.SubscribeResult{}, nil
}

func (s *Server) resourcesUnsubscribe(ctx context.Context, msg mcp.Message, request mcp.UnsubscribeRequest) (*mcp.UnsubscribeResult, error) {
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	s.subscriptions.Unsubscribe(sessionID, request.URI)
	return &mcp.UnsubscribeResult{}, nil
}

// ensureWatcher starts the file watcher if it hasn't been started yet
func (s *Server) ensureWatcher() error {
	s.watcherOnce.Do(func() {
		workflowsPath := filepath.Join(".", workflowsDir)

		// Ensure the workflows directory exists
		if err := os.MkdirAll(workflowsPath, 0755); err != nil {
			s.watcherInitErr = err
			return
		}

		// Create a filter that only accepts .md files
		filter := func(relPath string, info os.FileInfo) bool {
			if info.IsDir() {
				return true // Always allow directories
			}
			return filepath.Ext(relPath) == ".md"
		}

		// Create watcher with depth 0 (only watch workflows directory, not subdirectories)
		s.watcher = fswatch.NewWatcher(workflowsPath, 0, filter, s.handleFileEvents)
		if err := s.watcher.Start(); err != nil {
			s.watcherInitErr = err
			return
		}

		log.Debugf(context.Background(), "started workflow watcher for %s", workflowsPath)
	})

	return s.watcherInitErr
}

// handleFileEvents processes filesystem events from the watcher
func (s *Server) handleFileEvents(events []fswatch.Event) {
	for _, event := range events {
		// Convert filename to workflow URI
		workflowName := strings.TrimSuffix(event.Path, ".md")
		uri := fmt.Sprintf("workflow:///%s", workflowName)

		switch event.Type {
		case fswatch.EventDelete:
			// Send updated notification and list changed
			s.subscriptions.SendResourceUpdatedNotification(uri)
			s.subscriptions.AutoUnsubscribe(uri)
			s.subscriptions.SendListChangedNotification()

		case fswatch.EventCreate:
			// New workflow created - send list changed
			s.subscriptions.SendListChangedNotification()

		case fswatch.EventWrite:
			// Workflow modified - send updated notification
			s.subscriptions.SendResourceUpdatedNotification(uri)
		}
	}
}
