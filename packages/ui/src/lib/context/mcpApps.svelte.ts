import { getContext, setContext } from 'svelte';
import type { Client } from '@modelcontextprotocol/sdk/client';
import type { ToolDef } from '$lib/types';

interface McpAppsContext {
	readonly client: Client | undefined;
	readonly tools: ToolDef[];
	ensureClient(): Promise<Client | undefined>;
}

const MCP_APPS_KEY = Symbol('mcp-apps');

export function setMcpAppsContext(ctx: McpAppsContext) {
	setContext(MCP_APPS_KEY, ctx);
}

const DEFAULT_CONTEXT: McpAppsContext = {
	client: undefined,
	tools: [],
	ensureClient: () => Promise.resolve(undefined),
};

export function getMcpAppsContext(): McpAppsContext {
	return getContext<McpAppsContext>(MCP_APPS_KEY) ?? DEFAULT_CONTEXT;
}
