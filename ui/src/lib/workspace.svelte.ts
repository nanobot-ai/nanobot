import { type Workspace } from './types';
import { WorkspaceMimeType } from './types';
import { SimpleClient } from './mcpclient';

/**
 * WorkspaceService - Manages workspace operations via MCP API
 * Uses SimpleClient to communicate with the backend for workspace CRUD operations
 */
export class WorkspaceService {
	private client: SimpleClient;

	constructor(opts?: { client?: SimpleClient }) {
		this.client = opts?.client || new SimpleClient();
	}

	/**
	 * List all workspaces by querying resources with nanobot://workspaces/{id} URIs
	 * Note: resources/list only returns metadata, not content
	 */
	async listWorkspaces(): Promise<Workspace[]> {
		const result = await this.client.listResources({
			prefix: 'nanobot://workspaces/'
		});

		if (!result.resources) {
			return [];
		}

		// Filter for workspace resources and convert metadata to Workspace objects
		const workspaces: Workspace[] = [];
		for (const resource of result.resources) {
			// Check if this is a workspace resource
			if (
				resource.uri?.startsWith('nanobot://workspaces/') &&
				resource.mimeType === WorkspaceMimeType
			) {
				try {
					// Extract workspace ID from URI (nanobot://workspaces/{id})
					const id = resource.uri.split('/').pop() || '';

					// Build workspace from resource metadata
					const workspace: Workspace = {
						id,
						name: resource.name,
						created: resource.annotations?.lastModified || '',
						icons: resource.icons
					};

					// Extract vendor-scoped metadata from _meta['ai.nanobot']
					const nanobotMeta = resource._meta?.['ai.nanobot'] as Record<string, unknown> | undefined;
					if (nanobotMeta) {
						if (typeof nanobotMeta.order === 'number') {
							workspace.order = nanobotMeta.order;
						}
						if (typeof nanobotMeta.color === 'string') {
							workspace.color = nanobotMeta.color;
						}
					}

					workspaces.push(workspace);
				} catch (error) {
					console.error('Failed to parse workspace resource:', error);
				}
			}
		}

		return workspaces;
	}

	/**
	 * Create a new workspace
	 */
	async createWorkspace(workspace: Omit<Workspace, 'id' | 'created'>): Promise<Workspace> {
		return await this.client.callMCPTool<Workspace>('create_workspace', {
			payload: workspace
		});
	}

	/**
	 * Update an existing workspace
	 */
	async updateWorkspace(
		workspaceId: string,
		updates: Partial<Omit<Workspace, 'id' | 'created'>>
	): Promise<Workspace> {
		return await this.client.callMCPTool<Workspace>('update_workspace', {
			payload: {
				uri: `nanobot://workspaces/${workspaceId}`,
				...updates
			}
		});
	}

	/**
	 * Delete a workspace
	 */
	async deleteWorkspace(workspaceId: string): Promise<void> {
		await this.client.callMCPTool('delete_workspace', {
			payload: {
				uri: `nanobot://workspaces/${workspaceId}`
			}
		});
	}
}

// Default instance
export const workspaceService = new WorkspaceService();
