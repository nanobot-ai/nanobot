package meta

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools mcp.ServerTools
	data  *sessiondata.Data
}

func NewServer(data *sessiondata.Data) *Server {
	s := &Server{
		data: data,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("list_chats", "Returns all previous chat threads", s.listChats),
		mcp.NewServerTool("update_chat", "Update fields of a give chat thread", s.updateChat),
		mcp.NewServerTool("create_chat", "Create a new chat thread", s.createChat),
		mcp.NewServerTool("delete_chat", "Delete an existing chat thread", s.deleteChat),
		mcp.NewServerTool("list_agents", "List available agents and their meta data", s.listAgents),
		//mcp.NewServerTool("set_visibility", "Make the current thread public or private", s.setVisibility),
		//mcp.NewServerTool("clone", "Clone the current session and return a new session ID", s.clone),
	)

	return s
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

func (s *Server) initialize(_ context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
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
