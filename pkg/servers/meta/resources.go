package meta

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/fileuri"
	"github.com/nanobot-ai/nanobot/pkg/fswatch"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"log/slog"
)

const sessionsDir = "sessions"

// resourcesList returns all resources (workflows + cross-session files).
func (s *Server) resourcesList(ctx context.Context, _ mcp.Message, _ mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	var resources []mcp.Resource

	// Add workflow resources
	workflowResources, err := s.listWorkflowResources(ctx)
	if err != nil {
		slog.Error("failed to list workflow resources", "error", err)
	} else {
		resources = append(resources, workflowResources...)
	}

	// Add cross-session file resources
	fileResources, err := s.listFileResourcesAllSessions(ctx)
	if err != nil {
		slog.Error("failed to list cross-session file resources", "error", err)
	} else {
		resources = append(resources, fileResources...)
	}

	return &mcp.ListResourcesResult{Resources: resources}, nil
}

// resourcesRead reads a resource by URI.
func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if strings.HasPrefix(request.URI, "workflow:///") {
		return s.readWorkflowResource(ctx, request.URI)
	} else if strings.HasPrefix(request.URI, "file:///") {
		return s.readFileResource(ctx, request.URI)
	}
	return nil, mcp.ErrRPCInvalidParams.WithMessage("unsupported resource URI: %s", request.URI)
}

// resourcesSubscribe subscribes to a resource.
func (s *Server) resourcesSubscribe(ctx context.Context, msg mcp.Message, request mcp.SubscribeRequest) (*mcp.SubscribeResult, error) {
	sessionID, _ := types.GetSessionAndAccountID(ctx)

	// Validate the URI
	if strings.HasPrefix(request.URI, "workflow:///") {
		workflowName := strings.TrimPrefix(request.URI, "workflow:///")
		workflowName = strings.TrimSuffix(workflowName, ".md")
		if workflowName == "" {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow name is required")
		}
		workflowPath := filepath.Join(".", skillformat.WorkflowsDir, workflowName, skillformat.SkillMainFile)
		if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", request.URI)
		}
	} else if strings.HasPrefix(request.URI, "file:///") {
		relPath, decodeErr := fileuri.Decode(request.URI)
		if decodeErr != nil {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("%v", decodeErr)
		}
		cleanPath := filepath.Clean(relPath)
		if strings.HasPrefix(cleanPath, skillformat.WorkflowsDir+string(filepath.Separator)) {
			// Workflow supporting file — verify it exists
			if _, statErr := os.Stat(filepath.Join(".", cleanPath)); os.IsNotExist(statErr) {
				return nil, mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", request.URI)
			}
		} else {
			// Session file — verify access
			if err := s.verifyFileResourceAccess(ctx, request.URI); err != nil {
				return nil, err
			}
		}
	} else {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("unsupported resource URI: %s", request.URI)
	}

	s.subscriptions.Subscribe(sessionID, msg.Session.Root(), request.URI)
	return &mcp.SubscribeResult{}, nil
}

// resourcesUnsubscribe unsubscribes from a resource.
func (s *Server) resourcesUnsubscribe(ctx context.Context, _ mcp.Message, request mcp.UnsubscribeRequest) (*mcp.UnsubscribeResult, error) {
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	s.subscriptions.Unsubscribe(sessionID, request.URI)
	return &mcp.UnsubscribeResult{}, nil
}

// listWorkflowResources reads the workflows/ directory and returns workflow resources.
func (s *Server) listWorkflowResources(ctx context.Context) ([]mcp.Resource, error) {
	workflowsPath := filepath.Join(".", skillformat.WorkflowsDir)

	entries, err := os.ReadDir(workflowsPath)
	if err != nil {
		// Directory doesn't exist - return empty list
		return nil, nil
	}

	var resources []mcp.Resource
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
				slog.Debug("failed to parse frontmatter for workflow", "workflow", entry.Name(), "error", err)
			}

			resourceMeta := make(map[string]any)
			if fm.Name != "" {
				resourceMeta["name"] = fm.Name
				resourceMeta["displayName"] = skillformat.DisplayName(fm.Name)
			}
			if fm.Metadata["createdAt"] != "" {
				resourceMeta["createdAt"] = fm.Metadata["createdAt"]
			}

			res := mcp.Resource{
				URI:         fmt.Sprintf("workflow:///%s", name),
				Name:        name,
				Description: fm.Description,
				MimeType:    "text/markdown",
			}
			if len(resourceMeta) > 0 {
				res.Meta = resourceMeta
			}

			resources = append(resources, res)
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
			resources = append(resources, mcp.Resource{
				URI:      fileuri.Encode(relPath),
				Name:     filepath.Base(relPath),
				MimeType: mimeType,
				Size:     info.Size(),
			})
			return nil
		})
	}

	return resources, nil
}

