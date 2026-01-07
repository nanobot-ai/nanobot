package tasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools        mcp.ServerTools
	toolsService *tools.Service
}

type RunTaskParams struct {
	URI    string         `json:"uri"`
	Params map[string]any `json:"params,omitempty"`
}

func NewServer(toolsService *tools.Service) *Server {
	s := &Server{
		toolsService: toolsService,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("run_task", "Run a task with resource URI and parameters", s.run),
	)

	return s
}

func (s *Server) isInWorkspace(ctx context.Context) bool {
	return types.GetWorkspaceID(ctx) != ""
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
	case "tools/list":
		if !s.isInWorkspace(ctx) {
			mcp.Invoke(ctx, msg, mcp.ServerTools{}.List)
		} else {
			mcp.Invoke(ctx, msg, s.tools.List)
		}
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

func (s *Server) run(ctx context.Context, params RunTaskParams) (string, error) {
	// Validate URI is not empty
	if params.URI == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("uri is required")
	}

	// Call CreateTask on nanobot.tasks.runner
	result, err := s.toolsService.Call(ctx, "nanobot.tasks.runner", "StartTask", map[string]any{
		"taskName":  params.URI,
		"sessionId": types.GetWorkspaceID(ctx),
		"arguments": params.Params,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	var text string
	for _, c := range result.Content {
		if c.Type == "text" {
			text += c.Text + "\n\n"
		}
	}
	text = strings.TrimSpace(text)

	if result.IsError {
		return "", fmt.Errorf("failed to create task: %s", text)
	}

	// Return the full result from CreateTask
	return text, nil
}
