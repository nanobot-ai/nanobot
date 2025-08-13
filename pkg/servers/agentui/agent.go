package agentui

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools   mcp.ServerTools
	data    *sessiondata.Data
	runtime Caller
}

type Caller interface {
	Call(ctx context.Context, server, tool string, args any, opts ...tools.CallOptions) (ret *types.CallResult, err error)
	GetClient(ctx context.Context, name string) (*mcp.Client, error)
}

func NewServer(d *sessiondata.Data, r Caller) *Server {
	s := &Server{
		data:    d,
		runtime: r,
	}

	s.tools = mcp.NewServerTools(
		getChatCall{s: s},
		setCurrentAgentCall{s: s},
		chatCall{s: s},
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
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage(msg.Method))
	}
}
func (s *Server) describeSession(ctx context.Context, args any) <-chan struct{} {
	result := make(chan struct{})
	var description string

	session := mcp.SessionFromContext(ctx)
	session = session.Parent
	session.Get(types.DescriptionSessionKey, &description)
	if description == "" {
		go func() {
			defer close(result)
			ret, err := s.runtime.Call(ctx, "nanobot.summary", "nanobot.summary", args)
			if err != nil {
				return
			}
			for _, content := range ret.Content {
				if content.Type == "text" {
					description = content.Text
					session.Set(types.DescriptionSessionKey, description)
					break
				}
			}
		}()
	} else {
		close(result)
	}

	return result
}

func (s *Server) initialize(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	//target, err := s.data.GetCurrentAgentTargetMapping(ctx)
	//if err != nil {
	//	return nil, err
	//}
	//
	//remoteClient, err := s.runtime.GetClient(ctx, target.MCPServer)
	//if err != nil {
	//	return nil, err
	//}
	//
	//tools, err := remoteClient.ListTools(ctx)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to list tools: %w", err)
	//}
	//
	//found := false
	//for _, tool := range tools.Tools {
	//	if tool.Name == "set_current_agent" {
	//		found = true
	//		break
	//	}
	//}
	//
	//if !found {
	//	delete(s.tools, "set_current_agent")
	//}
	//
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
