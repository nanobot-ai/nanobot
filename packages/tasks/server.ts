import { Server } from "@nanobot-ai/nanomcp";
import TaskExecute from "./src/tools/execute-task.ts";
import Task from "./src/tools/task.ts";

const server = new Server(
	{
		name: "Nanobot Tasks",
		version: "0.0.0",
	},
	{
		tools: {
			Task,
			// TaskExecute,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9014);
}
