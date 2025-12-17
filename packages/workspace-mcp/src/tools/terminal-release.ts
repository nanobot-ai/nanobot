import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	terminalId: z.string().describe("The terminal ID to release"),
});

export default createTool({
	title: "Release Terminal",
	description: "Destroy terminal session and free resources",
	messages: {
		invoking: "Releasing terminal",
		invoked: "Terminal released",
	},
	inputSchema: schema,
	async handler(args) {
		const { sessionId, terminalId } = args;

		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Release the process
			await sandbox.release(terminalId);

			return toolResult.text(`Terminal ${terminalId} has been released`);
		} catch (error) {
			return toolResult.error(
				`Failed to release terminal: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	},
});
