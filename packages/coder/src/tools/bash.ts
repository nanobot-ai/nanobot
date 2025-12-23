import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import {
	executeBashCommand,
	generateShellId,
	registerBackgroundShell,
} from "../lib/shell.js";

const schema = z.object({
	command: z.string().describe("The command to execute"),
	timeout: z
		.number()
		.describe("Optional timeout in milliseconds (max 600000)")
		.optional(),
	description: z
		.string()
		.describe(
			"Clear, concise description of what this command does in 5-10 words, in active voice. Examples:\nInput: ls\nOutput: List files in current directory\n\nInput: git status\nOutput: Show working tree status\n\nInput: npm install\nOutput: Install package dependencies\n\nInput: mkdir foo\nOutput: Create directory 'foo'",
		)
		.optional(),
	run_in_background: z
		.boolean()
		.describe(
			"Set to true to run this command in the background. Use BashOutput to read the output later.",
		)
		.optional()
		.default(false),
});

export default createTool({
	title: "Bash",
	description: () => {
		return (
			'Executes a given bash command in a persistent shell session with optional timeout, ensuring proper handling and security measures.\n\nIMPORTANT: This tool is for terminal operations like git, npm, docker, etc. DO NOT use it for file operations (reading, writing, editing, searching, finding files) - use the specialized tools for this instead.\n\nBefore executing the command, please follow these steps:\n\n1. Directory Verification:\n   - If the command will create new directories or files, first use `ls` to verify the parent directory exists and is the correct location\n   - For example, before running "mkdir foo/bar", first use `ls foo` to check that "foo" exists and is the intended parent directory\n\n2. Command Execution:\n   - Always quote file paths that contain spaces with double quotes (e.g., cd "path with spaces/file.txt")\n   - Examples of proper quoting:\n     - cd "/Users/name/My Documents" (correct)\n     - cd /Users/name/My Documents (incorrect - will fail)\n     - python "/path/with spaces/script.py" (correct)\n     - python /path/with spaces/script.py (incorrect - will fail)\n   - After ensuring proper quoting, execute the command.\n   - Capture the output of the command.\n\nUsage notes:\n  - The command argument is required.\n  - You can specify an optional timeout in milliseconds (up to 600000ms / 10 minutes). If not specified, commands will timeout after 120000ms (2 minutes).\n  - It is very helpful if you write a clear, concise description of what this command does in 5-10 words.\n  - If the output exceeds 30000 characters, output will be truncated before being returned to you.\n  - You can use the `run_in_background` parameter to run the command in the background, which allows you to continue working while the command runs. You can monitor the output using the Bash tool as it becomes available. You do not need to use \'&\' at the end of the command when using this parameter.\n  \n  - Avoid using Bash with the `find`, `grep`, `cat`, `head`, `tail`, `sed`, `awk`, or `echo` commands, unless explicitly instructed or when these commands are truly necessary for the task. Instead, always prefer using the dedicated tools for these commands:\n    - File search: Use Glob (NOT find or ls)\n    - Content search: Use Grep (NOT grep or rg)\n    - Read files: Use Read (NOT cat/head/tail)\n    - Edit files: Use Edit (NOT sed/awk)\n    - Write files: Use Write (NOT echo >/cat <<EOF)\n    - Communication: Output text directly (NOT echo/printf)\n  - When issuing multiple commands:\n    - If the commands are independent and can run in parallel, make multiple Bash tool calls in a single message. For example, if you need to run "git status" and "git diff", send a single message with two Bash tool calls in parallel.\n    - If the commands depend on each other and must run sequentially, use a single Bash call with \'&&\' to chain them together (e.g., `git add . && git commit -m "message" && git push`). For instance, if one operation must complete before another starts (like mkdir before cp, Write before Bash for git operations, or git add before git commit), run these operations sequentially instead.\n    - Use \';\' only when you need to run commands sequentially but don\'t care if earlier commands fail\n    - DO NOT use newlines to separate commands (newlines are ok in quoted strings)\n  - Try to maintain your current working directory throughout the session by using absolute paths and avoiding usage of `cd`. You may use `cd` if the User explicitly requests it.' +
			`\n\nThe current working directory is: /workspace`
		);
	},
	messages: {
		invoking: "Executing command",
		invoked: "Command executed",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { command, timeout, run_in_background } = args;

		try {
			const client = await ensureConnected(ctx.workspaceId);

			if (run_in_background) {
				// Start command in background
				const { terminalId } = await client.terminalCreate("bash", {
					args: ["-c", command],
				});

				const shellId = generateShellId();
				registerBackgroundShell(shellId, terminalId, command);

				return toolResult.text(
					`Command started in background.\nShell ID: ${shellId}\nUse BashOutput tool with this shell ID to check output.`,
				);
			}

			// Execute synchronously
			const result = await executeBashCommand(client, command, { timeout });

			if (result.timedOut) {
				return toolResult.error(result.output);
			}

			if (result.exitCode !== 0) {
				return toolResult.text(
					`Exit code ${result.exitCode}\n${result.output}`,
				);
			}

			return toolResult.text(
				result.output || "Command completed successfully with no output.",
			);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error executing command: ${errorMessage}`);
		}
	},
});
