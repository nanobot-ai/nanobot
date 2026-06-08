package mcp

import (
	"context"
	"testing"
)

// recordingCloser records the deleteSession values it was closed with.
type recordingCloser struct {
	closed []bool
}

func (r *recordingCloser) Close(deleteSession bool) {
	r.closed = append(r.closed, deleteSession)
}

// TestSessionCloseClosesAttributeClosers verifies that closing a session also
// closes any attribute value that owns a closable resource (such as a
// downstream MCP client factory). Without this, downstream clients stored in
// session attributes are orphaned when the session is torn down: their upstream
// MCP session is never deleted (no DELETE is sent) and their goroutines leak.
func TestSessionCloseClosesAttributeClosers(t *testing.T) {
	for _, deleteSession := range []bool{true, false} {
		sess := NewEmptySession(context.Background())

		rc := &recordingCloser{}
		sess.Set("clients/grocy", rc)
		// A non-closer attribute must be left untouched (and must not panic).
		sess.Set("config-hash", SavedString("abc123"))

		sess.Close(deleteSession)

		if len(rc.closed) != 1 {
			t.Fatalf("deleteSession=%v: expected closable attribute to be closed exactly once, got %d calls (%v)",
				deleteSession, len(rc.closed), rc.closed)
		}
		if rc.closed[0] != deleteSession {
			t.Fatalf("deleteSession=%v: expected closer to receive %v, got %v",
				deleteSession, deleteSession, rc.closed[0])
		}
	}
}
