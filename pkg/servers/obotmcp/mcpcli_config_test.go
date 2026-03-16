package obotmcp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestPrepareMCPCLIConfigReturnsRefreshFailureEvenWithCachedFile(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected refresh error, got nil")
	}
	if path2 != "" {
		t.Fatalf("config path = %q, want empty path on error", path2)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cached config: %v", err)
	}
	if string(current) != string(original) {
		t.Fatalf("cached config changed unexpectedly:\nold:\n%s\nnew:\n%s", original, current)
	}
}
