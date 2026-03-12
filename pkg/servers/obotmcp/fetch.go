package obotmcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

var ErrSearchNotConfigured = errors.New("MCP_SERVER_SEARCH_URL is not configured")

type connectedServerLister interface {
	ConnectedMCPServers(context.Context) ([]ConnectedServer, error)
}

type obotConnectedServerLister struct{}

func (obotConnectedServerLister) ConnectedMCPServers(ctx context.Context) ([]ConnectedServer, error) {
	return fetchConnectedMCPServers(ctx)
}

type connectedServersResult struct {
	ConnectedServers []ConnectedServer `json:"connected_servers"`
}

func fetchConnectedMCPServers(ctx context.Context) ([]ConnectedServer, error) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil, nil
	}

	envMap := session.GetEnvMap()
	searchURL := strings.TrimSpace(envMap["MCP_SERVER_SEARCH_URL"])
	if searchURL == "" {
		return nil, ErrSearchNotConfigured
	}

	headers := map[string]string{}
	if apiKey := strings.TrimSpace(envMap["MCP_API_KEY"]); apiKey != "" {
		headers["Authorization"] = "Bearer " + apiKey
	} else if apiKey := strings.TrimSpace(envMap["MCP_SERVER_SEARCH_API_KEY"]); apiKey != "" {
		headers["Authorization"] = "Bearer " + apiKey
	}

	clientCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client, err := mcp.NewClient(clientCtx, "obot-connected-servers", mcp.Server{
		BaseURL: searchURL,
		Headers: headers,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to obot-connected-servers: %w", err)
	}
	defer client.Close(true)

	result, err := client.Call(clientCtx, "obot_list_connected_mcp_servers", map[string]any{})
	if err != nil {
		return nil, err
	}

	return extractConnectedMCPServers(result)
}

func extractConnectedMCPServers(result *mcp.CallToolResult) ([]ConnectedServer, error) {
	if result == nil {
		return nil, nil
	}

	var payload connectedServersResult
	if result.StructuredContent != nil {
		if err := mcp.JSONCoerce(result.StructuredContent, &payload); err == nil {
			return payload.ConnectedServers, nil
		} else {
			return nil, fmt.Errorf("decode connected MCP servers from structured content: %w", err)
		}
	}

	var sawTextContent bool
	for _, content := range result.Content {
		if content.Type != "text" || strings.TrimSpace(content.Text) == "" {
			continue
		}
		sawTextContent = true

		if err := mcp.JSONCoerce(content.Text, &payload); err == nil {
			return payload.ConnectedServers, nil
		} else {
			return nil, fmt.Errorf("decode connected MCP servers from text content: %w", err)
		}
	}

	if sawTextContent {
		return nil, fmt.Errorf("decode connected MCP servers: no parseable text payload")
	}

	return nil, nil
}
