package workflows

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nanobot-ai/nanobot/pkg/fileuri"
	"github.com/nanobot-ai/nanobot/pkg/fswatch"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"log/slog"
)

type Server struct {
	watcher        *fswatch.Watcher
	subscriptions  *fswatch.SubscriptionManager
	watcherOnce    sync.Once
	watcherInitErr error
}

func NewServer() *Server {
	return &Server{
		subscriptions: fswatch.NewSubscriptionManager(context.Background()),
	}
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
	s.subscriptions.AddSession(sessionID, msg.Session.Root())

	// Start watcher when first session initializes
	if err := s.ensureWatcher(); err != nil {
		slog.Error("failed to start file watcher", "error", err)
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

// parseWorkflowURI extracts the workflow name from a workflow:///name URI
func parseWorkflowURI(uri string) (string, error) {
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
	workflowsPath := filepath.Join(".", skillformat.WorkflowsDir)

	entries, err := os.ReadDir(workflowsPath)
	if err != nil {
		// Directory doesn't exist or can't be read - return empty list
		return &mcp.ListResourcesResult{Resources: []mcp.Resource{}}, nil
	}

	var result []mcp.Resource
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		workflowDir := filepath.Join(workflowsPath, name)

		// Read the main workflow file from the subdirectory
		contentBytes, err := os.ReadFile(filepath.Join(workflowDir, skillformat.SkillMainFile))
		if err == nil {
			fm, _, err := skillformat.ParseFrontmatter(string(contentBytes))
			if err != nil {
				slog.Debug("failed to parse frontmatter for workflow", "workflow", name, "error", err)
			}

			resourceMeta := skillformat.FrontmatterToMeta(fm)

			res := mcp.Resource{
				URI:         fmt.Sprintf("workflow:///%s", name),
				Name:        name,
				Description: fm.Description,
				MimeType:    "text/markdown",
			}
			if len(resourceMeta) > 0 {
				res.Meta = resourceMeta
			}

			result = append(result, res)
		}

		// List supporting files in the workflow directory (even if SKILL.md doesn't exist yet)
		_ = filepath.WalkDir(workflowDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return nil
			}
			if filepath.Base(path) == skillformat.SkillMainFile {
				return nil
			}
			relPath, err := filepath.Rel(".", path)
			if err != nil {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			mimeType := mime.TypeByExtension(filepath.Ext(relPath))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			result = append(result, mcp.Resource{
				URI:      fileuri.Encode(relPath),
				Name:     filepath.Base(relPath),
				MimeType: mimeType,
				Size:     info.Size(),
			})
			return nil
		})
	}

	return &mcp.ListResourcesResult{Resources: result}, nil
}

func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if strings.HasPrefix(request.URI, "file:///") {
		return s.readWorkflowFile(request.URI)
	}

	workflowName, err := parseWorkflowURI(request.URI)
	if err != nil {
		return nil, err
	}

	workflowPath := filepath.Join(".", skillformat.WorkflowsDir, workflowName, skillformat.SkillMainFile)
	contentBytes, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", request.URI)
	}

	content := string(contentBytes)
	fm, _, err := skillformat.ParseFrontmatter(content)
	if err != nil {
		slog.Debug("failed to parse frontmatter for workflow", "workflow", workflowName, "error", err)
	}

	resourceMeta := skillformat.FrontmatterToMeta(fm)

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

// readWorkflowFile reads a supporting file from a workflow directory.
func (s *Server) readWorkflowFile(uri string) (*mcp.ReadResourceResult, error) {
	relPath, err := fileuri.Decode(uri)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("%v", err)
	}

	cleanPath := filepath.Clean(relPath)

	// Validate the path is under workflows/ and has no traversal
	if !strings.HasPrefix(cleanPath, skillformat.WorkflowsDir+string(filepath.Separator)) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("file not in workflows directory: %s", uri)
	}
	for _, segment := range strings.Split(cleanPath, string(filepath.Separator)) {
		if segment == ".." {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid file path: %s", uri)
		}
	}

	contentBytes, err := os.ReadFile(filepath.Join(".", cleanPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", uri)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(relPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if i := strings.IndexByte(mimeType, ';'); i >= 0 {
		mimeType = strings.TrimSpace(mimeType[:i])
	}

	rc := mcp.ResourceContent{
		URI:      uri,
		Name:     filepath.Base(relPath),
		MIMEType: mimeType,
	}

	if _, isImage := types.ImageMimeTypes[mimeType]; isImage {
		blob := base64.StdEncoding.EncodeToString(contentBytes)
		rc.Blob = &blob
	} else if _, isPDF := types.PDFMimeTypes[mimeType]; isPDF {
		blob := base64.StdEncoding.EncodeToString(contentBytes)
		rc.Blob = &blob
	} else {
		text := string(contentBytes)
		rc.Text = &text
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{rc},
	}, nil
}

