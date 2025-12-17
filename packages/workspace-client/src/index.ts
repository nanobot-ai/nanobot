/**
 * @nanobot-ai/workspace-client
 *
 * A TypeScript client library for the Workspace MCP server.
 * Provides typed methods for file system operations and terminal management.
 *
 * @example
 * ```typescript
 * import { WorkspaceClient } from "@nanobot-ai/workspace-client";
 *
 * const client = new WorkspaceClient({ url: "http://localhost:44100/mcp" });
 * await client.connect();
 *
 * const content = await client.readTextFile("/path/to/file.txt");
 * console.log(content);
 *
 * await client.close();
 * ```
 *
 * @packageDocumentation
 */

export { WorkspaceClient } from "./client.ts";
export { ensureConnected } from "./clients.ts";
export type {
	ListDirEntry,
	ListDirResult,
	ReadTextFileOptions,
	TerminalCreateOptions,
	TerminalCreateResult,
	TerminalOutputResult,
	TerminalWaitResult,
	WorkspaceClientConfig,
} from "./types.ts";
export { WorkspaceClientError } from "./types.ts";
