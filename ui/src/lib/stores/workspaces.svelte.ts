import type { Workspace } from '$lib/types';
import { WorkspaceItemStore } from './workspace-item.svelte';
import { WorkspaceService } from '$lib/workspace.svelte';
import { SvelteSet } from 'svelte/reactivity';

/**
 * WorkspaceStore - Manages workspace list and caches WorkspaceItemStore instances
 * Uses WorkspaceService for backend operations via MCP
 */
class WorkspaceStore {
	workspaces = $state<Workspace[]>([]);
	expandedWorkspaceIds = $state<SvelteSet<string>>(new SvelteSet());
	isLoading = $state(false);

	// Cache of workspace item stores by workspace ID
	// Using regular Map (not SvelteMap) because we don't need this cache to be reactive
	// and we don't want mutations during derived computations
	private itemStoreCache = new Map<string, WorkspaceItemStore>();
	private service: WorkspaceService;

	constructor(opts?: { service?: WorkspaceService }) {
		this.service = opts?.service || new WorkspaceService();
		this.loadExpandedState();
	}

	/**
	 * Get or create a WorkspaceItemStore for a specific workspace
	 * Uses untrack to allow store creation during template evaluation without triggering
	 * state_unsafe_mutation errors
	 */
	getItemStore(workspaceId: string): WorkspaceItemStore {
		// return untrack(() => {
		let store = this.itemStoreCache.get(workspaceId);
		if (!store) {
			store = new WorkspaceItemStore(workspaceId);
			this.itemStoreCache.set(workspaceId, store);
		}
		return store;
		// });
	}

	/**
	 * Clear cached item store for a workspace (useful when workspace is deleted)
	 * Calls close() on the store to cleanup resources before removing from cache
	 */
	clearItemStore(workspaceId: string): void {
		const store = this.itemStoreCache.get(workspaceId);
		if (store) {
			store.close();
			this.itemStoreCache.delete(workspaceId);
		}
	}

	/**
	 * Load workspaces from backend via WorkspaceService
	 */
	async load() {
		this.isLoading = true;
		try {
			this.workspaces = await this.service.listWorkspaces();
		} catch (error) {
			console.error('Failed to load workspaces:', error);
			this.workspaces = [];
		} finally {
			this.isLoading = false;
		}
	}

	/**
	 * Toggle workspace expanded state
	 */
	toggleWorkspace(workspaceId: string) {
		if (this.expandedWorkspaceIds.has(workspaceId)) {
			this.expandedWorkspaceIds.delete(workspaceId);
		} else {
			this.expandedWorkspaceIds.add(workspaceId);
		}
		this.saveExpandedState();
	}

	/**
	 * Check if workspace is expanded
	 */
	isWorkspaceExpanded(workspaceId: string): boolean {
		return this.expandedWorkspaceIds.has(workspaceId);
	}

	/**
	 * Create new workspace via WorkspaceService
	 */
	async createWorkspace(name: string): Promise<Workspace> {
		const newWorkspace = await this.service.createWorkspace({
			name,
			order: this.workspaces.length
		});

		this.workspaces = [...this.workspaces, newWorkspace];
		return newWorkspace;
	}

	/**
	 * Update workspace via WorkspaceService
	 */
	async updateWorkspace(workspaceId: string, data: Partial<Workspace>): Promise<Workspace> {
		const index = this.workspaces.findIndex((w) => w.id === workspaceId);
		if (index === -1) {
			throw new Error('Workspace not found');
		}

		// Filter out id and created from updates (backend doesn't allow these to be changed)
		const { id, created, ...updates } = data;
		void id; // make TS happy
		void created; // make TS happy

		const updatedWorkspace = await this.service.updateWorkspace(workspaceId, updates);

		this.workspaces = [
			...this.workspaces.slice(0, index),
			updatedWorkspace,
			...this.workspaces.slice(index + 1)
		];
		return updatedWorkspace;
	}

	/**
	 * Delete workspace via WorkspaceService
	 */
	async deleteWorkspace(workspaceId: string): Promise<void> {
		await this.service.deleteWorkspace(workspaceId);

		this.workspaces = this.workspaces.filter((w) => w.id !== workspaceId);

		// Clear cached item store for this workspace
		this.clearItemStore(workspaceId);

		// Update expanded state reactively
		this.expandedWorkspaceIds.delete(workspaceId);

		this.saveExpandedState();
	}

	/**
	 * Load expanded state from localStorage
	 */
	private loadExpandedState() {
		try {
			const workspaceState = localStorage.getItem('nanobot-expanded-workspaces');
			if (workspaceState) {
				const ids = JSON.parse(workspaceState) as string[];
				ids.forEach((id) => this.expandedWorkspaceIds.add(id));
			}
		} catch (error) {
			console.error('Failed to load expanded state:', error);
		}
	}

	/**
	 * Save expanded state to localStorage
	 */
	private saveExpandedState() {
		try {
			localStorage.setItem(
				'nanobot-expanded-workspaces',
				JSON.stringify([...this.expandedWorkspaceIds])
			);
		} catch (error) {
			console.error('Failed to save expanded state:', error);
		}
	}
}

export const workspaceStore = new WorkspaceStore();
