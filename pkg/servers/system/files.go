package system

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/fswatch"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

var maxWatchDepth = 2

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

// listFileResources returns all file resources in the working directory up to maxWatchDepth.
func (s *Server) listFileResources() ([]mcp.Resource, error) {
	var resources []mcp.Resource

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Walk directory tree
	err = filepath.WalkDir(cwd, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Get relative path
		relPath, err := filepath.Rel(cwd, path)
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
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}

// readFileResource reads a file resource by URI.
func (s *Server) readFileResource(uri string) (*mcp.ReadResourceResult, error) {
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
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid file path: cannot access files outside working directory")
	}

	// Read file
	content, err := os.ReadFile(relPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", uri)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine MIME type
	mimeType := mime.TypeByExtension(filepath.Ext(relPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	contentStr := string(content)
	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{
			{
				URI:      uri,
				Name:     filepath.Base(relPath),
				MIMEType: mimeType,
				Text:     &contentStr,
			},
		},
	}, nil
}

// subscribeFileResource subscribes to a file resource.
func (s *Server) subscribeFileResource(uri string) error {
	// Parse file:/// URI
	if !strings.HasPrefix(uri, "file:///") {
		return mcp.ErrRPCInvalidParams.WithMessage("invalid file URI, expected file:///path")
	}

	relPath := strings.TrimPrefix(uri, "file:///")
	if relPath == "" {
		return mcp.ErrRPCInvalidParams.WithMessage("file path is required")
	}

	// Verify file exists
	if _, err := os.Stat(relPath); os.IsNotExist(err) {
		return mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", uri)
	}

	return nil
}

// ensureFileWatcher starts the file watcher if not already started.
func (s *Server) ensureFileWatcher() error {
	s.fileWatcherOnce.Do(func() {
		cwd, err := os.Getwd()
		if err != nil {
			s.fileWatcherInitErr = fmt.Errorf("failed to get working directory: %w", err)
			return
		}

		s.fileWatcher = fswatch.NewWatcher(cwd, maxWatchDepth, fileFilter, s.handleFileEvents)
		if err := s.fileWatcher.Start(); err != nil {
			s.fileWatcherInitErr = err
			return
		}

		log.Debugf(context.Background(), "started file watcher for %s with max depth %d", cwd, maxWatchDepth)
	})

	return s.fileWatcherInitErr
}
