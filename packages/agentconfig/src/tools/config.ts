import { createTool, hooks, toolResult } from "@nanobot-ai/nanomcp";
import { amendConfig } from "../lib/agentconfig.ts";

const schema = hooks.AgentConfigHookSchema;

export default createTool({
	title: "Config",
	description: "Modifies the agent config based on the file system",
	inputSchema: schema,
	outputSchema: schema,
	async handler(config, ctx) {
		config = await amendConfig(ctx.workspaceId, config);
		console.log(`New Agent Config: ${JSON.stringify(config, null, "  ")}`);
		return toolResult.structured(config);
	},
});