func (s *Server) resourcesSubscribe(ctx context.Context, msg mcp.Message, request mcp.SubscribeRequest) (*mcp.SubscribeResult, error) {
	if strings.HasPrefix(request.URI, "file:///") {
		relPath, err := fileuri.Decode(request.URI)
		if err != nil {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("%v", err)
		}
		cleanPath := filepath.Clean(relPath)
		if !strings.HasPrefix(cleanPath, skillformat.WorkflowsDir+string(filepath.Separator)) {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("file not in workflows directory: %s", request.URI)
		}
		if _, err := os.Stat(filepath.Join(".", cleanPath)); os.IsNotExist(err) {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", request.URI)
		}
	} else {
		workflowName, err := parseWorkflowURI(request.URI)
		if err != nil {
			return nil, err
		}
		workflowPath := filepath.Join(".", skillformat.WorkflowsDir, workflowName, skillformat.SkillMainFile)
		if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", request.URI)
		}
	}

	sessionID, _ := types.GetSessionAndAccountID(ctx)
	s.subscriptions.Subscribe(sessionID, msg.Session.Root(), request.URI)
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
		workflowsPath := filepath.Join(".", skillformat.WorkflowsDir)

		// Ensure the workflows directory exists
		if err := os.MkdirAll(workflowsPath, 0755); err != nil {
			s.watcherInitErr = err
			return
		}

		// Depth 3 to watch nested files like <workflow>/scripts/analyze.py
		s.watcher = fswatch.NewWatcher(workflowsPath, 3, nil, s.handleFileEvents)
		if err := s.watcher.Start(); err != nil {
			s.watcherInitErr = err
			return
		}

		slog.Debug("started workflow watcher", "path", workflowsPath)
	})

	return s.watcherInitErr
}

// handleFileEvents processes filesystem events from the watcher
func (s *Server) handleFileEvents(events []fswatch.Event) {
	for _, event := range events {
		// Event paths are relative to the workflows dir, e.g. "code-review/SKILL.md"
		// or "code-review/scripts/analyze.py"
		parts := strings.SplitN(event.Path, string(filepath.Separator), 2)
		workflowName := parts[0]
		workflowURI := fmt.Sprintf("workflow:///%s", workflowName)

		// Determine if this is the main workflow file or a supporting file
		isMainFile := len(parts) == 2 && parts[1] == skillformat.SkillMainFile

		switch event.Type {
		case fswatch.EventDelete:
			if isMainFile {
				s.subscriptions.SendResourceUpdatedNotification(workflowURI)
				s.subscriptions.AutoUnsubscribe(workflowURI)
			} else if len(parts) == 2 {
				fileURI := fileuri.Encode(filepath.Join(skillformat.WorkflowsDir, event.Path))
				s.subscriptions.SendResourceUpdatedNotification(fileURI)
				s.subscriptions.AutoUnsubscribe(fileURI)
			}
			s.subscriptions.SendListChangedNotification()

		case fswatch.EventCreate:
			s.subscriptions.SendListChangedNotification()

		case fswatch.EventWrite:
			if isMainFile {
				s.subscriptions.SendResourceUpdatedNotification(workflowURI)
			} else if len(parts) == 2 {
				fileURI := fileuri.Encode(filepath.Join(skillformat.WorkflowsDir, event.Path))
				s.subscriptions.SendResourceUpdatedNotification(fileURI)
			}
		}
	}
}
