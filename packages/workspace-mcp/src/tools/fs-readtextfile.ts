import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	path: z.string().describe("The absolute file path"),
	line: z
		.number()
		.describe("The line number to start reading from (1-based)")
		.optional(),
	limit: z.number().describe("The maximum number of lines to read").optional(),
	binary: z
		.boolean()
		.describe("Read raw content as a data URI instead of text")
		.optional(),
});

const outputSchema = z.object({
	content: z.string().describe("The file content"),
	truncated: z.boolean().describe("Whether content was truncated"),
});

export default createTool({
	title: "Read File",
	description: "A read a file from the filesystem",
	messages: {
		invoking: "Reading file",
		invoked: "File read",
	},
	inputSchema: schema,
	outputSchema,
	async handler(args) {
		const { sessionId, path, line, limit, binary } = args;

		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Use base64 encoding for binary mode
			const encoding = binary ? "base64" : "utf-8";

			const result = await sandbox.readFile(path, {
				encoding,
				offset: line,
				limit,
			});

			if (result === null) {
				return toolResult.error(`Error: File not found at path: ${path}`, {
					fileNotFound: true,
				});
			}

			// Check if content was truncated
			const lines = result.content.split("\n");
			const truncated = limit !== undefined && lines.length >= limit;

			return toolResult.structured(result.content, {
				content: result.content,
				truncated,
			});
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error reading file: ${errorMessage}`);
		}
	},
});
