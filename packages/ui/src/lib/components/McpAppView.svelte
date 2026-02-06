<script lang="ts">
import { onMount } from "svelte";
import { AppRenderer } from "@mcp-ui/client";
import type { CallToolResult, ReadResourceResult, ResourceContents } from "@modelcontextprotocol/sdk/types.js";
import React from "react";
import ReactDOM from "react-dom/client";
import type {
	Attachment,
	ChatMessageItemToolCall,
	ChatResult,
} from "$lib/types";

interface Props {
	item: ChatMessageItemToolCall;
	resourceUri: string;
	onSend?: (
		message: string,
		attachments?: Attachment[],
	) => Promise<ChatResult | void>;
	onReadResource?: (
		uri: string,
		opts?: { abort?: AbortController }
	) => Promise<{ contents: ResourceContents[] }>;
}

const { item, resourceUri, onSend, onReadResource }: Props = $props();
let container: HTMLDivElement | undefined = $state();

// MCP Apps Sandbox Configuration
// Default: Uses same-origin (/sandbox.html) which triggers fallback mode
//          This simplifies setup but shows console warnings (expected/harmless)
// Production: Set VITE_SANDBOX_URL environment variable for cross-origin sandbox
//             Example: VITE_SANDBOX_URL=https://sandbox.example.com/sandbox.html
const SANDBOX_URL = import.meta.env.VITE_SANDBOX_URL || "http://localhost:8080/sandbox.html";
console.warn(`SANDBOX_URL=${SANDBOX_URL}`)

onMount(() => {
	if (!container || !resourceUri || !item.name) return;

	console.warn(`resourceUri: ${resourceUri}`);
	console.warn(`item: ${JSON.stringify(item, null, 2)}`);

	const toolInput = item.arguments ? JSON.parse(item.arguments) : {};
	const toolResult = item.output
		? {
				content: item.output.content || [],
				structuredContent: item.output.structuredContent,
				isError: item.output.isError ?? false,
			}
		: undefined;

	const root = ReactDOM.createRoot(container);
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
			onCallTool: async (params): Promise<CallToolResult> => {
				console.warn(`MCP App --> Call Tool: ${params.name}`);
				if (!onSend) return { content: [], isError: true };
				const uiAction = {
					type: "tool",
					payload: { toolName: params.name, params: params.arguments || {} },
				};
				const reply = await onSend(JSON.stringify(uiAction));
				if (reply) {
					for (const item of reply.message?.items || []) {
						if (item.type === "tool" && item.output) {
						    console.warn(`Good Call Tool Result: \n${JSON.stringify(item.output, null, 2)}`)
							return $state.snapshot(item.output) as CallToolResult;
						}
					}
					console.warn(`After For Call Tool Result: \n${JSON.stringify(reply, null, 2)}`)
				}
				console.warn(`After If Call Tool Result: \n${JSON.stringify(reply, null, 2)}`)
				return { content: [], isError: true };
			},
			onOpenLink: async ({ url }) => {
				console.warn(`MCP App --> Open Link: ${url}`);
				if (url.startsWith("http://") || url.startsWith("https://")) {
					window.open(url, "_blank", "noopener,noreferrer");
				}
				return { isError: false };
			},
			onReadResource: async (params): Promise<ReadResourceResult> => {
				console.warn(`MCP App --> Read Resource: ${params.uri}`);
				if (!onReadResource) {
					return { contents: [] };
				}
				const uri = params.uri as string;
				const result = await onReadResource(uri);
				return result as ReadResourceResult;
			},
		}),
	);
});
</script>

{#if resourceUri}
	<div bind:this={container} class="w-full min-h-50"></div>
{/if}
