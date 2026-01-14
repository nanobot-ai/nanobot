import * as path from "node:path";
import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { getTask, getTaskStep, getTasksDescription } from "../lib/task.ts";

const schema = z.object({
	taskName: z.string().describe("The task name to execute"),
	filename: z
		.string()
		.optional()
		.describe("Optional filename for the task step (e.g., 'step1.md')"),
	arguments: z
		.array(z.record(z.string()))
		.optional()
		.describe("Optional arguments for the task"),
});

export default createTool({
	title: "Execute Step Task",
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
		invoking: "Executing task step",
		invoked: "Task step execution started",
	},
	inputSchema: schema,
	async handler({ taskName, filename, arguments: taskArgs }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const task = filename
			? await getTaskStep(client, taskName, filename)
			: await getTask(client, taskName);

		if (!task.name) {
			return toolResult.error(`task step not found: ${taskName}`);
		}

		// Collect input arguments
		const args: Record<string, string> = {};
		for (const input of task.inputs || []) {
			const arg = taskArgs?.find((arg) => arg.name === input.name);
			if (arg) {
				args[input.name] = arg.value;
			} else if (input.default) {
				args[input.name] = input.default;
			} else {
				return toolResult.error(
					`Missing argument: ${input.name} for task step ${taskName}. The argument is described as: ${input.description}`,
				);
			}
		}

		// Create task execution record
		const sessionId = ctx.sessionId || "default";
		const taskDir = path.join(".nanobot", "status", sessionId);
		const taskFile = path.join(taskDir, "task.json");

		// Ensure the directory exists by trying to create it (writeTextFile will handle parent dirs)
		const taskRecord = {
			taskName: taskName,
			arguments: args,
			startedAt: new Date().toISOString(),
			next: task.next,
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
╔══════════════════════════════════════════════════════════════════════════════╗
║ CRITICAL: TASK CHAINING BEHAVIOR - READ THIS FIRST                          ║
╚══════════════════════════════════════════════════════════════════════════════╝

This task may be part of a CHAIN of tasks. Your job is NOT done when you finish
the instructions for THIS task step. You MUST follow this completion sequence:

  STEP 1: Call TaskStepStatus with the outcome (succeeded/failed/needs_input)

  STEP 2: READ the TaskStepStatus response - it will tell you what to do next:
          • If it says there's a NEXT task → complete this step and then immediately call ExecuteStepTask
          • If it says there's NO next task → you are done, stop here

  STEP 3: If there was a next task, you must execute that step next.

DO NOT STOP after calling TaskStepStatus! You must check its response and
continue to the next task if one exists. The chain continues until there are
no more tasks.

╔══════════════════════════════════════════════════════════════════════════════╗
║ EXECUTION GUIDELINES FOR THIS TASK STEP                                     ║
╚══════════════════════════════════════════════════════════════════════════════╝

1. Create a Plan: Before starting, analyze the task instructions and create a detailed step-by-step plan using the TodoWrite tool
2. Record the Plan: Use the TodoWrite tool to create todo items for each major step
3. Follow the Plan: Execute each step in order, marking each step as "in_progress" before starting it
4. Explain your steps as you are taking them. Every tool call and every step should have some explanation as to why you are doing it. Be clear and concise in your explanations.
5. Update Status Immediately: Mark each step as "completed" immediately after finishing it - do not batch completions
6. Complete the Entire Task: Try to finish the entire task in one go - only stop if you need required information from the user
7. Error Handling: If you encounter an error, attempt to fix it at least once before stopping
8. Status Updates: Ensure exactly ONE todo is "in_progress" at any time (not less, not more)

When you finish this task step, proceed to the TASK CHAINING sequence above. You must ensure that you have completed everything you were asked to complete in this
step before continuing with the TASK CHAINING sequence.

Begin executing the task now. Start by creating your plan with the TodoWrite tool.
</execution_guidelines>`;

		return toolResult.text(promptInstructions);
	},
});
