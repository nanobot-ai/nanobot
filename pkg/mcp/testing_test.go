package mcp

import (
	"context"
	"testing"
)

func TestTestSession(t *testing.T) {
	ctx := context.Background()
	session := TestSession(ctx)

	if session == nil {
		t.Fatal("TestSession returned nil")
	}

	// Test that Set and Get work properly
	testKey := "test_key"
	testValue := "test_value"
	session.Set(testKey, testValue)

	var retrievedValue string
	if !session.Get(testKey, &retrievedValue) {
		t.Fatal("Failed to retrieve value from session")
	}

	if retrievedValue != testValue {
		t.Errorf("Expected value %q, got %q", testValue, retrievedValue)
	}

	// Test that attributes are initialized
	if session.attributes == nil {
		t.Error("Session attributes are nil")
	}

	// Test that context is set
	if session.ctx == nil {
		t.Error("Session context is nil")
	}

	// Verify the context has the session set
	sessionFromCtx := SessionFromContext(session.ctx)
	if sessionFromCtx != session {
		t.Error("Session context does not contain the session")
	}
}
