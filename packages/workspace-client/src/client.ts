import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";
import {
	CallToolResultSchema,
	type EmbeddedResource,
	type ImageContent,
	type TextContent,
} from "@modelcontextprotocol/sdk/types.js";
import type {
	ListDirOptions,
	ListDirResult,
	ReadTextFileOptions,
	TerminalCreateOptions,
	TerminalCreateResult,
	TerminalOutputResult,
	TerminalWaitResult,
	WorkspaceClientConfig,
	WorkspaceClientError,
} from "./types.ts";

/**
 * Client for interacting with the Workspace MCP server
 *
 * Provides typed methods for file system operations and terminal management.
 *
 * @example
 * ```typescript
 * const client = new WorkspaceClient({ url: "http://localhost:44100/mcp" });
 * await client.connect();
 *
 * // Read a file
 * const content = await client.readTextFile("/path/to/file.txt");
 *
 * // Create and manage a terminal
 * const { terminalId } = await client.terminalCreate("npm", { args: ["test"] });
 * const result = await client.terminalWait(terminalId);
 * await client.terminalRelease(terminalId);
 *
 * await client.close();
 * ```
 */
export class WorkspaceClient {
	private client: Client;
	private transport: StreamableHTTPClientTransport;
	private sessionId: string;
	private connected = false;

	/**
	 * Creates a new WorkspaceClient instance
	 *
	 * @param config - Configuration options
	 */
	constructor(config: WorkspaceClientConfig) {
		this.sessionId =
			config.sessionId ||
			`session-${Date.now()}-${Math.random().toString(36).slice(2, 11)}`;

		this.transport = new StreamableHTTPClientTransport(new URL(config.url));
		this.client = new Client(
			{
				name: config.clientInfo?.name || "@nanobot-ai/workspace-client",
				version: config.clientInfo?.version || "1.0.0",
			},
			{
				capabilities: {},
			},
		);
	}

	/**
	 * Establishes connection to the MCP server
	 * Must be called before using any other methods
	 */
	async connect(): Promise<void> {
		if (this.connected) {
			return;
		}

		try {
			await this.client.connect(this.transport);
			this.connected = true;
		} catch (error) {
			throw createError(
				"Failed to connect to MCP server",
				"CONNECTION_ERROR",
				error,
			);
		}
	}

	/**
	 * Reads the contents of a text file
	 *
	 * @param path - Path to the file to read
	 * @param options - Optional reading options (line offset and limit)
	 * @returns The file contents as a string
	 *
	 * @example
	 * ```typescript
	 * // Read entire file
	 * const content = await client.readTextFile("/path/to/file.txt");
	 *
	 * // Read specific lines
	 * const partial = await client.readTextFile("/path/to/file.txt", {
	 *   line: 10,
	 *   limit: 20
	 * });
	 * ```
	 */
	async readTextFile(
		path: string,
		options?: ReadTextFileOptions,
	): Promise<string> {
		this.ensureConnected();

		const result = await this.client.callTool({
			name: "readTextFile",
			arguments: {
				sessionId: this.sessionId,
				path,
				...(options?.line !== undefined && { line: options.line }),
				...(options?.limit !== undefined && { limit: options.limit }),
				...(options?.binary !== undefined && { binary: options.binary }),
			},
		});

		if (
			options?.ignoreNotFound &&
			(result.structuredContent as { fileNotFound: boolean })?.fileNotFound
		) {
			return "";
		}

		throwIfError(result, `Failed to read file: ${path}`, "READ_FILE_ERROR");

		const content = result.content as Array<
			TextContent | ImageContent | EmbeddedResource
		>;
		const firstContent = content[0] as TextContent | undefined;
		return firstContent?.text || "";
	}

	/**
	 * Writes content to a text file
	 * Creates the file if it doesn't exist
	 *
	 * @param path - Path to the file to write
	 * @param content - Content to write to the file
	 *
	 * @example
	 * ```typescript
	 * await client.writeTextFile("/path/to/file.txt", "Hello, world!");
	 * ```
	 */
	async writeTextFile(path: string, content: string): Promise<void> {
		this.ensureConnected();

		try {
			const result = await this.client.callTool({
				name: "writeTextFile",
				arguments: {
					sessionId: this.sessionId,
					path,
					content,
				},
			});

			throwIfError(result, "Failed to write file", "WRITE_FILE_ERROR");
		} catch (error) {
			if (error instanceof Error && error.name === "WorkspaceClientError") {
				throw error;
			}
			throw createError("Failed to write file", "WRITE_FILE_ERROR", error);
		}
	}

