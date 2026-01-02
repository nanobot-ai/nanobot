import * as path from "node:path";
import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { getTask, getTasksDescription } from "../lib/task.ts";

const schema = z.object({
	taskName: z.string().describe("The task name to execute"),
	arguments: z
		.record(z.string())
		.optional()
		.describe("Optional arguments for the task"),
});

export default createTool({
	title: "Execute Task",
	description: async (ctx) => {
		const client = await ensureConnected(ctx.workspaceId);
		const tasksDescriptions = await getTasksDescription(client);
		return `Execute one of the following tasks directly (not in the background) with the given arguments. This tool will provide you with the task instructions to execute.

<available_tasks>
${tasksDescriptions}
</available_tasks>
`;
	},
	messages: {
		invoking: "Executing task",
		invoked: "Task execution started",
	},
	inputSchema: schema,
	visibility: ["system"],
	async handler({ taskName, arguments: taskArgs }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const task = await getTask(client, taskName);

		if (!task.name) {
			return toolResult.error(`task not found: ${taskName}`);
		}

		// Collect input arguments
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

		// Create task execution record
		const sessionId = ctx.sessionId || "default";
		const taskDir = path.join(".nanobot", "status", sessionId);
		const taskFile = path.join(taskDir, "task.json");

		// Ensure the directory exists by trying to create it (writeTextFile will handle parent dirs)
		const taskRecord = {
			taskName: task.name,
			arguments: args,
			startedAt: new Date().toISOString(),
		};

		try {
			await client.writeTextFile(taskFile, JSON.stringify(taskRecord, null, 2));
		} catch (error) {
			return toolResult.error(
				`Failed to create task execution record: ${error instanceof Error ? error.message : String(error)}`,
			);
		}

		// Build the prompt for the LLM
		const promptInstructions = `You are now executing the task: ${task.name}

${task.description ? `<task_description>\n${task.description}\n</task_description>\n` : ""}
<task_arguments>
The following arguments have been provided for this task:

${Object.entries(args)
	.map(([key, value]) => `- ${key}: ${value}`)
	.join("\n")}
</task_arguments>

<task_instructions>
${task.instructions}
</task_instructions>

<execution_guidelines>
IMPORTANT: You must follow these guidelines when executing this task:

1. Create a Plan: Before starting, analyze the task instructions and create a detailed step-by-step plan using the TodoWrite tool
2. Record the Plan: Use the TodoWrite tool to create todo items for each major step
3. Follow the Plan: Execute each step in order, marking each step as "in_progress" before starting it
4. Update Status Immediately: Mark each step as "completed" immediately after finishing it - do not batch completions
5. Complete the Entire Task: Try to finish the entire task in one go - only stop if you need required information from the user
6. Error Handling: If you encounter an error, attempt to fix it at least once before stopping
7. Status Updates: Ensure exactly ONE todo is "in_progress" at any time (not less, not more)
8. Report Final Status: When you finish the task (or cannot continue), use the TaskStatus tool to report:
   - "succeeded" if the task completed successfully
   - "failed" if you encountered an error that could not be resolved after attempting to fix it
   - "needs_input" if you need information from the user to continue

Begin executing the task now. Start by creating your plan with the TodoWrite tool.
</execution_guidelines>`;

		return toolResult.text(promptInstructions);
	},
});
