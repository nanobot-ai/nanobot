package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/obot-platform/nanobot/pkg/config"
	"github.com/obot-platform/nanobot/pkg/types"
)

func TestNanobotLLMConfigIncludesMiniMax(t *testing.T) {
	provider := (&Nanobot{}).llmConfig().LLMProviders["minimax"]
	if provider.Dialect != types.DialectOpenAIChatCompletions {
		t.Fatalf("expected MiniMax to use %q, got %q", types.DialectOpenAIChatCompletions, provider.Dialect)
	}
	if provider.APIKey != "${MINIMAX_API_KEY}" {
		t.Fatalf("expected MiniMax API key placeholder, got %q", provider.APIKey)
	}
	if provider.BaseURL != "https://api.minimax.io/v1" {
		t.Fatalf("expected MiniMax global API endpoint, got %q", provider.BaseURL)
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
