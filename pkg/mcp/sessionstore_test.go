package mcp

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestInMemorySessionStoreReapsIdleSessions(t *testing.T) {
	store := NewInMemorySessionStoreWithOptions(InMemorySessionStoreOptions{
		IdleTTL:      20 * time.Millisecond,
		ReapInterval: 5 * time.Millisecond,
		MaxSessions:  4,
	})
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close session store: %v", err)
		}
	})

	session := newSessionStoreTestSession(t, "idle")
	if err := store.Store(context.Background(), session.ID(), session); err != nil {
		t.Fatal(err)
	}
	store.Release(session)

	waitForSessionClose(t, session, time.Second)

	if _, found, err := store.Acquire(context.Background(), nil, session.ID()); err != nil {
		t.Fatal(err)
	} else if found {
		t.Fatal("expired session remained available")
	}
}

func TestInMemorySessionStoreDoesNotReapAcquiredSessions(t *testing.T) {
	store := NewInMemorySessionStoreWithOptions(InMemorySessionStoreOptions{
		IdleTTL:      20 * time.Millisecond,
		ReapInterval: 5 * time.Millisecond,
		MaxSessions:  4,
	})
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close session store: %v", err)
		}
	})

	session := newSessionStoreTestSession(t, "acquired")
	if err := store.Store(context.Background(), session.ID(), session); err != nil {
		t.Fatal(err)
	}
	store.Release(session)

	acquired, found, err := store.Acquire(context.Background(), nil, session.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("session was not found")
	}

	time.Sleep(4 * 20 * time.Millisecond)
	select {
	case <-session.GetSession().Context().Done():
		t.Fatal("acquired session was reaped")
	default:
	}

	store.Release(acquired)
	waitForSessionClose(t, session, time.Second)
}

func TestInMemorySessionStoreEvictsLeastRecentlyUsedIdleSessionAtCapacity(t *testing.T) {
	store := NewInMemorySessionStoreWithOptions(InMemorySessionStoreOptions{
		IdleTTL:      time.Hour,
		ReapInterval: time.Minute,
		MaxSessions:  2,
	})
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close session store: %v", err)
		}
	})

	first := newSessionStoreTestSession(t, "first")
	second := newSessionStoreTestSession(t, "second")
	third := newSessionStoreTestSession(t, "third")

	storeSessionThenRelease(t, store, first)
	time.Sleep(time.Millisecond)
	storeSessionThenRelease(t, store, second)

	touched, found, err := store.Acquire(context.Background(), nil, first.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("first session was not found")
	}
	store.Release(touched)

	if err := store.Store(context.Background(), third.ID(), third); err != nil {
		t.Fatal(err)
	}
	store.Release(third)

	waitForSessionClose(t, second, time.Second)
	select {
	case <-first.GetSession().Context().Done():
		t.Fatal("most recently used idle session was evicted")
	default:
	}

	if _, found, err := store.Acquire(context.Background(), nil, second.ID()); err != nil {
		t.Fatal(err)
	} else if found {
		t.Fatal("least recently used session remained available")
	}
}

func TestInMemorySessionStoreReturnsTypedCapacityErrorWhenAllSessionsAreAcquired(t *testing.T) {
	store := NewInMemorySessionStoreWithOptions(InMemorySessionStoreOptions{
		IdleTTL:      time.Hour,
		ReapInterval: time.Minute,
		MaxSessions:  2,
	})
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close session store: %v", err)
		}
	})

	first := newSessionStoreTestSession(t, "first")
	second := newSessionStoreTestSession(t, "second")
	third := newSessionStoreTestSession(t, "third")

	if err := store.Store(context.Background(), first.ID(), first); err != nil {
		t.Fatal(err)
	}
	if err := store.Store(context.Background(), second.ID(), second); err != nil {
		t.Fatal(err)
	}

	err := store.Store(context.Background(), third.ID(), third)
	if !errors.Is(err, ErrSessionCapacity) {
		t.Fatalf("expected ErrSessionCapacity, got %v", err)
	}
	var capacityErr *SessionCapacityError
	if !errors.As(err, &capacityErr) {
		t.Fatalf("expected *SessionCapacityError, got %T", err)
	}
	if capacityErr.MaxSessions != 2 || capacityErr.ActiveSessions != 2 {
		t.Fatalf("unexpected capacity details: %#v", capacityErr)
	}

	for _, session := range []*ServerSession{first, second} {
		select {
		case <-session.GetSession().Context().Done():
			t.Fatalf("acquired session %q was evicted", session.ID())
		default:
		}
	}

	third.Close(true)
}

