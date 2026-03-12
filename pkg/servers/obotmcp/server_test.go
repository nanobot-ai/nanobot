package obotmcp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestRefreshMCPServerConfig(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	s := NewServer(".nanobot")
	ctx := testContext(t)
	lister := fakeConnectedServerLister{servers: []ConnectedServer{{
		ID:         "gmail-1",
		Name:       "Gmail",
		ConnectURL: "https://obot.example.com/mcp-connect/gmail-1",
	}}}

	_, err = prepareMCPCLIConfig(ctx, s.configDir, true, lister)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := s.refreshMCPServerConfig(ctx, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["success"] != true {
		t.Fatalf("success = %v, want true", result["success"])
	}
	if _, ok := result["config_path"]; ok {
		t.Fatalf("config_path should not be returned: %#v", result)
	}
	if _, ok := result["server_count"]; ok {
		t.Fatalf("server_count should not be returned: %#v", result)
	}
	if result["message"] != "Refreshed mcp-cli config." {
		t.Fatalf("message = %q, want refresh summary", result["message"])
	}
}

func TestPrepareMCPCLIConfigWritesAgentScopedConfig(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	ctx := testContext(t)
	lister := fakeConnectedServerLister{servers: []ConnectedServer{
		{
			ID:         "gmail-1",
			Name:       "Gmail",
			Alias:      "gmail",
			ConnectURL: "https://obot.example.com/mcp-connect/gmail-1",
		},
		{
			ID:         "shared-1",
			Name:       "Shared Server",
			ConnectURL: "https://obot.example.com/mcp-connect/shared-1",
		},
	}}

	path, err := prepareMCPCLIConfig(ctx, ".nanobot", false, lister)
	if err != nil {
		t.Fatalf("PrepareMCPCLIConfig failed: %v", err)
	}

	expectedSuffix := filepath.Join(".nanobot", "mcp-cli", "config.json")
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Fatalf("config path = %q, want suffix %q", path, expectedSuffix)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read generated config: %v", err)
	}

	content := string(data)
	for _, expected := range []string{
		`"gmail"`,
		`"shared-server"`,
		`"https://obot.example.com/mcp-connect/gmail-1"`,
		`"https://obot.example.com/mcp-connect/shared-1"`,
		`"Authorization": "Bearer ${MCP_API_KEY}"`,
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("generated config missing %q:\n%s", expected, content)
		}
	}
}

func TestPrepareMCPCLIConfigKeepsCachedFileOnRefreshFailure(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	ctx := testContext(t)
	lister := fakeConnectedServerLister{servers: []ConnectedServer{{
		ID:         "gmail-1",
		Name:       "Gmail",
		ConnectURL: "https://obot.example.com/mcp-connect/gmail-1",
	}}}

	path, err := prepareMCPCLIConfig(ctx, ".nanobot", false, lister)
	if err != nil {
		t.Fatalf("initial PrepareMCPCLIConfig failed: %v", err)
	}

	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read initial config: %v", err)
	}

	path2, err := prepareMCPCLIConfig(ctx, ".nanobot", false, fakeConnectedServerLister{err: errors.New("search server unavailable")})
	if err != nil {
		t.Fatalf("PrepareMCPCLIConfig with cached file failed: %v", err)
	}
	if path2 != path {
		t.Fatalf("config path changed: %q != %q", path2, path)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cached config: %v", err)
	}
	if string(current) != string(original) {
		t.Fatalf("cached config changed unexpectedly:\nold:\n%s\nnew:\n%s", original, current)
	}
}

func TestPrepareMCPCLIConfigSkipsRefreshWhenCacheIsFresh(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	ctx := testContext(t)
	fetchCount := 0
	lister := fakeConnectedServerLister{
		count: &fetchCount,
		servers: []ConnectedServer{{
			ID:         "gmail-1",
			Name:       "Gmail",
			ConnectURL: "https://obot.example.com/mcp-connect/gmail-1",
		}},
	}

	path, err := prepareMCPCLIConfig(ctx, ".nanobot", false, lister)
	if err != nil {
		t.Fatalf("initial PrepareMCPCLIConfig failed: %v", err)
	}
	if fetchCount != 1 {
		t.Fatalf("fetch count after initial prepare = %d, want 1", fetchCount)
	}

	path2, err := prepareMCPCLIConfig(ctx, ".nanobot", false, lister)
	if err != nil {
		t.Fatalf("second PrepareMCPCLIConfig failed: %v", err)
	}
	if path2 != path {
		t.Fatalf("config path changed: %q != %q", path2, path)
	}
	if fetchCount != 1 {
		t.Fatalf("fetch count after fresh-cache prepare = %d, want 1", fetchCount)
	}
}

func TestPrepareMCPCLIConfigRefreshesWhenCacheIsStale(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	ctx := testContext(t)
	fetchCount := 0
	lister := fakeConnectedServerLister{
		count: &fetchCount,
		servers: []ConnectedServer{{
			ID:         "gmail-1",
			Name:       "Gmail",
			ConnectURL: "https://obot.example.com/mcp-connect/gmail-1",
		}},
	}

	path, err := prepareMCPCLIConfig(ctx, ".nanobot", false, lister)
	if err != nil {
		t.Fatalf("initial PrepareMCPCLIConfig failed: %v", err)
	}
	if fetchCount != 1 {
		t.Fatalf("fetch count after initial prepare = %d, want 1", fetchCount)
	}

	staleTime := time.Now().Add(-configRefreshInterval - time.Minute)
	if err := os.Chtimes(path, staleTime, staleTime); err != nil {
		t.Fatalf("failed to mark config stale: %v", err)
	}

	_, err = prepareMCPCLIConfig(ctx, ".nanobot", false, lister)
	if err != nil {
		t.Fatalf("stale PrepareMCPCLIConfig failed: %v", err)
	}
	if fetchCount != 2 {
		t.Fatalf("fetch count after stale-cache prepare = %d, want 2", fetchCount)
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
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

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

	configureIntegration(ctx, ".nanobot", agent, &params, fakeConnectedServerLister{servers: []ConnectedServer{
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
	if _, err := os.Stat(filepath.Join(".nanobot", "mcp-cli", "config.json")); err != nil {
		t.Fatalf("expected mcp-cli config to be prepared during integration setup: %v", err)
	}
}

func TestConfigureIntegrationCachesConfiguredMCPServersSnapshot(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

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

	configureIntegration(ctx, ".nanobot", agent, &params, lister)
	configureIntegration(ctx, ".nanobot", agent, &params, lister)

	if fetchCount != 1 {
		t.Fatalf("fetch count = %d, want 1", fetchCount)
	}
}
