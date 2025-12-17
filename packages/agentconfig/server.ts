import { Server } from "@nanobot-ai/nanomcp";
import Config from "./src/tools/config.js";

const server = new Server(
	{
		name: "Agent Configuration",
		version: "0.0.1",
	},
	{
		tools: {
			config: Config,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve();
}
