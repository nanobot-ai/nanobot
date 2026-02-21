<script lang="ts">
	import { Settings } from '@lucide/svelte';
	import { renderMarkdown } from '$lib/markdown';
	import type { Attachment, ChatResult, ChatMessageItemToolCall } from '$lib/types';
	import { getMcpAppsContext } from '$lib/context/mcpApps.svelte';
	import MessageItemApp from './MessageItemApp.svelte';

	interface Props {
		item: ChatMessageItemToolCall;
		onSend?: (message: string, attachments?: Attachment[]) => Promise<ChatResult | void>;
	}

	let { item, onSend }: Props = $props();

	const { tools } = getMcpAppsContext();
	const toolResourceUri = $derived(
		tools?.find((t) => t.name === item.name)?._meta?.ui?.resourceUri
	);

	function parseToolInput(input: string) {
		try {
			const parsed = JSON.parse(input);
			if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
				return { success: true, data: parsed };
			}
		} catch {
			// JSON parsing failed, fall back to string display
		}
		return { success: false, data: input };
	}

	function parseToolOutput(output: string) {
		try {
			const parsed = JSON.parse(output);
			const formattedJson = JSON.stringify(parsed, null, 2);
			const highlightedJson = renderMarkdown('```json\n' + formattedJson + '\n```');
			return { success: true, data: highlightedJson };
		} catch {
			// JSON parsing failed, fall back to string display
		}
		return { success: false, data: output };
	}
</script>

<div
	class="text collapse mt-3 mb-2 w-full border border-base-100 bg-base-100 hover:collapse-arrow hover:border-base-300"
>
	<input type="checkbox" />
	<div class="collapse-title">
		<div class="flex items-center gap-2">
			{#if item.output}
				<Settings class="h-4 w-4 text-primary/60" />
			{:else}
				<span class="loading loading-xs loading-spinner"></span>
			{/if}
			<span class="text-sm font-medium text-primary/60">Tool call: {item.name}</span>
		</div>
	</div>
	<div class="collapse-content">
		<div class="space-y-3 pt-2">
			{#if item.arguments}
				<div class="grid">
					<div class="mb-1 text-xs font-medium text-base-content/70">Input:</div>
					{#if parseToolInput(item.arguments).success}
						<div class="overflow-x-auto rounded bg-base-200 p-3">
							<table class="table w-full table-zebra table-xs">
								<thead>
									<tr>
										<th class="text-xs">Key</th>
										<th class="text-xs">Value</th>
									</tr>
								</thead>
								<tbody>
									{#each Object.entries(parseToolInput(item.arguments).data) as [key, value] (key)}
										<tr>
											<td class="font-mono text-xs">{key}</td>
											<td class="font-mono text-xs break-all">
												{typeof value === 'object' ? JSON.stringify(value) : String(value)}
											</td>
										</tr>
									{/each}
									{#if Object.keys(parseToolInput(item.arguments).data).length === 0}
										<tr>
											<td class="font-mono text-xs">No arguments</td>
										</tr>
									{/if}
								</tbody>
							</table>
						</div>
					{:else}
						<div class="overflow-x-auto rounded bg-base-200 p-3 font-mono text-sm">
							{item.arguments}
						</div>
					{/if}
				</div>
			{/if}
			{#if item.output}
				<div class="flex flex-col">
					<div class="mb-1 text-xs font-medium text-base-content/70">Output:</div>
					{#if item.output.isError}
						<div class="alert alert-error">
							<span>Tool execution failed</span>
						</div>
					{/if}
					{#if item.output.structuredContent}
						<div
							class="prose overflow-x-auto rounded border border-success/20 bg-success/10 p-3 prose-invert"
						>
							{@html parseToolOutput(JSON.stringify(item.output.structuredContent)).data}
						</div>
					{/if}
					{#if item.output.content}
						<!-- Render tool output content items -->
						{#each item.output.content as contentItem, i (i)}
							<div
								class="prose overflow-x-auto rounded border border-success/20 bg-success/10 p-3 prose-invert"
							>
								{#if contentItem.type === 'text' && 'text' in contentItem && parseToolOutput(contentItem.text).success}
									<!-- JSON Syntax Highlighted Display -->
									{@html parseToolOutput(contentItem.text).data}
								{:else if contentItem.type === 'text' && 'text' in contentItem && contentItem.text}
									{@html renderMarkdown(contentItem.text)}
								{:else}
									{@html parseToolOutput(JSON.stringify(contentItem)).data}
								{/if}
							</div>
						{/each}
					{/if}
				</div>
			{:else}
				<div class="flex items-center gap-2 text-xs text-base-content/50 italic">
					<span class="loading loading-xs loading-spinner"></span>
					Running...
				</div>
			{/if}
		</div>
	</div>
</div>

{#if toolResourceUri && item.output}
	<MessageItemApp {item} resourceUri={toolResourceUri} />
{/if}

<div class="flex w-full flex-wrap items-start justify-start gap-2 p-2">
	{#if item.output && item.output.content}
		{#each item.output.content as contentItem, i (i)}
			{#if contentItem.type === 'image'}
				<img
					src="data:{contentItem.mimeType};base64,{contentItem.data}"
					alt="Tool output"
					class="max-w-full rounded"
				/>
			{/if}
		{/each}
	{/if}
</div>
