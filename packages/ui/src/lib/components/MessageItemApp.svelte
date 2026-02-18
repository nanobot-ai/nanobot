<script lang="ts">
	import { AppRenderer } from '@mcp-ui/client';
	import type { CallToolResult } from '@modelcontextprotocol/sdk/types.js';
	import React from 'react';
	import ReactDOM from 'react-dom/client';
	import { onMount, onDestroy } from 'svelte';
	import { getMcpAppsContext } from '$lib/context/mcpApps.svelte';
	import type { ChatMessageItemToolCall } from '$lib/types';

	interface Props {
		item: ChatMessageItemToolCall;
		resourceUri: string;
	}

	let { item, resourceUri }: Props = $props();
	let container: HTMLDivElement;
	let reactRoot: ReactDOM.Root | undefined;

	const mcpApps = getMcpAppsContext();
	const sandboxUrl = new URL('/sandbox_proxy.html', window.location.origin);

	onMount(async () => {
		if (!container) return;
		const client = await mcpApps.ensureClient();
		reactRoot = ReactDOM.createRoot(container);

		let toolResult: CallToolResult | undefined;
		if (item.output) {
			toolResult = {
				content: (item.output.content ?? []).map((c) => {
					if (c.type === 'text' && 'text' in c) return { type: 'text' as const, text: c.text };
					if (c.type === 'image' && 'data' in c)
						return { type: 'image' as const, data: c.data, mimeType: c.mimeType };
					return { type: 'text' as const, text: JSON.stringify(c) };
				}),
				// Deep-clone to strip Svelte 5 reactivity Proxies — postMessage can't clone Proxy objects.
				structuredContent: item.output.structuredContent
					? JSON.parse(JSON.stringify(item.output.structuredContent))
					: undefined,
				isError: item.output.isError
			};
		}

		let toolInput: Record<string, unknown> | undefined;
		if (item.arguments) {
			try {
				toolInput = JSON.parse(item.arguments);
			} catch {
				/* ignore */
			}
		}

		reactRoot.render(
			React.createElement(AppRenderer, {
				client,
				toolName: item.name || '',
				toolResourceUri: resourceUri,
				sandbox: { url: sandboxUrl },
				toolInput,
				toolResult,
				onOpenLink: async ({ url }: { url: string }) => {
					window.open(url, '_blank');
					return {};
				},
				onError: (error: Error) => console.error('[MCP App Error]', error)
			})
		);
	});

	onDestroy(() => {
		reactRoot?.unmount();
	});
</script>

<div bind:this={container} class="w-full rounded border border-base-300 overflow-hidden"></div>
