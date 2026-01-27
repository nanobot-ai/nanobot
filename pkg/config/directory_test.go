package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestLoadFromDirectory_Simple(t *testing.T) {
	data, err := loadFromDirectory("testdata/directory-simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}

	// Check agent loaded
	if len(config.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(config.Agents))
	}

	agent, exists := config.Agents["main"]
	if !exists {
		t.Fatal("expected 'main' agent to exist")
	}

	if agent.Name != "Main Assistant" {
		t.Errorf("expected agent name 'Main Assistant', got '%s'", agent.Name)
	}

	if agent.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", agent.Model)
	}

	if agent.Instructions.Instructions == "" {
		t.Error("expected non-empty instructions")
	}

	if !strings.Contains(agent.Instructions.Instructions, "helpful assistant") {
		t.Errorf("expected instructions to contain 'helpful assistant', got: %s", agent.Instructions.Instructions)
	}

	// Check MCP server loaded
	if len(config.MCPServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(config.MCPServers))
	}

	server, exists := config.MCPServers["myserver"]
	if !exists {
		t.Fatal("expected 'myserver' MCP server to exist")
	}

	if server.BaseURL != "https://example.com/mcp" {
		t.Errorf("expected MCP server URL 'https://example.com/mcp', got '%s'", server.BaseURL)
	}

	// Check entrypoint set to main
	if len(config.Publish.Entrypoint) != 1 || config.Publish.Entrypoint[0] != "main" {
		t.Errorf("expected entrypoint ['main'], got %v", config.Publish.Entrypoint)
	}
}

func TestLoadFromDirectory_MultipleAgents(t *testing.T) {
	data, err := loadFromDirectory("testdata/directory-multiple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}

	// Check both agents loaded
	if len(config.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(config.Agents))
	}

	mainAgent, exists := config.Agents["main"]
	if !exists {
		t.Fatal("expected 'main' agent to exist")
	}

	if mainAgent.Name != "Main Agent" {
		t.Errorf("expected main agent name 'Main Agent', got '%s'", mainAgent.Name)
	}

	helperAgent, exists := config.Agents["helper"]
	if !exists {
		t.Fatal("expected 'helper-agent' agent to exist (from id field)")
	}

	if helperAgent.Name != "Helper Agent" {
		t.Errorf("expected helper agent name 'Helper Agent', got '%s'", helperAgent.Name)
	}

	// Check entrypoint set to first agent lexicographically (helper-agent)
	if len(config.Publish.Entrypoint) != 1 || config.Publish.Entrypoint[0] != "helper" {
		t.Errorf("expected entrypoint ['helper-agent'], got %v", config.Publish.Entrypoint)
	}
}

func TestLoadFromDirectory_JSON(t *testing.T) {
	data, err := loadFromDirectory("testdata/directory-json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}

	// Check MCP server loaded from JSON
	server, exists := config.MCPServers["jsonserver"]
	if !exists {
		t.Fatal("expected 'jsonserver' MCP server to exist")
	}

	if server.BaseURL != "https://example.com/json" {
		t.Errorf("expected MCP server URL 'https://example.com/json', got '%s'", server.BaseURL)
	}
}

func TestLoadFromDirectory_BothMCPFiles_Error(t *testing.T) {
	_, err := loadFromDirectory("testdata/directory-both-mcp")
	if err == nil {
		t.Fatal("expected error when both mcpServers.yaml and mcpServers.json exist")
	}

	if !strings.Contains(err.Error(), "both mcpServers.yaml and mcpServers.json found") {
		t.Errorf("expected error about both files, got: %v", err)
	}
}

func TestLoadFromDirectory_NoMCPServers(t *testing.T) {
	data, err := loadFromDirectory("testdata/directory-no-mcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}

	if len(config.MCPServers) != 0 {
		t.Errorf("expected 0 MCP servers, got %d", len(config.MCPServers))
	}
}

func TestLoadFromDirectory_WithREADME(t *testing.T) {
	data, err := loadFromDirectory("testdata/directory-with-readme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}

	// Should only have 1 agent (main.md), README.md should be ignored
	if len(config.Agents) != 1 {
		t.Errorf("expected 1 agent (README.md should be ignored), got %d", len(config.Agents))
	}

	_, hasMain := config.Agents["main"]
	if !hasMain {
		t.Error("expected 'main' agent to exist")
	}

	_, hasREADME := config.Agents["README"]
	if hasREADME {
		t.Error("README.md should not be loaded as an agent")
	}
}

func TestLoadFromDirectory_HiddenFiles(t *testing.T) {
	data, err := loadFromDirectory("testdata/directory-hidden")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}

	// Should only have 1 agent (main.md), not .hidden.md
	if len(config.Agents) != 1 {
		t.Errorf("expected 1 agent (hidden files should be ignored), got %d", len(config.Agents))
	}

	_, hasMain := config.Agents["main"]
	if !hasMain {
		t.Error("expected 'main' agent to exist")
	}

	_, hasHidden := config.Agents[".hidden"]
	if hasHidden {
		t.Error("hidden agent should not be loaded")
	}
}

func TestLoadFromDirectory_InvalidYAML_Error(t *testing.T) {
	_, err := loadFromDirectory("testdata/directory-invalid-yaml")
	if err == nil {
		t.Fatal("expected error for invalid YAML front-matter")
	}

	if !strings.Contains(err.Error(), "invalid YAML") && !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected error about invalid YAML, got: %v", err)
	}
}

func TestLoadFromDirectory_MissingServer_Error(t *testing.T) {
	_, err := loadFromDirectory("testdata/directory-missing-server")
	if err == nil {
		t.Fatal("expected error when agent references non-existent MCP server")
	}

	if !strings.Contains(err.Error(), "references MCP server") || !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("expected error about missing MCP server, got: %v", err)
	}
}