// readWorkflowResource reads a specific workflow by URI.
func (s *Server) readWorkflowResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	workflowName := strings.TrimPrefix(uri, "workflow:///")
	workflowName = strings.TrimSuffix(workflowName, ".md")
	if workflowName == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow name is required")
	}

	workflowPath := filepath.Join(".", skillformat.WorkflowsDir, workflowName, skillformat.SkillMainFile)
	contentBytes, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", uri)
	}

	content := string(contentBytes)
	fm, _, err := skillformat.ParseFrontmatter(content)
	if err != nil {
		slog.Debug("failed to parse frontmatter for workflow", "workflow", workflowName, "error", err)
	}

	resourceMeta := make(map[string]any)
	if fm.Name != "" {
		resourceMeta["name"] = fm.Name
		resourceMeta["displayName"] = skillformat.DisplayName(fm.Name)
	}
	if fm.Metadata["createdAt"] != "" {
		resourceMeta["createdAt"] = fm.Metadata["createdAt"]
	}

	rc := mcp.ResourceContent{
		URI:      uri,
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

// maxSessionFileDepth is the maximum depth for walking session directories.
const maxSessionFileDepth = 2

// sessionFileFilter determines if a file should be included in session file listing.
func sessionFileFilter(relPath string, info os.FileInfo) bool {
	if relPath == "." {
		return true
	}

	basename := filepath.Base(relPath)

	// Exclude common non-content directories/files
	excludedDirs := map[string]struct{}{
		"node_modules": {},
		".git":         {},
		".nanobot":     {},
	}
	excludedFiles := map[string]struct{}{
		".DS_Store": {},
	}

	if info.IsDir() {
		if _, excluded := excludedDirs[basename]; excluded {
			return false
		}
	} else {
		if _, excluded := excludedFiles[basename]; excluded {
			return false
		}
	}

	return true
}

// listFileResourcesAllSessions lists files across all sessions belonging to the current account.
func (s *Server) listFileResourcesAllSessions(ctx context.Context) ([]mcp.Resource, error) {
	mcpSession := mcp.SessionFromContext(ctx)
	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return nil, err
	}

	sessions, err := manager.DB.FindByAccount(ctx, "thread", accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	var resources []mcp.Resource
	for _, sess := range sessions {
		sessDir := filepath.Join(cwd, sessionsDir, sess.SessionID)

		// Skip if directory doesn't exist
		if _, err := os.Stat(sessDir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(sessDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			relPath, err := filepath.Rel(sessDir, path)
			if err != nil {
				return nil
			}

			if relPath == "." {
				return nil
			}

			depth := len(strings.Split(relPath, string(filepath.Separator)))
			if d.IsDir() && depth > maxSessionFileDepth {
				return filepath.SkipDir
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}
			if !sessionFileFilter(relPath, info) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if d.IsDir() {
				return nil
			}

			mimeType := mime.TypeByExtension(filepath.Ext(relPath))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}

			// URI format: file:///sessions/{sessionID}/{path}
			uri := fileuri.Encode(filepath.Join(sessionsDir, sess.SessionID, relPath))
			name := fmt.Sprintf("%s/%s", sess.SessionID, relPath)

			resources = append(resources, mcp.Resource{
				URI:      uri,
				Name:     name,
				MimeType: mimeType,
				Size:     info.Size(),
				Annotations: &mcp.Annotations{
					LastModified: info.ModTime(),
				},
			})

			return nil
		})
		if err != nil {
			slog.Error("failed to walk session directory", "path", sessDir, "error", err)
		}
	}

	return resources, nil
}

// verifyFileResourceAccess verifies that the file URI belongs to a session owned by the current account.
func (s *Server) verifyFileResourceAccess(ctx context.Context, relPath string) error {

	// Expected format: sessions/{sessionID}/...
	if !strings.HasPrefix(relPath, sessionsDir+"/") {
		return mcp.ErrRPCInvalidParams.WithMessage("invalid file URI format, expected file:///sessions/{sessionID}/path")
	}

	parts := strings.SplitN(strings.TrimPrefix(relPath, sessionsDir+"/"), "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return mcp.ErrRPCInvalidParams.WithMessage("invalid file URI format, expected file:///sessions/{sessionID}/path")
	}

	targetSessionID := parts[0]

	mcpSession := mcp.SessionFromContext(ctx)
	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return err
	}

	// Verify the target session belongs to this account
	if _, err := manager.DB.GetByIDByAccountID(ctx, targetSessionID, accountID); err != nil {
		return mcp.ErrRPCInvalidParams.WithMessage("session not found or access denied: %s", targetSessionID)
	}

	return nil
}

