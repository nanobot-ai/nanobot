package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/types"
	"sigs.k8s.io/yaml"
)

// hasMarkdownFiles checks if the agents/ subdirectory contains any .md files (non-hidden)
func hasMarkdownFiles(dirPath string) (bool, error) {
	entries, err := os.ReadDir(filepath.Join(dirPath, "agents"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip hidden files and README files
		if strings.HasPrefix(name, ".") || strings.EqualFold(name, "README.md") {
			continue
		}
		if strings.HasSuffix(name, ".md") {
			return true, nil
		}
	}
	return false, nil
}

type frontMatterAgent struct {
	Default     bool
	Mode        string
	types.Agent `json:",inline" yaml:",inline"`
}

// parseMarkdownAgent parses a single .md file into an Agent
// Returns: (agentID string, agent types.Agent, error)
func parseMarkdownAgent(filePath string) (string, frontMatterAgent, error) {
	// Read file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", frontMatterAgent{}, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	// Parse front-matter
	yamlContent, body, err := parseFrontMatter(fileContent)
	if err != nil {
		return "", frontMatterAgent{}, fmt.Errorf("error parsing front-matter in %s: %w", filePath, err)
	}

	var parsed frontMatterAgent
	if len(yamlContent) > 0 {
		if err := yaml.Unmarshal(yamlContent, &parsed); err != nil {
			return "", frontMatterAgent{}, fmt.Errorf("invalid YAML front-matter in %s: %w", filePath, err)
		}
	}

	// Use filename without .md extension
	basename := filepath.Base(filePath)
	agentID := strings.TrimSuffix(basename, ".md")

	if agentID == "" {
		return "", frontMatterAgent{}, fmt.Errorf("agent in file %s has empty ID", filePath)
	}

	// Set instructions to the markdown body
	parsed.Instructions = types.DynamicInstructions{
		Instructions: body,
	}

	return agentID, parsed, nil
}

// loadAgentsFromMarkdown scans the agents/ subdirectory for .md files and parses them as agent definitions
func loadAgentsFromMarkdown(config *types.Config, dirPath string) error {
	agentsDir := filepath.Join(dirPath, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error reading directory %s: %w", agentsDir, err)
	}

	var explicitDefaultAgent string
	subagents := make(map[string]struct{}) // Track which agents are subagents
	config.Agents = make(map[string]types.Agent)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden files and README files
		if strings.HasPrefix(name, ".") || strings.EqualFold(name, "README.md") {
			continue
		}

		// Only process .md files
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		filePath := filepath.Join(agentsDir, name)
		agentID, agent, err := parseMarkdownAgent(filePath)
		if err != nil {
			return err
		}

		if agent.Default {
			if explicitDefaultAgent != "" {
				return fmt.Errorf("multiple agents marked as default: '%s' and '%s'", explicitDefaultAgent, agentID)
			}
			explicitDefaultAgent = agentID
		}

		switch agent.Mode {
		case "", "chat", "primary", "all":
			config.Publish.Entrypoint = append(config.Publish.Entrypoint, agentID)
		case "subagent":
			subagents[agentID] = struct{}{}
			if agent.Default {
				return fmt.Errorf("agent '%s' in file %s cannot be both 'subagent' and 'default'", agentID, filePath)
			}
		default:
			return fmt.Errorf("invalid mode '%s' for agent '%s' in file %s", agent.Mode, agentID, filePath)
		}

		config.Agents[agentID] = agent.Agent
	}

	if len(config.Agents) == 0 {
		return fmt.Errorf("no agent .md files found in directory: %s", agentsDir)
	}

	// Auto-set default agent to first non-subagent lexicographically if no explicit default
	if explicitDefaultAgent == "" {
		for agentID := range config.Agents {
			if _, ok := subagents[agentID]; ok {
				continue
			}
			if explicitDefaultAgent == "" {
				explicitDefaultAgent = agentID
				continue
			}
			if agentID < explicitDefaultAgent {
				explicitDefaultAgent = agentID
			}
		}
	}

	if explicitDefaultAgent == "" {
		return fmt.Errorf("no valid default agent could be determined from directory: %s", agentsDir)
	}

	// Ensure the explicitDefaultAgent is the first in the entrypoint list
	idx := slices.Index(config.Publish.Entrypoint, explicitDefaultAgent)
	if idx > 0 {
		config.Publish.Entrypoint = append([]string{explicitDefaultAgent}, append(config.Publish.Entrypoint[:idx], config.Publish.Entrypoint[idx+1:]...)...)
	} else if idx == -1 {
		config.Publish.Entrypoint = append([]string{explicitDefaultAgent}, config.Publish.Entrypoint...)
	}

	return nil
}

