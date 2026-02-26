package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func withWorkingDir(t *testing.T, dir string) func() {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	return func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}
}

func newManagerAndContext(t *testing.T, accountID string) (*session.Manager, context.Context) {
	t.Helper()

	manager, err := session.NewManager(":memory:")
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}

	ctx := context.Background()
	mcpSession := mcp.NewEmptySession(ctx)
	mcpSession.Set(session.ManagerSessionKey, manager)
	mcpSession.Set(types.AccountIDSessionKey, accountID)
	ctx = mcp.WithSession(ctx, mcpSession)

	return manager, ctx
}

func TestListFilesScopedByAccount(t *testing.T) {
	manager, ctx := newManagerAndContext(t, "account-a")

	accountADir := t.TempDir()
	if err := os.WriteFile(filepath.Join(accountADir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatalf("failed to write account-a file: %v", err)
	}

	accountBDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(accountBDir, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatalf("failed to write account-b file: %v", err)
	}

	if err := manager.DB.Create(ctx, &session.Session{
		Type:        "thread",
		SessionID:   "session-a",
		AccountID:   "account-a",
		Description: "Session A",
		Cwd:         accountADir,
	}); err != nil {
		t.Fatalf("failed to create account-a session: %v", err)
	}
	if err := manager.DB.Create(ctx, &session.Session{
		Type:        "thread",
		SessionID:   "session-b",
		AccountID:   "account-b",
		Description: "Session B",
		Cwd:         accountBDir,
	}); err != nil {
		t.Fatalf("failed to create account-b session: %v", err)
	}

	s := &Server{}
	result, err := s.listFiles(ctx, struct{}{})
	if err != nil {
		t.Fatalf("listFiles() failed: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 file resource, got %d", len(result.Resources))
	}

	file := result.Resources[0]
	if got, _ := file.Meta["sessionId"].(string); got != "session-a" {
		t.Fatalf("file._meta.sessionId = %q, want %q", got, "session-a")
	}
	if file.Name != "a.txt" {
		t.Fatalf("file.Name = %q, want %q", file.Name, "a.txt")
	}
}

func TestListFilesSkipsMissingSessionDirectories(t *testing.T) {
	manager, ctx := newManagerAndContext(t, "account-a")

	existingDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(existingDir, "present.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := manager.DB.Create(ctx, &session.Session{
		Type:      "thread",
		SessionID: "missing-session",
		AccountID: "account-a",
		Cwd:       filepath.Join(t.TempDir(), "does-not-exist"),
	}); err != nil {
		t.Fatalf("failed to create missing session: %v", err)
	}
	if err := manager.DB.Create(ctx, &session.Session{
		Type:      "thread",
		SessionID: "existing-session",
		AccountID: "account-a",
		Cwd:       existingDir,
	}); err != nil {
		t.Fatalf("failed to create existing session: %v", err)
	}

	s := &Server{}
	result, err := s.listFiles(ctx, struct{}{})
	if err != nil {
		t.Fatalf("listFiles() failed: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 file resource, got %d", len(result.Resources))
	}
	if got, _ := result.Resources[0].Meta["sessionId"].(string); got != "existing-session" {
		t.Fatalf("got session %q, want %q", got, "existing-session")
	}
}

func TestListWorkflowsMissingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	s := &Server{}
	result, err := s.listWorkflows(context.Background(), struct{}{})
	if err != nil {
		t.Fatalf("listWorkflows() failed: %v", err)
	}
	if len(result.Resources) != 0 {
		t.Fatalf("expected 0 workflows, got %d", len(result.Resources))
	}
}

