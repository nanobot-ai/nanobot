import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { getBackgroundShell } from "../lib/shell.js";

const schema = z.object({
	bash_id: z
		.string()
		.describe("The ID of the background shell to retrieve output from"),
	filter: z
		.string()
		.describe(
			"Optional regular expression to filter the output lines. Only lines matching this regex will be included in the result. Any lines that do not match will no longer be available to read.",
		)
		.optional(),
});

export default createTool({
	title: "BashOutput",
	description:
		"\n- Retrieves output from a running or completed background bash shell\n- Takes a shell_id parameter identifying the shell\n- Always returns only new output since the last check\n- Returns stdout and stderr output along with shell status\n- Supports optional regex filtering to show only lines matching a pattern\n- Use this tool when you need to monitor or check the output of a long-running shell\n- Shell IDs can be found using the /bashes command\n",
	messages: {
		invoking: "Getting shell output",
		invoked: "Shell output retrieved",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { bash_id, filter } = args;

		try {
			const shell = getBackgroundShell(bash_id);
			if (!shell) {
				return toolResult.error(
					`No background shell found with ID: ${bash_id}`,
				);
			}

			const client = await ensureConnected(ctx.workspaceId);
			const result = await client.terminalOutput(shell.terminalId);

			let output = result.output;

			// Apply regex filter if provided
			if (filter) {
				try {
					const regex = new RegExp(filter);
					const lines = output.split("\n");
					const filteredLines = lines.filter((line) => regex.test(line));
					output = filteredLines.join("\n");
				} catch (error) {
					return toolResult.error(
						`Invalid regex filter: ${error instanceof Error ? error.message : String(error)}`,
					);
				}
			}

			let statusMsg = "";
			if (result.exitStatus !== undefined) {
				statusMsg = `\n\nExit code: ${result.exitStatus.exitCode}`;
				statusMsg += `\nExit signal: ${result.exitStatus.signal}`;
			}

			if (result.truncated) {
				statusMsg += "\n\nOutput was truncated due to size limits";
			}

			return toolResult.structured(output + statusMsg, result);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error getting shell output: ${errorMessage}`);
		}
	},
});