	/**
	 * Lists the contents of a directory
	 *
	 * @param path - Path to the directory to list
	 * @param options - Optional listing options (recursive, cursor, limit)
	 * @returns An object containing an array of directory entries
	 *
	 * @example
	 * ```typescript
	 * const { entries } = await client.listDir("/path/to/directory");
	 * for (const entry of entries) {
	 *   console.log(`${entry.name} - ${entry.isDirectory ? "DIR" : "FILE"} - ${entry.size} bytes`);
	 * }
	 * ```
	 */
	async listDir(
		path: string,
		options?: ListDirOptions,
	): Promise<ListDirResult> {
		this.ensureConnected();

		const result = await this.client.callTool(
			{
				name: "listDir",
				arguments: {
					sessionId: this.sessionId,
					path,
					cursor: options?.cursor,
					recursive: options?.recursive,
					limit: options?.limit,
				},
			},
			CallToolResultSchema,
		);

		if (
			options?.ignoreNotFound &&
			(result.structuredContent as { fileNotFound: boolean })?.fileNotFound
		) {
			return {
				entries: [],
			};
		}

		throwIfError(result, "Failed to list directory", "LIST_DIR_ERROR");

		if (!result.structuredContent) {
			throw createError(
				"Invalid response from listDir: missing structuredContent",
				"INVALID_RESPONSE",
			);
		}

		const structuredContent = result.structuredContent as {
			entries: Array<{
				name: string;
				isFile: boolean;
				isDirectory: boolean;
				size: number;
			}>;
			cursor?: string;
		};
		return {
			entries: structuredContent.entries,
			cursor: structuredContent.cursor,
		};
	}

	/**
	 * Resolves a path to an absolute path
	 * Respects the session's base directory if set
	 *
	 * @param path - Path to resolve (can be relative or absolute)
	 * @returns The resolved absolute path
	 *
	 * @example
	 * ```typescript
	 * const absolutePath = await client.resolvePath("./relative/path");
	 * console.log(absolutePath); // e.g., "/home/user/workspace/relative/path"
	 * ```
	 */
	async resolvePath(path: string): Promise<string> {
		this.ensureConnected();

		const result = await this.client.callTool({
			name: "resolvePath",
			arguments: {
				sessionId: this.sessionId,
				path,
			},
		});

		throwIfError(
			result,
			`Failed to resolve path: ${path}`,
			"RESOLVE_PATH_ERROR",
		);

		const content = result.content as Array<
			TextContent | ImageContent | EmbeddedResource
		>;
		const firstContent = content[0] as TextContent | undefined;
		return firstContent?.text || "";
	}

	/**
	 * Creates a new terminal session and executes a command
	 *
	 * @param command - The command to execute
	 * @param options - Optional terminal options (args, env, cwd, outputByteLimit)
	 * @returns An object containing the terminal ID
	 *
	 * @example
	 * ```typescript
	 * const { terminalId } = await client.terminalCreate("npm", {
	 *   args: ["test"],
	 *   cwd: "/path/to/project",
	 *   env: { NODE_ENV: "test" }
	 * });
	 * ```
	 */
	async terminalCreate(
		command: string,
		options?: TerminalCreateOptions,
	): Promise<TerminalCreateResult> {
		this.ensureConnected();

		const result = await this.client.callTool(
			{
				name: "terminalCreate",
				arguments: {
					sessionId: this.sessionId,
					command,
					...(options?.args && { args: options.args }),
					...(options?.env && { env: options.env }),
					...(options?.cwd && { cwd: options.cwd }),
					...(options?.outputByteLimit !== undefined && {
						outputByteLimit: options.outputByteLimit,
					}),
				},
			},
			CallToolResultSchema,
		);

		throwIfError(result, "Failed to create terminal", "TERMINAL_CREATE_ERROR");

		if (!result.structuredContent) {
			throw createError(
				"Invalid response from terminalCreate: missing structuredContent",
				"INVALID_RESPONSE",
			);
		}

		const structuredContent = result.structuredContent as {
			terminalId: string;
		};
		return {
			terminalId: structuredContent.terminalId,
		};
	}

	/**
	 * Gets the current output from a terminal session without waiting for completion
	 *
	 * @param terminalId - The terminal ID returned from terminalCreate
	 * @returns An object containing the output, truncation status, and optional exit status
	 *
	 * @example
	 * ```typescript
	 * const { output, exitStatus } = await client.terminalOutput(terminalId);
	 * console.log(output);
	 * ```
	 */
	async terminalOutput(terminalId: string): Promise<TerminalOutputResult> {
		this.ensureConnected();

		const result = await this.client.callTool(
			{
				name: "terminalOutput",
				arguments: {
					sessionId: this.sessionId,
					terminalId,
				},
			},
			CallToolResultSchema,
		);

		throwIfError(
			result,
			"Failed to get terminal output",
			"TERMINAL_OUTPUT_ERROR",
		);

		if (!result.structuredContent) {
			throw createError(
				"Invalid response from terminalOutput: missing structuredContent",
				"INVALID_RESPONSE",
			);
		}

		const structuredContent = result.structuredContent as {
			output: string;
			truncated: boolean;
			exitStatus?: { exitCode?: number; signal?: string };
		};
		return {
			output: structuredContent.output,
			truncated: structuredContent.truncated,
			exitStatus: structuredContent.exitStatus,
		};
	}

