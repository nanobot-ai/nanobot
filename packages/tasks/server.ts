import { Server } from "@nanobot-ai/nanomcp";
import ExecuteTask from "./src/tools/execute-task.ts";
import DispatchTask from "./src/tools/task.ts";
import TaskStatus from "./src/tools/task-status.ts";

const server = new Server(
	{
		name: "Nanobot Tasks",
		version: "0.0.0",
	},
	{
		tools: {
			DispatchTask,
			ExecuteTask,
			TaskStatus,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9014);
}
