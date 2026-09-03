package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/obot-platform/nanobot/pkg/config"
	"github.com/obot-platform/nanobot/pkg/types"
)

func TestNanobotLLMConfigIncludesMiniMax(t *testing.T) {
	cfg := (&Nanobot{}).llmConfig()
	provider, ok := cfg.LLMProviders["minimax"]
	if !ok {
		t.Fatal("expected MiniMax provider config")
	}
	if provider.Dialect != types.DialectOpenAIChatCompletions {
		t.Fatalf("expected MiniMax dialect %q, got %q", types.DialectOpenAIChatCompletions, provider.Dialect)
	}
	if provider.APIKey != "${MINIMAX_API_KEY}" {
		t.Fatalf("unexpected MiniMax API key placeholder: %q", provider.APIKey)
	}
	if provider.BaseURL != "${MINIMAX_BASE_URL}" {
		t.Fatalf("unexpected MiniMax base URL: %q", provider.BaseURL)
	}
	if provider.Region != "${MINIMAX_REGION}" {
		t.Fatalf("unexpected MiniMax region placeholder: %q", provider.Region)
	}

	endpoints := cfg.ProviderEndpoints["minimax"]
	if len(endpoints) != 2 {
		t.Fatalf("expected two MiniMax regional endpoints, got %d", len(endpoints))
	}
	if endpoints[0].Region != "global_en" || endpoints[0].OpenAIBaseURL != "https://api.minimax.io/v1" || endpoints[0].AnthropicBaseURL != "https://api.minimax.io/anthropic" || endpoints[0].DocsRoot != "https://platform.minimax.io/docs" {
		t.Fatalf("unexpected MiniMax global endpoint: %+v", endpoints[0])
	}
	if endpoints[1].Region != "cn_zh" || endpoints[1].OpenAIBaseURL != "https://api.minimaxi.com/v1" || endpoints[1].AnthropicBaseURL != "https://api.minimaxi.com/anthropic" || endpoints[1].DocsRoot != "https://platform.minimaxi.com/docs" {
		t.Fatalf("unexpected MiniMax China endpoint: %+v", endpoints[1])
	}

	m3 := cfg.Models["minimax"]["MiniMax-M3"]
	if m3.ContextWindow != 1_000_000 || m3.Pricing.Input != 0.6 || m3.Pricing.Output != 2.4 || m3.Pricing.CacheRead != 0.12 || m3.Pricing.CacheWrite != nil {
		t.Fatalf("unexpected MiniMax-M3 model config: %+v", m3)
	}
	if !slices.Equal(m3.InputModalities, []string{"text", "image", "video"}) || !slices.Equal(m3.Thinking, []string{"adaptive", "disabled"}) {
		t.Fatalf("unexpected MiniMax-M3 capabilities: %+v", m3)
	}

	m27 := cfg.Models["minimax"]["MiniMax-M2.7"]
	if m27.ContextWindow != 204_800 || m27.Pricing.Input != 0.3 || m27.Pricing.Output != 1.2 || m27.Pricing.CacheRead != 0.06 || m27.Pricing.CacheWrite == nil || *m27.Pricing.CacheWrite != 0.375 {
		t.Fatalf("unexpected MiniMax-M2.7 model config: %+v", m27)
	}
	if !slices.Equal(m27.InputModalities, []string{"text"}) || !slices.Equal(m27.Thinking, []string{"always_on"}) {
		t.Fatalf("unexpected MiniMax-M2.7 capabilities: %+v", m27)
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
