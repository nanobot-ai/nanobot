import { getContext, setContext } from 'svelte';
import type { SidebarStore } from '$lib/stores/sidebar.svelte';

const SIDEBAR_KEY = Symbol('sidebar');

export function setSidebarContext(sidebar: SidebarStore) {
	setContext(SIDEBAR_KEY, sidebar);
}

export function getSidebarContext(): SidebarStore {
	return getContext<SidebarStore>(SIDEBAR_KEY);
}
