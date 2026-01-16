import { Server } from "@nanobot-ai/nanomcp";
import ExecuteTaskStep from "./src/tools/execute-task-step.ts";
import TaskStepStatus from "./src/tools/task-step-status.ts";

const server = new Server(
	{
		name: "Nanobot Tasks",
		version: "0.0.0",
	},
	{
		tools: {
			ExecuteTaskStep,
			TaskStepStatus,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9014);
}
