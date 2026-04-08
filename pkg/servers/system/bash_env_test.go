package system

import (
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

func TestObotMCPBashEnvVarsAddsAPIKeyWithoutMCPCLIRefresh(t *testing.T) {
	server := NewServer("", ".nanobot")
	ctx := testContext(t)
	session := mcp.SessionFromContext(ctx)
	session.SetEnv(map[string]string{
		"MCP_API_KEY":           "token-123",
		"MCP_SERVER_SEARCH_URL": "https://search.example.com/mcp",
	})

	env, err := server.obotMCPBashEnvVars(ctx, "echo hi")
	if err != nil {
		t.Fatalf("obotMCPBashEnvVars failed: %v", err)
	}

	joined := strings.Join(env, "\n")
	if !strings.Contains(joined, "MCP_API_KEY=token-123") {
		t.Fatalf("MCP_API_KEY missing from env:\n%s", joined)
	}
	if strings.Contains(joined, "MCP_CONFIG_PATH=") {
		t.Fatalf("MCP_CONFIG_PATH should be absent when command does not invoke mcp-cli:\n%s", joined)
	}
}

func TestObotMCPBashEnvVarsReturnsRefreshError(t *testing.T) {
	server := NewServer("", ".nanobot")
	ctx := testContext(t)
	session := mcp.SessionFromContext(ctx)
	session.SetEnv(map[string]string{
		"MCP_API_KEY":           "token-123",
		"MCP_SERVER_SEARCH_URL": "://bad",
	})

	_, err := server.obotMCPBashEnvVars(ctx, "mcp-cli info gmail")
	if err == nil {
		t.Fatal("expected refresh error, got nil")
	}
}

func TestObotMCPBashEnvVarsSkipsConfigWhenRefreshPrerequisitesAreMissing(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "without search URL",
			env: map[string]string{
				"MCP_API_KEY": "token-123",
			},
		},
		{
			name: "without API key",
			env: map[string]string{
				"MCP_SERVER_SEARCH_URL": "https://search.example.com/mcp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer("", ".nanobot")
			ctx := testContext(t)
			session := mcp.SessionFromContext(ctx)
			session.SetEnv(tt.env)

			env, err := server.obotMCPBashEnvVars(ctx, "mcp-cli info gmail")
			if err != nil {
				t.Fatalf("obotMCPBashEnvVars failed: %v", err)
			}

			joined := strings.Join(env, "\n")
			if strings.Contains(joined, "MCP_CONFIG_PATH=") {
				t.Fatalf("MCP_CONFIG_PATH should be absent when refresh prerequisites are missing:\n%s", joined)
			}
		})
	}
}
