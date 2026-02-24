package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestNewRecordSetsThreadScopedCwd(t *testing.T) {
	m := &Manager{}
	record := m.newRecord("test-session", "account-1")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(wd, "sessions", "test-session")
	if record.Cwd != expected {
		t.Fatalf("expected cwd %q, got %q", expected, record.Cwd)
	}
}

func TestLoadAttributesSetsCwdSessionKey(t *testing.T) {
	m := &Manager{}

	serverSession, err := mcp.NewServerSession(context.Background(), mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}))
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close(false)

	stored := &Session{
		SessionID: "test-session",
		AccountID: "account-1",
		Cwd:       "/tmp/test-cwd",
	}

	m.loadAttributesFromRecord(stored, serverSession)

	var cwd string
	if !serverSession.GetSession().Get(types.CwdSessionKey, &cwd) {
		t.Fatal("expected cwd session key to be present")
	}
	if cwd != stored.Cwd {
		t.Fatalf("expected cwd %q, got %q", stored.Cwd, cwd)
	}
}

func TestManagerStoreAndDeleteEmitEvents(t *testing.T) {
	manager, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	accountID := "account-a"
	ctx := types.WithNanobotContext(context.Background(), types.Context{
		User: mcp.User{ID: accountID},
	})

	serverSession, err := mcp.NewServerSession(context.Background(), mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}))
	if err != nil {
		t.Fatalf("failed to create server session: %v", err)
	}
	defer serverSession.Close(false)
	serverSession.GetSession().Set(types.AccountIDSessionKey, accountID)

	events := make(chan SessionEvent, 4)
	manager.SubscribeEvents(func(event SessionEvent) {
		events <- event
	})

	if err := manager.Store(ctx, "chat-1", serverSession); err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	created := awaitSessionEvent(t, events)
	if created.Type != SessionEventCreated {
		t.Fatalf("created.Type = %q, want %q", created.Type, SessionEventCreated)
	}
	if created.SessionType != "thread" {
		t.Fatalf("created.SessionType = %q, want %q", created.SessionType, "thread")
	}
	if created.SessionID != "chat-1" {
		t.Fatalf("created.SessionID = %q, want %q", created.SessionID, "chat-1")
	}
	if created.AccountID != accountID {
		t.Fatalf("created.AccountID = %q, want %q", created.AccountID, accountID)
	}

	_, found, err := manager.LoadAndDelete(ctx, mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}), "chat-1")
	if err != nil {
		t.Fatalf("LoadAndDelete() failed: %v", err)
	}
	if !found {
		t.Fatal("expected LoadAndDelete() to find chat-1")
	}

	deleted := awaitSessionEvent(t, events)
	if deleted.Type != SessionEventDeleted {
		t.Fatalf("deleted.Type = %q, want %q", deleted.Type, SessionEventDeleted)
	}
	if deleted.SessionType != "thread" {
		t.Fatalf("deleted.SessionType = %q, want %q", deleted.SessionType, "thread")
	}
	if deleted.SessionID != "chat-1" {
		t.Fatalf("deleted.SessionID = %q, want %q", deleted.SessionID, "chat-1")
	}
	if deleted.AccountID != accountID {
		t.Fatalf("deleted.AccountID = %q, want %q", deleted.AccountID, accountID)
	}
}

