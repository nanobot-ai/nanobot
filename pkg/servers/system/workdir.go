package system

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func (s *Server) baseWorkdir() string {
	if s.baseCwd != "" {
		return s.baseCwd
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func (s *Server) defaultSessionWorkdir(sessionID string) string {
	return filepath.Join(s.baseWorkdir(), "sessions", types.SanitizeSessionDirectoryName(sessionID))
}

func (s *Server) sessionWorkdir(ctx context.Context) string {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return s.baseWorkdir()
	}

	root := session.Root()

	var cwd string
	if root.Get(types.CwdSessionKey, &cwd) && cwd != "" {
		return cwd
	}

	// Lazily assign a per-session default so callers can use it immediately.
	cwd = s.defaultSessionWorkdir(root.ID())
	root.Set(types.CwdSessionKey, cwd)
	return cwd
}

func (s *Server) ensureSessionWorkdir(ctx context.Context) (string, error) {
	workdir := s.sessionWorkdir(ctx)
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create session workdir %s: %w", workdir, err)
	}
	return workdir, nil
}

func resolvePath(basePath, candidate string) string {
	if filepath.IsAbs(candidate) {
		return candidate
	}
	// Keep all relative file operations scoped to the session workdir.
	return filepath.Join(basePath, candidate)
}
