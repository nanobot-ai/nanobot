package session

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
)

func TestHTTPServerReconnectStormBoundsLiveSessionResources(t *testing.T) {
	manager := NewManagerWithOptions(context.Background(), newTestStore(t), ManagerOptions{
		LiveSessionIdleTTL: time.Hour,
		MaxLiveSessions:    1,
	})
	t.Cleanup(func() {
		if err := manager.Close(); err != nil {
			t.Errorf("close manager: %v", err)
		}
	})

	var (
		activeChildren int32
		maxChildren    int32
		handlerCalls   int32
	)
	firstStarted := make(chan struct{})
	unblockFirst := make(chan struct{})
	handler := mcp.MessageHandlerFunc(func(ctx context.Context, msg mcp.Message) {
		atomic.AddInt32(&handlerCalls, 1)
		active := atomic.AddInt32(&activeChildren, 1)
		updateAtomicMax(&maxChildren, active)
		msg.Session.Set("test-child", &countingSessionResource{active: &activeChildren})
		close(firstStarted)
		<-unblockFirst
		if err := msg.Reply(ctx, testInitializeResult()); err != nil {
			t.Errorf("reply to initialize: %v", err)
		}
	})

	server, err := mcp.NewHTTPServer(nil, handler, mcp.HTTPServerOptions{
		SessionStore: manager,
	})
	if err != nil {
		t.Fatal(err)
	}

	const requests = 10
	results := make(chan int, requests)
	for range requests {
		go func() {
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, newInitializeRequest())
			results <- recorder.Code
		}()
	}

	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("first initialization did not start")
	}

	for range requests - 1 {
		select {
		case status := <-results:
			if status != http.StatusServiceUnavailable {
				t.Fatalf("reconnect storm status = %d, want %d", status, http.StatusServiceUnavailable)
			}
		case <-time.After(time.Second):
			t.Fatal("reconnect request was not rejected before creating a session")
		}
	}

	close(unblockFirst)
	select {
	case status := <-results:
		if status != http.StatusOK {
			t.Fatalf("admitted initialization status = %d, want %d", status, http.StatusOK)
		}
	case <-time.After(time.Second):
		t.Fatal("admitted initialization did not finish")
	}

	if got := atomic.LoadInt32(&handlerCalls); got != 1 {
		t.Fatalf("handler calls = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&maxChildren); got != 1 {
		t.Fatalf("maximum concurrent downstream resources = %d, want 1", got)
	}

	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&activeChildren); got != 0 {
		t.Fatalf("active downstream resources after shutdown = %d, want 0", got)
	}
}

func TestHTTPServerEventStreamLifetimeReleasesLiveSession(t *testing.T) {
	manager := NewManagerWithOptions(context.Background(), newTestStore(t), ManagerOptions{
		LiveSessionIdleTTL: 20 * time.Millisecond,
		MaxLiveSessions:    1,
	})
	t.Cleanup(func() {
		if err := manager.Close(); err != nil {
			t.Errorf("close manager: %v", err)
		}
	})

	var activeChildren int32
	handler := mcp.MessageHandlerFunc(func(ctx context.Context, msg mcp.Message) {
		atomic.AddInt32(&activeChildren, 1)
		msg.Session.Set("test-child", &countingSessionResource{active: &activeChildren})
		if err := msg.Reply(ctx, testInitializeResult()); err != nil {
			t.Errorf("reply to initialize: %v", err)
		}
	})
	server, err := mcp.NewHTTPServer(nil, handler, mcp.HTTPServerOptions{
		SessionStore:           manager,
		EventStreamMaxLifetime: 25 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}

	initializeRecorder := httptest.NewRecorder()
	server.ServeHTTP(initializeRecorder, newInitializeRequest())
	if initializeRecorder.Code != http.StatusOK {
		t.Fatalf("initialize status = %d: %s", initializeRecorder.Code, initializeRecorder.Body.String())
	}
	sessionID := initializeRecorder.Header().Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("initialize response did not include a session ID")
	}

	streamRequest := httptest.NewRequest(http.MethodGet, "http://example.test/mcp", nil)
	streamRequest.Header.Set("Mcp-Session-Id", sessionID)
	streamDone := make(chan struct{})
	go func() {
		defer close(streamDone)
		server.ServeHTTP(httptest.NewRecorder(), streamRequest)
	}()

	select {
	case <-streamDone:
	case <-time.After(time.Second):
		t.Fatal("event stream exceeded its configured maximum lifetime")
	}

	deadline := time.After(time.Second)
	for atomic.LoadInt32(&activeChildren) != 0 {
		select {
		case <-deadline:
			t.Fatal("session resource remained active after stream lifetime and idle TTL elapsed")
		case <-time.After(time.Millisecond):
		}
	}
}

