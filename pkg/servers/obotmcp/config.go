package obotmcp

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"unicode"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const configuredMCPServersPromptSessionKey = "configuredMCPServersPrompt"
const configuredMCPServersPromptFieldMaxRunes = 200

// ConfigureIntegration wires Obot MCP integration into the agent by adding the
// Obot discover server and the mcp-cli refresh tool to the agent's tools
// and a cached snapshot of the user's configured Obot MCP servers to the system prompt.
func ConfigureIntegration(ctx context.Context, agent *types.HookAgent, params *types.AgentConfigHook) {
	configureIntegration(ctx, agent, params, obotConnectedServerLister{})
}

func configureIntegration(ctx context.Context, agent *types.HookAgent, params *types.AgentConfigHook, lister connectedServerLister) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return
	}

	envMap := session.GetEnvMap()
	searchURL := strings.TrimSpace(envMap["MCP_SERVER_SEARCH_URL"])
	if searchURL == "" {
		return
	}

	mcpServer := types.AgentConfigHookMCPServer{URL: searchURL}
	if apiKey := envMap["MCP_SERVER_SEARCH_API_KEY"]; apiKey != "" {
		mcpServer.Headers = map[string]string{
			"Authorization": "Bearer " + apiKey,
		}
	}
	params.MCPServers["mcp-server-search"] = mcpServer
	agent.Tools = append(agent.Tools, "mcp-server-search")

	params.MCPServers["nanobot.obot-mcp-cli"] = types.AgentConfigHookMCPServer{}
	agent.Tools = append(agent.Tools, "nanobot.obot-mcp-cli/refreshMCPServerConfig")

	root := session.Root()
	var configuredServersPrompt string
	if root.Get(configuredMCPServersPromptSessionKey, &configuredServersPrompt) {
		agent.Instructions.Instructions += configuredServersPrompt
		return
	}

	servers, err := lister.ConnectedMCPServers(ctx)
	if err != nil {
		if errors.Is(err, ErrSearchNotConfigured) {
			root.Set(configuredMCPServersPromptSessionKey, mcp.SavedString(""))
		} else {
			slog.Warn("failed to fetch connected MCP servers during Obot integration setup", "error", err)
		}
		return
	}

	configuredServersPrompt = buildConfiguredMCPServersPrompt(servers)
	root.Set(configuredMCPServersPromptSessionKey, mcp.SavedString(configuredServersPrompt))
	agent.Instructions.Instructions += configuredServersPrompt
}

func buildConfiguredMCPServersPrompt(servers []ConnectedServer) string {
	var prompt strings.Builder
	prompt.WriteString("\n\n## Configured MCP Servers\n\n")
	prompt.WriteString("This is a snapshot of the user's configured MCP servers from when this session first built its system prompt. ")
	prompt.WriteString("The user's configured servers can change later based on actions taken in this thread or out of band. Use `mcp-cli` for an up-to-date list of the user's connected MCP servers and their tools.\n\n")

	entries := buildInventoryEntries(servers)
	if len(entries) == 0 {
		prompt.WriteString("- No MCP servers were configured at snapshot time.\n")
		return prompt.String()
	}

	for _, entry := range entries {
		prompt.WriteString("- `")
		prompt.WriteString(entry.Name)
		prompt.WriteString("`: ")
		displayName := sanitizeConfiguredMCPServerPromptField(entry.Server.Name)
		if displayName == "" {
			displayName = sanitizeConfiguredMCPServerPromptField(entry.Server.ID)
		}
		if displayName != "" {
			prompt.WriteString(displayName)
		} else {
			prompt.WriteString("Unnamed server")
		}
		if description := sanitizeConfiguredMCPServerPromptField(entry.Server.Description); description != "" {
			prompt.WriteString(" - ")
			prompt.WriteString(description)
		}
		prompt.WriteString("\n")
	}

	return prompt.String()
}

func sanitizeConfiguredMCPServerPromptField(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		switch {
		case r == '`':
			builder.WriteRune('\'')
		case unicode.IsControl(r):
			builder.WriteRune(' ')
		default:
			builder.WriteRune(r)
		}
	}

	value = strings.Join(strings.Fields(builder.String()), " ")
	runes := []rune(value)
	if len(runes) > configuredMCPServersPromptFieldMaxRunes {
		value = strings.TrimSpace(string(runes[:configuredMCPServersPromptFieldMaxRunes-3])) + "..."
	}

	return value
}
