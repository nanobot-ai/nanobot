import * as path from "node:path";
import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";

const schema = z.object({
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
	title: "Task Status",
	description:
		"Report the final status of a task execution. Use this tool after completing (or attempting to complete) a task to indicate whether it succeeded, failed, or needs user input to continue.",
	messages: {
		invoking: "Reporting task status",
		invoked: "Task status reported",
	},
	inputSchema: schema,
	visibility: ["app"],
	async enabled(ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const sessionId = ctx.sessionId || "default";
		const taskFile = path.join(".nanobot", "status", sessionId, "task.json");

		try {
			const existingContent = await client.readTextFile(taskFile, {
				ignoreNotFound: true,
			});

			// If no task.json found, tool is disabled
			if (!existingContent) {
				return false;
			}

			// Parse the task record
			const taskRecord = JSON.parse(existingContent) as Record<string, unknown>;

			// If task already has a terminal status, tool is disabled
			if (
				taskRecord.status === "succeeded" ||
				taskRecord.status === "failed" ||
				taskRecord.status === "needs_input"
			) {
				return false;
			}

			// Task exists and is not in terminal state, tool is enabled
			return true;
		} catch (_error) {
			// If there's an error reading/parsing, disable the tool
			return false;
		}
	},
	async handler({ status, explanation }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const sessionId = ctx.sessionId || "default";
		const taskFile = path.join(".nanobot", "status", sessionId, "task.json");

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

		return toolResult.structured(
			`Task status reported: ${statusEmoji[status]} ${status}`,
			{
				status,
				explanation,
				completedAt: taskRecord.completedAt,
			},
		);
	},
});