	/**
	 * Waits for a terminal command to complete
	 * Blocks until the command finishes execution
	 *
	 * @param terminalId - The terminal ID returned from terminalCreate
	 * @returns An object containing the exit code and optional signal
	 *
	 * @example
	 * ```typescript
	 * const { exitCode, signal } = await client.terminalWait(terminalId);
	 * if (exitCode === 0) {
	 *   console.log("Command succeeded");
	 * }
	 * ```
	 */
	async terminalWait(terminalId: string): Promise<TerminalWaitResult> {
		this.ensureConnected();

		const result = await this.client.callTool(
			{
				name: "terminalWait",
				arguments: {
					sessionId: this.sessionId,
					terminalId,
				},
			},
			CallToolResultSchema,
		);

		throwIfError(result, "Failed to wait for terminal", "TERMINAL_WAIT_ERROR");

		if (!result.structuredContent) {
			throw createError(
				"Invalid response from terminalWait: missing structuredContent",
				"INVALID_RESPONSE",
			);
		}

		const structuredContent = result.structuredContent as {
			exitCode: number;
			signal?: string;
		};
		return {
			exitCode: structuredContent.exitCode,
			signal: structuredContent.signal,
		};
	}

	/**
	 * Terminates a running terminal command
	 *
	 * @param terminalId - The terminal ID returned from terminalCreate
	 * @param signal - Optional signal to send (default: SIGTERM)
	 *
	 * @example
	 * ```typescript
	 * await client.terminalKill(terminalId);
	 * // or with custom signal
	 * await client.terminalKill(terminalId, "SIGKILL");
	 * ```
	 */
	async terminalKill(terminalId: string, signal?: string): Promise<void> {
		this.ensureConnected();

		const result = await this.client.callTool({
			name: "terminalKill",
			arguments: {
				sessionId: this.sessionId,
				terminalId,
				...(signal && { signal }),
			},
		});

		throwIfError(result, "Failed to kill terminal", "TERMINAL_KILL_ERROR");
	}

	/**
	 * Releases a terminal session and frees its resources
	 * Should be called when done with a terminal to prevent resource leaks
	 *
	 * @param terminalId - The terminal ID returned from terminalCreate
	 *
	 * @example
	 * ```typescript
	 * await client.terminalRelease(terminalId);
	 * ```
	 */
	async terminalRelease(terminalId: string): Promise<void> {
		this.ensureConnected();

		const result = await this.client.callTool({
			name: "terminalRelease",
			arguments: {
				sessionId: this.sessionId,
				terminalId,
			},
		});

		throwIfError(
			result,
			"Failed to release terminal",
			"TERMINAL_RELEASE_ERROR",
		);
	}

	/**
	 * Closes the connection to the MCP server and cleans up resources
	 *
	 * @example
	 * ```typescript
	 * await client.close();
	 * ```
	 */
	async close(): Promise<void> {
		if (!this.connected) {
			return;
		}

		try {
			await this.client.close();
			this.connected = false;
		} catch (error) {
			throw createError("Failed to close connection", "CLOSE_ERROR", error);
		}
	}

	/**
	 * Gets the session ID being used by this client
	 */
	getSessionId(): string {
		return this.sessionId;
	}

	/**
	 * Checks if the client is currently connected
	 */
	isConnected(): boolean {
		return this.connected;
	}

	private ensureConnected(): void {
		if (!this.connected) {
			throw createError(
				"Client is not connected. Call connect() first.",
				"NOT_CONNECTED",
			);
		}
	}
}

function throwIfError(
	result: Record<string, unknown>,
	message: string,
	code: string,
): void {
	if (result.isError) {
		const content = Array.isArray(result.content) && result.content?.[0];
		const firstContent = content?.[0]?.text;
		throw createError(
			`${message}: ${firstContent?.text || "Unknown error"}`,
			code,
		);
	}
}

function createError(
	message: string,
	code: string,
	cause?: unknown,
): WorkspaceClientError {
	const error = new Error(
		`${message} (${code}) ${cause instanceof Error ? cause.message : String(cause ?? "")}`,
	) as WorkspaceClientError;
	error.name = "WorkspaceClientError";
	error.code = code;
	error.cause = cause;
	return error;
}