func TestInMemorySessionStoreCloseStopsReaperAndClosesSessions(t *testing.T) {
	store := NewInMemorySessionStoreWithOptions(InMemorySessionStoreOptions{
		IdleTTL:      time.Hour,
		ReapInterval: time.Minute,
		MaxSessions:  4,
	})
	session := newSessionStoreTestSession(t, "shutdown")
	if err := store.Store(context.Background(), session.ID(), session); err != nil {
		t.Fatal(err)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("second close was not idempotent: %v", err)
	}

	select {
	case <-store.done:
	case <-time.After(time.Second):
		t.Fatal("reaper did not stop")
	}
	waitForSessionClose(t, session, time.Second)

	err := store.Store(context.Background(), "after-close", newSessionStoreTestSession(t, "after-close"))
	if !errors.Is(err, ErrSessionStoreClosed) {
		t.Fatalf("expected ErrSessionStoreClosed, got %v", err)
	}
}

func TestHTTPServerReturnsServiceUnavailableAtSessionCapacity(t *testing.T) {
	store := NewInMemorySessionStoreWithOptions(InMemorySessionStoreOptions{
		IdleTTL:      time.Hour,
		ReapInterval: time.Minute,
		MaxSessions:  1,
	})
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close session store: %v", err)
		}
	})

	blocking := newSessionStoreTestSession(t, "blocking")
	if err := store.Store(context.Background(), blocking.ID(), blocking); err != nil {
		t.Fatal(err)
	}

	var handlerCalls int32
	handler := MessageHandlerFunc(func(ctx context.Context, msg Message) {
		atomic.AddInt32(&handlerCalls, 1)
		if err := msg.Reply(ctx, InitializeResult{
			ProtocolVersion: "2025-06-18",
			ServerInfo: ServerInfo{
				Name:    "test",
				Version: "1",
			},
		}); err != nil {
			t.Errorf("reply to initialize: %v", err)
		}
	})
	server, err := NewHTTPServer(nil, handler, HTTPServerOptions{SessionStore: store})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "http://example.test/mcp", bytes.NewBufferString(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`,
	))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d: %s", http.StatusServiceUnavailable, recorder.Code, recorder.Body.String())
	}
	if retryAfter := recorder.Header().Get("Retry-After"); retryAfter == "" {
		t.Fatal("capacity response did not include Retry-After")
	}
	if got := atomic.LoadInt32(&handlerCalls); got != 0 {
		t.Fatalf("initialize handler ran %d times despite capacity rejection", got)
	}
}

func storeSessionThenRelease(t *testing.T, store *InMemorySessionStore, session *ServerSession) {
	t.Helper()
	if err := store.Store(context.Background(), session.ID(), session); err != nil {
		t.Fatal(err)
	}
	store.Release(session)
}

func newSessionStoreTestSession(t *testing.T, id string) *ServerSession {
	t.Helper()
	session, err := NewExistingServerSession(
		context.Background(),
		SessionState{ID: id},
		MessageHandlerFunc(func(context.Context, Message) {}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if session.GetSession().IsActive() {
			session.Close(true)
		}
	})
	return session
}

func waitForSessionClose(t *testing.T, session *ServerSession, timeout time.Duration) {
	t.Helper()
	select {
	case <-session.GetSession().Context().Done():
	case <-time.After(timeout):
		t.Fatalf("session %q was not closed within %s", session.ID(), timeout)
	}
}
