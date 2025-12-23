/**
 * Docker-based sandbox implementation using Docker CLI
 *
 * This module provides a SandboxDriver implementation that uses Docker containers
 * to create isolated execution environments. Each sandbox runs in its own container.
 *
 * @example
 * ```typescript
 * import { Manager } from "./manager.js";
 * import { DockerDriver } from "./docker.js";
 * import { createMemoryFileStorage } from "@remix-run/file-storage/memory";
 *
 * // Create a manager with the Docker driver
 * const driver = new DockerDriver({
 *   image: "ubuntu:latest",
 *   workdir: "/workspace",
 * });
 *
 * const manager = new Manager(createMemoryFileStorage(), {
 *   drivers: { docker: driver },
 *   defaultDriver: "docker",
 * });
 *
 * // Create and use a sandbox
 * const sandbox = await manager.createSandbox();
 * await sandbox.writeFile("hello.txt", "Hello, World!");
 * const content = await sandbox.readFile("hello.txt");
 * console.log(content); // { content: "Hello, World!", encoding: "utf-8" }
 *
 * // Execute commands
 * const procId = await sandbox.execute("echo", ["Hello from Docker!"]);
 * await sandbox.wait(procId);
 * const output = await sandbox.output(procId);
 * console.log(output.output); // "Hello from Docker!\n"
 *
 * // Cleanup
 * await manager.deleteSandbox(sandbox.id);
 * ```
 */

import { spawn } from "node:child_process";
import { existsSync } from "node:fs";
import { cp, mkdir, rm } from "node:fs/promises";
import { join } from "node:path";
import type { Encoding, Sandbox, SandboxMetadata } from "../sandbox.js";
import type { SandboxConfig, SandboxDriver } from "./manager.js";

export interface DockerDriverConfig {
	image?: string;
	workdir?: string;
	env?: Record<string, string>;
	localDataDir?: string;
}

interface DockerSandboxConfig {
	image: string;
	workdir: string;
	localDataDir?: string;
}

/**
 * Process state is persisted to disk in /.data/.processes/<processId>/
 * Each process directory contains:
 * - meta.json: { command, args, startTime, exitCode, signal, outputByteLimit, truncated }
 * - stdout.txt: Process stdout output
 * - stderr.txt: Process stderr output
 * - exitcode.txt: Process exit code
 * - signal.txt: Signal that terminated process (if killed)
 * - pid.txt: Process PID for kill operations
 *
 * The sb_execute script manages all process state and output.
 */
interface ProcessMetadata {
	command: string;
	args: string[];
	startTime: number;
	exitCode: number | null;
	signal?: string | null;
	outputByteLimit?: number;
	truncated?: boolean;
}

export class DockerDriver implements SandboxDriver {
	readonly name = "docker";
	private config: Required<DockerDriverConfig>;

	constructor(config?: DockerDriverConfig) {
		this.config = {
			image: config?.image ?? "ubuntu:latest",
			workdir: config?.workdir ?? "/workspace",
			env: config?.env ?? {},
			localDataDir:
				config?.localDataDir ?? join(process.cwd(), "docker-sandbox"),
		};
	}

