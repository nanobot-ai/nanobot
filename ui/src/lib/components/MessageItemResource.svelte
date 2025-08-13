<script lang="ts">
	import { FileText } from '@lucide/svelte';
	import type { ChatMessageItemResource } from '$lib/types';

	interface Props {
		item: ChatMessageItemResource;
	}

	let { item }: Props = $props();
</script>

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
		<pre class="mt-2 overflow-x-auto rounded bg-base-100 p-2 text-xs">{item.resource.text}</pre>
	{/if}
</div>
