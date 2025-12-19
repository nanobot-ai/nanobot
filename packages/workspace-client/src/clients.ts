import { WorkspaceClient } from "@nanobot-ai/workspace-client";

/**
 * Map of workspace clients by workspaceId
 * Each workspaceId has its own dedicated client instance
 */
const clientsByWorkspaceId = new Map<string, WorkspaceClient>();

/**
 * Get or create a workspace client for the given workspaceId
 *
 * @param workspaceId - The workspace identifier (used as sessionId)
 * @param url - Optional workspace server URL (defaults to WORKSPACE_URL env var or localhost:44100)
 * @returns WorkspaceClient instance for the workspaceId
 */
export function getWorkspaceClient(
	workspaceId: string,
	url?: string,
): WorkspaceClient {
	let client = clientsByWorkspaceId.get(workspaceId);

	if (!client) {
		const serverUrl =
			url || process.env.WORKSPACE_URL || "http://localhost:5173/mcp";
		client = new WorkspaceClient({
			url: serverUrl,
			sessionId: workspaceId,
		});
		clientsByWorkspaceId.set(workspaceId, client);
	}

	return client;
}

/**
 * Ensure the workspace client is connected for the given workspaceId
 *
 * @param workspaceId - The workspace identifier
 * @param url - Optional workspace server URL
 * @returns Connected WorkspaceClient instance
 */
export async function ensureConnected(
	workspaceId: string,
): Promise<WorkspaceClient> {
	const client = getWorkspaceClient(workspaceId);
	if (!client.isConnected()) {
		await client.connect();
	}
	return client;
}

/**
 * Close and remove a workspace client
 *
 * @param workspaceId - The workspace identifier
 */
export async function closeWorkspaceClient(workspaceId: string): Promise<void> {
	const client = clientsByWorkspaceId.get(workspaceId);
	if (client) {
		await client.close();
		clientsByWorkspaceId.delete(workspaceId);
	}
}

/**
 * Close all workspace clients
 */
export async function closeAllClients(): Promise<void> {
	const closePromises = Array.from(clientsByWorkspaceId.entries()).map(
		async ([workspaceId, client]) => {
			await client.close();
			clientsByWorkspaceId.delete(workspaceId);
		},
	);
	await Promise.all(closePromises);
}
