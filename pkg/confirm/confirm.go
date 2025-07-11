package confirm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/pkg/browser"
)

const Timeout = 15 * time.Minute

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (*Service) Confirm(ctx context.Context, session *mcp.Session, target types.TargetMapping[mcp.Tool], funcCall *types.ToolCall) (*types.CallResult, error) {
	for session.Parent != nil {
		session = session.Parent
	}

	if session.InitializeRequest.Capabilities.Elicitation == nil {
		return nil, nil
	}

	approvedCalls := make(map[string]any)
	if !session.Get("approvedCalls", &approvedCalls) {
		session.Set("approvedCalls", approvedCalls)
	}

	approveCallKey := fmt.Sprintf("%s/%s", target.MCPServer, target.TargetName)
	if _, exists := approvedCalls[approveCallKey]; exists {
		// If the call is already approved, we can skip confirmation
		return nil, nil
	}

	req := types.ToolCallConfirm{
		MCPServer:  target.MCPServer,
		Tool:       target.Target,
		Invocation: funcCall,
	}

	meta, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal confirmation request: %w", err)
	}

	elicit := mcp.ElicitRequest{
		Message: req.Message(),
		RequestedSchema: mcp.PrimitiveSchema{
			Type: "object",
			Properties: map[string]mcp.PrimitiveProperty{
				"answer": {
					Type:      "enum",
					Enum:      []string{"yes", "no", "always"},
					EnumNames: []string{"Yes", "No", "Always"},
				},
			},
		},
		Meta: meta,
	}

	var elicitResponse mcp.ElicitResult

	if err := session.Exchange(ctx, "elicitation/create", elicit, &elicitResponse); err != nil {
		return nil, fmt.Errorf("failed to elicit confirmation: %w", err)
	}

	var answer string
	switch elicitResponse.Action {
	case "reject":
		answer = "no"
	case "cancel":
		return nil, fmt.Errorf("user has cancelled call to tool %s on server %s", target.TargetName, target.MCPServer)
	case "accept":
		s, ok := elicitResponse.Content["answer"].(string)
		if !ok {
			return nil, fmt.Errorf("expected 'answer' field in elicit response content, got: %v", elicitResponse.Content["answer"])
		}
		answer = s
	}

	switch answer {
	case "yes":
		return nil, nil
	case "always":
		approvedCalls[approveCallKey] = true
		return nil, nil
	case "no":
		return &types.CallResult{
			IsError: true,
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Tool call to %s on server %s was declined by the user.", target.TargetName, target.MCPServer),
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected answer from user: %s", answer)
	}
}

func (*Service) HandleAuthURL(ctx context.Context, mcpServerName, url string) (bool, error) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return false, fmt.Errorf("no session found in context")
	}

	for session.Parent != nil {
		session = session.Parent
	}

	if session.InitializeRequest.Capabilities.Elicitation == nil {
		return false, nil
	}

	elicit := mcp.ElicitRequest{
		Message: fmt.Sprintf("MCP server %s requires authoriziation, please visit the following URL to continue: %s", mcpServerName, url),
		RequestedSchema: mcp.PrimitiveSchema{
			Type: "object",
			Properties: map[string]mcp.PrimitiveProperty{
				"answer": {
					Type:      "enum",
					Enum:      []string{"open", "user", "no"},
					EnumNames: []string{"Open for me", "I'll open it myself", "Decline"},
				},
			},
		},
	}

	var elicitResponse mcp.ElicitResult

	if err := session.Exchange(ctx, "elicitation/create", elicit, &elicitResponse); err != nil {
		return false, fmt.Errorf("failed to elicit confirmation: %w", err)
	}

	var answer string
	switch elicitResponse.Action {
	case "reject":
		answer = "no"
	case "cancel":
		return false, fmt.Errorf("user has declined authorization for server %s", mcpServerName)
	case "accept":
		s, ok := elicitResponse.Content["answer"].(string)
		if !ok {
			return false, fmt.Errorf("expected 'answer' field in elicit response content, got: %v", elicitResponse.Content["answer"])
		}
		answer = s
	}

	switch answer {
	case "open":
		_, _ = fmt.Fprintf(os.Stderr, "Opening browser to %s. If there is an issue paste this link into a browser manually\n", url)
		return true, browser.OpenURL(url)
	case "user":
		_, _ = fmt.Fprintf(os.Stderr, "Naviate to %s.\n", url)
		return true, nil
	case "no":
		return false, fmt.Errorf("user has declined authorization for server %s", mcpServerName)
	default:
		return false, fmt.Errorf("unexpected answer from user: %s", answer)
	}
}
