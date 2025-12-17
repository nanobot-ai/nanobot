import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const inputSchema = z.object({
	uri: z.string().describe("The workspace URI"),
});

const outputSchema = z.object({
	workspaceId: z.string().describe("The workspace ID"),
});

export default createTool({
	title: "Create Session",
	description:
		"Create a new workspace session with an isolated sandbox environment",
	messages: {
		invoking: "Creating session",
		invoked: "Session created",
	},
	inputSchema,
	outputSchema,
	async handler({ uri }) {
		try {
			// Create the sandbox for this session
			const sb = await sessionSandboxManager.getSandbox(uri);
			return toolResult.structured(`Session created with ID: ${sb.id}`, {
				workspaceId: sb.id,
			});
		} catch (error) {
			return toolResult.error(
				`Failed to create session: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	},
});
