import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	terminalId: z.string().describe("The terminal ID to get output from"),
});

const outputSchema = z.object({
	output: z.string().describe("The terminal output"),
	truncated: z
		.boolean()
		.describe("Whether output was truncated due to byte limit"),
	exitStatus: z
		.object({
			exitCode: z.number().describe("Exit code of the command").nullable(),
			signal: z
				.string()
				.describe("Signal that terminated the command")
				.nullable(),
		})
		.optional(),
});

export default createTool({
	title: "Get Terminal Output",
	description:
		"Retrieve current output from a terminal without waiting for completion",
	messages: {
		invoking: "Getting terminal output",
		invoked: "Terminal output retrieved",
	},
	inputSchema: schema,
	outputSchema,
	async handler(args) {
		const { sessionId, terminalId } = args;

		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Get the process output
			const result = await sandbox.output(terminalId);

			let responseText = `Terminal: ${terminalId}\n`;
			responseText += `Output (${result.truncated ? "truncated" : "complete"}):\n`;
			responseText += result.output;

			const exitStatus = {
				exitCode: result.exitCode,
				signal: result.signal ?? null,
			};

			responseText += "\n\nExit Status:\n";
			responseText += `Exit Code: ${exitStatus.exitCode}\n`;
			if (exitStatus.signal) {
				responseText += `Signal: ${exitStatus.signal}\n`;
			}

			return toolResult.structured(responseText, {
				output: result.output,
				truncated: result.truncated,
				exitStatus,
			});
		} catch (error) {
			return toolResult.error(
				`Failed to get terminal output: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	},
});
