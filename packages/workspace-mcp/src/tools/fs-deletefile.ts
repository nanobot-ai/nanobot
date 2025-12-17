import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	path: z.string().describe("The absolute file path to delete"),
});

export default createTool({
	title: "Delete File",
	description: "Delete a file from the filesystem",
	messages: {
		invoking: "Deleting file",
		invoked: "File deleted",
	},
	inputSchema: schema,
	async handler({ sessionId, path }) {
		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Delete the file
			await sandbox.deleteFile(path);

			return toolResult.text(`Successfully deleted file: ${path}`);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error deleting file: ${errorMessage}`);
		}
	},
});
