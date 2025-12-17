import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { readTextFile } from "../lib/workspace-client.js";

const schema = z.object({
	file_path: z.string().describe("The absolute path to the file to read"),
	offset: z
		.number()
		.describe(
			"The line number to start reading from. Only provide if the file is too large to read at once",
		)
		.optional(),
	limit: z
		.number()
		.describe(
			"The number of lines to read. Only provide if the file is too large to read at once.",
		)
		.optional(),
});

export default createTool({
	title: "Read",
	description:
		"Reads a file from the local filesystem. You can access any file directly by using this tool.\nAssume this tool is able to read all files on the machine. If the User provides a path to a file assume that path is valid. It is okay to read a file that does not exist; an error will be returned.\n\nUsage:\n- The file_path parameter must be an absolute path, not a relative path\n- By default, it reads up to 2000 lines starting from the beginning of the file\n- You can optionally specify a line offset and limit (especially handy for long files), but it's recommended to read the whole file by not providing these parameters\n- Any lines longer than 2000 characters will be truncated\n- Results are returned using cat -n format, with line numbers starting at 1\n- This tool allows Claude Code to read images (eg PNG, JPG, etc). When reading an image file the contents are presented visually as Claude Code is a multimodal LLM.\n- This tool can read PDF files (.pdf). PDFs are processed page by page, extracting both text and visual content for analysis.\n- This tool can read Jupyter notebooks (.ipynb files) and returns all cells with their outputs, combining code, text, and visualizations.\n- This tool can only read files, not directories. To read a directory, use an ls command via the Bash tool.\n- You can call multiple tools in a single response. It is always better to speculatively read multiple potentially useful files in parallel.\n- You will regularly be asked to read screenshots. If the user provides a path to a screenshot, ALWAYS use this tool to view the file at the path. This tool will work with all temporary file paths.\n- If you read a file that exists but has empty contents you will receive a system reminder warning in place of file contents.",
	messages: {
		invoking: "Reading file",
		invoked: "File read",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { file_path, offset, limit } = args;

		try {
			const content = await readTextFile(ctx.workspaceId, file_path, {
				line: offset,
				limit: limit,
				convert: true,
			});
			return toolResult.text(content);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error reading file: ${errorMessage}`);
		}
	},
});
