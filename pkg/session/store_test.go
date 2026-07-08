package session

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := NewStoreFromDSN(filepath.Join(t.TempDir(), "nanobot.db"))
	if err != nil {
		t.Fatalf("NewStoreFromDSN: %v", err)
	}
	return store
}

func createTestSession(t *testing.T, store *Store, id string, updatedAt time.Time) {
	t.Helper()

	if err := store.Create(context.Background(), &Session{
		SessionID: id,
		AccountID: "account-1",
		State: State{
			ID: id,
		},
	}); err != nil {
		t.Fatalf("Create(%s): %v", id, err)
	}

	if err := store.db.Model(&Session{}).
		Where("session_id = ?", id).
		UpdateColumn("updated_at", updatedAt).
		Error; err != nil {
		t.Fatalf("UpdateColumn(%s): %v", id, err)
	}
}

func TestDeleteSessionsUpdatedBefore(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now()
	cutoff := now.Add(-7 * 24 * time.Hour)

	createTestSession(t, store, "old", now.Add(-8*24*time.Hour))
	createTestSession(t, store, "fresh", now.Add(-6*24*time.Hour))
	createTestSession(t, store, "excluded", now.Add(-8*24*time.Hour))

	if err := store.AddWorkflowRun(ctx, "old", "workflow:///old"); err != nil {
		t.Fatalf("AddWorkflowRun(old): %v", err)
	}
	if err := store.AddWorkflowRun(ctx, "excluded", "workflow:///excluded"); err != nil {
		t.Fatalf("AddWorkflowRun(excluded): %v", err)
	}

	deleted, err := store.DeleteSessionsUpdatedBefore(ctx, cutoff, "excluded")
	if err != nil {
		t.Fatalf("DeleteSessionsUpdatedBefore: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	if _, err := store.Get(ctx, "old"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("old session lookup err = %v, want gorm.ErrRecordNotFound", err)
	}
	var oldRows int64
	if err := store.db.Unscoped().Model(&Session{}).Where("session_id = ?", "old").Count(&oldRows).Error; err != nil {
		t.Fatalf("count old session rows: %v", err)
	}
	if oldRows != 0 {
		t.Fatalf("old session should be hard-deleted, found %d unscoped rows", oldRows)
	}
	if _, err := store.Get(ctx, "fresh"); err != nil {
		t.Fatalf("fresh session should remain: %v", err)
	}
	if _, err := store.Get(ctx, "excluded"); err != nil {
		t.Fatalf("excluded session should remain: %v", err)
	}

	workflowURIs, err := store.ListWorkflowURIs(ctx, "old", "excluded")
	if err != nil {
		t.Fatalf("ListWorkflowURIs: %v", err)
	}
	if len(workflowURIs["old"]) != 0 {
		t.Fatalf("old workflow runs should be deleted, got %v", workflowURIs["old"])
	}
	if got := workflowURIs["excluded"]; len(got) != 1 || got[0] != "workflow:///excluded" {
		t.Fatalf("excluded workflow runs = %v, want [workflow:///excluded]", got)
	}
}
