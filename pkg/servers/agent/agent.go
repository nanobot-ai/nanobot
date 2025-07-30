package agent

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools      mcp.ServerTools
	data       *sessiondata.Data
	multiAgent bool
	runtime    Caller
}

type Caller interface {
	Call(ctx context.Context, server, tool string, args any, opts ...tools.CallOptions) (ret *types.CallResult, err error)
}

func NewServer(d *sessiondata.Data, r Caller) *Server {
	s := &Server{
		data:    d,
		runtime: r,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("get_chat", "Returns the contents of the current thread", s.getChat),
		mcp.NewServerTool("set_current_agent", "Set the current agent the user is chatting with", s.setAgent),
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

type chatCall struct {
	s *Server
}

func (c chatCall) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        types.AgentTool,
		Description: types.AgentToolDescription,
		InputSchema: types.ChatInputSchema,
	}
}

func (c chatCall) Invoke(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	currentAgent := c.s.data.CurrentAgent(ctx)

	description := c.s.describeSession(ctx, payload.Arguments)

	result, err := c.s.runtime.Call(ctx, currentAgent, types.AgentTool, payload.Arguments, tools.CallOptions{
		ProgressToken: msg.ProgressToken(),
		LogData: map[string]any{
			"mcpToolName": payload.Name,
		},
	})
	if err != nil {
		return nil, err
	}

	if payload.Name == types.AgentTool && result.ChatResponse && result.Agent != "" {
		c.s.data.SetCurrentAgent(ctx, result.Agent)
	}

	mcpResult := mcp.CallToolResult{
		IsError: result.IsError,
		Content: result.Content,
	}

	if description != nil {
		<-description
	}

	err = msg.Reply(ctx, mcpResult)
	return &mcpResult, err
}

func (s *Server) describeSession(ctx context.Context, args any) <-chan struct{} {
	result := make(chan struct{})
	var description string

	session := mcp.SessionFromContext(ctx)
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
	agents, err := s.data.Agents(ctx)
	if err != nil {
		return nil, err
	}
	if len(agents) <= 1 {
		delete(s.tools, "set_current_agent")
	} else {
		s.multiAgent = true
	}
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