func TestListWorkflowsIncludesMetadataAndMarkdownOnly(t *testing.T) {
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	workflowsPath := filepath.Join(tempDir, workflowsDir)
	if err := os.MkdirAll(workflowsPath, 0o755); err != nil {
		t.Fatalf("failed to create workflows directory: %v", err)
	}

	content := `---
name: Test Workflow
description: Workflow description.
createdAt: 2026-02-25T00:00:00Z
---

# Body
`
	if err := os.WriteFile(filepath.Join(workflowsPath, "test.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write workflow file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workflowsPath, "ignore.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("failed to write non-workflow file: %v", err)
	}

	s := &Server{}
	result, err := s.listWorkflows(context.Background(), struct{}{})
	if err != nil {
		t.Fatalf("listWorkflows() failed: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(result.Resources))
	}

	workflow := result.Resources[0]
	if workflow.URI != "workflow:///test" {
		t.Fatalf("workflow.URI = %q, want %q", workflow.URI, "workflow:///test")
	}
	if workflow.Name != "test" {
		t.Fatalf("workflow.Name = %q, want %q", workflow.Name, "test")
	}
	if workflow.Description != "Workflow description." {
		t.Fatalf("workflow.Description = %q, want %q", workflow.Description, "Workflow description.")
	}
	if workflow.MimeType != "text/markdown" {
		t.Fatalf("workflow.MimeType = %q, want %q", workflow.MimeType, "text/markdown")
	}
	if workflow.Meta == nil {
		t.Fatal("workflow.Meta should not be nil")
	}
	if workflow.Meta["name"] != "Test Workflow" {
		t.Fatalf("workflow.Meta[name] = %v, want %q", workflow.Meta["name"], "Test Workflow")
	}
	if workflow.Meta["createdAt"] != "2026-02-25T00:00:00Z" {
		t.Fatalf("workflow.Meta[createdAt] = %v, want %q", workflow.Meta["createdAt"], "2026-02-25T00:00:00Z")
	}
}

func TestListFilesFallbackSessionDirectory(t *testing.T) {
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	manager, ctx := newManagerAndContext(t, "account-a")

	sessionDir := filepath.Join(tempDir, "sessions", types.SanitizeSessionDirectoryName("session-no-cwd"))
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("failed to create session directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "fallback.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("failed to write fallback file: %v", err)
	}

	if err := manager.DB.Create(ctx, &session.Session{
		Type:      "thread",
		SessionID: "session-no-cwd",
		AccountID: "account-a",
		Cwd:       "",
	}); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	s := &Server{}
	result, err := s.listFiles(ctx, struct{}{})
	if err != nil {
		t.Fatalf("listFiles() failed: %v", err)
	}
	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 file resource, got %d", len(result.Resources))
	}
	if result.Resources[0].Name != "fallback.txt" {
		t.Fatalf("file.Name = %q, want %q", result.Resources[0].Name, "fallback.txt")
	}
}

func TestDefaultSessionCwd(t *testing.T) {
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	got := defaultSessionCwd("session-id")
	want := filepath.Join(tempDir, "sessions", "session-id")
	if runtime.GOOS == "windows" {
		got = filepath.Clean(got)
		want = filepath.Clean(want)
	}
	if got != want {
		t.Fatalf("defaultSessionCwd() = %q, want %q", got, want)
	}
}

func TestMetaToolsOnlyExposeUpdateChat(t *testing.T) {
	s := NewServer(nil)

	tools, err := s.tools.List(context.Background(), mcp.Message{}, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("tools.List() failed: %v", err)
	}

	if len(tools.Tools) != 1 {
		t.Fatalf("expected exactly one meta tool, got %d", len(tools.Tools))
	}
	if tools.Tools[0].Name != "update_chat" {
		t.Fatalf("tool name = %q, want %q", tools.Tools[0].Name, "update_chat")
	}
}

func TestInitializeAdvertisesResourceSubscriptions(t *testing.T) {
	baseCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mcpSession := mcp.NewEmptySession(baseCtx)
	mcpSession.Set(types.SessionInitSessionKey, types.SessionInitHook{
		Meta: map[string]any{
			"ui": true,
		},
	})
	mcpSession.Set(types.AccountIDSessionKey, "account-a")

	ctx := mcp.WithSession(baseCtx, mcpSession)
	s := NewServer(nil)
	result, err := s.initialize(ctx, mcp.Message{Session: mcpSession}, mcp.InitializeRequest{
		ProtocolVersion: "2024-11-05",
	})
	if err != nil {
		t.Fatalf("initialize() failed: %v", err)
	}

	if result.Capabilities.Resources == nil {
		t.Fatal("expected resources capability")
	}
	if !result.Capabilities.Resources.Subscribe {
		t.Fatal("expected resources subscribe capability")
	}
	if !result.Capabilities.Resources.ListChanged {
		t.Fatal("expected resources listChanged capability")
	}
}

func TestResourcesListIncludesAgentAndChatResources(t *testing.T) {
	manager, ctx := newManagerAndContext(t, "account-a")
	mcpSession := mcp.SessionFromContext(ctx)
	mcpSession.Set(types.ConfigSessionKey, types.Config{
		Publish: types.Publish{
			Entrypoint: []string{"assistant"},
		},
		Agents: map[string]types.Agent{
			"assistant": {
				HookAgent: types.HookAgent{
					Name:        "Assistant",
					Description: "Primary assistant",
				},
			},
		},
	})

	if err := manager.DB.Create(ctx, &session.Session{
		Type:        "thread",
		SessionID:   "chat-1",
		AccountID:   "account-a",
		Description: "Chat One",
		Cwd:         t.TempDir(),
		State: session.State(mcp.SessionState{
			Attributes: map[string]any{
				types.CurrentAgentSessionKey: "assistant",
			},
		}),
		Config: session.ConfigWrapper(types.Config{
			Publish: types.Publish{
				Entrypoint: []string{"assistant", "reviewer"},
			},
		}),
	}); err != nil {
		t.Fatalf("failed to create chat session: %v", err)
	}

	s := NewServer(sessiondata.NewData(nil))
	result, err := s.resourcesList(ctx, mcp.Message{}, mcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("resourcesList() failed: %v", err)
	}

	var (
		agentResource *mcp.Resource
		chatResource  *mcp.Resource
	)
	for i := range result.Resources {
		res := &result.Resources[i]
		switch res.URI {
		case "agent:///assistant":
			agentResource = res
		case "chat:///threads/chat-1":
			chatResource = res
		}
	}

	if agentResource == nil {
		t.Fatal("expected agent resource for assistant")
	}
	if agentResource.MimeType != types.AgentMimeType {
		t.Fatalf("agent mimeType = %q, want %q", agentResource.MimeType, types.AgentMimeType)
	}
	if got, _ := agentResource.Meta["id"].(string); got != "assistant" {
		t.Fatalf("agent _meta.id = %q, want %q", got, "assistant")
	}

	if chatResource == nil {
		t.Fatal("expected chat resource for chat-1")
	}
	if chatResource.MimeType != types.SessionMimeType {
		t.Fatalf("chat mimeType = %q, want %q", chatResource.MimeType, types.SessionMimeType)
	}
	if got, _ := chatResource.Meta["id"].(string); got != "chat-1" {
		t.Fatalf("chat _meta.id = %q, want %q", got, "chat-1")
	}
	if got, _ := chatResource.Meta["currentAgentId"].(string); got != "assistant" {
		t.Fatalf("chat _meta.currentAgentId = %q, want %q", got, "assistant")
	}
	available, ok := chatResource.Meta["availableAgentIds"].([]any)
	if !ok {
		t.Fatalf("chat _meta.availableAgentIds has unexpected type %T", chatResource.Meta["availableAgentIds"])
	}
	if len(available) != 2 || available[0] != "assistant" || available[1] != "reviewer" {
		t.Fatalf("chat _meta.availableAgentIds = %v, want [assistant reviewer]", available)
	}
}

func TestResourcesReadChat(t *testing.T) {
	manager, ctx := newManagerAndContext(t, "account-a")
	if err := manager.DB.Create(ctx, &session.Session{
		Type:        "thread",
		SessionID:   "chat-1",
		AccountID:   "account-a",
		Description: "Chat One",
		Cwd:         t.TempDir(),
		State: session.State(mcp.SessionState{
			Attributes: map[string]any{
				types.CurrentAgentSessionKey: "assistant",
			},
		}),
		Config: session.ConfigWrapper(types.Config{
			Publish: types.Publish{
				Entrypoint: []string{"assistant", "reviewer"},
			},
		}),
	}); err != nil {
		t.Fatalf("failed to create chat session: %v", err)
	}

	s := &Server{}
	result, err := s.resourcesRead(ctx, mcp.Message{}, mcp.ReadResourceRequest{
		URI: "chat:///threads/chat-1",
	})
	if err != nil {
		t.Fatalf("resourcesRead() failed: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("expected one content item, got %d", len(result.Contents))
	}

	content := result.Contents[0]
	if content.MIMEType != types.SessionMimeType {
		t.Fatalf("content MIMEType = %q, want %q", content.MIMEType, types.SessionMimeType)
	}
	if content.Text == nil {
		t.Fatal("expected chat resource text content")
	}

	var chat types.Chat
	if err := json.Unmarshal([]byte(*content.Text), &chat); err != nil {
		t.Fatalf("failed to unmarshal chat payload: %v", err)
	}
	if chat.ID != "chat-1" {
		t.Fatalf("chat.ID = %q, want %q", chat.ID, "chat-1")
	}
	if chat.CurrentAgentID != "assistant" {
		t.Fatalf("chat.CurrentAgentID = %q, want %q", chat.CurrentAgentID, "assistant")
	}
	if len(chat.AvailableAgentIDs) != 2 || chat.AvailableAgentIDs[0] != "assistant" || chat.AvailableAgentIDs[1] != "reviewer" {
		t.Fatalf("chat.AvailableAgentIDs = %v, want [assistant reviewer]", chat.AvailableAgentIDs)
	}
}

func TestResourcesReadAgent(t *testing.T) {
	_, ctx := newManagerAndContext(t, "account-a")
	mcpSession := mcp.SessionFromContext(ctx)
	mcpSession.Set(types.ConfigSessionKey, types.Config{
		Publish: types.Publish{
			Entrypoint: []string{"assistant"},
		},
		Agents: map[string]types.Agent{
			"assistant": {
				HookAgent: types.HookAgent{
					Name:        "Assistant",
					Description: "Primary assistant",
				},
			},
		},
	})

	s := NewServer(sessiondata.NewData(nil))
	result, err := s.resourcesRead(ctx, mcp.Message{}, mcp.ReadResourceRequest{
		URI: "agent:///assistant",
	})
	if err != nil {
		t.Fatalf("resourcesRead() failed: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("expected one content item, got %d", len(result.Contents))
	}

	content := result.Contents[0]
	if content.MIMEType != types.AgentMimeType {
		t.Fatalf("content MIMEType = %q, want %q", content.MIMEType, types.AgentMimeType)
	}
	if content.Text == nil {
		t.Fatal("expected agent resource text content")
	}

	var agent types.AgentDisplay
	if err := json.Unmarshal([]byte(*content.Text), &agent); err != nil {
		t.Fatalf("failed to unmarshal agent payload: %v", err)
	}
	if agent.ID != "assistant" {
		t.Fatalf("agent.ID = %q, want %q", agent.ID, "assistant")
	}
}

func TestResourcesReadFileScopedByAccount(t *testing.T) {
	manager, ctx := newManagerAndContext(t, "account-a")

	accountADir := t.TempDir()
	aPath := filepath.Join(accountADir, "a.txt")
	if err := os.WriteFile(aPath, []byte("a"), 0o644); err != nil {
		t.Fatalf("failed to write account-a file: %v", err)
	}

	accountBDir := t.TempDir()
	bPath := filepath.Join(accountBDir, "b.txt")
	if err := os.WriteFile(bPath, []byte("b"), 0o644); err != nil {
		t.Fatalf("failed to write account-b file: %v", err)
	}

	if err := manager.DB.Create(ctx, &session.Session{
		Type:      "thread",
		SessionID: "session-a",
		AccountID: "account-a",
		Cwd:       accountADir,
	}); err != nil {
		t.Fatalf("failed to create account-a session: %v", err)
	}
	if err := manager.DB.Create(ctx, &session.Session{
		Type:      "thread",
		SessionID: "session-b",
		AccountID: "account-b",
		Cwd:       accountBDir,
	}); err != nil {
		t.Fatalf("failed to create account-b session: %v", err)
	}

	s := &Server{}
	okResult, err := s.resourcesRead(ctx, mcp.Message{}, mcp.ReadResourceRequest{
		URI: fileURI(aPath),
	})
	if err != nil {
		t.Fatalf("resourcesRead() for account-a file failed: %v", err)
	}
	if len(okResult.Contents) != 1 || okResult.Contents[0].Text == nil || *okResult.Contents[0].Text != "a" {
		t.Fatalf("unexpected content for account-a file: %#v", okResult.Contents)
	}

	_, err = s.resourcesRead(ctx, mcp.Message{}, mcp.ReadResourceRequest{
		URI: fileURI(bPath),
	})
	if err == nil {
		t.Fatal("expected reading account-b file to fail for account-a context")
	}
}

func TestResourcesReadWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	workflowsPath := filepath.Join(tempDir, workflowsDir)
	if err := os.MkdirAll(workflowsPath, 0o755); err != nil {
		t.Fatalf("failed to create workflows directory: %v", err)
	}

	content := `---
name: Test Workflow
createdAt: 2026-02-25T00:00:00Z
---

# Body
`
	if err := os.WriteFile(filepath.Join(workflowsPath, "test.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write workflow file: %v", err)
	}

	s := &Server{}
	result, err := s.resourcesRead(context.Background(), mcp.Message{}, mcp.ReadResourceRequest{
		URI: "workflow:///test",
	})
	if err != nil {
		t.Fatalf("resourcesRead() failed: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("expected one content item, got %d", len(result.Contents))
	}
	if result.Contents[0].MIMEType != "text/markdown" {
		t.Fatalf("workflow MIMEType = %q, want %q", result.Contents[0].MIMEType, "text/markdown")
	}
	if result.Contents[0].Text == nil || !strings.Contains(*result.Contents[0].Text, "# Body") {
		t.Fatalf("expected markdown workflow content, got %#v", result.Contents[0].Text)
	}
}

func TestResourcesSubscribeAndUnsubscribeFile(t *testing.T) {
	manager, ctx := newManagerAndContext(t, "account-a")
	dir := t.TempDir()
	filePath := filepath.Join(dir, "watched.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := manager.DB.Create(ctx, &session.Session{
		Type:      "thread",
		SessionID: "session-a",
		AccountID: "account-a",
		Cwd:       dir,
	}); err != nil {
		t.Fatalf("failed to create account-a session: %v", err)
	}

	mcpSession := mcp.SessionFromContext(ctx)
	msg := mcp.Message{Session: mcpSession}

	s := NewServer(nil)
	if _, err := s.resourcesSubscribe(ctx, msg, mcp.SubscribeRequest{
		URI: fileURI(filePath),
	}); err != nil {
		t.Fatalf("resourcesSubscribe() failed: %v", err)
	}

	if _, err := s.resourcesUnsubscribe(ctx, msg, mcp.UnsubscribeRequest{
		URI: fileURI(filePath),
	}); err != nil {
		t.Fatalf("resourcesUnsubscribe() failed: %v", err)
	}
}

func TestResourcesSubscribeRejectsUnknownResource(t *testing.T) {
	_, ctx := newManagerAndContext(t, "account-a")
	msg := mcp.Message{Session: mcp.SessionFromContext(ctx)}
	s := NewServer(nil)

	_, err := s.resourcesSubscribe(ctx, msg, mcp.SubscribeRequest{
		URI: "chat:///threads/does-not-exist",
	})
	if err == nil {
		t.Fatal("expected resourcesSubscribe() to reject unknown resource")
	}
}

func TestSessionEventCreatedSendsAccountScopedListChanged(t *testing.T) {
	s := NewServer(nil)

	sessionA := newCaptureSession(t)
	defer sessionA.session.Close(false)
	sessionB := newCaptureSession(t)
	defer sessionB.session.Close(false)

	s.sessionLock.Lock()
	s.sessions["a"] = trackedSession{
		session:   sessionA.session,
		accountID: "account-a",
	}
	s.sessions["b"] = trackedSession{
		session:   sessionB.session,
		accountID: "account-b",
	}
	s.sessionLock.Unlock()

	s.handleSessionEvent(session.SessionEvent{
		Type:        session.SessionEventCreated,
		SessionType: "thread",
		SessionID:   "chat-1",
		AccountID:   "account-a",
	})

	msg, ok := sessionA.read(2 * time.Second)
	if !ok {
		t.Fatal("expected list_changed notification for account-a session")
	}
	if msg.Method != "notifications/resources/list_changed" {
		t.Fatalf("message.Method = %q, want %q", msg.Method, "notifications/resources/list_changed")
	}

	if msg, ok := sessionB.read(150 * time.Millisecond); ok {
		t.Fatalf("unexpected message for account-b session: %s", msg.Method)
	}
}

func TestSessionEventDeletedSendsUpdatedAndAutoUnsubscribes(t *testing.T) {
	s := NewServer(nil)

	sessionA := newCaptureSession(t)
	defer sessionA.session.Close(false)

	s.sessionLock.Lock()
	s.sessions["a"] = trackedSession{
		session:   sessionA.session,
		accountID: "account-a",
	}
	s.sessionLock.Unlock()

	uri := chatURI("chat-1")
	s.subscriptions.AddSession("a", sessionA.session)
	s.subscriptions.Subscribe("a", sessionA.session, uri)

	s.handleSessionEvent(session.SessionEvent{
		Type:        session.SessionEventDeleted,
		SessionType: "thread",
		SessionID:   "chat-1",
		AccountID:   "account-a",
	})

	updated, ok := sessionA.read(2 * time.Second)
	if !ok {
		t.Fatal("expected resources/updated notification for deleted chat")
	}
	if updated.Method != "notifications/resources/updated" {
		t.Fatalf("updated.Method = %q, want %q", updated.Method, "notifications/resources/updated")
	}

	var payload struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(updated.Params, &payload); err != nil {
		t.Fatalf("failed to unmarshal updated payload: %v", err)
	}
	if payload.URI != uri {
		t.Fatalf("updated URI = %q, want %q", payload.URI, uri)
	}

	listChanged, ok := sessionA.read(2 * time.Second)
	if !ok {
		t.Fatal("expected list_changed notification for deleted chat")
	}
	if listChanged.Method != "notifications/resources/list_changed" {
		t.Fatalf("listChanged.Method = %q, want %q", listChanged.Method, "notifications/resources/list_changed")
	}

	s.subscriptions.SendResourceUpdatedNotification(uri)
	if msg, ok := sessionA.read(150 * time.Millisecond); ok {
		t.Fatalf("expected chat URI to be auto-unsubscribed, got message: %s", msg.Method)
	}
}

func TestSessionEventUpdatedSendsResourceUpdated(t *testing.T) {
	s := NewServer(nil)

	sessionA := newCaptureSession(t)
	defer sessionA.session.Close(false)

	uri := chatURI("chat-1")
	s.subscriptions.AddSession("a", sessionA.session)
	s.subscriptions.Subscribe("a", sessionA.session, uri)

	s.handleSessionEvent(session.SessionEvent{
		Type:        session.SessionEventUpdated,
		SessionType: "thread",
		SessionID:   "chat-1",
		AccountID:   "account-a",
	})

	updated, ok := sessionA.read(2 * time.Second)
	if !ok {
		t.Fatal("expected resources/updated notification for updated chat")
	}
	if updated.Method != "notifications/resources/updated" {
		t.Fatalf("updated.Method = %q, want %q", updated.Method, "notifications/resources/updated")
	}

	var payload struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(updated.Params, &payload); err != nil {
		t.Fatalf("failed to unmarshal updated payload: %v", err)
	}
	if payload.URI != uri {
		t.Fatalf("updated URI = %q, want %q", payload.URI, uri)
	}

	if msg, ok := sessionA.read(150 * time.Millisecond); ok {
		t.Fatalf("unexpected extra notification for updated chat: %s", msg.Method)
	}
}

type captureSession struct {
	session  *mcp.Session
	messages <-chan mcp.Message
}

func newCaptureSession(t *testing.T) captureSession {
	t.Helper()
	wire := &captureWire{
		sessionID: fmt.Sprintf("session-%d", atomic.AddUint64(&captureSessionCounter, 1)),
		sent:      make(chan mcp.Message, 16),
	}

	session, err := mcp.NewSession(context.Background(), "test", mcp.Server{}, mcp.ClientOption{
		Wire: wire,
	})
	if err != nil {
		t.Fatalf("failed to create capture session: %v", err)
	}

	return captureSession{
		session:  session,
		messages: wire.sent,
	}
}

func (s captureSession) read(timeout time.Duration) (mcp.Message, bool) {
	select {
	case msg := <-s.messages:
		return msg, true
	case <-time.After(timeout):
		return mcp.Message{}, false
	}
}

var captureSessionCounter uint64

type captureWire struct {
	sessionID string
	sent      chan mcp.Message
	ctx       context.Context
	cancel    context.CancelCauseFunc
}

func (w *captureWire) Close(bool) {
	if w.cancel != nil {
		w.cancel(fmt.Errorf("closed"))
	}
}

func (w *captureWire) Wait() {
	if w.ctx == nil {
		return
	}
	<-w.ctx.Done()
}

func (w *captureWire) Start(ctx context.Context, _ mcp.WireHandler) error {
	w.ctx, w.cancel = context.WithCancelCause(ctx)
	return nil
}

func (w *captureWire) Send(ctx context.Context, req mcp.Message) error {
	select {
	case w.sent <- req:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *captureWire) SessionID() string {
	return w.sessionID
}
