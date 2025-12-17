import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { getBackgroundShell, removeBackgroundShell } from "../lib/shell.js";

const schema = z.object({
	shell_id: z.string().describe("The ID of the background shell to kill"),
});

export default createTool({
	title: "KillShell",
	description:
		"\n- Kills a running background bash shell by its ID\n- Takes a shell_id parameter identifying the shell to kill\n- Returns a success or failure status \n- Use this tool when you need to terminate a long-running shell\n- Shell IDs can be found using the /bashes command\n",
	messages: {
		invoking: "Killing shell",
		invoked: "Shell killed",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { shell_id } = args;

		try {
			const shell = getBackgroundShell(shell_id);
			if (!shell) {
				return toolResult.error(
					`No background shell found with ID: ${shell_id}`,
				);
			}

			const client = await ensureConnected(ctx.workspaceId);
			await client.terminalKill(shell.terminalId);
			await client.terminalRelease(shell.terminalId);

			// Remove from registry
			removeBackgroundShell(shell_id);

			return toolResult.text(`Successfully killed shell: ${shell_id}`);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error killing shell: ${errorMessage}`);
		}
	},
});
