package session

import (
	"testing"
)

func TestStore_CreateAndList(t *testing.T) {
	store, err := NewStoreFromDSN("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	testSession := &Session{
		Type:     "test",
		ParentID: "parent123",
		Config:   ConfigWrapper{},
		State: MetadataWrapper{
			"key1": "value1",
			"key2": 123,
		},
	}

	err = store.Create(testSession)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if testSession.ID == 0 {
		t.Error("Expected session ID to be set after creation")
	}

	sessions, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	session := sessions[0]
	if session.Type != "test" {
		t.Errorf("Expected type 'test', got '%s'", session.Type)
	}
	if session.ParentID != "parent123" {
		t.Errorf("Expected parentID 'parent123', got '%s'", session.ParentID)
	}
	if len(session.Config.Agents) != 0 {
		t.Errorf("Expected empty config agents, got %v", session.Config.Agents)
	}
	if session.State["key1"] != "value1" {
		t.Errorf("Expected metadata key1 'value1', got '%v'", session.State["key1"])
	}
}

func TestStore_CreateMultiple(t *testing.T) {
	store, err := NewStoreFromDSN(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	session1 := &Session{Type: "type1", ParentID: "parent1"}
	session2 := &Session{Type: "type2", ParentID: "parent2"}

	if err := store.Create(session1); err != nil {
		t.Fatalf("Failed to create session1: %v", err)
	}
	if err := store.Create(session2); err != nil {
		t.Fatalf("Failed to create session2: %v", err)
	}

	sessions, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions, got %d", len(sessions))
	}
}
