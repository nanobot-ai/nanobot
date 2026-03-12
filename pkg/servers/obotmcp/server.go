package obotmcp

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	configDir string
	tools     mcp.ServerTools
}

func NewServer(configDir string) *Server {
	s := &Server{
		configDir: configDir,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("refreshMCPServerConfig", `Refreshes the agent-scoped mcp-cli configuration from the user's currently connected Obot MCP servers.

Use this after connecting a new MCP server in Obot when you need it to appear immediately in mcp-cli instead of waiting for the local cache to expire.

The refresh is agent-scoped and affects future mcp-cli commands for the current agent.`, s.refreshMCPServerConfig),
	)

	return s
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
	case "notifications/cancelled":
		mcp.HandleCancelled(ctx, msg)
	case "tools/list":
		mcp.Invoke(ctx, msg, s.tools.List)
	case "tools/call":
		mcp.Invoke(ctx, msg, s.tools.Call)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%v", msg.Method))
	}
}

func (s *Server) initialize(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools: &mcp.ToolsServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}

func (s *Server) refreshMCPServerConfig(ctx context.Context, _ struct{}) (map[string]any, error) {
	if _, err := PrepareMCPCLIConfig(ctx, s.configDir, true); err != nil {
		return nil, err
	}

	return map[string]any{
		"success": true,
		"message": "Refreshed mcp-cli config.",
	}, nil
}

func PrepareMCPCLIConfig(ctx context.Context, configDir string, force bool) (string, error) {
	return prepareMCPCLIConfig(ctx, configDir, force, obotConnectedServerLister{})
}

func prepareMCPCLIConfig(ctx context.Context, configDir string, force bool, lister connectedServerLister) (string, error) {
	configPath := mcpCLIConfigPath(configDir)
	if !force && !configNeedsRefresh(configPath) {
		return configPath, nil
	}

	servers, err := lister.ConnectedMCPServers(ctx)
	if err != nil {
		if _, statErr := os.Stat(configPath); statErr == nil {
			slog.Warn("failed to refresh mcp-cli config; using cached config",
				"path", configPath,
				"error", err)
			return configPath, nil
		}
		if !errors.Is(err, ErrSearchNotConfigured) {
			return "", err
		}
		servers = nil
	}

	return prepareMCPCLIConfigWithServers(configPath, servers)
}

func prepareMCPCLIConfigWithServers(configPath string, servers []ConnectedServer) (string, error) {
	if err := writeMCPCLIConfig(configPath, buildConfig(servers)); err != nil {
		return "", err
	}

	return configPath, nil
}
