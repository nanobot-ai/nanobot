import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";

const schema = z.object({
	file_path: z
		.string()
		.describe(
			"The absolute path to the file to write (must be absolute, not relative)",
		),
	content: z.string().describe("The content to write to the file"),
});

export default createTool({
	title: "Write",
	description:
		"Writes a file to the local filesystem.\n\nUsage:\n- This tool will overwrite the existing file if there is one at the provided path.\n- If this is an existing file, you MUST use the Read tool first to read the file's contents. This tool will fail if you did not read the file first.\n- ALWAYS prefer editing existing files in the codebase. NEVER write new files unless explicitly required.\n- NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.\n- Only use emojis if the user explicitly requests it. Avoid writing emojis to files unless asked.",
	messages: {
		invoking: "Writing file",
		invoked: "File written",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { file_path, content } = args;

		try {
			const client = await ensureConnected(ctx.workspaceId);
			await client.writeTextFile(file_path, content);
			return toolResult.text(`Successfully wrote to file: ${file_path}`);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error writing file: ${errorMessage}`);
		}
	},
});
