package system

import (
	"context"
	"os"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/obotmcp"
)

func (s *Server) augmentBashEnvForObotMCPIntegration(ctx context.Context, command string) ([]string, error) {
	env := os.Environ()

	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return env, nil
	}

	envMap := session.GetEnvMap()
	if apiKey := strings.TrimSpace(envMap["MCP_API_KEY"]); apiKey != "" {
		env = append(env, "MCP_API_KEY="+apiKey)
	}

	// The mcp-cli portion will catch some bash commands that aren't actually executing mcp-cli, but that's ok.
	// prepareMCPCLIConfig won't do the work on every invocation. It does it based on the last mod time of the config
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

	return env, nil
}
