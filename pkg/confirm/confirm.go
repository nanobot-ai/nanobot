package confirm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/obotmcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const Timeout = 15 * time.Minute

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (*Service) HandleAuthURL(ctx context.Context, req mcp.AuthURLRequest) (bool, error) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return false, fmt.Errorf("no session found in context")
	}

	for session.Parent != nil {
		session = session.Parent
	}

	lookupCtx := mcp.WithSession(ctx, session)
	var candidates []string
	if u := strings.TrimSpace(req.ConnectURL); u != "" {
		candidates = append(candidates, u)
	}
	if u := obotmcp.ResourceURLFromAuthorizeURL(req.URL); u != "" {
		candidates = append(candidates, u)
	}
	var displayNames []string
	for _, s := range []string{req.DisplayName, req.ServerKey} {
		s = strings.TrimSpace(s)
		if s != "" {
			displayNames = append(displayNames, s)
		}
	}
	mcpID, catalogID := obotmcp.LookupObotConnectedServerIDs(lookupCtx, candidates, displayNames)

	meta := map[string]any{
		types.MetaPrefix + "oauth-url":   req.URL,
		types.MetaPrefix + "server-name": req.DisplayName,
	}
	if mcpID != "" {
		meta[types.MetaPrefix+"mcp-server-id"] = mcpID
	}
	if catalogID != "" {
		meta[types.MetaPrefix+"catalog-entry-id"] = catalogID
	}
	metaStr, _ := json.Marshal(meta)

	elicit := mcp.ElicitRequest{
		Message: fmt.Sprintf("MCP server %s requires authorization, please visit the following URL to continue: %s", req.DisplayName, req.URL),
		RequestedSchema: mcp.PrimitiveSchema{
			Type:       "object",
			Properties: map[string]mcp.PrimitiveProperty{},
		},
		Meta: metaStr,
	}

	var elicitResponse mcp.ElicitResult

	if err := session.Exchange(ctx, "elicitation/create", elicit, &elicitResponse); err != nil {
		return false, fmt.Errorf("failed to elicit confirmation: %w", err)
	}

	switch elicitResponse.Action {
	case "accept":
		return true, nil
	case "reject":
		return false, fmt.Errorf("user has rejected authorization for server %s", req.DisplayName)
	default:
		return false, fmt.Errorf("user has canceled authorization for server %s", req.DisplayName)
	}
}
