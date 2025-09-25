<script lang="ts">
	import { ExternalLink } from '@lucide/svelte';
	import type { ChatMessageItemResourceLink } from '$lib/types';

	interface Props {
		item: ChatMessageItemResourceLink;
	}

	let { item }: Props = $props();
</script>

<div class="mb-3 rounded-lg bg-base-200 p-3">
	<div class="flex items-center gap-2 text-sm">
		<ExternalLink class="h-4 w-4 text-primary" />
		<span>{item.name || item.uri}</span>
	</div>
	{#if item.description}
		<p class="mt-1 text-xs opacity-70">{item.description}</p>
	{/if}
</div>
