import { Client, createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { getTask } from "../lib/task.ts";

const schema = z.object({
	taskName: z.string().describe("The task name to start"),
	arguments: z
		.record(z.string())
		.optional()
		.describe("Optional arguments for the task"),
});

export default createTool({
	title: "Dispatch Task",
	description: async () => {
		return `Execute a task in the background with the given arguments.`;
	},
	messages: {
		invoking: "Dispatching task",
		invoked: "Task dispatched",
	},
	inputSchema: schema,
	async handler({ taskName, arguments: taskArgs }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const task = await getTask(
			client,
			taskName.replace(/^workspace:\/\/+/, ""),
		);
		if (!task.id) {
			return toolResult.error(
				`task "${taskName}" not found in workspace ${ctx.workspaceId}`,
			);
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
					`Missing argument: ${input.name} for task "${taskName}". The argument is described as: ${input.description}`,
				);
			}
		}

		const chatClient = new Client({
			baseUrl: "http://localhost:8099",
			path: "/mcp",
			sessionId: "new",
			workspaceId: ctx.workspaceId,
		});

		await chatClient.callMCPTool("chat", {
			payload: {
				type: "tool",
				arguments: {
					name: "ExecuteTaskStep",
					arguments: {
						name: task.id,
						params: args,
					},
				},
			},
			async: true,
			progressToken: crypto.randomUUID(),
		});
		const { id } = await chatClient.getSessionDetails();

		return toolResult.structured(
			`Task step ${task.name} dispatched with ${args}. Task ID: ${id}`,
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
