package config

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools mcp.ServerTools
}

func NewServer() *Server {
	s := &Server{}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("config", "Modifies the agent config based on the file system", s.config),
	)

	return s
}

func (s *Server) config(ctx context.Context, params types.AgentConfigHook) (types.AgentConfigHook, error) {
	if params.Agent != nil {
		params.Agent.MCPServers = append(params.Agent.MCPServers, "nanobot.coder")
		if params.MCPServers == nil {
			params.MCPServers = make(map[string]types.AgentConfigHookMCPServer, 1)
		}
		params.MCPServers["nanobot.coder"] = types.AgentConfigHookMCPServer{}
	}
	return params, nil
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
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