// loadMCPServers loads MCP server definitions from mcp-servers.yaml or mcp-servers.json.
// Deprecated: define mcpServers in nanobot.yaml instead.
func loadMCPServers(config *types.Config, dirPath string) error {
	yamlPath := filepath.Join(dirPath, "mcp-servers.yaml")
	jsonPath := filepath.Join(dirPath, "mcp-servers.json")

	_, yamlErr := os.Stat(yamlPath)
	_, jsonErr := os.Stat(jsonPath)

	yamlExists := yamlErr == nil
	jsonExists := jsonErr == nil

	// Error if both exist
	if yamlExists && jsonExists {
		return fmt.Errorf("both mcp-servers.yaml and mcp-servers.json found in %s, only one is allowed", dirPath)
	}

	// If neither exists, return empty map (valid case)
	if !yamlExists && !jsonExists {
		return nil
	}

	// Only apply deprecated file if nanobot.yaml did not already define mcpServers
	if len(config.MCPServers) > 0 {
		slog.Warn("mcp-servers.yaml/mcp-servers.json is deprecated and ignored when mcpServers is defined in nanobot.yaml", "dir", dirPath)
		return nil
	}

	filePath := yamlPath
	if jsonExists {
		filePath = jsonPath
	}

	slog.Warn("mcp-servers.yaml/mcp-servers.json is deprecated; define mcpServers in nanobot.yaml instead", "path", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filePath, err)
	}
	if err := yaml.Unmarshal(data, &config.MCPServers); err != nil {
		return fmt.Errorf("error parsing %s: %w", filePath, err)
	}

	// Validate that all referenced MCP servers exist
	for agentID, agent := range config.Agents {
		for _, serverRef := range agent.MCPServers {
			// Remove a tool reference if it exists (e.g., "server-name/tool-name" -> "server-name")
			serverRef, _, _ = strings.Cut(serverRef, "/")
			if _, exists := config.MCPServers[serverRef]; !exists {
				return fmt.Errorf("agent '%s' references MCP server '%s' which is not defined", agentID, serverRef)
			}
		}
	}

	return nil
}

// loadFromDirectory builds a config from a directory of markdown agent files, optionally
// merged with a base nanobot.yaml. Markdown agents take precedence over YAML-defined agents.
func loadFromDirectory(dirPath string, baseYAML []byte) ([]byte, error) {
	var config types.Config

	// Parse nanobot.yaml as base config if provided
	if len(baseYAML) > 0 {
		if err := yaml.Unmarshal(baseYAML, &config); err != nil {
			return nil, fmt.Errorf("error parsing nanobot.yaml in %s: %w", dirPath, err)
		}
	}

	// Save YAML-defined agents so we can merge them back after markdown takes precedence
	yamlAgents := config.Agents

	// Reset fields that markdown will rebuild
	config.Agents = nil

	// Load agents from .md files (these take precedence over YAML agents)
	if err := loadAgentsFromMarkdown(&config, dirPath); err != nil {
		return nil, err
	}

	// Merge back YAML-only agents that markdown did not define
	for id, agent := range yamlAgents {
		if _, exists := config.Agents[id]; !exists {
			if config.Agents == nil {
				config.Agents = make(map[string]types.Agent)
			}
			config.Agents[id] = agent
		}
	}

	// Load MCP servers from deprecated mcp-servers.yaml/json if present
	if err := loadMCPServers(&config, dirPath); err != nil {
		return nil, err
	}

	return json.Marshal(config)
}
