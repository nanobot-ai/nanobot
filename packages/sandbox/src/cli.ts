#!/usr/bin/env node

/**
 * Sandbox CLI - Manage and interact with sandboxes
 */

import { homedir } from "node:os";
import { join } from "node:path";
import { createFsFileStorage } from "@remix-run/file-storage/fs";
import { Command } from "commander";
import { DockerDriver } from "./lib/docker.js";
import { Manager } from "./lib/manager.js";

// Get image from environment variable or use default
const SANDBOX_IMAGE = process.env.SANDBOX_IMAGE || "sandbox-test";

// Setup storage in user's home directory
const storageDir = join(homedir(), ".sandbox");
const storage = createFsFileStorage(storageDir);

// Function to create manager with specified image
function createManager(image: string) {
	const driver = new DockerDriver({
		image,
		workdir: "/workspace",
		localDataDir: join(storageDir, "data"),
	});

	return new Manager(storage, {
		drivers: { docker: driver },
		defaultDriver: "docker",
	});
}

// Helper to get manager from command context
function getManager(command?: Command): Manager {
	const image = command?.parent?.opts().image || SANDBOX_IMAGE;
	return createManager(image);
}

const program = new Command();

program
	.name("sb")
	.description("Sandbox management and execution CLI")
	.version("0.0.1")
	.option(
		"-i, --image <image>",
		"Docker image to use for sandboxes",
		SANDBOX_IMAGE,
	);

// Create sandbox
program
	.command("create")
	.description("Create a new sandbox")
	.option("-b, --base-uri <uri>", "Base URI or parent sandbox ID")
	.option("-m, --meta <json>", "Metadata as JSON string")
	.action(async (options, command) => {
		try {
			const manager = getManager(command);
			const opts: { baseUri?: string; meta?: Record<string, unknown> } = {};
			if (options.baseUri) opts.baseUri = options.baseUri;
			if (options.meta) opts.meta = JSON.parse(options.meta);

			const sandbox = await manager.createSandbox(opts);
			console.log(`Created sandbox: ${sandbox.id}`);
			console.log(`Metadata:`, await sandbox.getMeta());
		} catch (error) {
			console.error("Error creating sandbox:", error);
			process.exit(1);
		}
	});

// List sandboxes
program
	.command("list")
	.alias("ls")
	.description("List all sandboxes")
	.action(async (_options, command) => {
		try {
			const manager = getManager(command);
			const sandboxes = await manager.listSandboxes();
			if (sandboxes.length === 0) {
				console.log("No sandboxes found");
				return;
			}

			console.log(`Found ${sandboxes.length} sandbox(es):\n`);
			for (const sandbox of sandboxes) {
				const meta = await sandbox.getMeta();
				console.log(`  ${sandbox.id}`);
				if (Object.keys(meta).length > 0) {
					console.log(`    Meta: ${JSON.stringify(meta)}`);
				}
			}
		} catch (error) {
			console.error("Error listing sandboxes:", error);
			process.exit(1);
		}
	});

// Get sandbox info
program
	.command("info <id>")
	.description("Get information about a sandbox")
	.action(async (id: string, command) => {
		try {
			const manager = getManager(command);
			const sandbox = await manager.getSandbox(id);
			if (!sandbox) {
				console.error(`Sandbox ${id} not found`);
				process.exit(1);
			}

			const meta = await sandbox.getMeta();
			console.log(`Sandbox: ${sandbox.id}`);
			console.log(`Metadata:`, JSON.stringify(meta, null, 2));
		} catch (error) {
			console.error("Error getting sandbox info:", error);
			process.exit(1);
		}
	});

// Delete sandbox
program
	.command("delete <id>")
	.alias("rm")
	.description("Delete a sandbox")
	.action(async (id: string, command) => {
		try {
			const manager = getManager(command);
			await manager.deleteSandbox(id);
			console.log(`Deleted sandbox: ${id}`);
		} catch (error) {
			console.error("Error deleting sandbox:", error);
			process.exit(1);
		}
	});

// Read file from sandbox
program
	.command("read <id> <path>")
	.description("Read a file from a sandbox")
	.option("-e, --encoding <encoding>", "Encoding: utf-8 or base64", "utf-8")
	.option("-o, --offset <offset>", "Line offset", "1")
	.option("-l, --limit <limit>", "Line limit")
	.action(
		async (
			id: string,
			path: string,
			options: { encoding?: string; offset?: string; limit?: string },
			command,
		) => {
			try {
				const manager = getManager(command);
				const sandbox = await manager.getSandbox(id);
				if (!sandbox) {
					console.error(`Sandbox ${id} not found`);
					process.exit(1);
				}

				const encoding =
					options.encoding === "base64" ? "base64" : ("utf-8" as const);
				const offset = options.offset ? Number.parseInt(options.offset, 10) : 1;
				const limit = options.limit
					? Number.parseInt(options.limit, 10)
					: undefined;

				const result = await sandbox.readFile(path, {
					encoding,
					offset,
					limit,
				});
				if (result === null) {
					console.error(`File not found: ${path}`);
					process.exit(1);
				}
				console.log(result.content);
			} catch (error) {
				console.error("Error reading file:", error);
				process.exit(1);
			}
		},
	);

