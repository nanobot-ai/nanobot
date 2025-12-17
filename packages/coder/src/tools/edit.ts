import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { performEdit } from "../lib/edit.js";

const schema = z.object({
	file_path: z.string().describe("The absolute path to the file to modify"),
	old_string: z.string().describe("The text to replace"),
	new_string: z
		.string()
		.describe(
			"The text to replace it with (must be different from old_string)",
		),
	replace_all: z
		.boolean()
		.describe("Replace all occurences of old_string (default false)")
		.optional()
		.default(false),
});

export default createTool({
	title: "Edit",
	description:
		"Performs exact string replacements in files. \n\nUsage:\n- You must use your `Read` tool at least once in the conversation before editing. This tool will error if you attempt an edit without reading the file. \n- When editing text from Read tool output, ensure you preserve the exact indentation (tabs/spaces) as it appears AFTER the line number prefix. The line number prefix format is: spaces + line number + tab. Everything after that tab is the actual file content to match. Never include any part of the line number prefix in the old_string or new_string.\n- ALWAYS prefer editing existing files in the codebase. NEVER write new files unless explicitly required.\n- Only use emojis if the user explicitly requests it. Avoid adding emojis to files unless asked.\n- The edit will FAIL if `old_string` is not unique in the file. Either provide a larger string with more surrounding context to make it unique or use `replace_all` to change every instance of `old_string`. \n- Use `replace_all` for replacing and renaming strings across the file. This parameter is useful if you want to rename a variable for instance.",
	messages: {
		invoking: "Editing file",
		invoked: "File edited",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { file_path, old_string, new_string, replace_all } = args;

		try {
			const client = await ensureConnected(ctx.workspaceId);

			// Read the current file content
			const content = await client.readTextFile(file_path);

			// Perform the edit
			const result = performEdit(
				content,
				old_string,
				new_string,
				replace_all ?? false,
			);

			if (!result.success) {
				return toolResult.error(`Edit failed: ${result.error}`);
			}

			// Write the new content back
			if (result.newContent) {
				await client.writeTextFile(file_path, result.newContent);
			}

			return toolResult.text(`Successfully edited file: ${file_path}`);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error editing file: ${errorMessage}`);
		}
	},
});
