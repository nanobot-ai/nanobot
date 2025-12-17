import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID to delete"),
});

export default createTool({
	title: "Delete Session",
	description: "Delete a session and free all its resources",
	messages: {
		invoking: "Deleting session",
		invoked: "Session deleted",
	},
	inputSchema: schema,
	async handler({ sessionId }) {
		try {
			await sessionSandboxManager.deleteSandbox(sessionId);

			return toolResult.text(`Session ${sessionId} has been deleted`);
		} catch (error) {
			return toolResult.error(
				`Failed to delete session: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	},
});
