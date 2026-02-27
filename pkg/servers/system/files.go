package system

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/fswatch"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

var (
	// fileCreatedAt gets a file's creation time.
	fileCreatedAt func(relPath string, info os.FileInfo) time.Time
	maxWatchDepth = 2
)

func init() {
	maxWatchDepthEnv := os.Getenv("NANOBOT_FILE_WATCH_MAX_DEPTH")
	if maxWatchDepthEnv != "" {
		if depth, err := strconv.Atoi(maxWatchDepthEnv); err == nil && depth > 0 {
			maxWatchDepth = depth
		}
	}
}

var (
	// Common directories/patterns to exclude from watching
	excludedDirs = map[string]struct{}{
		"node_modules": {},
		"vendor":       {},
		"__pycache__":  {},
		"dist":         {},
		"build":        {},
		"bin":          {},
		".git":         {},
		".svn":         {},
		".jj":          {},
		".vscode":      {},
		".idea":        {},
		".nanobot":     {},
		"sessions":     {},
	}

	excludedFiles = map[string]struct{}{
		"nanobot.db":         {},
		"nanobot.db-journal": {},
		".DS_Store":          {},
	}
)

// fileFilter determines if a file or directory should be included in file watching.
func fileFilter(relPath string, info os.FileInfo) bool {
	if relPath == "." {
		return true
	}

	basename := filepath.Base(relPath)

	// Check if basename is an excluded directory or file
	if info.IsDir() {
		if _, excluded := excludedDirs[basename]; excluded {
			return false
		}
	} else {
		if _, excluded := excludedFiles[basename]; excluded {
			return false
		}
	}

	// Check parent path components for directory exclusions
	parts := strings.Split(relPath, string(filepath.Separator))

	// Check all parts except the last (which we already checked above)
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		// Exclude specific directories
		if _, excluded := excludedDirs[part]; excluded {
			return false
		}
	}

	return true
}

// fileResourceMeta returns a given file's resource metadata.
func fileResourceMeta(relPath string, info os.FileInfo) map[string]any {
	if info == nil {
		return nil
	}

	var (
		modifiedAt = info.ModTime()
		meta       = make(map[string]any)
	)
	if !modifiedAt.IsZero() {
		meta["modifiedAt"] = formatTimestamp(modifiedAt)
	}

	if fileCreatedAt != nil {
		if createdAt := fileCreatedAt(relPath, info); !createdAt.IsZero() && !createdAt.After(modifiedAt) {
			meta["createdAt"] = formatTimestamp(createdAt)
		}
	}

	return meta
}

func formatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// handleFileEvents processes filesystem events from the watcher.
func (s *Server) handleFileEvents(events []fswatch.Event) {
	for _, event := range events {
		uri := "file:///" + event.Path

		switch event.Type {
		case fswatch.EventDelete:
			// Send updated notification and auto-unsubscribe
			s.subscriptions.SendResourceUpdatedNotification(uri)
			s.subscriptions.AutoUnsubscribe(uri)
			s.subscriptions.SendListChangedNotification()

		case fswatch.EventCreate:
			// New file created - send list changed
			s.subscriptions.SendListChangedNotification()

		case fswatch.EventWrite:
			// File modified - send updated notification
			s.subscriptions.SendResourceUpdatedNotification(uri)
		}
	}
}

// listFileResources returns all file resources in the session directory up to maxWatchDepth.
func (s *Server) listFileResources(ctx context.Context) ([]mcp.Resource, error) {
	var resources []mcp.Resource

	sessionID, _ := types.GetSessionAndAccountID(ctx)
	if sessionID == "" {
		return nil, nil
	}

	dir := sessionDir(sessionID)

	// If session directory doesn't exist yet, return empty list
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	// Walk directory tree
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}

		// Skip the root
		if relPath == "." {
			return nil
		}

		// Get depth
		depth := len(strings.Split(relPath, string(filepath.Separator)))
		if d.IsDir() && depth > maxWatchDepth {
			return filepath.SkipDir
		}

		// Apply filter
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if !fileFilter(relPath, info) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories (we only list files as resources)
		if d.IsDir() {
			return nil
		}

		// Determine MIME type
		mimeType := mime.TypeByExtension(filepath.Ext(relPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		resources = append(resources, mcp.Resource{
			URI:      "file:///" + relPath,
			Name:     relPath,
			MimeType: mimeType,
			Size:     info.Size(),
			Meta:     fileResourceMeta(relPath, info),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}

// readFileResource reads a file resource by URI, resolved against the session directory.
func (s *Server) readFileResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	// Parse file:/// URI
	if !strings.HasPrefix(uri, "file:///") {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid file URI, expected file:///path")
	}

	relPath := strings.TrimPrefix(uri, "file:///")
	if relPath == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("file path is required")
	}

	// Prevent directory traversal attacks
	cleanPath := filepath.Clean(relPath)
	if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid file path: cannot access files outside session directory")
	}

	// Resolve against session directory
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	if sessionID == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("session not found")
	}
	absPath := filepath.Join(sessionDir(sessionID), relPath)

	// Open file once to get both content and metadata
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

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Determine MIME type
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
		Meta:     fileResourceMeta(relPath, info),
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

// subscribeFileResource subscribes to a file resource.
func (s *Server) subscribeFileResource(ctx context.Context, uri string) error {
	// Parse file:/// URI
	if !strings.HasPrefix(uri, "file:///") {
		return mcp.ErrRPCInvalidParams.WithMessage("invalid file URI, expected file:///path")
	}

	relPath := strings.TrimPrefix(uri, "file:///")
	if relPath == "" {
		return mcp.ErrRPCInvalidParams.WithMessage("file path is required")
	}

	// Prevent directory traversal
	if filepath.IsAbs(relPath) {
		return mcp.ErrRPCInvalidParams.WithMessage("invalid file path: cannot access files outside session directory")
	}
	for _, segment := range strings.Split(relPath, "/") {
		if segment == ".." {
			return mcp.ErrRPCInvalidParams.WithMessage("invalid file path: cannot access files outside session directory")
		}
	}

	// Resolve against session directory
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	if sessionID == "" {
		return mcp.ErrRPCInvalidParams.WithMessage("session not found")
	}
	absPath := filepath.Join(sessionDir(sessionID), relPath)

	// Verify file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", uri)
	}

	return nil
}

// ensureFileWatcher starts the file watcher for a session's directory if not already started.
func (s *Server) ensureFileWatcher(sessionID string) error {
	if sessionID == "" {
		return nil
	}

	s.fileWatchersMu.Lock()
	defer s.fileWatchersMu.Unlock()

	if _, ok := s.fileWatchers[sessionID]; ok {
		return nil
	}

	dir, err := ensureSessionDir(sessionID)
	if err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	watcher := fswatch.NewWatcher(dir, maxWatchDepth, fileFilter, s.handleFileEvents)
	if err := watcher.Start(); err != nil {
		return err
	}

	log.Debugf(context.Background(), "started file watcher for session %s at %s with max depth %d", sessionID, dir, maxWatchDepth)

	s.fileWatchers[sessionID] = watcher
	return nil
}