func TestManagerSubscribeEventsUnsubscribe(t *testing.T) {
	manager, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	accountID := "account-a"
	ctx := context.Background()

	serverSession, err := mcp.NewServerSession(context.Background(), mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}))
	if err != nil {
		t.Fatalf("failed to create server session: %v", err)
	}
	defer serverSession.Close(false)
	serverSession.GetSession().Set(types.AccountIDSessionKey, accountID)

	events := make(chan SessionEvent, 1)
	unsubscribe := manager.SubscribeEvents(func(event SessionEvent) {
		events <- event
	})
	unsubscribe()

	if err := manager.Store(ctx, "chat-2", serverSession); err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	select {
	case event := <-events:
		t.Fatalf("unexpected event after unsubscribe: %+v", event)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestManagerUpdateThreadDescriptionEmitsUpdatedEvent(t *testing.T) {
	manager, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	accountID := "account-a"
	ctx := context.Background()

	if err := manager.DB.Create(ctx, &Session{
		Type:        "thread",
		SessionID:   "chat-rename",
		AccountID:   accountID,
		Description: "Before",
	}); err != nil {
		t.Fatalf("failed to create thread session: %v", err)
	}

	events := make(chan SessionEvent, 1)
	manager.SubscribeEvents(func(event SessionEvent) {
		events <- event
	})

	updated, changed, err := manager.UpdateThreadDescription(ctx, "chat-rename", accountID, "After")
	if err != nil {
		t.Fatalf("UpdateThreadDescription() failed: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	if updated.Description != "After" {
		t.Fatalf("updated.Description = %q, want %q", updated.Description, "After")
	}

	event := awaitSessionEvent(t, events)
	if event.Type != SessionEventUpdated {
		t.Fatalf("event.Type = %q, want %q", event.Type, SessionEventUpdated)
	}
	if event.SessionType != "thread" {
		t.Fatalf("event.SessionType = %q, want %q", event.SessionType, "thread")
	}
	if event.SessionID != "chat-rename" {
		t.Fatalf("event.SessionID = %q, want %q", event.SessionID, "chat-rename")
	}
	if event.AccountID != accountID {
		t.Fatalf("event.AccountID = %q, want %q", event.AccountID, accountID)
	}
}

func TestManagerUpdateThreadDescriptionNoChangeNoEvent(t *testing.T) {
	manager, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	accountID := "account-a"
	ctx := context.Background()

	if err := manager.DB.Create(ctx, &Session{
		Type:        "thread",
		SessionID:   "chat-rename",
		AccountID:   accountID,
		Description: "Same",
	}); err != nil {
		t.Fatalf("failed to create thread session: %v", err)
	}

	events := make(chan SessionEvent, 1)
	manager.SubscribeEvents(func(event SessionEvent) {
		events <- event
	})

	_, changed, err := manager.UpdateThreadDescription(ctx, "chat-rename", accountID, "Same")
	if err != nil {
		t.Fatalf("UpdateThreadDescription() failed: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false for identical description")
	}

	select {
	case event := <-events:
		t.Fatalf("unexpected event for no-op rename: %+v", event)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestManagerStoreDescriptionChangeEmitsUpdatedEvent(t *testing.T) {
	manager, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	accountID := "account-a"
	ctx := context.Background()

	serverSession, err := mcp.NewServerSession(context.Background(), mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}))
	if err != nil {
		t.Fatalf("failed to create server session: %v", err)
	}
	defer serverSession.Close(false)
	serverSession.GetSession().Set(types.AccountIDSessionKey, accountID)
	serverSession.GetSession().Set(types.DescriptionSessionKey, "Before")

	events := make(chan SessionEvent, 4)
	manager.SubscribeEvents(func(event SessionEvent) {
		events <- event
	})

	if err := manager.Store(ctx, "chat-store-update", serverSession); err != nil {
		t.Fatalf("Store() create failed: %v", err)
	}
	created := awaitSessionEvent(t, events)
	if created.Type != SessionEventCreated {
		t.Fatalf("created.Type = %q, want %q", created.Type, SessionEventCreated)
	}

	serverSession.GetSession().Set(types.DescriptionSessionKey, "After")
	if err := manager.Store(ctx, "chat-store-update", serverSession); err != nil {
		t.Fatalf("Store() update failed: %v", err)
	}

	updated := awaitSessionEvent(t, events)
	if updated.Type != SessionEventUpdated {
		t.Fatalf("updated.Type = %q, want %q", updated.Type, SessionEventUpdated)
	}
	if updated.SessionType != "thread" {
		t.Fatalf("updated.SessionType = %q, want %q", updated.SessionType, "thread")
	}
	if updated.SessionID != "chat-store-update" {
		t.Fatalf("updated.SessionID = %q, want %q", updated.SessionID, "chat-store-update")
	}
	if updated.AccountID != accountID {
		t.Fatalf("updated.AccountID = %q, want %q", updated.AccountID, accountID)
	}
}

func awaitSessionEvent(t *testing.T, events <-chan SessionEvent) SessionEvent {
	t.Helper()
	select {
	case event := <-events:
		return event
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for session event")
		return SessionEvent{}
	}
}
