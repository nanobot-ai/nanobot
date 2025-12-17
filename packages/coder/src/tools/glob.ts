import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { executeBashCommand } from "../lib/shell.js";

const schema = z.object({
	pattern: z.string().describe("The glob pattern to match files against"),
	path: z
		.string()
		.describe(
			'The directory to search in. If not specified, the current working directory will be used. IMPORTANT: Omit this field to use the default directory. DO NOT enter "undefined" or "null" - simply omit it for the default behavior. Must be a valid directory path if provided.',
		)
		.optional(),
});

export default createTool({
	title: "Glob",
	description:
		'- Fast file pattern matching tool that works with any codebase size\n- Supports glob patterns like "**/*.js" or "src/**/*.ts"\n- Returns matching file paths sorted by modification time\n- Use this tool when you need to find files by name patterns\n- When you are doing an open ended search that may require multiple rounds of globbing and grepping, use the Agent tool instead\n- You can call multiple tools in a single response. It is always better to speculatively perform multiple searches in parallel if they are potentially useful.',
	messages: {
		invoking: "Searching files",
		invoked: "File search complete",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { pattern, path } = args;

		try {
			const client = await ensureConnected(ctx.workspaceId);
			const searchPath = path || ".";

			// Use ripgrep's --files mode with --glob pattern
			// Note: --files doesn't support JSON output, so we parse plain text
			const rgArgs: string[] = ["rg", "--files"];

			// Add glob pattern
			rgArgs.push(`--glob ${JSON.stringify(pattern)}`);

			// Add path if specified
			if (path) {
				rgArgs.push(JSON.stringify(searchPath));
			}

			// Add sort by modification time (most recent first)
			const command = `${rgArgs.join(" ")} | xargs -r ls -t 2>/dev/null || true`;

			const result = await executeBashCommand(client, command);

			if (!result.output.trim()) {
				return toolResult.text("No files found matching pattern");
			}

			const files = result.output
				.trim()
				.split("\n")
				.filter((line) => line.length > 0);

			if (files.length === 0) {
				return toolResult.text("No files found matching pattern");
			}

			return toolResult.text(files.join("\n"));
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error searching files: ${errorMessage}`);
		}
	},
});
