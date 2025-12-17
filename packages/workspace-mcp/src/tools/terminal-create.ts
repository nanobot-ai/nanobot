import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";
import { sessionSandboxManager } from "../lib/shared.js";

const schema = z.object({
	sessionId: z.string().min(1).describe("The session ID"),
	command: z.string().describe("The executable command to run"),
	args: z.array(z.string()).describe("Command-line arguments array").optional(),
	env: z
		.record(z.string(), z.string())
		.describe("Environment variable pairs")
		.optional(),
	cwd: z.string().describe("Working directory path").optional(),
	outputByteLimit: z
		.number()
		.describe("Maximum retained output size in bytes")
		.optional(),
});

const outputSchema = z.object({
	terminalId: z
		.string()
		.describe("The unique identifier for the created terminal"),
});

export default createTool({
	title: "Create Terminal",
	description: "Create a new terminal session and start running a command",
	messages: {
		invoking: "Creating terminal",
		invoked: "Terminal created",
	},
	inputSchema: schema,
	outputSchema,
	async handler(args) {
		const {
			sessionId,
			command,
			args: cmdArgs,
			env,
			cwd,
			outputByteLimit,
		} = args;

		try {
			const sandbox = await sessionSandboxManager.getSandbox(sessionId);

			// Execute the command (returns process ID)
			const terminalId = await sandbox.execute(command, cmdArgs ?? [], {
				cwd,
				env,
				outputByteLimit,
			});

			return toolResult.structured(`Terminal created with ID: ${terminalId}`, {
				terminalId,
			});
		} catch (error) {
			return toolResult.error(
				`Failed to create terminal: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	},
});