func TestHasMarkdownFiles(t *testing.T) {
	// Directory with .md files
	hasMd, err := hasMarkdownFiles("testdata/directory-simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hasMd {
		t.Error("expected directory-simple to have .md files")
	}

	// Directory that doesn't exist
	_, err = hasMarkdownFiles("testdata/nonexistent")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestParseMarkdownAgent_IDFromFilename(t *testing.T) {
	agentID, agent, err := parseMarkdownAgent("testdata/directory-simple/main.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use filename (main.md -> main)
	if agentID != "main" {
		t.Errorf("expected agent ID 'main', got '%s'", agentID)
	}

	if agent.Name != "Main Assistant" {
		t.Errorf("expected name 'Main Assistant', got '%s'", agent.Name)
	}
}

func TestParseMarkdownAgent_AllFields(t *testing.T) {
	agentID, agent, err := parseMarkdownAgent("testdata/directory-simple/main.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agentID != "main" {
		t.Errorf("expected agent ID 'main', got '%s'", agentID)
	}

	// Check various fields from front-matter
	if agent.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", agent.Model)
	}

	if len(agent.MCPServers) != 1 || agent.MCPServers[0] != "myserver" {
		t.Errorf("expected mcpServers ['myserver'], got %v", agent.MCPServers)
	}

	if len(agent.Tools) != 1 || agent.Tools[0] != "myserver" {
		t.Errorf("expected tools ['myserver'], got %v", agent.Tools)
	}

	if agent.Temperature == nil || agent.Temperature.String() != "0.7" {
		t.Errorf("expected temperature 0.7, got %v", agent.Temperature)
	}

	if agent.MaxTokens != 1000 {
		t.Errorf("expected maxTokens 1000, got %d", agent.MaxTokens)
	}

	// Check instructions from body
	if !strings.Contains(agent.Instructions.Instructions, "helpful assistant") {
		t.Errorf("expected instructions to contain 'helpful assistant', got: %s", agent.Instructions.Instructions)
	}
}

func TestLoadFromDirectory_DefaultAgent_Single(t *testing.T) {
	data, err := loadFromDirectory("testdata/directory-default-single")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}

	// Check agent loaded
	if len(config.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(config.Agents))
	}

	agent, exists := config.Agents["agent"]
	if !exists {
		t.Fatal("expected 'agent' agent to exist")
	}

	if agent.Name != "Single Default Agent" {
		t.Errorf("expected agent name 'Single Default Agent', got '%s'", agent.Name)
	}

	// Check entrypoint set to the default agent
	if len(config.Publish.Entrypoint) != 1 || config.Publish.Entrypoint[0] != "agent" {
		t.Errorf("expected entrypoint ['agent'], got %v", config.Publish.Entrypoint)
	}
}

func TestLoadFromDirectory_DefaultAgent_Explicit(t *testing.T) {
	data, err := loadFromDirectory("testdata/directory-default-explicit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}

	// Check both agents loaded
	if len(config.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(config.Agents))
	}

	alphaAgent, exists := config.Agents["alpha"]
	if !exists {
		t.Fatal("expected 'alpha' agent to exist")
	}

	if alphaAgent.Name != "Alpha Agent" {
		t.Errorf("expected alpha agent name 'Alpha Agent', got '%s'", alphaAgent.Name)
	}

	zuluAgent, exists := config.Agents["zulu"]
	if !exists {
		t.Fatal("expected 'zulu' agent to exist")
	}

	if zuluAgent.Name != "Zulu Agent" {
		t.Errorf("expected zulu agent name 'Zulu Agent', got '%s'", zuluAgent.Name)
	}

	// Check entrypoint set to explicitly marked default agent (zulu), not lexicographically first (alpha)
	if len(config.Publish.Entrypoint) != 1 || config.Publish.Entrypoint[0] != "zulu" {
		t.Errorf("expected entrypoint ['zulu'] (explicitly marked default), got %v", config.Publish.Entrypoint)
	}
}

func TestLoadFromDirectory_DefaultAgent_MultipleDefaults_Error(t *testing.T) {
	_, err := loadFromDirectory("testdata/directory-default-multiple-error")
	if err == nil {
		t.Fatal("expected error when multiple agents are marked as default")
	}

	if !strings.Contains(err.Error(), "multiple agents marked as default") {
		t.Errorf("expected error about multiple default agents, got: %v", err)
	}

	// Error should mention both agent IDs
	if !strings.Contains(err.Error(), "first") || !strings.Contains(err.Error(), "second") {
		t.Errorf("expected error to mention both 'first' and 'second' agents, got: %v", err)
	}
}

func TestParseMarkdownAgent_DefaultField(t *testing.T) {
	agentID, agent, err := parseMarkdownAgent("testdata/directory-default-single/agent.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agentID != "agent" {
		t.Errorf("expected agent ID 'agent', got '%s'", agentID)
	}

	if agent.Default != true {
		t.Errorf("expected Default field to be true, got %v", agent.Default)
	}

	if agent.Agent.Name != "Single Default Agent" {
		t.Errorf("expected name 'Single Default Agent', got '%s'", agent.Agent.Name)
	}
}

func TestParseMarkdownAgent_NoDefaultField(t *testing.T) {
	agentID, agent, err := parseMarkdownAgent("testdata/directory-simple/main.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agentID != "main" {
		t.Errorf("expected agent ID 'main', got '%s'", agentID)
	}

	// Default should be false when not specified
	if agent.Default != false {
		t.Errorf("expected Default field to be false when not specified, got %v", agent.Default)
	}
}
