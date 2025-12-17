import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	path: z.string().describe("The absolute directory path"),
	cursor: z.string().describe("Cursor for pagination").optional(),
	recursive: z.boolean().describe("List directories recursively").optional(),
	limit: z.number().describe("Maximum number of entries to return").optional(),
});

const entrySchema = z.object({
	name: z.string(),
	isFile: z.boolean(),
	isDirectory: z.boolean(),
	size: z.number(),
	skipped: z.boolean().optional(),
});

const outputSchema = z.object({
	entries: z.array(entrySchema),
	cursor: z.string().optional(),
});

export default createTool({
	title: "List Directory",
	description: "List the contents of a directory",
	messages: {
		invoking: "Listing directory",
		invoked: "Directory listed",
	},
	inputSchema: schema,
	outputSchema,
	async handler(args) {
		const { sessionId, path, cursor, recursive, limit } = args;

		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Read the directory contents with options
			const result = await sandbox.readdir(path, {
				cursor,
				recursive,
				limit,
			});

			// Handle empty directory (sandbox returns {entries: []} for not found)
			if (result.entries.length === 0 && !cursor) {
				return toolResult.structured(
					`Directory ${path} is empty or not found`,
					{
						entries: [],
						cursor: undefined,
					},
				);
			}

			const formattedEntries = result.entries
				.map((entry) => entry.name + (entry.isDirectory ? "/" : ""))
				.join("\n");

			let message = `Listed ${result.entries.length} entries in ${path}`;
			if (recursive) {
				message += " (recursive)";
			}
			if (result.cursor) {
				message += "\n(More entries available - use cursor for next page)";
			}
			message += `\n\n${formattedEntries}`;

			return toolResult.structured(message, {
				entries: result.entries,
				cursor: result.cursor,
			});
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error listing directory: ${errorMessage}`);
		}
	},
});
