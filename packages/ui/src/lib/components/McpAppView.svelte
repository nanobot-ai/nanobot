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
	onToolCall?: (
		toolName: string,
		args: Record<string, unknown>,
		opts?: { abort?: AbortController }
	) => Promise<CallToolResult>;
}

const { item, resourceUri, onSend, onReadResource, onToolCall }: Props = $props();
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
				if (!onToolCall) {
					console.error('onToolCall callback not provided');
					return { content: [], isError: true };
				}
				try {
					const result = await onToolCall(
						params.name,
						(params.arguments as Record<string, unknown>) || {}
					);
					console.warn(`Call Tool Result: \n${JSON.stringify(result, null, 2)}`);
					return result
				} catch (error) {
					console.error(`Tool call failed:`, error);
					return {
						content: [{
							type: 'text',
							text: `Tool call failed: ${error instanceof Error ? error.message : String(error)}`
						}],
						isError: true
					} as CallToolResult;
				}
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
			onError: (error) => {
			  console.error(`MCP App --> Error: ${error}`);

			}
		}),
	);
});
</script>

{#if resourceUri}
	<div bind:this={container} class="w-full min-h-50"></div>
{/if}
