package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/obot-platform/nanobot/pkg/config"
	"github.com/obot-platform/nanobot/pkg/mcp"
)

func TestExternalUIRequest(t *testing.T) {
	for _, test := range []struct {
		path string
		want bool
	}{
		{path: "/mcp"},
		{path: "/mcp/chat"},
		{path: "/mcp/ui", want: true},
	} {
		t.Run(test.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, test.path, nil)
			if got := isExternalUIRequest(mcp.WithRequest(t.Context(), req)); got != test.want {
				t.Fatalf("isExternalUIRequest() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNanobotConfigPathsDefault(t *testing.T) {
	n := &Nanobot{}

	paths := n.ConfigPaths()
	if len(paths) != 1 || paths[0] != config.DefaultConfigPath {
		t.Fatalf("expected default config path [.nanobot/], got %v", paths)
	}
}

func TestRuntimeConfigDirDefaultsWhenNoPathsExist(t *testing.T) {
	configDir := runtimeConfigDir([]string{"./missing", "https://example.com/nanobot.yaml"})
	if configDir != config.DefaultConfigPath {
		t.Fatalf("expected default config dir %q, got %q", config.DefaultConfigPath, configDir)
	}
}

func TestRuntimeConfigDirKeepsDefaultForExistingDirectory(t *testing.T) {
	dir := t.TempDir()

	configDir := runtimeConfigDir([]string{"./missing", dir})
	if configDir != config.DefaultConfigPath {
		t.Fatalf("expected default config dir %q, got %q", config.DefaultConfigPath, configDir)
	}
}

func TestRuntimeConfigDirReturnsFileParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nanobot.yaml")
	if err := os.WriteFile(path, []byte("agents: {}\n"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	configDir := runtimeConfigDir([]string{path})
	if configDir != dir {
		t.Fatalf("expected config dir %q, got %q", dir, configDir)
	}
}
