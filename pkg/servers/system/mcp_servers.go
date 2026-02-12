package system

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

// reservedServerNames contains names that cannot be used for dynamic MCP servers
var reservedServerNames = map[string]struct{}{
	"nanobot.system": {},
}

// DynamicMCPServersSessionKey is the session key for storing dynamically added MCP servers
const DynamicMCPServersSessionKey = "dynamicMCPServers"

// DynamicMCPServers stores dynamically added MCP servers for a session
type DynamicMCPServers map[string]types.AgentConfigHookMCPServer

// Serialize implements mcp.Serializable
func (d DynamicMCPServers) Serialize() (any, error) {
	return d, nil
}

// Deserialize implements mcp.Deserializable
func (d *DynamicMCPServers) Deserialize(data any) (any, error) {
	result := make(DynamicMCPServers)
	if err := mcp.JSONCoerce(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddMCPServerParams are the parameters for the addMCPServer tool
type AddMCPServerParams struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// RemoveMCPServerParams are the parameters for the removeMCPServer tool
type RemoveMCPServerParams struct {
	Name string `json:"name"`
}

func (s *Server) addMCPServer(ctx context.Context, params AddMCPServerParams) (map[string]any, error) {
	if params.URL == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("url is required")
	}
	if params.Name == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("name is required")
	}

	// Validate server name
	if strings.Contains(params.Name, "/") {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("server name must not contain '/'")
	}
	if _, exists := reservedServerNames[params.Name]; exists {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("server name '%s' is reserved", params.Name)
	}

	// Validate URL format
	parsedURL, err := url.Parse(params.URL)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid URL: %v", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("URL must use http or https scheme")
	}

	// Get session
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil, mcp.ErrRPCInternal.WithMessage("no session found")
	}

	// Validate that the URL hostname+port matches the MCP_SERVER_SEARCH_URL host (Obot host)
	envMap := session.GetEnvMap()
	if searchURL := envMap["MCP_SERVER_SEARCH_URL"]; searchURL != "" {
		searchParsed, err := url.Parse(searchURL)
		if err == nil {
			if parsedURL.Host != searchParsed.Host {
				return nil, mcp.ErrRPCInvalidParams.WithMessage("URL host %q does not match the allowed host %q", parsedURL.Host, searchParsed.Host)
			}
		}
	}

	// Use MCP_API_KEY from the environment as the Bearer token for dynamic servers
	var headers map[string]string

	if apiKey := session.GetEnvMap()["MCP_API_KEY"]; apiKey != "" {
		headers = map[string]string{
			"Authorization": "Bearer " + apiKey,
		}
	}

	// Create the new server config
	newServer := types.AgentConfigHookMCPServer{
		URL:     params.URL,
		Headers: headers,
	}

	// Get or create dynamic servers map from session
	var dynamicServers DynamicMCPServers
	if !session.Get(DynamicMCPServersSessionKey, &dynamicServers) {
		dynamicServers = make(DynamicMCPServers)
	}

	// Add new server to map
	dynamicServers[params.Name] = newServer

	// Save back to session
	session.Set(DynamicMCPServersSessionKey, dynamicServers)

	result := map[string]any{
		"success": true,
		"name":    params.Name,
		"url":     params.URL,
		"message": fmt.Sprintf("Successfully added MCP server '%s'. The server's tools will be available in the next agent turn.", params.Name),
	}

	return result, nil
}

func (s *Server) removeMCPServer(ctx context.Context, params RemoveMCPServerParams) (map[string]any, error) {
	if params.Name == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("name is required")
	}

	// Get session
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil, mcp.ErrRPCInternal.WithMessage("no session found")
	}

	// Get dynamic servers map from session
	var dynamicServers DynamicMCPServers
	if session.Get(DynamicMCPServersSessionKey, &dynamicServers) {
		// Delete server from map (no-op if it doesn't exist)
		delete(dynamicServers, params.Name)

		// Save back to session
		session.Set(DynamicMCPServersSessionKey, dynamicServers)
	}

	return map[string]any{
		"success": true,
		"name":    params.Name,
		"message": fmt.Sprintf("Successfully removed MCP server '%s'. The server's tools will no longer be available in the next agent turn.", params.Name),
	}, nil
}