// Write file to sandbox
program
	.command("write <id> <path>")
	.description("Write content to a file in sandbox (reads from stdin)")
	.option("-e, --encoding <encoding>", "Encoding: utf-8 or base64", "utf-8")
	.option("-c, --content <content>", "Content to write (alternative to stdin)")
	.action(
		async (
			id: string,
			path: string,
			options: { encoding?: string; content?: string },
			command,
		) => {
			try {
				const manager = getManager(command);
				const sandbox = await manager.getSandbox(id);
				if (!sandbox) {
					console.error(`Sandbox ${id} not found`);
					process.exit(1);
				}

				const encoding =
					options.encoding === "base64" ? "base64" : ("utf-8" as const);

				let content: string;
				if (options.content) {
					content = options.content;
				} else {
					// Read from stdin
					const chunks: Buffer[] = [];
					for await (const chunk of process.stdin) {
						chunks.push(chunk);
					}
					content = Buffer.concat(chunks).toString("utf-8");
				}

				await sandbox.writeFile(path, content, { encoding });
				console.log(`Wrote to ${path}`);
			} catch (error) {
				console.error("Error writing file:", error);
				process.exit(1);
			}
		},
	);

// Delete file from sandbox
program
	.command("delete-file <id> <path>")
	.description("Delete a file from sandbox")
	.action(async (id: string, path: string, command) => {
		try {
			const manager = getManager(command);
			const sandbox = await manager.getSandbox(id);
			if (!sandbox) {
				console.error(`Sandbox ${id} not found`);
				process.exit(1);
			}

			await sandbox.deleteFile(path);
			console.log(`Deleted ${path}`);
		} catch (error) {
			console.error("Error deleting file:", error);
			process.exit(1);
		}
	});

// Execute command in sandbox
program
	.command("exec <id> <command> [args...]")
	.description("Execute a command in sandbox")
	.option("-w, --cwd <cwd>", "Working directory")
	.option("-e, --env <json>", "Environment variables as JSON")
	.action(
		async (
			id: string,
			command: string,
			args: string[],
			options: { cwd?: string; env?: string },
			cmd,
		) => {
			try {
				const manager = getManager(cmd);
				const sandbox = await manager.getSandbox(id);
				if (!sandbox) {
					console.error(`Sandbox ${id} not found`);
					process.exit(1);
				}

				const execOpts: { cwd?: string; env?: Record<string, string> } = {};
				if (options.cwd) execOpts.cwd = options.cwd;
				if (options.env) execOpts.env = JSON.parse(options.env);

				const processId = await sandbox.execute(command, args, execOpts);
				console.log(`Started process: ${processId}`);

				// Wait for completion
				const waitResult = await sandbox.wait(processId);

				// Get output
				const result = await sandbox.output(processId);
				console.log(result.output);

				// Release process
				await sandbox.release(processId);

				// Show signal if process was killed
				if (waitResult.signal) {
					console.log(`\nProcess terminated by signal: ${waitResult.signal}`);
				}

				// Exit with same code
				process.exit(result.exitCode);
			} catch (error) {
				console.error("Error executing command:", error);
				process.exit(1);
			}
		},
	);

// Get process output
program
	.command("output <id> <processId>")
	.description("Get output from a process")
	.action(async (id: string, processId: string, command) => {
		try {
			const manager = getManager(command);
			const sandbox = await manager.getSandbox(id);
			if (!sandbox) {
				console.error(`Sandbox ${id} not found`);
				process.exit(1);
			}

			const result = await sandbox.output(processId);
			console.log(result.output);
			console.log(
				`\nExit code: ${result.exitCode}${result.truncated ? " (truncated)" : ""}${result.signal ? ` (signal: ${result.signal})` : ""}`,
			);
		} catch (error) {
			console.error("Error getting output:", error);
			process.exit(1);
		}
	});

// Wait for process
program
	.command("wait <id> <processId>")
	.description("Wait for a process to complete")
	.action(async (id: string, processId: string, command) => {
		try {
			const manager = getManager(command);
			const sandbox = await manager.getSandbox(id);
			if (!sandbox) {
				console.error(`Sandbox ${id} not found`);
				process.exit(1);
			}

			const result = await sandbox.wait(processId);
			console.log(`Process ${processId} completed`);
			console.log(
				`Exit code: ${result.exitCode}${result.signal ? ` (signal: ${result.signal})` : ""}`,
			);
		} catch (error) {
			console.error("Error waiting for process:", error);
			process.exit(1);
		}
	});

// Kill process
program
	.command("kill <id> <processId>")
	.description("Kill a running process")
	.action(async (id: string, processId: string, command) => {
		try {
			const manager = getManager(command);
			const sandbox = await manager.getSandbox(id);
			if (!sandbox) {
				console.error(`Sandbox ${id} not found`);
				process.exit(1);
			}

			await sandbox.kill(processId);
			console.log(`Killed process ${processId}`);
		} catch (error) {
			console.error("Error killing process:", error);
			process.exit(1);
		}
	});

// Release process
program
	.command("release <id> <processId>")
	.description("Release a process (cleanup)")
	.action(async (id: string, processId: string, command) => {
		try {
			const manager = getManager(command);
			const sandbox = await manager.getSandbox(id);
			if (!sandbox) {
				console.error(`Sandbox ${id} not found`);
				process.exit(1);
			}

			await sandbox.release(processId);
			console.log(`Released process ${processId}`);
		} catch (error) {
			console.error("Error releasing process:", error);
			process.exit(1);
		}
	});

// Show help if no arguments provided
if (process.argv.length === 2) {
	program.help();
}

program.parse(process.argv);
