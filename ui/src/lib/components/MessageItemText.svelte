<script lang="ts">
	import { renderMarkdown } from '$lib/markdown';
	import type { ChatMessageItemText } from '$lib/types';

	interface Props {
		item: ChatMessageItemText;
		role: 'user' | 'assistant';
	}

	let { item, role }: Props = $props();

	const renderedContent = $derived(role === 'assistant' ? renderMarkdown(item.text) : item.text);
</script>

<div class="prose mb-3 w-full max-w-none rounded-lg bg-base-200 p-3 prose-invert">
	{@html renderedContent}
</div>
