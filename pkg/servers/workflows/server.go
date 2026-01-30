package workflows

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
	case "resources/list":
		mcp.Invoke(ctx, msg, s.resourcesList)
	case "resources/read":
		mcp.Invoke(ctx, msg, s.resourcesRead)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%v", msg.Method))
	}
}

func (s *Server) initialize(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Resources: &mcp.ResourcesServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}

// parseWorkflowDescription extracts description from workflow markdown content.
// Looks for: # Workflow: <name>\n\n<description paragraph>
func parseWorkflowDescription(content string) string {
	lines := strings.Split(content, "\n")

	// Find the "# Workflow:" line
	startIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# Workflow:") {
			startIdx = i + 1
			break
		}
	}

	if startIdx == -1 || startIdx >= len(lines) {
		return ""
	}

	// Skip empty lines after header
	for startIdx < len(lines) && strings.TrimSpace(lines[startIdx]) == "" {
		startIdx++
	}

	if startIdx >= len(lines) {
		return ""
	}

	// Collect description paragraph (until empty line or next header)
	var desc []string
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			break
		}
		desc = append(desc, trimmed)
	}

	return strings.Join(desc, " ")
}

func (s *Server) resourcesList(ctx context.Context, _ mcp.Message, _ mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	workflowsDir := filepath.Join(".", "workflows")

	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		// Directory doesn't exist or can't be read - return empty list
		return &mcp.ListResourcesResult{Resources: []mcp.Resource{}}, nil
	}

	var result []mcp.Resource
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")

		// Read file to extract description
		content, err := os.ReadFile(filepath.Join(workflowsDir, entry.Name()))
		if err != nil {
			// Skip files we can't read
			continue
		}

		description := parseWorkflowDescription(string(content))

		result = append(result, mcp.Resource{
			URI:         fmt.Sprintf("workflow:///%s", name),
			Name:        name,
			Description: description,
			MimeType:    "text/markdown",
		})
	}

	return &mcp.ListResourcesResult{Resources: result}, nil
}

func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Parse the URI - expecting workflow:///name format
	if !strings.HasPrefix(request.URI, "workflow:///") {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid workflow URI format, expected workflow:///name")
	}

	workflowName := strings.TrimPrefix(request.URI, "workflow:///")
	if workflowName == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow name is required")
	}

	// Add .md extension if not present
	if !strings.HasSuffix(workflowName, ".md") {
		workflowName = workflowName + ".md"
	}

	workflowPath := filepath.Join(".", "workflows", workflowName)
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", request.URI)
	}

	contentStr := string(content)
	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{
			{
				URI:      request.URI,
				Name:     strings.TrimSuffix(filepath.Base(workflowName), ".md"),
				MIMEType: "text/markdown",
				Text:     &contentStr,
			},
		},
	}, nil
}
