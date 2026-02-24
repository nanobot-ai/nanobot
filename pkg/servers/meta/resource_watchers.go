package meta

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/fswatch"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

const fileWatchMaxDepth = 64

func (s *Server) ensureWatchers(ctx context.Context) error {
	if err := s.ensureWorkflowWatcher(ctx); err != nil {
		return err
	}
	return s.ensureFileWatchers(ctx)
}

func (s *Server) ensureWorkflowWatcher(ctx context.Context) error {
	s.workflowsWatcherLock.Lock()
	defer s.workflowsWatcherLock.Unlock()

	if s.workflowsWatcher != nil {
		return nil
	}

	workflowsPath := filepath.Join(".", workflowsDir)
	if err := os.MkdirAll(workflowsPath, 0o755); err != nil {
		return err
	}

	filter := func(relPath string, info os.FileInfo) bool {
		if info.IsDir() {
			return true
		}
		return filepath.Ext(relPath) == ".md"
	}

	watcher := fswatch.NewWatcher(workflowsPath, 0, filter, s.handleWorkflowEvents)
	if err := watcher.Start(); err != nil {
		return err
	}

	s.workflowsWatcher = watcher
	log.Debugf(ctx, "started meta workflow watcher for %s", workflowsPath)
	return nil
}

func (s *Server) ensureFileWatchers(ctx context.Context) error {
	mcpSession := mcp.SessionFromContext(ctx)
	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		// Context may be missing session manager during isolated tests/boot paths.
		return nil
	}

	sessions, err := manager.DB.FindByAccount(ctx, "thread", accountID)
	if err != nil {
		return err
	}

	desiredRoots := make(map[string]struct{}, len(sessions))
	for _, chatSession := range sessions {
		cwd := strings.TrimSpace(chatSession.Cwd)
		if cwd == "" {
			cwd = defaultSessionCwd(chatSession.SessionID)
		}

		absCwd, err := filepath.Abs(cwd)
		if err != nil {
			continue
		}
		info, err := os.Stat(absCwd)
		if err != nil || !info.IsDir() {
			continue
		}

		desiredRoots[absCwd] = struct{}{}
	}

	s.fileWatchersLock.Lock()
	defer s.fileWatchersLock.Unlock()

	for root, watcher := range s.fileWatchers {
		if _, ok := desiredRoots[root]; ok {
			continue
		}
		if err := watcher.Close(); err != nil {
			log.Debugf(ctx, "failed to close stale file watcher for %s: %v", root, err)
		}
		delete(s.fileWatchers, root)
	}

	for root := range desiredRoots {
		if _, ok := s.fileWatchers[root]; ok {
			continue
		}

		watcher := fswatch.NewWatcher(root, fileWatchMaxDepth, nil, s.handleFileEvents(root))
		if err := watcher.Start(); err != nil {
			log.Debugf(ctx, "failed to start file watcher for %s: %v", root, err)
			continue
		}
		s.fileWatchers[root] = watcher
		log.Debugf(ctx, "started meta file watcher for %s", root)
	}

	return nil
}

func (s *Server) handleFileEvents(root string) fswatch.EventHandler {
	return func(events []fswatch.Event) {
		for _, event := range events {
			path := filepath.Join(root, event.Path)
			uri := fileURI(path)

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
}

func (s *Server) handleWorkflowEvents(events []fswatch.Event) {
	for _, event := range events {
		workflowName := strings.TrimSuffix(filepath.ToSlash(event.Path), ".md")
		if workflowName == "" || workflowName == "." {
			continue
		}

		uri := workflowURIPrefix + workflowName
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

func (s *Server) closeWatchers() {
	s.fileWatchersLock.Lock()
	for root, watcher := range s.fileWatchers {
		_ = watcher.Close()
		delete(s.fileWatchers, root)
	}
	s.fileWatchersLock.Unlock()

	s.workflowsWatcherLock.Lock()
	if s.workflowsWatcher != nil {
		_ = s.workflowsWatcher.Close()
		s.workflowsWatcher = nil
	}
	s.workflowsWatcherLock.Unlock()
}
