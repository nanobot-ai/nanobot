package config

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/types"
	"sigs.k8s.io/yaml"
)

// unmarshalConfig parses either YAML or JSON bytes into a types.Config.
func unmarshalConfig(t *testing.T, data []byte) types.Config {
	t.Helper()
	// Try JSON first (loadFromDirectory returns JSON), fall back to YAML.
	var cfg types.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		obj := map[string]any{}
		if err := yaml.Unmarshal(data, &obj); err != nil {
			t.Fatalf("failed to unmarshal config: %v", err)
		}
		out, err := json.Marshal(obj)
		if err != nil {
			t.Fatalf("failed to re-marshal config: %v", err)
		}
		if err := json.Unmarshal(out, &cfg); err != nil {
			t.Fatalf("failed to unmarshal config after YAML round-trip: %v", err)
		}
	}
	return cfg
}

// TestResourceRead_FileSkipsDirectoryLogic verifies that when the resource URL
// points directly to a YAML file, markdown agents are NOT loaded even though
// the parent directory contains an agents/ subdirectory with .md files.
func TestResourceRead_FileSkipsDirectoryLogic(t *testing.T) {
	r := &resource{
		resourceType: "path",
		url:          "testdata/directory-yaml-and-markdown/nanobot.yaml",
	}

	data, err := r.read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := unmarshalConfig(t, data)

	// The file contains only mcpServers — agents come from the markdown files
	// in agents/, which should NOT be loaded when pointing to a file.
	if len(cfg.Agents) != 0 {
		t.Errorf("expected 0 agents when reading file directly (markdown should be skipped), got %d: %v", len(cfg.Agents), cfg.Agents)
	}

	if _, ok := cfg.MCPServers["myserver"]; !ok {
		t.Error("expected 'myserver' to be present from the YAML file")
	}
}

// TestResourceRead_DirectoryLoadsMarkdown verifies that when the resource URL
// points to a directory containing an agents/ subdirectory with .md files,
// those markdown agents are loaded and merged.
func TestResourceRead_DirectoryLoadsMarkdown(t *testing.T) {
	r := &resource{
		resourceType: "path",
		url:          "testdata/directory-yaml-and-markdown",
	}

	data, err := r.read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := unmarshalConfig(t, data)

	if _, ok := cfg.Agents["main"]; !ok {
		t.Error("expected 'main' agent to be loaded from markdown when URL is a directory")
	}
}

// TestResourceRead_DirectoryYAMLOnly verifies that when the resource URL points
// to a directory that has nanobot.yaml but no agents/ markdown files, the YAML
// config is returned as-is.
func TestResourceRead_DirectoryYAMLOnly(t *testing.T) {
	r := &resource{
		resourceType: "path",
		url:          "testdata/resolver-yaml-only",
	}

	data, err := r.read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := unmarshalConfig(t, data)

	if _, ok := cfg.Agents["myagent"]; !ok {
		t.Error("expected 'myagent' from nanobot.yaml")
	}

	if agent := cfg.Agents["myagent"]; agent.Name != "YAML Only Agent" {
		t.Errorf("expected agent name 'YAML Only Agent', got %q", agent.Name)
	}
}

// TestResourceRead_FileNotExist verifies that a non-existent file path returns
// an error rather than silently succeeding.
func TestResourceRead_FileNotExist(t *testing.T) {
	r := &resource{
		resourceType: "path",
		url:          "testdata/nonexistent/nanobot.yaml",
	}

	_, err := r.read(context.Background())
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

// TestResourceRead_FileVsDirectory_AgentCount contrasts reading the same content
// as a file versus as a directory to confirm the file path skips markdown loading.
func TestResourceRead_FileVsDirectory_AgentCount(t *testing.T) {
	fileResource := &resource{
		resourceType: "path",
		url:          "testdata/directory-yaml-and-markdown/nanobot.yaml",
	}
	dirResource := &resource{
		resourceType: "path",
		url:          "testdata/directory-yaml-and-markdown",
	}

	fileData, err := fileResource.read(context.Background())
	if err != nil {
		t.Fatalf("file read error: %v", err)
	}
	dirData, err := dirResource.read(context.Background())
	if err != nil {
		t.Fatalf("dir read error: %v", err)
	}

	fileCfg := unmarshalConfig(t, fileData)
	dirCfg := unmarshalConfig(t, dirData)

	if len(fileCfg.Agents) != 0 {
		t.Errorf("file path: expected 0 agents, got %d", len(fileCfg.Agents))
	}
	if len(dirCfg.Agents) == 0 {
		t.Error("directory path: expected agents from markdown, got none")
	}

	// Both should have the MCP server defined in nanobot.yaml
	if _, ok := fileCfg.MCPServers["myserver"]; !ok {
		t.Error("file path: expected 'myserver' MCP server")
	}
	if !strings.Contains(string(dirData), "myserver") {
		t.Error("directory path: expected 'myserver' MCP server in output")
	}
}
