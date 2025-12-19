import AgentConfig from "@nanobot-ai/agentconfig";
import Coder from "@nanobot-ai/coder";
import { mergeConfig, Server } from "@nanobot-ai/nanomcp";
import Tasks from "@nanobot-ai/tasks";
import WorkspaceMcp from "@nanobot-ai/workspace-mcp";
import WorkspaceResources from "@nanobot-ai/workspace-resources";

const server = new Server(
	{
		name: "Nanobot Services",
		version: "0.0.1",
	},
	mergeConfig(AgentConfig.config, WorkspaceMcp.config),
);

server.router.map("/mcp/tasks", Tasks.mcp.handler);
server.router.map("/mcp/coder", Coder.mcp.handler);
server.router.map("/mcp/workspace-resources", WorkspaceResources.mcp.handler);

export default server;

if (import.meta.main) {
	await server.serve(5174);
}
