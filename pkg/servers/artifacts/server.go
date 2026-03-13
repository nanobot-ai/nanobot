package artifacts

import (
	"context"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	obotconfig "github.com/nanobot-ai/nanobot/pkg/servers/obot"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools mcp.ServerTools
}

func NewServer() *Server {
	s := &Server{}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("publishArtifact",
			"Publish a local workflow as a shareable artifact to the Obot registry. "+
				"Reads the workflow directory, validates the SKILL.md, creates a ZIP, and uploads it.",
			s.publishArtifact),
		mcp.NewServerTool("searchArtifacts",
			"Search the Obot registry for published artifacts (workflows) by keyword query. "+
				"This searches the REMOTE registry only — it does NOT find locally installed workflows. "+
				"To find installed workflows, read the local `workflows/` directory instead.",
			s.searchArtifacts),
		mcp.NewServerTool("installArtifact",
			"Download and install a published artifact from the Obot registry into the local workspace.",
			s.installArtifact),
	)

	return s
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
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

type obotConfig struct {
	baseURL    string
	authHeader string
}

func getObotConfig(ctx context.Context) (obotConfig, error) {
	cfg, err := obotconfig.GetConfig(ctx)
	if err != nil {
		return obotConfig{}, fmt.Errorf("%w — artifact tools require an Obot platform connection", err)
	}
	return obotConfig{baseURL: cfg.BaseURL, authHeader: cfg.AuthHeader}, nil
}
