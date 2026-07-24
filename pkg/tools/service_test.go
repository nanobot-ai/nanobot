package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
)

func TestClientFactoryCloseIsTerminal(t *testing.T) {
	var createCount atomic.Int32
	factory := newClientFactory(func(*mcp.SessionState) (*mcp.Client, error) {
		createCount.Add(1)
		return nil, nil
	})

	factory.Close(true)
	client, err := factory.get(context.Background(), "env")

	if !errors.Is(err, errClientFactoryClosed) {
		t.Fatalf("get error = %v, want %v", err, errClientFactoryClosed)
	}
	if client != nil {
		t.Fatal("closed factory returned a client")
	}
	if got := createCount.Load(); got != 0 {
		t.Fatalf("client create count = %d, want 0", got)
	}
}

func TestClientFactoryCloseWithoutDeleteRemainsReusable(t *testing.T) {
	var createCount atomic.Int32
	factory := newClientFactory(func(*mcp.SessionState) (*mcp.Client, error) {
		createCount.Add(1)
		return nil, nil
	})

	factory.Close(false)
	if _, err := factory.get(context.Background(), "env"); err != nil {
		t.Fatalf("get after non-deleting close returned error: %v", err)
	}
	if got := createCount.Load(); got != 1 {
		t.Fatalf("client create count = %d, want 1", got)
	}
}

func TestClientFactoryRejectsCanceledOwnerBeforeResume(t *testing.T) {
	ownerCtx, cancel := context.WithCancelCause(context.Background())
	ownerErr := errors.New("owner session deleted")
	cancel(ownerErr)

	var createCount atomic.Int32
	factory := newClientFactory(func(*mcp.SessionState) (*mcp.Client, error) {
		createCount.Add(1)
		return nil, nil
	})
	factory.state.oldState = &mcp.SessionState{ID: "persisted-downstream-session"}

	client, err := factory.get(ownerCtx, "env")

	if !errors.Is(err, errClientFactoryClosed) || !errors.Is(err, ownerErr) {
		t.Fatalf("get error = %v, want closed factory and owner errors", err)
	}
	if client != nil {
		t.Fatal("factory with canceled owner returned a client")
	}
	if got := createCount.Load(); got != 0 {
		t.Fatalf("client create count = %d, want 0", got)
	}
}

func TestClientFactoriesRegisterAtomically(t *testing.T) {
	session := mcp.NewEmptySession(context.Background())
	start := make(chan struct{})
	states := make(chan *clientFactoryState, 2)

	for range 2 {
		factory := newClientFactory(func(*mcp.SessionState) (*mcp.Client, error) {
			return nil, errors.New("unexpected client creation")
		})
		go func() {
			<-start
			if !session.GetOrSetSessionCloser("clients/downstream", &factory, &factory) {
				states <- nil
				return
			}
			states <- factory.state
		}()
	}
	close(start)

	first, second := <-states, <-states
	if first == nil || second == nil {
		t.Fatal("client factory registration unexpectedly failed")
	}
	if first != second {
		t.Fatal("concurrent registrations resolved to different client factories")
	}

	session.Close(true)
	closedFactory := clientFactory{state: first}
	if _, err := closedFactory.get(context.Background(), "env"); !errors.Is(err, errClientFactoryClosed) {
		t.Fatalf("get error after session deletion = %v, want %v", err, errClientFactoryClosed)
	}
}

func TestGetClientDoesNotHydrateAfterParentDelete(t *testing.T) {
	var requestCount atomic.Int32
	downstream := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		rw.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(downstream.Close)

	parent, err := mcp.NewServerSession(context.Background(), mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}))
	if err != nil {
		t.Fatal(err)
	}
	parent.GetSession().Set("clients/downstream", mcp.SessionState{ID: "persisted-downstream-session"})

	service := NewToolsService()
	clientCtx := types.WithConfig(parent.GetSession().Context(), types.Config{
		MCPServers: map[string]mcp.Server{
			"downstream": {BaseURL: downstream.URL},
		},
	})
	parent.Close(true)

	client, err := service.GetClient(clientCtx, "downstream")

	if !errors.Is(err, errClientFactoryClosed) {
		t.Fatalf("GetClient error = %v, want %v", err, errClientFactoryClosed)
	}
	if client != nil {
		t.Fatal("GetClient returned a client after parent deletion")
	}
	if got := requestCount.Load(); got != 0 {
		t.Fatalf("downstream request count = %d, want 0", got)
	}
}

