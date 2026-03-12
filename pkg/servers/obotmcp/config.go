package obotmcp

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const configuredMCPServersPromptSessionKey = "configuredMCPServersPrompt"

// ConfigureIntegration injects the Obot discovery MCP server and appends a snapshot of the user's configured
// Obot MCP servers to the system prompt.
func ConfigureIntegration(ctx context.Context, configDir string, agent *types.HookAgent, params *types.AgentConfigHook) {
	configureIntegration(ctx, configDir, agent, params, obotConnectedServerLister{})
}

func configureIntegration(ctx context.Context, configDir string, agent *types.HookAgent, params *types.AgentConfigHook, lister connectedServerLister) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return
	}

	envMap := session.GetEnvMap()
	searchURL := envMap["MCP_SERVER_SEARCH_URL"]
	if searchURL == "" {
		return
	}

	params.MCPServers["nanobot.obot-mcp-cli"] = types.AgentConfigHookMCPServer{}
	agent.Tools = append(agent.Tools, "nanobot.obot-mcp-cli/refreshMCPServerConfig")

	mcpServer := types.AgentConfigHookMCPServer{URL: searchURL}
	if apiKey := envMap["MCP_SERVER_SEARCH_API_KEY"]; apiKey != "" {
		mcpServer.Headers = map[string]string{
			"Authorization": "Bearer " + apiKey,
		}
	}

	params.MCPServers["mcp-server-search"] = mcpServer
	agent.Tools = append(agent.Tools, "mcp-server-search")

	configPath := mcpCLIConfigPath(configDir)
	needsConfigRefresh := configNeedsRefresh(configPath)

	root := session.Root()
	var configuredServersPrompt string
	needsPromptGenerated := !root.Get(configuredMCPServersPromptSessionKey, &configuredServersPrompt)

	if !needsConfigRefresh && !needsPromptGenerated {
		agent.Instructions.Instructions += configuredServersPrompt
		return
	}

	var (
		servers []ConnectedServer
		err     error
	)
	if needsConfigRefresh || needsPromptGenerated {
		servers, err = lister.ConnectedMCPServers(ctx)
		if errors.Is(err, ErrSearchNotConfigured) {
			if needsPromptGenerated {
				// Obot mcp search server not configured, so we don't need to generate a prompt
				root.Set(configuredMCPServersPromptSessionKey, mcp.SavedString(""))
			} else {
				agent.Instructions.Instructions += configuredServersPrompt
			}
			return
		}
	}

	if needsConfigRefresh {
		if err != nil {
			slog.Warn("skipping mcp-cli config refresh during Obot integration setup because connected server fetch failed",
				"path", configPath,
				"error", err)
		} else {
			if _, writeErr := prepareMCPCLIConfigWithServers(configPath, servers); writeErr != nil {
				slog.Warn("failed to prepare mcp-cli config during Obot integration setup", "error", writeErr)
			}
		}
	}

	if needsPromptGenerated {
		if err != nil {
			slog.Warn("failed to build configured MCP servers prompt snapshot", "error", err)
		} else {
			configuredServersPrompt = buildConfiguredMCPServersPrompt(servers)
			root.Set(configuredMCPServersPromptSessionKey, mcp.SavedString(configuredServersPrompt))
		}
	}
	agent.Instructions.Instructions += configuredServersPrompt
}

func buildConfiguredMCPServersPrompt(servers []ConnectedServer) string {
	var prompt strings.Builder
	prompt.WriteString("\n\n## Configured MCP Servers\n\n")
	prompt.WriteString("This is a snapshot of the user's configured MCP servers from when this session first built its system prompt. ")
	prompt.WriteString("It can change later if the user connects or configures new MCP servers. If that happens, call `refreshMCPServerConfig`.\n\n")

	entries := buildInventoryEntries(servers)
	if len(entries) == 0 {
		prompt.WriteString("- No MCP servers were configured at snapshot time.\n")
		return prompt.String()
	}

	for _, entry := range entries {
		prompt.WriteString("- `")
		prompt.WriteString(entry.Name)
		prompt.WriteString("`: ")
		if entry.Server.Name != "" {
			prompt.WriteString(entry.Server.Name)
		} else {
			prompt.WriteString(entry.Server.ID)
		}
		if entry.Server.Description != "" {
			prompt.WriteString(" - ")
			prompt.WriteString(entry.Server.Description)
		}
		prompt.WriteString("\n")
	}

	return prompt.String()
}
