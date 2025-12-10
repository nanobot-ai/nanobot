package config

import (
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

var DefaultConfig = types.Config{
	Agents: map[string]types.Agent{
		"main": {},
	},
}

var UI = types.Config{
	Agents: map[string]types.Agent{
		"nanobot.summary.agent": {
			Chat: new(bool),
			Instructions: types.DynamicInstructions{
				Instructions: `- you will generate a short title based on the first message a user begins a conversation with
- ensure it is not more than 80 characters long
- the title should be a summary of the user's message
- do not use quotes or colons`,
			},
		},
	},
	Publish: types.Publish{
		MCPServers: []string{"nanobot.meta", "nanobot.resources", "nanobot.agentui"},
	},
	MCPServers: map[string]mcp.Server{
		"nanobot.meta":      {},
		"nanobot.resources": {},
		"nanobot.agentui":   {},
	},
}