func TestHTTPServerDeleteClosesDownstreamHTTPSession(t *testing.T) {
	const (
		downstreamSessionID = "downstream-session-2"
		passthroughHeader   = "X-Downstream-Auth"
		passthroughValue    = "session-token"
	)

	type deleteRequest struct {
		sessionID   string
		passthrough string
	}
	deleteRequests := make(chan deleteRequest, 1)
	var deleteCount atomic.Int32
	var initializeCount atomic.Int32
	downstream := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodDelete:
			deleteCount.Add(1)
			select {
			case deleteRequests <- deleteRequest{
				sessionID:   req.Header.Get(mcp.SessionIDHeader),
				passthrough: req.Header.Get(passthroughHeader),
			}:
			default:
			}
			rw.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			// The client may probe for an event stream after initialization.
			rw.WriteHeader(http.StatusMethodNotAllowed)
		case http.MethodPost:
			var msg mcp.Message
			if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
			if msg.Method != "initialize" {
				rw.WriteHeader(http.StatusAccepted)
				return
			}

			result, err := json.Marshal(mcp.InitializeResult{
				ProtocolVersion: "2025-06-18",
				ServerInfo: mcp.ServerInfo{
					Name:    "downstream",
					Version: "test",
				},
			})
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			rw.Header().Set(mcp.SessionIDHeader, fmt.Sprintf("downstream-session-%d", initializeCount.Add(1)))
			rw.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(rw).Encode(mcp.Message{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Result:  result,
			})
		default:
			rw.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(downstream.Close)

	parent, err := mcp.NewServerSession(context.Background(), mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Close(false) })

	service := NewToolsService()
	clientCtx := types.WithConfig(parent.GetSession().Context(), types.Config{
		MCPServers: map[string]mcp.Server{
			"downstream": {
				BaseURL:            downstream.URL,
				PassthroughHeaders: []string{passthroughHeader},
			},
		},
	})
	incoming := httptest.NewRequest(http.MethodPost, "http://nanobot.example/mcp", nil)
	incoming.Header.Set(passthroughHeader, passthroughValue)
	clientCtx = mcp.WithRequest(clientCtx, incoming)
	downstreamClient, err := service.GetClient(clientCtx, "downstream")
	if err != nil {
		t.Fatalf("failed to create downstream client: %v", err)
	}
	if got := downstreamClient.Session.ID(); got != "downstream-session-1" {
		t.Fatalf("first downstream session ID = %q, want %q", got, "downstream-session-1")
	}

	// Force the factory through its copied-value replacement path. The session
	// attribute must keep ownership of the replacement client so DELETE closes
	// this second remote session rather than the stale first one.
	downstreamClient.Close(false)
	downstreamClient, err = service.GetClient(clientCtx, "downstream")
	if err != nil {
		t.Fatalf("failed to replace downstream client: %v", err)
	}
	if got := downstreamClient.Session.ID(); got != downstreamSessionID {
		t.Fatalf("replacement downstream session ID = %q, want %q", got, downstreamSessionID)
	}
	if got := deleteCount.Load(); got != 0 {
		t.Fatalf("downstream DELETE count before parent deletion = %d, want 0", got)
	}

	sessions := mcp.NewInMemorySessionStore()
	if err := sessions.Store(context.Background(), parent.ID(), parent); err != nil {
		t.Fatal(err)
	}
	front, err := mcp.NewHTTPServer(nil, mcp.MessageHandlerFunc(func(context.Context, mcp.Message) {}), mcp.HTTPServerOptions{
		SessionStore: sessions,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "http://nanobot.example/mcp", nil)
	req.Header.Set(mcp.SessionIDHeader, parent.ID())
	rw := httptest.NewRecorder()
	front.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("front DELETE status = %d, want %d; body: %s", rw.Code, http.StatusOK, rw.Body.String())
	}
	select {
	case got := <-deleteRequests:
		if got.sessionID != downstreamSessionID {
			t.Fatalf("downstream DELETE session ID = %q, want %q", got.sessionID, downstreamSessionID)
		}
		if got.passthrough != passthroughValue {
			t.Fatalf("downstream DELETE passthrough header = %q, want %q", got.passthrough, passthroughValue)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for downstream DELETE")
	}
	if got := deleteCount.Load(); got != 1 {
		t.Fatalf("downstream DELETE count = %d, want 1", got)
	}
}

func testConfig() types.Config {
	return types.Config{
		Agents: map[string]types.Agent{
			"test-agent": {
				HookAgent: types.HookAgent{
					MaxTokens: 1234,
				},
			},
		},
	}
}

