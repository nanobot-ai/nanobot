package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"sigs.k8s.io/yaml"
)

// hasMarkdownFiles checks if a directory contains any .md files (non-hidden)
func hasMarkdownFiles(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
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
	Default bool
	types.Agent
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

// loadAgentsFromMarkdown scans a directory for .md files and parses them as agent definitions
func loadAgentsFromMarkdown(dirPath string) (map[string]types.Agent, string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, "", fmt.Errorf("error reading directory %s: %w", dirPath, err)
	}

	var explicitDefaultAgent string
	agents := make(map[string]types.Agent)
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

		filePath := filepath.Join(dirPath, name)
		agentID, agent, err := parseMarkdownAgent(filePath)
		if err != nil {
			return nil, "", err
		}

		if agent.Default {
			if explicitDefaultAgent != "" {
				return nil, "", fmt.Errorf("multiple agents marked as default: '%s' and '%s'", explicitDefaultAgent, agentID)
			}
			explicitDefaultAgent = agentID
		}

		agents[agentID] = agent.Agent
	}

	if len(agents) == 0 {
		return nil, "", fmt.Errorf("no agent .md files found in directory: %s", dirPath)
	}

	return agents, explicitDefaultAgent, nil
}

// loadMCPServers loads MCP server definitions from mcpServers.yaml or mcpServers.json
func loadMCPServers(dirPath string) (map[string]mcp.Server, error) {
	yamlPath := filepath.Join(dirPath, "mcpServers.yaml")
	jsonPath := filepath.Join(dirPath, "mcpServers.json")

	_, yamlErr := os.Stat(yamlPath)
	_, jsonErr := os.Stat(jsonPath)

	yamlExists := yamlErr == nil
	jsonExists := jsonErr == nil

	// Error if both exist
	if yamlExists && jsonExists {
		return nil, fmt.Errorf("both mcpServers.yaml and mcpServers.json found in %s, only one is allowed", dirPath)
	}

	// If neither exists, return empty map (valid case)
	if !yamlExists && !jsonExists {
		return make(map[string]mcp.Server), nil
	}

	var servers map[string]mcp.Server

	if yamlExists {
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			return nil, fmt.Errorf("error reading mcpServers.yaml: %w", err)
		}
		if err := yaml.Unmarshal(data, &servers); err != nil {
			return nil, fmt.Errorf("error parsing mcpServers.yaml: %w", err)
		}
	} else if jsonExists {
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("error reading mcpServers.json: %w", err)
		}
		if err := json.Unmarshal(data, &servers); err != nil {
			return nil, fmt.Errorf("error parsing mcpServers.json: %w", err)
		}
	}

	if servers == nil {
		servers = make(map[string]mcp.Server)
	}

	return servers, nil
}

func loadFromDirectory(dirPath string) ([]byte, error) {
	// Load agents from .md files
	agents, defaultAgent, err := loadAgentsFromMarkdown(dirPath)
	if err != nil {
		return nil, err
	}

	// Load MCP servers
	mcpServers, err := loadMCPServers(dirPath)
	if err != nil {
		return nil, err
	}

	// Validate that all referenced MCP servers exist
	for agentID, agent := range agents {
		for _, serverRef := range agent.MCPServers {
			if _, exists := mcpServers[serverRef]; !exists {
				return nil, fmt.Errorf("agent '%s' references MCP server '%s' which is not defined", agentID, serverRef)
			}
		}
	}

	// Build config
	config := types.Config{
		Agents:     agents,
		MCPServers: mcpServers,
	}

	// Auto-set entrypoint to first agent lexicographically if no entrypoint is set
	if defaultAgent == "" && len(agents) > 0 {
		for agentID := range agents {
			if defaultAgent == "" || agentID < defaultAgent {
				defaultAgent = agentID
			}
		}
	}
	config.Publish.Entrypoint = types.StringList{defaultAgent}

	return json.Marshal(config)
}
