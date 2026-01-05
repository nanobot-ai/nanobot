import { Server } from "@nanobot-ai/nanomcp";
import ExecuteTask from "./src/tools/execute-task.ts";
import StartTask from "./src/tools/start-task.ts";

const server = new Server(
	{
		name: "Nanobot Tasks",
		version: "0.0.0",
	},
	{
		tools: {
			StartTask,
			ExecuteTask,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9014);
}
