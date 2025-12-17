import {
	description as todoDescription,
	inputSchema as todoSchema,
} from "@nanobot-ai/coder/tools/todowrite.ts";
import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { zodToJsonSchema } from "zod-to-json-schema";
import { getTask, getTasks } from "../lib/task.ts";

const schema = z.object({
	taskName: z.string().describe("The task name to start"),
});

export default createTool({
	title: "Start Task",
	description: async (ctx) => {
		const client = await ensureConnected(ctx.workspaceId);
		const tasks = await getTasks(client);
		const available = tasks
			.map((s) => `${s.name}: ${s.description}`)
			.join("\n");
		return `Execute a task within the main conversation

<available_tasks>
${available}
</available_tasks>

The output of this tool will contain a reference to the next file to read for the steps of the task.
`;
	},
	messages: {
		invoking: "Loading task",
		invoked: "Task loaded",
	},
	inputSchema: schema,
	async handler({ taskName }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const task = await getTask(client, taskName);
		if (!task.name) {
			return toolResult.error(`task not found: ${taskName}`);
		}

		const jsonSchema = zodToJsonSchema(todoSchema);
		console.log(`Schema: ${JSON.stringify(jsonSchema)}`);

		let todoContent = "";
		const resp = await ctx.sample?.({
			maxTokens: 100_000,
			systemPrompt:
				"Given the input text in <input>...</input>, create an appropriate todo list and call the TodoWrite tool",
			tools: [
				{
					name: "TodoWrite",
					description: todoDescription,
					// @ts-expect-error
					inputSchema: jsonSchema,
				},
			],
			toolChoice: {
				mode: "required",
			},
			messages: [
				{
					role: "user",
					content: {
						type: "text",
						text: `<input>${task.instructions}</input>`,
					},
				},
			],
		});

		console.log(`Resp: ${JSON.stringify(resp, null, 2)}`);
		if (Array.isArray(resp?.content)) {
			for (const messagePart of resp.content) {
				if (messagePart.type === "tool_use" && messagePart.input) {
					todoContent = JSON.stringify(messagePart.input);
				}
			}
		} else if (resp?.content?.type === "tool_use" && resp.content.input) {
			todoContent = JSON.stringify(resp.content.input);
		}
		if (todoContent) {
			await client.writeTextFile(".nanobot/status/todo.json", todoContent);
			todoContent = `Your current todo content is:\n\n${todoContent}`;
		}
		return toolResult.text(
			`<command-message>The "${task.name}" task is loading</command-message>\n<command-name>${task.name}</command-name>`,
			`Your task instructions are:

<instructions>${task.instructions}</instructions>

Follow the instructions, track your progress with TodoWrite.
${todoContent}`,
		);
	},
});
