import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	path: z.string().describe("The absolute file path"),
	content: z.string().describe("The text content to write to the file"),
	encoding: z
		.enum(["utf-8", "base64"])
		.optional()
		.default("utf-8")
		.describe("The encoding to use for the file content"),
});

export default createTool({
	title: "Write File",
	description: "Write or update a text file in the filesystem",
	messages: {
		invoking: "Writing file",
		invoked: "File written",
	},
	inputSchema: schema,
	async handler({ sessionId, path, content, encoding }) {
		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Write the file (creates it if it doesn't exist)
			await sandbox.writeFile(path, content, { encoding });

			return toolResult.text(`Successfully wrote to file: ${path}`);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error writing file: ${errorMessage}`);
		}
	},
});
