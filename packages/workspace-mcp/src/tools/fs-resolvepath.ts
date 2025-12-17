import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	path: z
		.string()
		.describe("The path to resolve (can be relative or absolute)"),
});

const outputSchema = z.object({
	path: z.string().describe("The resolved absolute path"),
});

export default createTool({
	title: "Resolve Path",
	description:
		"Resolve a path to an absolute path, respecting the session's base directory if set",
	messages: {
		invoking: "Resolving path",
		invoked: "Path resolved",
	},
	inputSchema: schema,
	outputSchema,
	async handler(args) {
		const { sessionId, path } = args;

		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Resolve the path
			const resolvedPath = sandbox.resolvePath(path);

			return toolResult.structured(resolvedPath, {
				path: resolvedPath,
			});
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error resolving path: ${errorMessage}`);
		}
	},
});
