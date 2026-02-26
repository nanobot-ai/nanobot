package workflows

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestInitialize_AccessBySessionType(t *testing.T) {
	tests := []struct {
		name        string
		meta        map[string]any
		expectTools bool
	}{
		{
			name:        "ui session has tools",
			meta:        map[string]any{"ui": true},
			expectTools: true,
		},
		{
			name:        "chat session has tools",
			meta:        map[string]any{"chat": true},
			expectTools: true,
		},
		{
			name:        "non-ui non-chat session has no tools",
			meta:        map[string]any{},
			expectTools: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewToolsServer()
			baseCtx := context.Background()
			session := mcp.NewEmptySession(baseCtx)
			session.Set(types.SessionInitSessionKey, types.SessionInitHook{
				Meta: tt.meta,
			})
			ctx := mcp.WithSession(baseCtx, session)

			result, err := s.initialize(ctx, mcp.Message{}, mcp.InitializeRequest{
				ProtocolVersion: "2024-11-05",
			})
			if err != nil {
				t.Fatalf("initialize() failed: %v", err)
			}

			hasTools := result.Capabilities.Tools != nil
			if hasTools != tt.expectTools {
				t.Fatalf("tools capability mismatch: got %v, want %v", hasTools, tt.expectTools)
			}
		})
	}
}

func TestRecordWorkflowRun_DeduplicatesURI(t *testing.T) {
	s := NewToolsServer()
	ctx := context.Background()
	session := mcp.NewEmptySession(ctx)
	ctx = mcp.WithSession(ctx, session)

	uri := "workflow:///test-workflow"
	if _, err := s.recordWorkflowRun(ctx, struct {
		URI string `json:"uri"`
	}{URI: uri}); err != nil {
		t.Fatalf("first recordWorkflowRun() failed: %v", err)
	}

	if _, err := s.recordWorkflowRun(ctx, struct {
		URI string `json:"uri"`
	}{URI: uri}); err != nil {
		t.Fatalf("second recordWorkflowRun() failed: %v", err)
	}

	var uris []string
	session.Root().Get(types.WorkflowURIsSessionKey, &uris)
	if len(uris) != 1 {
		t.Fatalf("expected one recorded URI, got %d: %v", len(uris), uris)
	}
	if uris[0] != uri {
		t.Errorf("recorded URI = %q, want %q", uris[0], uri)
	}
}

func TestDeleteWorkflow_RemovesFile(t *testing.T) {
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	workflowsPath := filepath.Join(tempDir, workflowsDir)
	if err := os.MkdirAll(workflowsPath, 0755); err != nil {
		t.Fatalf("failed to create workflows directory: %v", err)
	}

	workflowPath := filepath.Join(workflowsPath, "to-delete.md")
	if err := os.WriteFile(workflowPath, []byte("# test"), 0644); err != nil {
		t.Fatalf("failed to write workflow file: %v", err)
	}

	s := NewToolsServer()
	if _, err := s.deleteWorkflow(context.Background(), struct {
		URI string `json:"uri"`
	}{URI: "workflow:///to-delete"}); err != nil {
		t.Fatalf("deleteWorkflow() failed: %v", err)
	}

	if _, err := os.Stat(workflowPath); !os.IsNotExist(err) {
		t.Fatalf("expected workflow file to be deleted, stat err: %v", err)
	}
}
