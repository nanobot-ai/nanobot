import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { getTask, getTasks } from "../lib/task.ts";

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
		const tasks = await getTasks(client);
		const available = tasks
			.map((s) => `${s.name}: ${s.description}`)
			.join("\n");
		return `Execute a task within the main conversation

<available_tasks>
${available}
</available_tasks>

The output of this tool will contain the exact instructions to be followed to complete the task.
`;
	},
	messages: {
		invoking: "Loading task",
		invoked: "Task loaded",
	},
	inputSchema: schema,
	async handler({ taskName, arguments: taskArgs }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const task = await getTask(client, taskName);
		if (!task.name) {
			return toolResult.error(`task not found: ${taskName}`);
		}

		// Format arguments if provided
		let argsSection = "";
		if (taskArgs && Object.keys(taskArgs).length > 0) {
			argsSection = `

## Task Arguments

The following arguments have been provided for this task:

${Object.entries(taskArgs)
	.map(([key, value]) => `- **${key}**: ${value}`)
	.join("\n")}

Use these arguments as context when completing the task.
`;
		}

		// Build the comprehensive prompt
		const prompt = `# Task: ${task.name}

You have been assigned a task to complete. Follow these instructions carefully.

## Task Instructions

${task.instructions}
${argsSection}

## CRITICAL: Task Execution Requirements

You MUST follow this process to complete the task:

### 1. Create a Detailed Plan

**IMMEDIATELY** after reading these instructions, you must use the TodoWrite tool to create a comprehensive, step-by-step plan. Your plan should:

- Break down the task into specific, actionable items
- Include ALL steps needed to complete the task
- Use clear, descriptive task names in two forms:
  - **content**: Imperative form (e.g., "Implement user authentication")
  - **activeForm**: Present continuous form (e.g., "Implementing user authentication")
- Set initial status for all tasks as "pending"

### 2. Execute Tasks Sequentially

- Mark EXACTLY ONE task as "in_progress" before starting work on it
- Complete the task fully
- Mark it as "completed" IMMEDIATELY after finishing
- Then move to the next task

### 3. Task Completion Standards

**ONLY** mark a task as completed when you have FULLY accomplished it:

- ✅ All code is written and working
- ✅ Tests pass (if applicable)
- ✅ No errors or blockers remain
- ✅ Implementation is complete, not partial

**NEVER** mark a task as completed if:
- ❌ Tests are failing
- ❌ Implementation is partial or incomplete
- ❌ You encountered unresolved errors
- ❌ Files or dependencies are missing

### 4. Update Progress Continuously

- Use TodoWrite to update task status as you work
- Keep exactly ONE task "in_progress" at any time
- Mark tasks completed immediately after finishing each one
- If you discover new work, add new tasks to the plan

### 5. Stay Focused

- Complete ALL tasks in your plan before considering the work done
- Do not skip tasks or mark them completed prematurely
- If blocked, create a new task describing what needs resolution
- Be thorough and meticulous

## Working Directory

The base directory for this task is: ${task.baseDir}

All file paths should be relative to this directory unless otherwise specified.

---

**BEGIN YOUR WORK NOW:**

1. First, use the TodoWrite tool to create your complete plan
2. Then, start executing tasks one by one
3. Mark each as completed before moving to the next
4. Ensure all requirements are met before finishing`;

		return toolResult.text(prompt);
	},
});
