import AgentConfig from "@nanobot-ai/agentconfig";
import Coder from "@nanobot-ai/coder";
import { mergeConfig, Server } from "@nanobot-ai/nanomcp";
import Tasks from "@nanobot-ai/tasks";
import WorkspaceMcp from "@nanobot-ai/workspace-mcp";

const server = new Server(
	{
		name: "Nanobot Services",
		version: "0.0.1",
	},
	mergeConfig(
		AgentConfig.config,
		Coder.config,
		Tasks.config,
		WorkspaceMcp.config,
	),
);

export default server;

if (import.meta.main) {
	await server.serve(5174);
}
