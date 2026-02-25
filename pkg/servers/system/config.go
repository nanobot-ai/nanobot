package system

import (
	"context"
	"maps"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

var allowedPermsToTools = map[string][]string{
	"bash":            {"bash"},
	"read":            {"read"},
	"write":           {"write", "edit"},
	"edit":            {"edit"},
	"glob":            {"glob"},
	"grep":            {"grep"},
	"todoWrite":       {"todoWrite"},
	"webFetch":        {"webFetch"},
	"skills":          {"getSkill"},
	"mcpServers":      {"addMCPServer", "removeMCPServer"},
	"askUserQuestion": {"askUserQuestion"},
}

func (s *Server) config(ctx context.Context, params types.AgentConfigHook) (types.AgentConfigHook, error) {
	if agent := params.Agent; agent != nil {
		for _, perm := range agent.Permissions.Allowed(maps.Keys(allowedPermsToTools)) {
			for _, tool := range allowedPermsToTools[perm] {
				agent.MCPServers = append(agent.MCPServers, "nanobot.system/"+tool)
			}

			if perm == "skills" {
				// Get all available skills (built-in + user-defined)
				skillsList, err := s.listSkills(ctx, struct{}{})
				if err != nil {
					// If we can't list skills, log but don't fail the hook
					continue
				}

				// Build the skills prompt
				if len(skillsList.Skills) > 0 {
					var skillsPrompt strings.Builder
					skillsPrompt.WriteString("\n\n## Available Skills\n\n")
					skillsPrompt.WriteString("Skills provide detailed instructions for specific tasks. ")
					skillsPrompt.WriteString("When your task fits one of the skills below, call getSkill('skill-name') to load its instructions.\n\n")

					for _, skill := range skillsList.Skills {
						skillsPrompt.WriteString("- **")
						skillsPrompt.WriteString(skill.Name)
						skillsPrompt.WriteString("**: ")
						skillsPrompt.WriteString(skill.Description)
						skillsPrompt.WriteString("\n")
					}

					// Append to agent instructions
					agent.Instructions.Instructions += skillsPrompt.String()
				}

				// Make workflow tools available to agents with skills permission.
				agent.MCPServers = append(agent.MCPServers, "nanobot.workflow-tools")
			}
		}

		if params.MCPServers == nil {
			params.MCPServers = make(map[string]types.AgentConfigHookMCPServer, 3)
		}
		params.MCPServers["nanobot.system"] = types.AgentConfigHookMCPServer{}
		params.MCPServers["nanobot.workflows"] = types.AgentConfigHookMCPServer{}
		params.MCPServers["nanobot.workflow-tools"] = types.AgentConfigHookMCPServer{}

		session := mcp.SessionFromContext(ctx)

		// Configure MCP search server if environment variables are set
		if agent.Name != "nanobot.summary" && session != nil {
			envMap := session.GetEnvMap()
			if searchURL := envMap["MCP_SERVER_SEARCH_URL"]; searchURL != "" {
				mcpServer := types.AgentConfigHookMCPServer{
					URL: searchURL,
				}

				// Add authentication header if API key is provided
				if apiKey := envMap["MCP_SERVER_SEARCH_API_KEY"]; apiKey != "" {
					mcpServer.Headers = map[string]string{
						"Authorization": "Bearer " + apiKey,
					}
				}

				params.MCPServers["mcp-server-search"] = mcpServer

				// Also add to the agent's MCP server list so tools get fetched
				agent.Tools = append(agent.Tools, "mcp-server-search")
			}

			var dynamicServers DynamicMCPServers
			if session.Get(DynamicMCPServersSessionKey, &dynamicServers) {
				for name, server := range dynamicServers {
					// Skip dynamic servers that would overwrite existing MCP server definitions
					if _, exists := params.MCPServers[name]; exists {
						continue
					}
					params.MCPServers[name] = server
					agent.MCPServers = append(agent.MCPServers, name)
				}
			}
		}
	}

	return params, nil
}
