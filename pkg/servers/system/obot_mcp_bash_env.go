package system

import (
	"context"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/obotmcp"
)

const mcpCLIClientName = "MCP_CLIENT_NAME=Obot Agent MCP cli"

// obotMCPBashEnvVars returns Obot MCP-related environment variables for bash commands.
// When the command appears to invoke mcp-cli, it may also prepare or refresh the
// agent-scoped mcp-cli config so MCP_CONFIG_PATH can be set correctly.
func (s *Server) obotMCPBashEnvVars(ctx context.Context, command string) ([]string, error) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil, nil
	}

	var env []string
	envMap := session.GetEnvMap()
	if apiKey := strings.TrimSpace(envMap["MCP_API_KEY"]); apiKey != "" {
		env = append(env, "MCP_API_KEY="+apiKey)
	}

	// The mcp-cli portion will catch some bash commands that aren't actually executing mcp-cli, but that's ok.
	// The connected server list is cached globally, so the actual fetch only happens every 15 minutes.
	if !strings.Contains(command, "mcp-cli") || strings.TrimSpace(envMap["MCP_SERVER_SEARCH_URL"]) == "" ||
		strings.TrimSpace(envMap["MCP_API_KEY"]) == "" {
		return env, nil
	}

	configPath, err := obotmcp.PrepareMCPCLIConfig(ctx, s.configDir, false)
	if err != nil {
		return nil, err
	}
	if configPath != "" {
		env = append(env, "MCP_CONFIG_PATH="+configPath)
	}
	env = append(env, mcpCLIClientName)

	return env, nil
}
