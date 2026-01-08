package mcp

import (
	"context"
	"sync"
)

// TestSession creates a new Session for testing purposes.
// It initializes the unexported fields necessary for the session to function
// in test environments without requiring a wire connection.
func TestSession(ctx context.Context) *Session {
	s := &Session{
		attributes: make(map[string]any),
		lock:       sync.Mutex{},
	}
	s.ctx, s.cancel = context.WithCancelCause(WithSession(ctx, s))
	return s
}
