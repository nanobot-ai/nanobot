import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	terminalId: z.string().describe("The terminal ID to wait for"),
});

const outputSchema = z.object({
	exitCode: z.number().describe("Exit code of the command"),
	signal: z.string().describe("Signal that terminated the command").nullable(),
});

export default createTool({
	title: "Wait for Terminal Exit",
	description: "Block until the terminal command finishes execution",
	messages: {
		invoking: "Waiting for terminal to exit",
		invoked: "Terminal exited",
	},
	inputSchema: schema,
	outputSchema,
	async handler(args) {
		const { sessionId, terminalId } = args;

		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Wait for the process to complete
			const result = await sandbox.wait(terminalId);

			let responseText = `Terminal ${terminalId} has exited\n`;
			responseText += `Exit Code: ${result.exitCode}\n`;
			if (result.signal) {
				responseText += `Signal: ${result.signal}\n`;
			}

			return toolResult.structured(responseText, {
				exitCode: result.exitCode,
				signal: result.signal ?? null,
			});
		} catch (error) {
			return toolResult.error(
				`Failed to wait for terminal exit: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	},
});
