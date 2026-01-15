// Shared workspace state that persists across page navigations
import { WorkspaceInstance, WorkspaceService } from '$lib/workspace.svelte';
import { SvelteMap } from 'svelte/reactivity';

const workspaceServiceState = $state<{ current: WorkspaceService }>({
	current: new WorkspaceService()
});

// Shared map of workspace instances (workspaceId -> WorkspaceInstance)
const workspaceDataState = $state<{ current: SvelteMap<string, WorkspaceInstance> }>({
	current: new SvelteMap<string, WorkspaceInstance>()
});

// Loading state for individual workspaces
const loadingWorkspaceState = $state<{ current: SvelteMap<string, boolean> }>({
	current: new SvelteMap<string, boolean>()
});

// Initialize and load workspaces
workspaceServiceState.current.load();

export function getWorkspaceService(): WorkspaceService {
	return workspaceServiceState.current;
}

export function getWorkspaceData(): SvelteMap<string, WorkspaceInstance> {
	return workspaceDataState.current;
}

export function getLoadingWorkspace(): SvelteMap<string, boolean> {
	return loadingWorkspaceState.current;
}

export function reload(): void {
	workspaceServiceState.current.load();
}

export async function loadWorkspaceInstance(workspaceId: string): Promise<WorkspaceInstance> {
	const service = workspaceServiceState.current;
	const data = workspaceDataState.current;
	const loading = loadingWorkspaceState.current;

	let instance = data.get(workspaceId);

	if (!instance) {
		loading.set(workspaceId, true);
		instance = service.getWorkspace(workspaceId) as WorkspaceInstance;
		data.set(workspaceId, instance);

		try {
			await instance.load();
		} finally {
			loading.set(workspaceId, false);
		}
	}

	return instance;
}

export function isWorkspaceLoading(workspaceId: string): boolean {
	return loadingWorkspaceState.current.get(workspaceId) ?? false;
}

export function clearWorkspaceInstance(workspaceId: string): void {
	workspaceDataState.current.delete(workspaceId);
	loadingWorkspaceState.current.delete(workspaceId);
}

export function clearAllWorkspaceInstances(): void {
	workspaceDataState.current.clear();
	loadingWorkspaceState.current.clear();
}
