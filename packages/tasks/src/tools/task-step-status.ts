import * as path from "node:path";
import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";

const schema = z.object({
	taskName: z.string().describe("The task name"),
	status: z
		.enum(["succeeded", "failed", "needs_input"])
		.describe(
			"The status of the task execution: 'succeeded' if completed successfully, 'failed' if an error occurred that could not be resolved, 'needs_input' if user input is required to continue",
		),
	explanation: z
		.string()
		.describe("A clear explanation of why the task has this status"),
});

export default createTool({
	title: "Task Step Status",
	description:
		"Report the final status of a task execution. Use this tool after completing (or attempting to complete) a task to indicate whether it succeeded, failed, or needs user input to continue.",
	messages: {
		invoking: "Reporting task status",
		invoked: "Task status reported",
	},
	inputSchema: schema,
	async handler({ taskName, status, explanation }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);

		const sessionId = ctx.sessionId || "default";
		const taskFile = path.join(".nanobot", sessionId, "status", "task.json");

		// Read the existing task record
		let taskRecord: Record<string, unknown>;
		try {
			const existingContent = await client.readTextFile(taskFile, {
				ignoreNotFound: true,
			});

			if (!existingContent) {
				return toolResult.error(
					"No task execution record found. You must execute a task first using the ExecuteTask tool before reporting status.",
				);
			}

			taskRecord = JSON.parse(existingContent);
		} catch (error) {
			return toolResult.error(
				`Failed to read task execution record: ${error instanceof Error ? error.message : String(error)}`,
			);
		}

		// Update the task record with status information
		taskRecord.status = status;
		taskRecord.explanation = explanation;
		taskRecord.completedAt = new Date().toISOString();

		try {
			await client.writeTextFile(taskFile, JSON.stringify(taskRecord, null, 2));
			await client.deleteFile(path.join(".nanobot", "mcp.json"));
		} catch (error) {
			return toolResult.error(
				`Failed to update task execution record: ${error instanceof Error ? error.message : String(error)}`,
			);
		}

		const statusEmoji = {
			succeeded: "✓",
			failed: "✗",
			needs_input: "?",
		};

		let nextTaskText =
			"\n\nUse all the information from ALL the task step executions to give a summary of the steps you executed. Be concise and informative.";
		if (taskRecord.next) {
			nextTaskText = `\n\nYou must now use the ExecuteTaskStep tool to execute the next task step: taskName: ${taskName}, filename: ${taskRecord.next}`;
		}

		return toolResult.structured(
			`Task status reported: ${statusEmoji[status]} ${status}${nextTaskText}`,
			{
				status,
				explanation,
				completedAt: taskRecord.completedAt,
			},
		);
	},
});
