import { Client, createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { getTask, getTasksDescription } from "../lib/task.ts";

const schema = z.object({
	taskName: z.string().describe("The task name to start"),
	arguments: z
		.record(z.string())
		.optional()
		.describe("Optional arguments for the task"),
});

export default createTool({
	title: "Start Task",
	description: async (ctx) => {
		const client = await ensureConnected(ctx.workspaceId);
		const tasksDescriptions = await getTasksDescription(client);
		return `Execute one of the following tasks in the background with the given arguments.

<available_tasks>
${tasksDescriptions}
</available_tasks>
`;
	},
	messages: {
		invoking: "Dispatching task",
		invoked: "Task dispatched",
	},
	inputSchema: schema,
	async handler({ taskName, arguments: taskArgs }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const task = await getTask(client, taskName);
		if (!task.name) {
			return toolResult.error(`task not found: ${taskName}`);
		}

		// collect input
		const args: Record<string, string> = {};
		for (const input of task.inputs || []) {
			const arg = taskArgs?.[input.name];
			if (arg) {
				args[input.name] = arg;
			} else if (input.default) {
				args[input.name] = input.default;
			} else {
				return toolResult.error(
					`Missing argument: ${input.name} for task ${taskName}. The argument is described as: ${input.description}`,
				);
			}
		}

		const chatInput = {
			type: "tool",
			payload: {
				name: task.name,
				arguments: args,
			},
		};

		const chatClient = new Client({
			baseUrl: "http://localhost:5173",
			path: "/mcp",
			sessionId: "new",
			workspaceId: ctx.workspaceId,
		});

		await chatClient.callMCPTool("chat", {
			payload: JSON.stringify(chatInput),
			async: true,
		});
		const { id } = await chatClient.getSessionDetails();

		return toolResult.structured(
			`Task ${task.name} dispatched with ${args}. Task ID: ${id}`,
			{
				task: {
					name: task.name,
					arguments: args,
				},
				id,
				created_at: new Date().toISOString(),
			},
		);
	},
});