	async createSandbox(
		config: SandboxConfig,
		opts?: { recreate?: boolean },
	): Promise<Record<string, unknown>> {
		const containerName = `sandbox-${config.id}`;

		// Create local directory for this sandbox
		const localDataDir = join(this.config.localDataDir, config.id);

		// If parentId is set, copy parent's data directory
		if (config.parentId) {
			const parentDataDir = join(this.config.localDataDir, config.parentId);

			if (existsSync(parentDataDir)) {
				if (!opts?.recreate) {
					// Copy parent directory to new sandbox directory
					await cp(parentDataDir, localDataDir, { recursive: true });
					console.log(
						`Copied parent data from ${parentDataDir} to ${localDataDir}`,
					);
				}
			} else {
				// Parent directory doesn't exist, create empty one
				await mkdir(localDataDir, { recursive: true });
				console.warn(
					`Parent directory ${parentDataDir} not found, created empty directory`,
				);
			}
		} else {
			// No parent, create empty directory
			await mkdir(localDataDir, { recursive: true });
		}

		// Set permissions based on current user
		const currentUid = process.getuid ? process.getuid() : null;

		if (currentUid === 0 || currentUid === 10000) {
			// Running as root or as UID 10000, set ownership to 10000
			try {
				await new Promise<void>((resolve, reject) => {
					const proc = spawn("chown", ["-R", "10000:10000", localDataDir]);
					let stderr = "";

					proc.stderr.on("data", (data) => {
						stderr += data.toString();
					});

					proc.on("close", (code) => {
						if (code !== 0) {
							reject(new Error(`chown failed: ${stderr}`));
						} else {
							resolve();
						}
					});

					proc.on("error", (error) => {
						reject(error);
					});
				});
			} catch (error) {
				console.warn(
					`Warning: Could not set ownership to UID 10000 for ${localDataDir}:`,
					error,
				);
			}
		} else {
			// Not running as root or UID 10000, make directory world-writable
			try {
				await new Promise<void>((resolve, reject) => {
					const proc = spawn("chmod", ["-R", "777", localDataDir]);
					let stderr = "";

					proc.stderr.on("data", (data) => {
						stderr += data.toString();
					});

					proc.on("close", (code) => {
						if (code !== 0) {
							reject(new Error(`chmod failed: ${stderr}`));
						} else {
							resolve();
						}
					});

					proc.on("error", (error) => {
						reject(error);
					});
				});
			} catch (error) {
				console.warn(
					`Warning: Could not set permissions on ${localDataDir}:`,
					error,
				);
			}
		}

		// Build docker run command
		const args = [
			"run",
			"-d", // Detached mode
			"--name",
			containerName,
			"-w",
			this.config.workdir,
			// Bind mount local directory to /.data in container
			"-v",
			`${localDataDir}:/.data:rw`,
			// Add FUSE device
			"--device",
			"/dev/fuse",
			// Add SYS_ADMIN capability for FUSE
			"--cap-add",
			"SYS_ADMIN",
		];

		if (config.baseUri) {
			if (!existsSync(config.baseUri)) {
				throw new Error(`Base URI ${config.baseUri} does not exist`);
			}

			args.push("-v", `${config.baseUri}:/base:ro`);
			// TODO: remove this, it's a hack to make overlay work until agentfs is fixed
			args.push("-v", `${config.baseUri}:/workspace`);
		}

		// Add environment variables
		for (const [key, value] of Object.entries(this.config.env)) {
			args.push("-e", `${key}=${value}`);
		}

		// Build configuration JSON to pass to entrypoint
		const entrypointConfig = {
			id: config.id,
			baseUri: config.baseUri,
			workdir: this.config.workdir,
			meta: config.meta,
		};

		// Pass JSON config as argument to entrypoint
		args.push(this.config.image, JSON.stringify(entrypointConfig));

		console.log(
			`Starting Docker container ${this.config.image} docker ${args.join(" ")}...`,
		);
		await execDocker(args);

		return {
			image: this.config.image,
			workdir: this.config.workdir,
			localDataDir,
		};
	}

	async deleteSandbox(config: SandboxConfig): Promise<void> {
		const driverConfig = config.driverConfig as DockerSandboxConfig | undefined;
		const containerId = `sandbox-${config.id}`;

		try {
			// Stop the container
			await execDocker(["stop", containerId]);
		} catch (_error) {
			// Container might already be stopped, continue to remove
		}

		try {
			// Remove the container
			await execDocker(["rm", containerId]);
		} catch (_error) {
			// Container might already be removed, ignore
		}

		// Clean up local data directory
		if (driverConfig?.localDataDir) {
			try {
				await rm(driverConfig.localDataDir, { recursive: true, force: true });
			} catch (_error) {
				// Directory might not exist or already be removed, ignore
			}
		}
	}

	async loadSandbox(
		config: SandboxConfig,
		_opts?: { create?: boolean },
	): Promise<Sandbox> {
		const driverConfig = config.driverConfig as DockerSandboxConfig | undefined;
		if (!driverConfig) {
			throw new Error(`Sandbox ${config.id} not created yet`);
		}

		if (!(await this.#containerExists(config.id))) {
			await this.createSandbox(config, { recreate: true });
		}

		return new DockerSandbox(config.id, driverConfig, config.meta ?? {});
	}

	/**
	 * Check if the container exists
	 */
	async #containerExists(id: string): Promise<boolean> {
		try {
			const output = await execDocker(["inspect", `sandbox-${id}`]);
			const inspection = JSON.parse(output);

			// Check if container is running
			if (inspection[0]?.State?.Running !== true) {
				await execDocker(["rm", "-f", `sandbox-${id}`]);
				return false;
			}

			return true;
		} catch (_error) {
			return false;
		}
	}
}

