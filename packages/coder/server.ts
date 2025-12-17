import { Server } from "@nanobot-ai/nanomcp";
import config from "./nanomcp.config";

const server = new Server(
	{
		name: "test",
		version: "1.0.0",
	},
	config,
);

if (import.meta.main) {
	await server.serve(9013);
}
