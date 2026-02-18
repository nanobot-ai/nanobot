package config

import (
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

var UI = types.Config{
	Agents: map[string]types.Agent{
		"nanobot.summary": {
			HookAgent: types.HookAgent{
				Chat: new(bool),
				Instructions: types.DynamicInstructions{
					Instructions: `- you will generate a short title based on the first message a user begins a conversation with
- ensure it is not more than 80 characters long
- the title should be a summary of the user's message
- do not use quotes or colons`,
				},
			},
		},
		"nanobot.compact": {
			HookAgent: types.HookAgent{
				Name:      "Compaction Summarizer",
				Model:     "claude-3-7-sonnet-latest",
				MaxTokens: 1_024,
				Instructions: types.DynamicInstructions{
					Instructions: `You condense Nanobot conversations when the history nears the context window.
- Produce a succinct but information-dense markdown summary with headings for Context, Work Completed, Pending Items, and Tool Archives.
- Preserve key facts, decisions, constraints, file paths, URLs, and numerical details needed to continue the work without rereading the raw history.
- For each truncated or external tool artifact, emit a bullet in Tool Archives like "• Tool archive <id>: ~/.nanobot/tool-output/<id> (use Read/Grep to recover details)".
- Keep the summary under 900 words and prefer bullet lists over paragraphs.
- Do not invent information; if something was omitted or truncated, say so explicitly.`,
				},
			},
		},
	},
	Publish: types.Publish{
		MCPServers: []string{"nanobot.meta", "nanobot.workflows"},
		Resources:  []string{"nanobot.system"},
	},
	MCPServers: map[string]mcp.Server{
		"nanobot.meta":      {},
		"nanobot.workflows": {},
		"nanobot.system":    {},
	},
}
