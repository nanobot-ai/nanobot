import { Server } from "@nanobot-ai/nanomcp";
import {
	listResource,
	listResourceTemplates,
	readResource,
} from "./src/lib/resources.ts";

const server = new Server(
	{
		name: "Nanobot Workspace Resource MCP Server",
		version: "0.0.1",
	},
	{
		resources: {
			list: listResource,
			read: readResource,
			templates: listResourceTemplates,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9016);
}
