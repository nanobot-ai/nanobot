package obotmcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const testSessionID = "test-session-123"

type fakeConnectedServerLister struct {
	servers []ConnectedServer
	err     error
	count   *int
}

func (f fakeConnectedServerLister) ConnectedMCPServers(context.Context) ([]ConnectedServer, error) {
	if f.count != nil {
		*f.count = *f.count + 1
	}
	return f.servers, f.err
}

type sequenceConnectedServerLister struct {
	results [][]ConnectedServer
	errors  []error
	count   *int
}

func (s sequenceConnectedServerLister) ConnectedMCPServers(context.Context) ([]ConnectedServer, error) {
	call := 0
	if s.count != nil {
		*s.count = *s.count + 1
		call = *s.count - 1
	}

	var servers []ConnectedServer
	if call < len(s.results) {
		servers = s.results[call]
	}

	var err error
	if call < len(s.errors) {
		err = s.errors[call]
	}

	return servers, err
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	handler := mcp.MessageHandlerFunc(func(ctx context.Context, msg mcp.Message) {})
	serverSession, err := mcp.NewExistingServerSession(context.Background(),
		mcp.SessionState{ID: testSessionID}, handler)
	if err != nil {
		t.Fatalf("failed to create server session: %v", err)
	}
	return mcp.WithSession(context.Background(), serverSession.GetSession())
}