// readFileResource reads a file resource (workflow supporting files or cross-session files).
func (s *Server) readFileResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	relPath, err := fileuri.Decode(uri)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("%v", err)
	}

	// Prevent directory traversal: reject absolute paths and any ".." segments.
	if filepath.IsAbs(relPath) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid file path")
	}
	for _, segment := range strings.Split(relPath, "/") {
		if segment == ".." {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid file path")
		}
	}

	cleanPath := filepath.Clean(relPath)

	// Workflow supporting files don't need session access verification
	if !strings.HasPrefix(cleanPath, skillformat.WorkflowsDir+string(filepath.Separator)) {
		if err := s.verifyFileResourceAccess(ctx, relPath); err != nil {
			return nil, err
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	absPath := filepath.Join(cwd, relPath)

	f, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", uri)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
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
		rc.Blob = new(base64.StdEncoding.EncodeToString(content))
	} else if _, isPDF := types.PDFMimeTypes[mimeType]; isPDF {
		rc.Blob = new(base64.StdEncoding.EncodeToString(content))
	} else {
		rc.Text = new(string(content))
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{rc},
	}, nil
}

// ensureWatchers starts the workflow and sessions directory watchers.
func (s *Server) ensureWatchers() error {
	s.watcherOnce.Do(func() {
		// Watch workflows directory
		workflowsPath := filepath.Join(".", skillformat.WorkflowsDir)
		if err := os.MkdirAll(workflowsPath, 0755); err != nil {
			s.watcherInitErr = fmt.Errorf("failed to create workflows directory: %w", err)
			return
		}

		// Depth 3 to watch nested files like <workflow>/scripts/analyze.py
		s.workflowWatcher = fswatch.NewWatcher(workflowsPath, 3, nil, s.handleWorkflowEvents)
		if err := s.workflowWatcher.Start(); err != nil {
			s.watcherInitErr = fmt.Errorf("failed to start workflow watcher: %w", err)
			return
		}

		slog.Debug("started meta workflow watcher", "path", workflowsPath)

		// Watch sessions directory
		cwd, err := os.Getwd()
		if err != nil {
			slog.Error("failed to get working directory for sessions watcher", "error", err)
			return
		}

		sessionsPath := filepath.Join(cwd, sessionsDir)
		if err := os.MkdirAll(sessionsPath, 0755); err != nil {
			slog.Error("failed to create sessions directory", "path", sessionsPath, "error", err)
			return
		}

		s.sessionsWatcher = fswatch.NewWatcher(sessionsPath, 3, sessionFileFilter, s.handleSessionFileEvents)
		if err := s.sessionsWatcher.Start(); err != nil {
			slog.Error("failed to start sessions watcher", "error", err)
			return
		}

		slog.Debug("started meta sessions watcher", "path", sessionsPath)
	})

	return s.watcherInitErr
}

// handleWorkflowEvents processes filesystem events from the workflow watcher.
func (s *Server) handleWorkflowEvents(events []fswatch.Event) {
	for _, event := range events {
		// Event paths are relative to the workflows dir, e.g. "code-review/SKILL.md"
		// or "code-review/scripts/analyze.py"
		parts := strings.SplitN(event.Path, string(filepath.Separator), 2)
		workflowName := parts[0]
		workflowURI := fmt.Sprintf("workflow:///%s", workflowName)

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

// handleSessionFileEvents processes filesystem events from the sessions watcher.
func (s *Server) handleSessionFileEvents(events []fswatch.Event) {
	for _, event := range events {
		// event.Path is relative to the sessions directory, e.g. "{sessionID}/file.txt"
		uri := fileuri.Encode(filepath.Join(sessionsDir, event.Path))

		switch event.Type {
		case fswatch.EventDelete:
			s.subscriptions.SendResourceUpdatedNotification(uri)
			s.subscriptions.AutoUnsubscribe(uri)
			s.subscriptions.SendListChangedNotification()
		case fswatch.EventCreate:
			s.subscriptions.SendListChangedNotification()
		case fswatch.EventWrite:
			s.subscriptions.SendResourceUpdatedNotification(uri)
		}
	}
}
