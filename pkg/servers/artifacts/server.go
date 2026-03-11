package artifacts

import (
	"context"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
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
			"Search the Obot registry for published artifacts (workflows) by keyword query.",
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
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return obotConfig{}, fmt.Errorf("no session found")
	}

	envMap := session.GetEnvMap()

	baseURL := envMap["OBOT_URL"]
	if baseURL == "" {
		return obotConfig{}, fmt.Errorf("OBOT_URL is not configured — artifact tools require an Obot platform connection")
	}

	var authHeader string
	if apiKey := envMap["MCP_API_KEY"]; apiKey != "" {
		authHeader = "Bearer " + apiKey
	}

	return obotConfig{baseURL: baseURL, authHeader: authHeader}, nil
}
