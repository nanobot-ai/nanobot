package workflows

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type ToolsServer struct {
	tools mcp.ServerTools
}

func NewToolsServer() *ToolsServer {
	s := &ToolsServer{}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("recordWorkflowRun", "Record that a workflow was executed in the current chat session", s.recordWorkflowRun),
		mcp.NewServerTool("deleteWorkflow", "Delete a workflow by its URI", s.deleteWorkflow),
	)

	return s
}

func (s *ToolsServer) OnMessage(ctx context.Context, msg mcp.Message) {
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

func (s *ToolsServer) initialize(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
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

func (s *ToolsServer) recordWorkflowRun(ctx context.Context, data struct {
	URI string `json:"uri"`
}) (*map[string]string, error) {
	mcpSession := mcp.SessionFromContext(ctx).Root()

	var uris []string
	mcpSession.Get(types.WorkflowURIsSessionKey, &uris)

	// Deduplicate: only append if URI is not already recorded.
	var found bool
	for _, u := range uris {
		if u == data.URI {
			found = true
			break
		}
	}
	if !found {
		uris = append(uris, data.URI)
	}

	mcpSession.Set(types.WorkflowURIsSessionKey, uris)

	return &map[string]string{"uri": data.URI}, nil
}

func (s *ToolsServer) deleteWorkflow(_ context.Context, data struct {
	URI string `json:"uri"`
}) (*struct{}, error) {
	workflowName, err := parseWorkflowURI(data.URI)
	if err != nil {
		return nil, err
	}

	workflowPath := filepath.Join(".", workflowsDir, workflowName+".md")
	if err := os.Remove(workflowPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to delete workflow: %w", err)
	}

	return &struct{}{}, nil
}
