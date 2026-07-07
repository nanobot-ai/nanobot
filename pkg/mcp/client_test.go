package mcp

import (
	"context"
	"strings"
	"testing"
)

func TestClientListToolsNoToolsDoesNotNeedSession(t *testing.T) {
	tools, err := (&Client{noTools: true}).ListTools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools.Tools) != 0 {
		t.Fatalf("expected no tools, got %#v", tools.Tools)
	}
}

func TestClientCallNoToolsDoesNotNeedSession(t *testing.T) {
	_, err := (&Client{noTools: true}).Call(context.Background(), "hidden", nil)
	if err == nil || !strings.Contains(err.Error(), "no tools allowed") {
		t.Fatalf("expected no tools error, got %v", err)
	}
}

func TestClientCallToolOverridesDisableAbsentTool(t *testing.T) {
	_, err := (&Client{toolOverrides: ToolOverrides{"allowed": {}}}).Call(context.Background(), "hidden", nil)
	if err == nil || !strings.Contains(err.Error(), `tool "hidden" not found`) {
		t.Fatalf("expected tool not found error, got %v", err)
	}
}

func TestClientCallToolOverridesDisableOriginalRenamedTool(t *testing.T) {
	_, err := (&Client{toolOverrides: ToolOverrides{"original": {Name: "renamed"}}}).Call(context.Background(), "original", nil)
	if err == nil || !strings.Contains(err.Error(), `tool "original" not found`) {
		t.Fatalf("expected tool not found error, got %v", err)
	}
}

func TestClientCallEmptyToolOverridesDoesNotDisableTools(t *testing.T) {
	s := NewEmptySession(t.Context())
	_, err := (&Client{Session: s, toolOverrides: ToolOverrides{}}).Call(context.Background(), "someTool", nil)
	if err == nil || strings.Contains(err.Error(), `tool "someTool" not found`) {
		t.Fatalf("expected call to reach session exchange (not be blocked by empty ToolOverrides), got %v", err)
	}
}
