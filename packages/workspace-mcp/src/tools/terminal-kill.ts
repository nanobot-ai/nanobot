import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	terminalId: z.string().describe("The terminal ID to kill"),
	signal: z.string().describe("Signal to send (default: SIGTERM)").optional(),
});

export default createTool({
	title: "Kill Terminal",
	description: "Terminate a running terminal command",
	messages: {
		invoking: "Killing terminal",
		invoked: "Terminal killed",
	},
	inputSchema: schema,
	async handler(args) {
		const { sessionId, terminalId, signal } = args;

		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Kill the process
			await sandbox.kill(terminalId, signal);

			return toolResult.text(
				`Terminal ${terminalId} has been killed with signal ${signal || "SIGTERM"}`,
			);
		} catch (error) {
			return toolResult.error(
				`Failed to kill terminal: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	},
});
