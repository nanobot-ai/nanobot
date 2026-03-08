package system

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestConfigHook_ImportsObotConfiguredServers(t *testing.T) {
	t.Helper()

	const (
		userToken = "user-token"
		mcpToken  = "ok1-test-mcp-token"
	)

	var authHeaders []string
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		authHeaders = append(authHeaders, req.Header.Get("Authorization"))
		rw.Header().Set("Content-Type", "application/json")

		switch req.URL.Path {
		case "/api/mcp-servers":
			_, _ = rw.Write([]byte(`{"items":[{"id":"github-id","alias":"GitHub Enterprise","configured":true,"connectURL":"https://obot.example.com/mcp-connect/github","manifest":{"name":"GitHub","shortDescription":"Query issues and repositories"}},{"id":"not-configured","configured":false,"connectURL":"https://obot.example.com/mcp-connect/skip","manifest":{"name":"Skip Me"}}]}`))
		case "/api/mcp-server-instances":
			_, _ = rw.Write([]byte(`{"items":[{"mcpServerID":"gmail-id","connectURL":"https://obot.example.com/mcp-connect/gmail-instance"}]}`))
		case "/api/all-mcps/servers":
			_, _ = rw.Write([]byte(`{"items":[{"id":"gmail-id","manifest":{"name":"Gmail","shortDescription":"Read and send email"}}]}`))
		default:
			http.NotFound(rw, req)
		}
	}))
	defer server.Close()

	s := NewServer("")
	ctx := context.Background()
	session := mcp.NewEmptySession(ctx)
	session.Set(mcp.SessionEnvMapKey, map[string]string{
		"OBOT_URL":          server.URL,
		"ANTHROPIC_API_KEY": userToken,
		"MCP_API_KEY":       mcpToken,
	})
	ctx = mcp.WithSession(ctx, session)

	result, err := s.config(ctx, types.AgentConfigHook{
		Agent: &types.HookAgent{
			Name: "test-agent",
			Instructions: types.DynamicInstructions{
				Instructions: "Base instructions.",
			},
		},
	})
	if err != nil {
		t.Fatalf("config() error = %v", err)
	}

	if len(authHeaders) != 3 {
		t.Fatalf("expected 3 Obot API calls, got %d", len(authHeaders))
	}

	for i, header := range authHeaders {
		if header != "Bearer "+userToken {
			t.Fatalf("request %d auth header = %q, want %q", i, header, "Bearer "+userToken)
		}
	}

	githubServer, ok := result.MCPServers["github-enterprise"]
	if !ok {
		t.Fatalf("expected imported server %q to exist", "github-enterprise")
	}
	if githubServer.URL != "https://obot.example.com/mcp-connect/github" {
		t.Fatalf("github-enterprise URL = %q", githubServer.URL)
	}
	if got := githubServer.Headers["Authorization"]; got != "Bearer "+mcpToken {
		t.Fatalf("github-enterprise auth header = %q, want %q", got, "Bearer "+mcpToken)
	}

	gmailServer, ok := result.MCPServers["gmail"]
	if !ok {
		t.Fatalf("expected imported server %q to exist", "gmail")
	}
	if gmailServer.URL != "https://obot.example.com/mcp-connect/gmail-instance" {
		t.Fatalf("gmail URL = %q", gmailServer.URL)
	}
	if got := gmailServer.Headers["Authorization"]; got != "Bearer "+mcpToken {
		t.Fatalf("gmail auth header = %q, want %q", got, "Bearer "+mcpToken)
	}

	if !strings.Contains(result.Agent.Instructions.Instructions, "## Connected MCP Servers") {
		t.Fatalf("expected connected MCP servers section in instructions")
	}
	if !strings.Contains(result.Agent.Instructions.Instructions, "github-enterprise") {
		t.Fatalf("expected imported server name in instructions, got %q", result.Agent.Instructions.Instructions)
	}

	for _, expectedServer := range []string{"github-enterprise", "gmail"} {
		found := false
		for _, serverName := range result.Agent.MCPServers {
			if serverName == expectedServer {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected agent MCPServers to include %q", expectedServer)
		}
	}
}

func TestConfigHook_ObotImportGracefulOnFailure(t *testing.T) {
	s := NewServer("")
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	session := mcp.NewEmptySession(ctx)
	session.Set(mcp.SessionEnvMapKey, map[string]string{
		"OBOT_URL":          server.URL,
		"ANTHROPIC_API_KEY": "user-token",
		"MCP_API_KEY":       "ok1-test-mcp-token",
	})
	ctx = mcp.WithSession(ctx, session)

	result, err := s.config(ctx, types.AgentConfigHook{
		Agent: &types.HookAgent{
			Name: "test-agent",
		},
	})
	if err != nil {
		t.Fatalf("config() error = %v", err)
	}

	if _, ok := result.MCPServers["github-enterprise"]; ok {
		t.Fatalf("did not expect imported Obot servers when API discovery fails")
	}
}
