<script lang="ts">
import { AppRenderer } from "@mcp-ui/client";
import type {
	CallToolResult,
	ReadResourceResult,
} from "@modelcontextprotocol/sdk/types.js";
import React from "react";
import ReactDOM from "react-dom/client";
import { getContext } from "svelte";
import type { SimpleClient } from "$lib/mcpclient";
import type { ChatMessageItemToolCall } from "$lib/types";

interface Props {
	item: ChatMessageItemToolCall;
}

let { item }: Props = $props();
let container: HTMLDivElement | undefined = $state();
let root: ReactDOM.Root | null = null;

const mcpClient = getContext<SimpleClient>("mcpClient");
const SANDBOX_URL =
	import.meta.env.VITE_SANDBOX_URL || "http://localhost:8081/sandbox.html";

const resourceUri = $derived(item._meta?.ui?.resourceUri);

$effect(() => {
	if (!container || !resourceUri || !item.name) return;

	const toolInput = item.arguments ? JSON.parse(item.arguments) : {};
	const toolResult = item.output
		? {
				content: item.output.content || [],
				structuredContent: item.output.structuredContent,
				isError: item.output.isError ?? false,
			}
		: undefined;

	root = ReactDOM.createRoot(container);
	root.render(
		React.createElement(AppRenderer, {
			toolName: item.name,
			toolResourceUri: resourceUri,
			sandbox: { url: new URL(SANDBOX_URL) },
			toolInput,
			toolResult: toolResult as CallToolResult | undefined,
			hostContext: {
				theme:
					document.documentElement.dataset.theme === "dark" ? "dark" : "light",
				platform: "web",
			},
			// Callback handlers instead of MCP client
			onReadResource: async ({ uri }): Promise<ReadResourceResult> => {
				const result = await mcpClient.readResource(uri);
				return result as ReadResourceResult;
			},
			onCallTool: async (params): Promise<CallToolResult> => {
				const result = (await mcpClient.exchange("tools/call", {
					name: params.name,
					arguments: params.arguments,
				})) as CallToolResult;

				return result;
			},
			onOpenLink: async ({ url }) => {
				if (url.startsWith("http://") || url.startsWith("https://")) {
					window.open(url, "_blank", "noopener,noreferrer");
				}
				return { isError: false };
			},
			onMessage: async (params) => {
				console.log("MCP App message:", params);
				return { isError: false };
			},
		}),
	);

	return () => {
		root?.unmount();
		root = null;
	};
});
</script>

{#if resourceUri}
	<div bind:this={container} class="w-full min-h-50"></div>
{/if}
