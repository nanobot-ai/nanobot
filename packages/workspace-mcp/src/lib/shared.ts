import { homedir } from "node:os";
import { join } from "node:path";
import type { Sandbox } from "@nanobot-ai/sandbox";
import { DockerDriver, SandboxManager } from "@nanobot-ai/sandbox";
import { createFsFileStorage } from "@remix-run/file-storage/fs";

// Get Docker image from environment variable or use default
const SANDBOX_IMAGE = process.env.SANDBOX_IMAGE || "sandbox-test";

// Setup storage in user's home directory
const storageDir = join(homedir(), ".workspace-mcp");
const storage = createFsFileStorage(storageDir);

// Create Docker driver
const driver = new DockerDriver({
	image: SANDBOX_IMAGE,
	workdir: "/workspace",
	localDataDir: join(storageDir, "data"),
});

// Create sandbox manager
const manager = new SandboxManager(storage, {
	drivers: { docker: driver },
	defaultDriver: "docker",
});

/**
 * Session-based sandbox manager
 * Maps session IDs to sandbox instances
 */
class SessionSandboxManager {
	/**
	 * Get an existing sandbox for a given session, or create a new one if it doesn't exist'
	 */
	async getSandbox(sessionId: string): Promise<Sandbox> {
		const parts = sessionId.split("?");
		const params: Record<string, string> = {};
		let baseUri: string | undefined;
		let parentId: string | undefined;
		const parsed = new URLSearchParams(parts[1] ?? "");
		for (const [key, value] of parsed) {
			if (key === "baseUri") {
				baseUri = value;
			} else if (key === "parentId") {
				parentId = value;
			} else {
				params[key] = value;
			}
		}
		// Try to load existing sandbox from storage
		const existingSandbox = await manager.getSandbox(parts[0]);
		if (existingSandbox) {
			return existingSandbox;
		}

		return await manager.createSandbox({
			id: parts[0],
			parentId,
			baseUri,
			meta: params,
		});
	}

	/**
	 * Delete a session sandbox
	 */
	async deleteSandbox(sessionId: string): Promise<void> {
		await manager.deleteSandbox(sessionId);
	}

	/**
	 * List all sessions
	 */
	async listSessions(): Promise<string[]> {
		const sandboxes = await manager.listSandboxes();
		return sandboxes.map((sb) => sb.id);
	}
}

/**
 * Shared session sandbox manager instance used across all tools
 */
export const sessionSandboxManager = new SessionSandboxManager();