func TestAddMCPServer_ValidatesURL(t *testing.T) {
	s := NewServer("")

	tests := []struct {
		name    string
		url     string
		wantErr string
	}{
		{name: "empty URL", url: "", wantErr: "url is required"},
		{name: "invalid scheme", url: "ftp://example.com/mcp", wantErr: "URL must use http or https scheme"},
		{name: "no scheme", url: "example.com/mcp", wantErr: "URL must use http or https scheme"},
		{name: "valid https URL", url: "https://obot.example.com/mcp"},
		{name: "valid http URL", url: "http://obot.example.com/mcp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			session := mcp.NewEmptySession(ctx)
			session.Set(mcp.SessionEnvMapKey, map[string]string{
				"MCP_SERVER_SEARCH_URL": "https://obot.example.com/search",
			})
			ctx = mcp.WithSession(ctx, session)

			_, err := s.addMCPServer(ctx, AddMCPServerParams{URL: tt.url, Name: "test-server"})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if got := err.Error(); !strings.Contains(got, tt.wantErr) {
					t.Errorf("error = %q, want to contain %q", got, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddMCPServer_ValidatesHostMatch(t *testing.T) {
	s := NewServer("")

	tests := []struct {
		name      string
		serverURL string
		searchURL string
		wantErr   bool
	}{
		{name: "matching host", serverURL: "https://obot.example.com/mcp/server1", searchURL: "https://obot.example.com/search"},
		{name: "mismatching host", serverURL: "https://evil.example.com/mcp", searchURL: "https://obot.example.com/search", wantErr: true},
		{name: "matching host with port", serverURL: "https://obot.example.com:8443/mcp", searchURL: "https://obot.example.com:8443/search"},
		{name: "mismatching port", serverURL: "https://obot.example.com:9999/mcp", searchURL: "https://obot.example.com:8443/search", wantErr: true},
		{name: "no search URL configured", serverURL: "https://anything.example.com/mcp", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			session := mcp.NewEmptySession(ctx)
			envMap := map[string]string{}
			if tt.searchURL != "" {
				envMap["MCP_SERVER_SEARCH_URL"] = tt.searchURL
			}
			session.Set(mcp.SessionEnvMapKey, envMap)
			ctx = mcp.WithSession(ctx, session)

			_, err := s.addMCPServer(ctx, AddMCPServerParams{URL: tt.serverURL, Name: "test-server"})
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddMCPServer_ValidatesName(t *testing.T) {
	s := NewServer("")

	tests := []struct {
		name    string
		srvName string
		wantErr string
	}{
		{name: "empty name", wantErr: "name is required"},
		{name: "name with slash", srvName: "my/server", wantErr: "must not contain '/'"},
		{name: "reserved name nanobot.system", srvName: "nanobot.system", wantErr: "reserved"},
		{name: "valid name", srvName: "my-custom-server"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			session := mcp.NewEmptySession(ctx)
			session.Set(mcp.SessionEnvMapKey, map[string]string{"MCP_SERVER_SEARCH_URL": "https://obot.example.com/search"})
			ctx = mcp.WithSession(ctx, session)

			_, err := s.addMCPServer(ctx, AddMCPServerParams{URL: "https://obot.example.com/mcp", Name: tt.srvName})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if got := err.Error(); !strings.Contains(got, tt.wantErr) {
					t.Errorf("error = %q, want to contain %q", got, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRemoveMCPServer_NonExistentIsNotError(t *testing.T) {
	s := NewServer("")
	ctx := context.Background()
	session := mcp.NewEmptySession(ctx)
	ctx = mcp.WithSession(ctx, session)

	result, err := s.removeMCPServer(ctx, RemoveMCPServerParams{Name: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["success"] != true {
		t.Error("expected success to be true")
	}
}

func TestConfigureIntegrationAddsConfiguredMCPServersSnapshot(t *testing.T) {
	ctx := context.Background()
	session := mcp.NewEmptySession(ctx)
	session.Set(mcp.SessionEnvMapKey, map[string]string{
		"MCP_SERVER_SEARCH_URL": "https://search.example.com/mcp",
	})
	ctx = mcp.WithSession(ctx, session)

	agent := &types.HookAgent{
		Name: "test-agent",
		Instructions: types.DynamicInstructions{
			Instructions: "You are a helpful assistant.",
		},
	}
	params := types.AgentConfigHook{
		MCPServers: map[string]types.AgentConfigHookMCPServer{},
		Agent:      agent,
	}

	configureIntegration(ctx, agent, &params, fakeConnectedServerLister{servers: []ConnectedServer{
		{
			ID:          "gmail-1",
			Name:        "Gmail",
			Description: "Email access",
			ConnectURL:  "https://obot.example.com/mcp-connect/gmail-1",
		},
	}})

	if !strings.Contains(agent.Instructions.Instructions, "## Configured MCP Servers") {
		t.Fatal("expected configured MCP servers snapshot in instructions")
	}
	if !strings.Contains(agent.Instructions.Instructions, "`gmail`: Gmail - Email access") {
		t.Fatalf("expected sanitized mcp-cli server name and description in snapshot, got:\n%s", agent.Instructions.Instructions)
	}
}

func TestConfigureIntegrationCachesConfiguredMCPServersSnapshot(t *testing.T) {
	ctx := context.Background()
	session := mcp.NewEmptySession(ctx)
	session.Set(mcp.SessionEnvMapKey, map[string]string{
		"MCP_SERVER_SEARCH_URL": "https://search.example.com/mcp",
	})
	ctx = mcp.WithSession(ctx, session)

	fetchCount := 0
	lister := fakeConnectedServerLister{
		count: &fetchCount,
		servers: []ConnectedServer{{
			ID:         "gmail-1",
			Name:       "Gmail",
			ConnectURL: "https://obot.example.com/mcp-connect/gmail-1",
		}},
	}

	agent := &types.HookAgent{
		Name: "test-agent",
		Instructions: types.DynamicInstructions{
			Instructions: "You are a helpful assistant.",
		},
	}
	params := types.AgentConfigHook{
		MCPServers: map[string]types.AgentConfigHookMCPServer{},
		Agent:      agent,
	}

	configureIntegration(ctx, agent, &params, lister)
	configureIntegration(ctx, agent, &params, lister)

	if fetchCount != 1 {
		t.Fatalf("fetch count = %d, want 1", fetchCount)
	}
}

func TestConfigureIntegrationDoesNotCacheEmptySnapshotOnTransientFetchError(t *testing.T) {
	ctx := context.Background()
	session := mcp.NewEmptySession(ctx)
	session.Set(mcp.SessionEnvMapKey, map[string]string{
		"MCP_SERVER_SEARCH_URL": "https://search.example.com/mcp",
	})
	ctx = mcp.WithSession(ctx, session)

	fetchCount := 0
	lister := sequenceConnectedServerLister{
		count: &fetchCount,
		errors: []error{
			errors.New("temporary search failure"),
			nil,
		},
		results: [][]ConnectedServer{
			nil,
			{{
				ID:         "gmail-1",
				Name:       "Gmail",
				ConnectURL: "https://obot.example.com/mcp-connect/gmail-1",
			}},
		},
	}

	agent := &types.HookAgent{
		Name: "test-agent",
		Instructions: types.DynamicInstructions{
			Instructions: "You are a helpful assistant.",
		},
	}
	params := types.AgentConfigHook{
		MCPServers: map[string]types.AgentConfigHookMCPServer{},
		Agent:      agent,
	}

	configureIntegration(ctx, agent, &params, lister)
	if strings.Contains(agent.Instructions.Instructions, "## Configured MCP Servers") {
		t.Fatalf("did not expect configured MCP servers snapshot after transient failure, got:\n%s", agent.Instructions.Instructions)
	}

	configureIntegration(ctx, agent, &params, lister)
	if fetchCount != 2 {
		t.Fatalf("fetch count = %d, want 2", fetchCount)
	}
	if !strings.Contains(agent.Instructions.Instructions, "## Configured MCP Servers") {
		t.Fatalf("expected configured MCP servers snapshot after retry, got:\n%s", agent.Instructions.Instructions)
	}
}
