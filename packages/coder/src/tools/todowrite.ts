import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";

const todoItemSchema = z.object({
	content: z.string().min(1),
	status: z.enum(["pending", "in_progress", "completed"]),
	activeForm: z.string().min(1),
});

export const inputSchema = z.object({
	todos: z.array(todoItemSchema).describe("The updated todo list"),
});

export const description =
	'Use this tool to create and manage a structured task list for your current coding session. This helps you track progress, organize complex tasks, and demonstrate thoroughness to the user.\nIt also helps the user understand the progress of the task and overall progress of their requests.\n\n## When to Use This Tool\nUse this tool proactively in these scenarios:\n\n1. Complex multi-step tasks - When a task requires 3 or more distinct steps or actions\n2. Non-trivial and complex tasks - Tasks that require careful planning or multiple operations\n3. User explicitly requests todo list - When the user directly asks you to use the todo list\n4. User provides multiple tasks - When users provide a list of things to be done (numbered or comma-separated)\n5. After receiving new instructions - Immediately capture user requirements as todos\n6. When you start working on a task - Mark it as in_progress BEFORE beginning work. Ideally you should only have one todo as in_progress at a time\n7. After completing a task - Mark it as completed and add any new follow-up tasks discovered during implementation\n\n## When NOT to Use This Tool\n\nSkip using this tool when:\n1. There is only a single, straightforward task\n2. The task is trivial and tracking it provides no organizational benefit\n3. The task can be completed in less than 3 trivial steps\n4. The task is purely conversational or informational\n\nNOTE that you should not use this tool if there is only one trivial task to do. In this case you are better off just doing the task directly.\n\n## Task States and Management\n\n1. **Task States**: Use these states to track progress:\n   - pending: Task not yet started\n   - in_progress: Currently working on (limit to ONE task at a time)\n   - completed: Task finished successfully\n\n   **IMPORTANT**: Task descriptions must have two forms:\n   - content: The imperative form describing what needs to be done (e.g., "Run tests", "Build the project")\n   - activeForm: The present continuous form shown during execution (e.g., "Running tests", "Building the project")\n\n2. **Task Management**:\n   - Update task status in real-time as you work\n   - Mark tasks complete IMMEDIATELY after finishing (don\'t batch completions)\n   - Exactly ONE task must be in_progress at any time (not less, not more)\n   - Complete current tasks before starting new ones\n   - Remove tasks that are no longer relevant from the list entirely\n\n3. **Task Completion Requirements**:\n   - ONLY mark a task as completed when you have FULLY accomplished it\n   - If you encounter errors, blockers, or cannot finish, keep the task as in_progress\n   - When blocked, create a new task describing what needs to be resolved\n   - Never mark a task as completed if:\n     - Tests are failing\n     - Implementation is partial\n     - You encountered unresolved errors\n     - You couldn\'t find necessary files or dependencies\n\n4. **Task Breakdown**:\n   - Create specific, actionable items\n   - Break complex tasks into smaller, manageable steps\n   - Use clear, descriptive task names\n   - Always provide both forms:\n     - content: "Fix authentication bug"\n     - activeForm: "Fixing authentication bug"\n\nWhen in doubt, use this tool. Being proactive with task management demonstrates attentiveness and ensures you complete all requirements successfully.\n';

export default createTool({
	title: "TodoWrite",
	description,
	messages: {
		invoking: "Updating todos",
		invoked: "Todos updated",
	},
	inputSchema,
	async handler(args, ctx) {
		const { todos } = args;

		try {
			// Validate that there's exactly one in_progress task if any tasks are in_progress
			const inProgressCount = todos.filter(
				(t) => t.status === "in_progress",
			).length;
			if (inProgressCount > 1) {
				return toolResult.error(
					"Error: Only one task can be in_progress at a time",
				);
			}

			// Get workspace client
			const client = await ensureConnected(ctx.workspaceId);

			// Write the todo list to ${AGENT_HOME}/.todo.json
			const todoFilePath = ".nanobot/status/todo.json";
			const todoContent = JSON.stringify(todos, null, 2);
			await client.writeTextFile(todoFilePath, todoContent);

			return toolResult.text(`Todo list updated:\n\n${todoContent}`);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Failed to write todo list: ${errorMessage}`);
		}
	},
});
