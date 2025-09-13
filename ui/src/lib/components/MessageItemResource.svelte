<script lang="ts">
	import { FileText, AlertTriangle } from '@lucide/svelte';
	import type { ChatMessageItemResource } from '$lib/types';

	interface Props {
		item: ChatMessageItemResource;
	}

	let { item }: Props = $props();

	const isError = item.resource.mimeType === 'application/vnd.nanobot.error+json';
</script>

{#if isError}
	<div class="mb-3 rounded-lg border border-error/20 bg-error/10 p-3">
		<div class="mb-2 flex items-center gap-2 text-sm">
			<AlertTriangle class="h-4 w-4 text-error" />
			<span class="font-medium text-error">Error</span>
		</div>
		{#if item.resource.text}
			<pre
				class="mt-2 rounded bg-base-100 p-2 text-xs break-all whitespace-pre-wrap text-error">{item
					.resource.text}</pre>
		{/if}
	</div>
{:else}
	<div class="mb-3 rounded-lg bg-base-200 p-3">
		<div class="mb-2 flex items-center gap-2 text-sm">
			<FileText class="h-4 w-4 text-secondary" />
			<span class="font-medium">Resource</span>
			<span class="badge badge-sm">{item.resource.mimeType}</span>
		</div>
		<a
			href={item.resource.uri}
			class="link text-sm link-primary"
			target="_blank"
			rel="noopener noreferrer"
		>
			{item.resource.uri}
		</a>
		{#if item.resource.text}
			<pre class="mt-2 rounded bg-base-100 p-2 text-xs break-all whitespace-pre-wrap">{item.resource
					.text}</pre>
		{/if}
	</div>
{/if}