class DockerSandbox implements Sandbox {
	readonly id: string;
	private readonly workdir: string;
	private readonly meta: SandboxMetadata;
	private readonly driverConfig: DockerSandboxConfig;

	constructor(id: string, config: DockerSandboxConfig, meta: SandboxMetadata) {
		this.id = id;
		this.workdir = config.workdir;
		this.meta = meta;
		this.driverConfig = config;
	}

	async getMeta(): Promise<SandboxMetadata> {
		return this.meta;
	}

	resolvePath(path: string): string {
		return this.#_resolvePath(path);
	}

	#_resolvePath(path: string): string {
		if (path.startsWith("/")) {
			return path;
		}
		return `${this.workdir}/${path}`;
	}

	private async getProcessMetadata(
		processId: string,
	): Promise<ProcessMetadata | null> {
		const processDir = `/.data/.processes/${processId}`;
		try {
			const metaJson = await this.execInContainer([
				"sb_read",
				`${processDir}/meta.json`,
				"utf-8",
			]);
			return JSON.parse(metaJson) as ProcessMetadata;
		} catch (_error) {
			return null;
		}
	}

	private async getProcessExitCode(processId: string): Promise<number | null> {
		const processDir = `/.data/.processes/${processId}`;
		try {
			const exitCodeStr = await this.execInContainer([
				"sb_read",
				`${processDir}/exitcode.txt`,
				"utf-8",
			]);
			return Number.parseInt(exitCodeStr.trim(), 10);
		} catch (_error) {
			return null;
		}
	}

	async readFile(
		path: string,
		opts?: { encoding?: Encoding; limit?: number; offset?: number },
	): Promise<{
		content: string;
		encoding: Encoding;
	} | null> {
		const encoding = opts?.encoding ?? "utf-8";
		const absolutePath = this.#_resolvePath(path);

		try {
			const args = ["sb_read", absolutePath, encoding];

			if (opts?.offset) {
				args.push(String(opts.offset));
			} else {
				args.push("1");
			}

			if (opts?.limit) {
				args.push(String(opts.limit));
			}

			const content = await this.execInContainer(args);

			return {
				content: encoding === "base64" ? content.trim() : content,
				encoding,
			};
		} catch (error) {
			// Check if the error is a "file not found" error
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			if (
				errorMessage.includes("No such file") ||
				errorMessage.includes("not found")
			) {
				return null;
			}
			// For other errors, still throw
			throw new Error(`Failed to read file ${path}: ${errorMessage}`);
		}
	}

	async writeFile(
		path: string,
		content: string,
		opts?: { encoding?: Encoding },
	): Promise<void> {
		const encoding = opts?.encoding ?? "utf-8";
		const absolutePath = this.#_resolvePath(path);

		try {
			// sb_write will create directories as needed
			await this.execInContainer(["sb_write", absolutePath, encoding], content);
		} catch (error) {
			throw new Error(
				`Failed to write file ${path}: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	}

	async deleteFile(path: string): Promise<void> {
		const absolutePath = this.#_resolvePath(path);

		try {
			await this.execInContainer(["rm", "-f", absolutePath]);
		} catch (error) {
			throw new Error(
				`Failed to delete file ${path}: ${error instanceof Error ? error.message : String(error)}`,
			);
		}
	}

	async readdir(
		path: string,
		opts?: {
			cursor?: string;
			recursive?: boolean;
			limit?: number;
		},
	): Promise<{
		entries: Array<{
			name: string;
			isFile: boolean;
			isDirectory: boolean;
			size: number;
			skipped?: boolean;
		}>;
		cursor?: string;
	}> {
		const absolutePath = this.#_resolvePath(path);

		try {
			// Calculate offset from cursor
			const offset = opts?.cursor ? opts.cursor : "0";
			const limit = opts?.limit ? String(opts.limit) : "";
			const recursive = opts?.recursive ? "1" : "0";

			// Call sb_readdir script
			const args = ["sb_readdir", absolutePath, offset, limit, recursive];
			const output = await this.execInContainer(args);

			// Parse the output
			const lines = output
				.trim()
				.split("\n")
				.filter((line) => line.length > 0);

			// Check if we got more entries than the limit (indicates more pages)
			let hasMore = false;
			if (opts?.limit && lines.length > opts.limit) {
				hasMore = true;
				lines.pop(); // Remove the extra entry
			}

			const entries = lines.map((line) => {
				const [type, sizeStr, fullPath, skipped] = line.split("|");
				const size = Number.parseInt(sizeStr, 10) || 0;

				// Get relative path from the directory being listed
				let name = fullPath;
				if (fullPath.startsWith(`${absolutePath}/`)) {
					name = fullPath.substring(absolutePath.length + 1);
				} else if (fullPath === absolutePath) {
					name = ".";
				}

				return {
					name,
					isFile: type === "f",
					isDirectory: type === "d",
					size,
					skipped: skipped === "1" ? true : undefined,
				};
			});

			// Calculate next cursor if there are more entries
			const nextCursor = hasMore
				? String(Number.parseInt(offset, 10) + entries.length)
				: undefined;

			return {
				entries,
				cursor: nextCursor,
			};
		} catch (error) {
			// Check if the error is a "not found" or "not a directory" error
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			if (
				errorMessage.includes("not found") ||
				errorMessage.includes("not a directory")
			) {
				return { entries: [], cursor: undefined };
			}
			throw new Error(`Failed to read directory ${path}: ${errorMessage}`);
		}
	}

	async execute(
		command: string,
		args: string[],
		opts?: {
			cwd?: string;
			env?: Record<string, string>;
			outputByteLimit?: number;
		},
	): Promise<string> {
		// Generate unique process ID
		const processId = `proc-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

		// Default output byte limit to 0 (unlimited)
		const outputByteLimit = opts?.outputByteLimit ?? 0;

		// Build sb_execute command with process ID and output byte limit
		const execArgs = [
			"sb_execute",
			processId,
			String(outputByteLimit),
			command,
			...args,
		];

		// Start the process in background
		const dockerArgs = ["exec", "-d"];

		// Add working directory
		if (opts?.cwd) {
			dockerArgs.push("--workdir", this.#_resolvePath(opts.cwd));
		}

		// Add environment variables
		if (opts?.env) {
			for (const [key, value] of Object.entries(opts.env)) {
				dockerArgs.push("--env", `${key}=${value}`);
			}
		}

		const containerId = `sandbox-${this.id}`;
		dockerArgs.push(containerId, ...execArgs);

		// Execute in detached mode
		await execDocker(dockerArgs);

		return processId;
	}

	async kill(id: string, signal?: string): Promise<void> {
		const metadata = await this.getProcessMetadata(id);
		if (!metadata) {
			throw new Error(`Process ${id} not found`);
		}

		// Default to SIGTERM if no signal specified
		const killSignal = signal ?? "SIGTERM";

		// Read PID from pid.txt file
		const processDir = `/.data/.processes/${id}`;
		let pid: string | null = null;
		try {
			pid = (
				await this.execInContainer([
					"sb_read",
					`${processDir}/pid.txt`,
					"utf-8",
				])
			).trim();
		} catch (_error) {
			// PID file might not exist
		}

		// Try to kill the process by PID if available, otherwise by command name
		try {
			if (pid) {
				await this.execInContainer(["kill", `-${killSignal}`, pid]);
			} else {
				await this.execInContainer([
					"pkill",
					`-${killSignal}`,
					"-f",
					metadata.command,
				]);
			}

			// Update metadata with signal information
			const processDir = `/.data/.processes/${id}`;
			await this.execInContainer([
				"sh",
				"-c",
				`echo "${killSignal}" > "${processDir}/signal.txt"`,
			]);

			// Update the metadata JSON with signal
			const updatedMeta = { ...metadata, signal: killSignal };
			await this.execInContainer(
				["sb_write", `${processDir}/meta.json`, "utf-8"],
				JSON.stringify(updatedMeta, null, 2),
			);
		} catch (_error) {
			// Process might already be done or not found
		}
	}

	async output(id: string): Promise<{
		output: string;
		truncated: boolean;
		exitCode: number;
		signal?: string;
	}> {
		const metadata = await this.getProcessMetadata(id);
		if (!metadata) {
			throw new Error(`Process ${id} not found`);
		}

		const processDir = `/.data/.processes/${id}`;

		// Read stdout and stderr
		let stdout = "";
		let stderr = "";

		try {
			stdout = await this.execInContainer([
				"sb_read",
				`${processDir}/stdout.txt`,
				"utf-8",
			]);
		} catch (_error) {
			// File might not exist yet
		}

		try {
			stderr = await this.execInContainer([
				"sb_read",
				`${processDir}/stderr.txt`,
				"utf-8",
			]);
		} catch (_error) {
			// File might not exist yet
		}

		// Get exit code from file or metadata
		let exitCode = await this.getProcessExitCode(id);
		if (exitCode === null) {
			exitCode = metadata.exitCode ?? -1;
		}

		// Get signal from metadata (if process was killed)
		const signal = metadata.signal ?? undefined;

		// Get truncated flag from metadata
		const truncated = metadata.truncated ?? false;

		return {
			output: stdout + stderr,
			truncated,
			exitCode,
			signal,
		};
	}

	async wait(id: string): Promise<{ exitCode: number; signal?: string }> {
		const metadata = await this.getProcessMetadata(id);
		if (!metadata) {
			throw new Error(`Process ${id} not found`);
		}

		// Poll for exit code file to exist (sb_execute writes this when done)
		const maxWait = 3600000; // 1 hour
		const pollInterval = 100; // 100ms
		const startTime = Date.now();

		while (Date.now() - startTime < maxWait) {
			const exitCode = await this.getProcessExitCode(id);
			if (exitCode !== null) {
				// Refresh metadata to get signal information
				const updatedMetadata = await this.getProcessMetadata(id);
				const signal = updatedMetadata?.signal ?? undefined;
				return { exitCode, signal };
			}
			await new Promise((resolve) => setTimeout(resolve, pollInterval));
		}

		throw new Error(`Process ${id} timed out after ${maxWait}ms`);
	}

	async release(id: string): Promise<void> {
		const metadata = await this.getProcessMetadata(id);
		if (!metadata) {
			throw new Error(`Process ${id} not found`);
		}

		// Remove the process directory
		const processDir = `/.data/.processes/${id}`;
		try {
			await this.execInContainer(["rm", "-rf", processDir]);
		} catch (_error) {
			// Directory might not exist or already removed
		}
	}

	private async execInContainer(
		args: string[],
		stdin?: string,
		opts?: { cwd?: string; env?: Record<string, string> },
	): Promise<string> {
		return new Promise((resolve, reject) => {
			const dockerArgs = ["exec", "-i"];

			// Add environment variables via Docker --env
			if (opts?.env) {
				for (const [key, value] of Object.entries(opts.env)) {
					dockerArgs.push("--env", `${key}=${value}`);
				}
			}

			// Add working directory via Docker --workdir
			if (opts?.cwd) {
				const absoluteCwd = this.#_resolvePath(opts.cwd);
				dockerArgs.push("--workdir", absoluteCwd);
			}

			const containerId = `sandbox-${this.id}`;
			dockerArgs.push(containerId, ...args);

			const proc = spawn("docker", dockerArgs);

			let stdout = "";
			let stderr = "";

			proc.stdout.on("data", (data) => {
				stdout += data.toString();
			});

			proc.stderr.on("data", (data) => {
				stderr += data.toString();
			});

			proc.on("close", (code) => {
				if (code !== 0) {
					reject(
						new Error(
							`Command failed with exit code ${code}: ${stderr || stdout}`,
						),
					);
				} else {
					resolve(stdout);
				}
			});

			proc.on("error", (error) => {
				reject(error);
			});

			// Write stdin if provided
			if (stdin !== undefined) {
				proc.stdin.write(stdin);
				proc.stdin.end();
			}
		});
	}
}

async function execDocker(args: string[]): Promise<string> {
	return new Promise((resolve, reject) => {
		const proc = spawn("docker", args);

		let stdout = "";
		let stderr = "";

		proc.stdout.on("data", (data) => {
			stdout += data.toString();
		});

		proc.stderr.on("data", (data) => {
			stderr += data.toString();
		});

		proc.on("close", (code) => {
			if (code !== 0) {
				reject(new Error(`Docker command failed: ${stderr || stdout}`));
			} else {
				resolve(stdout);
			}
		});

		proc.on("error", (error) => {
			reject(error);
		});
	});
}