func TestConvertToSampleRequestWithFileAttachments(t *testing.T) {
	s := &Service{}

	req, err := s.convertToSampleRequest(testConfig(), "test-agent", map[string]any{
		"prompt": "Check the attached files",
		"attachments": []any{
			map[string]any{"url": "file:///notes/todo.md"},
			map[string]any{"url": "file:///notes/todo.md"},
			map[string]any{"url": "file:///docs/spec%20draft.md"},
		},
	})
	if err != nil {
		t.Fatalf("convertToSampleRequest returned error: %v", err)
	}

	if len(req.Messages) != 2 {
		t.Fatalf("expected 2 messages (prompt+preview and hidden attachment context), got %d", len(req.Messages))
	}

	first := req.Messages[0]
	if len(first.Content) != 3 {
		t.Fatalf("expected prompt message to include text + 2 previews, got %d items", len(first.Content))
	}
	if first.Content[0].Type != "text" || first.Content[0].Text != "Check the attached files" {
		t.Fatalf("unexpected first content item: %#v", first.Content[0])
	}
	if first.Content[1].Type != "resource_link" || first.Content[1].URI != "file:///notes/todo.md" {
		t.Fatalf("unexpected first attachment preview: %#v", first.Content[1])
	}
	if first.Content[2].Type != "resource_link" || first.Content[2].URI != "file:///docs/spec%20draft.md" {
		t.Fatalf("unexpected second attachment preview: %#v", first.Content[2])
	}

	second := req.Messages[1]
	if len(second.Content) != 1 || second.Content[0].Type != "text" {
		t.Fatalf("expected hidden attachment text message, got %#v", second.Content)
	}
	if second.Content[0].Meta == nil || second.Content[0].Meta[types.AttachmentMetaKey] != true {
		t.Fatalf("expected attachment meta key %q=true, got %#v", types.AttachmentMetaKey, second.Content[0].Meta)
	}
	if !strings.Contains(second.Content[0].Text, `"notes/todo.md"`) ||
		!strings.Contains(second.Content[0].Text, `"docs/spec draft.md"`) {
		t.Fatalf("expected hidden message to include decoded paths, got: %q", second.Content[0].Text)
	}
}

func TestConvertToSampleRequestDataAttachmentStillInlined(t *testing.T) {
	s := &Service{}

	req, err := s.convertToSampleRequest(testConfig(), "test-agent", map[string]any{
		"prompt": "Review this image",
		"attachments": []any{
			map[string]any{
				"url":      "data:image/png;base64,ZmFrZQ==",
				"name":     "test.png",
				"mimeType": "image/png",
			},
		},
	})
	if err != nil {
		t.Fatalf("convertToSampleRequest returned error: %v", err)
	}

	if len(req.Messages) != 2 {
		t.Fatalf("expected 2 messages (prompt + data attachment), got %d", len(req.Messages))
	}

	dataMsg := req.Messages[1]
	if len(dataMsg.Content) != 1 {
		t.Fatalf("expected 1 content item for data attachment, got %d", len(dataMsg.Content))
	}
	if dataMsg.Content[0].Type != "image" || dataMsg.Content[0].Data != "ZmFrZQ==" {
		t.Fatalf("unexpected data attachment item: %#v", dataMsg.Content[0])
	}
}

func TestConvertToSampleRequestRejectsInvalidAttachmentURL(t *testing.T) {
	s := &Service{}

	_, err := s.convertToSampleRequest(testConfig(), "test-agent", map[string]any{
		"prompt": "Check this",
		"attachments": []any{
			map[string]any{"url": "https://example.com/file.pdf"},
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid attachment URL, got nil")
	}
	if !strings.Contains(err.Error(), "only data URI and file:/// URIs are supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddHookMutationContent(t *testing.T) {
	response := &types.CallResult{
		Meta: map[string]any{
			mcp.HookMutationsMetaKey: map[string]any{
				"request": map[string]any{
					"mutated": true,
					"reasons": []any{"request normalized", "request redacted"},
				},
				"response": map[string]any{
					"mutated": true,
					"reasons": []any{"response filtered"},
				},
			},
		},
		Content: []mcp.Content{
			{Type: "text", Text: "tool output"},
		},
	}

	addHookMutationContent(response)

	if len(response.Content) != 3 {
		t.Fatalf("unexpected content length: %#v", response.Content)
	}
	if response.Content[0].Text != "MCP request was mutated by hooks. Reasons: request normalized; request redacted" {
		t.Fatalf("request mutation content was not prepended: %#v", response.Content)
	}
	if response.Content[1].Text != "tool output" {
		t.Fatalf("original tool output was not preserved in the middle: %#v", response.Content)
	}
	if response.Content[2].Text != "MCP response was mutated by hooks. Reasons: response filtered" {
		t.Fatalf("response mutation content was not appended: %#v", response.Content)
	}
}

func TestAddHookMutationContentOmitsUnmutatedDirections(t *testing.T) {
	response := &types.CallResult{
		Meta: map[string]any{
			mcp.HookMutationsMetaKey: map[string]mcp.HookMutation{
				"request": {Mutated: true, Reasons: []string{"request normalized"}},
			},
		},
		Content: []mcp.Content{{Type: "text", Text: "tool output"}},
	}

	addHookMutationContent(response)

	if len(response.Content) != 2 {
		t.Fatalf("unexpected content length: %#v", response.Content)
	}
	if response.Content[0].Text != "MCP request was mutated by hooks. Reasons: request normalized" {
		t.Fatalf("request mutation content was not prepended: %#v", response.Content)
	}
	if response.Content[1].Text != "tool output" {
		t.Fatalf("original tool output was not preserved after request mutation content: %#v", response.Content)
	}
}
