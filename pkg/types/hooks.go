package types

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

// AgentConfigHook is a hook that can be used to configure the agent.
// Hook Name = "config"
type AgentConfigHook struct {
	Agent      *Agent                              `json:"agent,omitempty"`
	Meta       map[string]any                      `json:"_meta,omitempty"`
	SessionID  string                              `json:"sessionId,omitempty"`
	MCPServers map[string]AgentConfigHookMCPServer `json:"mcpServers,omitempty"`
}

type AgentConfigHookMCPServer struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func (a AgentConfigHookMCPServer) ToMCPServer() mcp.Server {
	return mcp.Server{
		BaseURL: a.URL,
		Headers: a.Headers,
	}
}

// AgentRequestHook is a hook that can be used to modify the request before it is sent to the MCP server.
// Hook Name = "request"
type AgentRequestHook struct {
	Request  *CompletionRequest  `json:"request,omitempty"`
	Response *CompletionResponse `json:"response,omitempty"`
}

// AgentResponseHook is a hook that can be used to modify the response before it is returned to the agent.
// Hook Name = "response"
type AgentResponseHook = AgentRequestHook

type SessionInitHook struct {
	URL       string         `json:"url"`
	SessionID string         `json:"sessionId"`
	Meta      map[string]any `json:"_meta,omitempty"`
}

func (s *SessionInitHook) Serialize() (any, error) {
	return s, nil
}

func (s *SessionInitHook) Deserialize(data any) (any, error) {
	return s, mcp.JSONCoerce(data, &s)
}

func IsUISession(ctx context.Context) bool {
	session := mcp.SessionFromContext(ctx)
	var sessionInit SessionInitHook
	session.Get(SessionInitSessionKey, &sessionInit)
	isUI, _ := sessionInit.Meta["ui"].(bool)
	return isUI
}

func GetWorkspaceID(ctx context.Context) string {
	var sessionInit SessionInitHook
	mcp.SessionFromContext(ctx).Get(SessionInitSessionKey, &sessionInit)
	workspace, _ := sessionInit.Meta["workspace"].(map[string]any)
	id, _ := workspace["id"].(string)
	return id
}
