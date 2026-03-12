package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

func TestAugmentBashEnvForMCPCLIAddsConfigPath(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	server := NewServer(".nanobot")
	ctx := testContext(t)
	configPath := filepath.Join(tmpDir, ".nanobot", "mcp-cli", "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("{\"mcpServers\":{}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := mcp.SessionFromContext(ctx)
	session.SetEnv(map[string]string{
		"MCP_API_KEY":           "token-123",
		"MCP_SERVER_SEARCH_URL": "https://search.example.com/mcp",
	})

	env, err := server.augmentBashEnvForObotMCPIntegration(ctx, "mcp-cli info gmail")
	if err != nil {
		t.Fatalf("augmentBashEnvForMCPCLI failed: %v", err)
	}

	joined := strings.Join(env, "\n")
	if !strings.Contains(joined, "MCP_API_KEY=token-123") {
		t.Fatalf("MCP_API_KEY missing from env:\n%s", joined)
	}
	expectedSuffix := filepath.Join(".nanobot", "mcp-cli", "config.json")
	foundConfigPath := false
	for _, entry := range env {
		if !strings.HasPrefix(entry, "MCP_CONFIG_PATH=") {
			continue
		}
		foundConfigPath = true
		if !strings.HasSuffix(strings.TrimPrefix(entry, "MCP_CONFIG_PATH="), expectedSuffix) {
			t.Fatalf("MCP_CONFIG_PATH = %q, want suffix %q", entry, expectedSuffix)
		}
	}
	if !foundConfigPath {
		t.Fatalf("MCP_CONFIG_PATH missing from env:\n%s", joined)
	}
}

func TestAugmentBashEnvForMCPCLISkipsConfigWithoutSearchURL(t *testing.T) {
	server := NewServer(".nanobot")
	ctx := testContext(t)
	session := mcp.SessionFromContext(ctx)
	session.SetEnv(map[string]string{
		"MCP_API_KEY": "token-123",
	})

	env, err := server.augmentBashEnvForObotMCPIntegration(ctx, "mcp-cli info gmail")
	if err != nil {
		t.Fatalf("augmentBashEnvForMCPCLI failed: %v", err)
	}

	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "MCP_CONFIG_PATH=") {
		t.Fatalf("MCP_CONFIG_PATH should be absent when MCP_SERVER_SEARCH_URL is missing:\n%s", joined)
	}
}

func TestAugmentBashEnvForMCPCLISkipsConfigWithoutAPIKey(t *testing.T) {
	server := NewServer(".nanobot")
	ctx := testContext(t)
	session := mcp.SessionFromContext(ctx)
	session.SetEnv(map[string]string{
		"MCP_SERVER_SEARCH_URL": "https://search.example.com/mcp",
	})

	env, err := server.augmentBashEnvForObotMCPIntegration(ctx, "mcp-cli info gmail")
	if err != nil {
		t.Fatalf("augmentBashEnvForMCPCLI failed: %v", err)
	}

	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "MCP_CONFIG_PATH=") {
		t.Fatalf("MCP_CONFIG_PATH should be absent when MCP_API_KEY is missing:\n%s", joined)
	}
}