func TestManagerReserveEvictsIdleSessionAtCapacity(t *testing.T) {
	manager := NewManagerWithOptions(context.Background(), newTestStore(t), ManagerOptions{
		LiveSessionIdleTTL: time.Hour,
		MaxLiveSessions:    1,
	})
	t.Cleanup(func() {
		if err := manager.Close(); err != nil {
			t.Errorf("close manager: %v", err)
		}
	})

	var activeChildren int32 = 1
	session := newManagerTestSession(t, "idle")
	session.GetSession().Set("test-child", &countingSessionResource{active: &activeChildren})
	if err := manager.Store(context.Background(), session.ID(), session); err != nil {
		t.Fatal(err)
	}
	manager.Release(session)

	release, err := manager.Reserve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	release()

	select {
	case <-session.GetSession().Context().Done():
	case <-time.After(time.Second):
		t.Fatal("idle session was not closed when capacity was reserved")
	}
	if got := atomic.LoadInt32(&activeChildren); got != 0 {
		t.Fatalf("active downstream resources = %d, want 0", got)
	}
}

func TestManagerAcquireClosesLoadedSessionOnAccountMismatch(t *testing.T) {
	store := newTestStore(t)
	createTestSession(t, store, "other-account", time.Now())

	manager := NewManagerWithOptions(context.Background(), store, ManagerOptions{})
	t.Cleanup(func() {
		if err := manager.Close(); err != nil {
			t.Errorf("close manager: %v", err)
		}
	})

	var loaded *mcp.ServerSession
	newServerSession := manager.newServerSession
	manager.newServerSession = func(ctx context.Context, state mcp.SessionState, server mcp.MessageHandler) (*mcp.ServerSession, error) {
		session, err := newServerSession(ctx, state, server)
		loaded = session
		return session, err
	}

	ctx := types.WithNanobotContext(context.Background(), types.Context{
		User: mcp.User{ID: "not-account-1"},
	})
	session, found, err := manager.Acquire(ctx, mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}), "other-account")
	if err != nil {
		t.Fatal(err)
	}
	if found || session != nil {
		t.Fatal("session from another account was returned")
	}
	if loaded == nil {
		t.Fatal("persisted session was not loaded")
	}
	select {
	case <-loaded.GetSession().Context().Done():
	case <-time.After(time.Second):
		t.Fatal("rejected loaded session was not closed")
	}
}

type countingSessionResource struct {
	active *int32
	once   sync.Once
}

func (r *countingSessionResource) Close(bool) {
	r.once.Do(func() {
		atomic.AddInt32(r.active, -1)
	})
}

func newInitializeRequest() *http.Request {
	return httptest.NewRequest(http.MethodPost, "http://example.test/mcp", bytes.NewBufferString(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`,
	))
}

func testInitializeResult() mcp.InitializeResult {
	return mcp.InitializeResult{
		ProtocolVersion: "2025-06-18",
		ServerInfo: mcp.ServerInfo{
			Name:    "test",
			Version: "1",
		},
	}
}

func newManagerTestSession(t *testing.T, id string) *mcp.ServerSession {
	t.Helper()
	session, err := mcp.NewExistingServerSession(
		context.Background(),
		mcp.SessionState{ID: id},
		mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}),
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

func updateAtomicMax(target *int32, candidate int32) {
	for {
		current := atomic.LoadInt32(target)
		if candidate <= current || atomic.CompareAndSwapInt32(target, current, candidate) {
			return
		}
	}
}
