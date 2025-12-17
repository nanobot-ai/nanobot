/**
 * Configuration options for the WorkspaceClient
 */
export type WorkspaceClientConfig = {
	/**
	 * The URL of the MCP server endpoint
	 * @example "http://localhost:44100/mcp"
	 */
	url: string;

	/**
	 * Optional session ID. If not provided, a unique session ID will be generated.
	 */
	sessionId?: string;

	/**
	 * Optional client information
	 */
	clientInfo?: {
		name: string;
		version: string;
	};
};

/**
 * Options for reading a text file
 */
export type ReadTextFileOptions = {
	/**
	 * Line number to start reading from (0-based)
	 */
	line?: number;

	/**
	 * Maximum number of lines to read
	 */
	limit?: number;

	/**
	 * If the file is not found, return an empty string instead of throwing an error.
	 */
	ignoreNotFound?: boolean;

	/**
	 * Read raw content as a data URI instead of a string.
	 */
	binary?: boolean;
};

export type ListDirOptions = {
	/**
	 * If the directory is not found, return an empty array instead of throwing an error.
	 */
	ignoreNotFound?: boolean;

	/**
	 * Cursor for pagination. Use the cursor returned from a previous call to get the next page of results.
	 */
	cursor?: string;

	/**
	 * If true, recursively list all files and directories in subdirectories.
	 * When recursive is true, entry names will be relative paths from the base directory.
	 */
	recursive?: boolean;

	/**
	 * Maximum number of entries to return. If there are more entries, a cursor will be returned
	 * that can be used to fetch the next page of results.
	 */
	limit?: number;
};

/**
 * Entry in a directory listing
 */
export type ListDirEntry = {
	/**
	 * Name of the file or directory
	 */
	name: string;

	/**
	 * Whether this entry is a file
	 */
	isFile: boolean;

	/**
	 * Whether this entry is a directory
	 */
	isDirectory: boolean;

	/**
	 * Size of the file in bytes
	 */
	size: number;
};

/**
 * Result from listing a directory
 */
export type ListDirResult = {
	/**
	 * Array of directory entries
	 */
	entries: ListDirEntry[];
	cursor?: string;
};

/**
 * Options for creating a terminal session
 */
export type TerminalCreateOptions = {
	/**
	 * Command arguments
	 */
	args?: string[];

	/**
	 * Environment variables for the command
	 */
	env?: Record<string, string>;

	/**
	 * Working directory for the command
	 */
	cwd?: string;

	/**
	 * Maximum output size in bytes
	 */
	outputByteLimit?: number;
};

/**
 * Result from creating a terminal session
 */
export type TerminalCreateResult = {
	/**
	 * Unique identifier for the terminal session
	 */
	terminalId: string;
};

/**
 * Result from getting terminal output
 */
export type TerminalOutputResult = {
	/**
	 * The current output from the terminal
	 */
	output: string;

	/**
	 * Whether the output was truncated due to size limits
	 */
	truncated: boolean;

	/**
	 * Exit status if the command has completed
	 */
	exitStatus?: {
		exitCode?: number;
		signal?: string;
	};
};

/**
 * Result from waiting for terminal completion
 */
export type TerminalWaitResult = {
	/**
	 * Exit code of the command
	 */
	exitCode: number;

	/**
	 * Signal that terminated the command, if any
	 */
	signal?: string;
};

/**
 * Custom error class for workspace client errors
 */
export class WorkspaceClientError extends Error {
	constructor(
		message: string,
		public code?: string,
		public cause?: unknown,
	) {
		super(message);
		this.name = "WorkspaceClientError";
	}
}
