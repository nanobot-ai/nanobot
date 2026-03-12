package system

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/obotmcp"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
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
	"askUserQuestion": {"askUserQuestion"},
}

func (s *Server) config(ctx context.Context, params types.AgentConfigHook) (types.AgentConfigHook, error) {
	if agent := params.Agent; agent != nil && agent.Name != "nanobot.summary" {
		for _, perm := range agent.Permissions.Allowed(maps.Keys(allowedPermsToTools)) {
			for _, tool := range allowedPermsToTools[perm] {
				agent.Tools = append(agent.Tools, "nanobot.system/"+tool)
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

					if session := mcp.SessionFromContext(ctx); session != nil {
						if envMap := session.GetEnvMap(); envMap["OBOT_URL"] != "" {
							skillsPrompt.WriteString("\nWhen you need a new skill that is not already installed, use the searchSkills tool to search Obot.\n")
						}
					}
					// Append to agent instructions
					agent.Instructions.Instructions += skillsPrompt.String()
				}

				// Make workflow and artifact tools available to agents with skills permission.
				agent.Tools = append(agent.Tools, "nanobot.workflow-tools")
				agent.Tools = append(agent.Tools, "nanobot.artifacts")

				session := mcp.SessionFromContext(ctx)
				if session != nil {
					if envMap := session.GetEnvMap(); envMap["OBOT_URL"] != "" {
						agent.Tools = append(agent.Tools, "nanobot.skills")
					}
				}
			}
		}

		// Inject session directory and workflow directory paths into agent instructions
		if params.SessionID != "" {
			absSessionDir := sessionDir(params.SessionID)
			cwd, err := os.Getwd()
			if err != nil {
				return params, fmt.Errorf("failed to get working directory: %w", err)
			}
			absWorkflowDir := filepath.Join(cwd, skillformat.WorkflowsDir)
			absSkillsDir := filepath.Join(cwd, s.configDir, "skills")

			agent.Instructions.Instructions += fmt.Sprintf(`

## File Paths

Always use absolute file paths when using Read, Write, Edit, Glob, Grep, and Bash tools.

Your session directory is: %s
This is where your files for this session live. The Bash tool defaults to this as its working directory.

Workflow files must always be stored in: %s
Do NOT put workflow files in the session directory.

Skill files are stored in: %s
Do NOT put skill files in the session directory or workflow directory.
`, absSessionDir, absWorkflowDir, absSkillsDir)
		}

		if params.MCPServers == nil {
			params.MCPServers = make(map[string]types.AgentConfigHookMCPServer, 4)
		}
		params.MCPServers["nanobot.system"] = types.AgentConfigHookMCPServer{}
		params.MCPServers["nanobot.workflows"] = types.AgentConfigHookMCPServer{}
		params.MCPServers["nanobot.workflow-tools"] = types.AgentConfigHookMCPServer{}
		params.MCPServers["nanobot.artifacts"] = types.AgentConfigHookMCPServer{}
		session := mcp.SessionFromContext(ctx)
		if session != nil {
			if envMap := session.GetEnvMap(); envMap["OBOT_URL"] != "" && agent.Permissions != nil && agent.Permissions.IsAllowed("skills") {
				params.MCPServers["nanobot.skills"] = types.AgentConfigHookMCPServer{}
			}
		}

		obotmcp.ConfigureIntegration(ctx, s.configDir, agent, &params)
	}

	return params, nil
}
