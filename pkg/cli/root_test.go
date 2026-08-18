package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/obot-platform/nanobot/pkg/config"
	"github.com/obot-platform/nanobot/pkg/types"
)

func TestNanobotConfigPathsDefault(t *testing.T) {
	n := &Nanobot{}

	paths := n.ConfigPaths()
	if len(paths) != 1 || paths[0] != config.DefaultConfigPath {
		t.Fatalf("expected default config path [.nanobot/], got %v", paths)
	}
}

func TestNanobotLLMConfigBuiltinOrcaRouter(t *testing.T) {
	n := &Nanobot{}

	cfg := n.llmConfig()
	orcarouter, ok := cfg.LLMProviders["orcarouter"]
	if !ok {
		t.Fatalf("expected built-in orcarouter provider, got %v", cfg.LLMProviders)
	}
	if orcarouter.Dialect != types.DialectOpenAIResponses {
		t.Errorf("orcarouter dialect: got %q, want %q", orcarouter.Dialect, types.DialectOpenAIResponses)
	}
	if orcarouter.APIKey != "${ORCAROUTER_API_KEY}" {
		t.Errorf("orcarouter apiKey: got %q, want %q", orcarouter.APIKey, "${ORCAROUTER_API_KEY}")
	}
	if orcarouter.BaseURL != "https://api.orcarouter.ai/v1" {
		t.Errorf("orcarouter baseURL: got %q, want %q", orcarouter.BaseURL, "https://api.orcarouter.ai/v1")
	}
}

func TestNanobotLLMConfigBuiltinProviders(t *testing.T) {
	n := &Nanobot{}

	cfg := n.llmConfig()
	for _, name := range []string{"openai", "anthropic", "orcarouter"} {
		if _, ok := cfg.LLMProviders[name]; !ok {
			t.Errorf("expected built-in provider %q", name)
		}
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
