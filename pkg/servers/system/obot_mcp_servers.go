package system

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"log/slog"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const (
	obotConfiguredMCPServersSessionKey = "obotConfiguredMCPServers"
	obotConfiguredMCPServersCacheTTL   = 30 * time.Second
)

var obotServerNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type obotConfiguredMCPServersCache struct {
	ExpiresAt          time.Time         `json:"expiresAt,omitempty"`
	Servers            DynamicMCPServers `json:"servers,omitempty"`
	ServerDescriptions map[string]string `json:"serverDescriptions,omitempty"`
}

type obotListResponse[T any] struct {
	Items []T `json:"items"`
}

type obotMCPManifest struct {
	Name             string `json:"name,omitempty"`
	ShortDescription string `json:"shortDescription,omitempty"`
	Description      string `json:"description,omitempty"`
}

type obotMCPServer struct {
	ID         string          `json:"id,omitempty"`
	Alias      string          `json:"alias,omitempty"`
	Configured bool            `json:"configured,omitempty"`
	ConnectURL string          `json:"connectURL,omitempty"`
	Manifest   obotMCPManifest `json:"manifest,omitempty"`
}

type obotMCPServerInstance struct {
	MCPServerID string `json:"mcpServerID,omitempty"`
	ConnectURL  string `json:"connectURL,omitempty"`
}

func (s *Server) getObotConfiguredMCPServers(ctx context.Context) (DynamicMCPServers, map[string]string, error) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil, nil, nil
	}

	if cachedServers, cachedDescriptions, ok := getCachedObotConfiguredMCPServers(session); ok {
		return cachedServers, cachedDescriptions, nil
	}

	envMap := session.GetEnvMap()
	apiBaseURL := obotAPIBaseURL(envMap)
	userToken := obotUserToken(envMap)
	if apiBaseURL == "" || userToken == "" {
		return nil, nil, nil
	}

	mcpAccessToken := obotMCPAccessToken(envMap, userToken)
	servers, descriptions, err := fetchObotConfiguredMCPServers(ctx, apiBaseURL, userToken, mcpAccessToken)
	if err != nil {
		return nil, nil, err
	}

	session.Set(obotConfiguredMCPServersSessionKey, obotConfiguredMCPServersCache{
		ExpiresAt:          time.Now().Add(obotConfiguredMCPServersCacheTTL),
		Servers:            servers,
		ServerDescriptions: descriptions,
	})

	return servers, descriptions, nil
}

func getCachedObotConfiguredMCPServers(session *mcp.Session) (DynamicMCPServers, map[string]string, bool) {
	var cache obotConfiguredMCPServersCache
	if !session.Get(obotConfiguredMCPServersSessionKey, &cache) {
		return nil, nil, false
	}
	if time.Now().After(cache.ExpiresAt) {
		return nil, nil, false
	}
	return cache.Servers, cache.ServerDescriptions, true
}

func fetchObotConfiguredMCPServers(ctx context.Context, apiBaseURL, userToken, mcpAccessToken string) (DynamicMCPServers, map[string]string, error) {
	httpClient := &http.Client{Timeout: 15 * time.Second}

	var (
		configuredServers obotListResponse[obotMCPServer]
		serverInstances   obotListResponse[obotMCPServerInstance]
		allServers        obotListResponse[obotMCPServer]
	)

	if err := obotGet(ctx, httpClient, apiBaseURL, userToken, "/mcp-servers", &configuredServers); err != nil {
		return nil, nil, err
	}
	if err := obotGet(ctx, httpClient, apiBaseURL, userToken, "/mcp-server-instances", &serverInstances); err != nil {
		return nil, nil, err
	}
	if err := obotGet(ctx, httpClient, apiBaseURL, userToken, "/all-mcps/servers", &allServers); err != nil {
		return nil, nil, err
	}

	result := DynamicMCPServers{}
	descriptions := map[string]string{}
	usedNames := map[string]int{}
	for _, server := range configuredServers.Items {
		if !server.Configured || strings.TrimSpace(server.ConnectURL) == "" {
			continue
		}

		name := uniqueObotServerName(server.Alias, server.Manifest.Name, server.ID, usedNames)
		result[name] = types.AgentConfigHookMCPServer{
			URL:     server.ConnectURL,
			Headers: obotMCPHeaders(mcpAccessToken),
		}
		if description := obotServerDescription(server); description != "" {
			descriptions[name] = description
		}
	}

	serverByID := make(map[string]obotMCPServer, len(allServers.Items))
	for _, server := range allServers.Items {
		serverByID[server.ID] = server
	}

	for _, instance := range serverInstances.Items {
		if strings.TrimSpace(instance.ConnectURL) == "" {
			continue
		}

		server := serverByID[instance.MCPServerID]
		name := uniqueObotServerName(server.Alias, server.Manifest.Name, instance.MCPServerID, usedNames)
		result[name] = types.AgentConfigHookMCPServer{
			URL:     instance.ConnectURL,
			Headers: obotMCPHeaders(mcpAccessToken),
		}
		if description := obotServerDescription(server); description != "" {
			descriptions[name] = description
		}
	}

	if len(result) == 0 {
		return nil, nil, nil
	}

	return result, descriptions, nil
}

