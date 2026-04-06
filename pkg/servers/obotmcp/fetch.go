package obotmcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

const connectedServersCacheTTL = 15 * time.Minute

var ErrSearchNotConfigured = errors.New("MCP_SERVER_SEARCH_URL is not configured")

type connectedServerLister interface {
	ConnectedMCPServers(context.Context) ([]ConnectedServer, error)
}

type connectedServersCache struct {
	mu      sync.Mutex
	servers []ConnectedServer
	fetched time.Time
}

func (c *connectedServersCache) get() ([]ConnectedServer, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.fetched) >= connectedServersCacheTTL {
		return nil, false
	}
	return c.servers, true
}

func (c *connectedServersCache) set(servers []ConnectedServer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.servers = servers
	c.fetched = time.Now()
}

func (c *connectedServersCache) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fetched = time.Time{}
}

// This cache is intentionally process-global because the runtime is single-user and
// all sessions should share the same connected-server snapshot and refresh cadence.
var globalConnectedServersCache connectedServersCache

type obotConnectedServerLister struct{}

func (obotConnectedServerLister) ConnectedMCPServers(ctx context.Context) ([]ConnectedServer, error) {
	if cached, ok := globalConnectedServersCache.get(); ok {
		return cached, nil
	}

	servers, err := fetchConnectedMCPServers(ctx)
	if err != nil {
		return nil, err
	}

	globalConnectedServersCache.set(servers)
	return servers, nil
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
	if result.StructuredContent == nil {
		return nil, fmt.Errorf("decode connected MCP servers: missing structured content")
	}

	var payload connectedServersResult
	if err := mcp.JSONCoerce(result.StructuredContent, &payload); err != nil {
		return nil, fmt.Errorf("decode connected MCP servers from structured content: %w", err)
	}

	return payload.ConnectedServers, nil
}
