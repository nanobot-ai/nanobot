import { getContext, setContext } from 'svelte';

const LAYOUT_CONTEXT_KEY = Symbol('layout');

export interface LayoutContext {
	isSidebarCollapsed: boolean;
	isMobileSidebarOpen: boolean;
}

export function setLayoutContext(context: LayoutContext) {
	setContext(LAYOUT_CONTEXT_KEY, context);
}

export function getLayoutContext(): LayoutContext {
	return getContext<LayoutContext>(LAYOUT_CONTEXT_KEY);
}