func obotGet(ctx context.Context, httpClient *http.Client, apiBaseURL, userToken, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(apiBaseURL, "/")+path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Obot API %s returned %s", path, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func obotAPIBaseURL(envMap map[string]string) string {
	for _, key := range []string{"OBOT_URL", "OBOT_SERVER_URL"} {
		if value := strings.TrimSpace(envMap[key]); value != "" {
			value = strings.TrimSuffix(value, "/")
			if strings.HasSuffix(value, "/api") {
				return value
			}
			return value + "/api"
		}
	}
	return ""
}

func obotUserToken(envMap map[string]string) string {
	for _, key := range []string{"OBOT_TOKEN", "ANTHROPIC_API_KEY", "OPENAI_API_KEY"} {
		value := strings.TrimSpace(envMap[key])
		if value == "" || strings.HasPrefix(value, "ok1-") {
			continue
		}
		return value
	}
	return ""
}

func obotMCPAccessToken(envMap map[string]string, fallback string) string {
	if value := strings.TrimSpace(envMap["MCP_API_KEY"]); value != "" {
		return value
	}
	return fallback
}

func obotMCPHeaders(token string) map[string]string {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

func obotServerDescription(server obotMCPServer) string {
	if server.Manifest.ShortDescription != "" {
		return server.Manifest.ShortDescription
	}
	return server.Manifest.Description
}

func uniqueObotServerName(alias, manifestName, fallback string, used map[string]int) string {
	name := alias
	if strings.TrimSpace(name) == "" {
		name = manifestName
	}
	if strings.TrimSpace(name) == "" {
		name = fallback
	}
	if strings.TrimSpace(name) == "" {
		name = "obot-server"
	}

	name = strings.ToLower(strings.TrimSpace(name))
	name = obotServerNameSanitizer.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-._")
	if name == "" {
		name = "obot-server"
	}

	if _, reserved := reservedServerNames[name]; reserved {
		name = "obot-" + name
	}
	for _, prefix := range reservedServerNamePrefixes {
		if strings.HasPrefix(name, prefix) {
			name = "obot-" + name
			break
		}
	}

	base := name
	if _, exists := used[name]; !exists {
		used[name] = 1
		return name
	}

	for i := 2; ; i++ {
		name = fmt.Sprintf("%s-%d", base, i)
		if _, exists := used[name]; !exists {
			used[name] = 1
			return name
		}
	}
}

func appendConnectedMCPServerInstructions(agent *types.HookAgent, descriptions map[string]string) {
	if len(descriptions) == 0 {
		return
	}

	serverNames := make([]string, 0, len(descriptions))
	for name := range descriptions {
		serverNames = append(serverNames, name)
	}
	slices.Sort(serverNames)

	var builder strings.Builder
	builder.WriteString("\n\n## Connected MCP Servers\n\n")
	builder.WriteString("The user already has these MCP servers connected for this session:\n")
	for _, name := range serverNames {
		builder.WriteString("- ")
		builder.WriteString(name)
		if description := strings.TrimSpace(descriptions[name]); description != "" {
			builder.WriteString(": ")
			builder.WriteString(description)
		}
		builder.WriteString("\n")
	}
	builder.WriteString("Use the connected tools when they are relevant.\n")

	agent.Instructions.Instructions += builder.String()
}

func mergeSessionMCPServers(agent *types.HookAgent, mcpServers map[string]types.AgentConfigHookMCPServer, imported DynamicMCPServers, descriptions map[string]string, existing map[string]struct{}) map[string]string {
	if agent == nil || len(imported) == 0 {
		return nil
	}

	addedDescriptions := map[string]string{}
	for name, server := range imported {
		if _, exists := existing[name]; exists {
			continue
		}
		existing[name] = struct{}{}
		mcpServers[name] = server
		agent.MCPServers = append(agent.MCPServers, name)
		if description := descriptions[name]; description != "" {
			addedDescriptions[name] = description
		}
	}

	return addedDescriptions
}

func logObotConfiguredServerError(err error) {
	if err == nil {
		return
	}
	slog.Debug("failed to load configured MCP servers from Obot", "error", err)
}
