import { Server } from "@nanobot-ai/nanomcp";
import StartTask from "./src/tools/task.ts";

const server = new Server(
	{
		name: "Nanobot Tasks",
		version: "0.0.0",
	},
	{
		tools: {
			StartTask,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9014);
}
