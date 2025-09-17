<script lang="ts">
	import type { ChatMessage, ChatMessageItemResource } from '$lib/types';
	import { isUIResource } from '@mcp-ui/client';
	import MessageItemUI from '$lib/components/MessageItemUI.svelte';

	interface Props {
		messages: ChatMessage[];
		onSendMessage?: (message: string) => void;
		isLoading?: boolean;
	}

	let { messages, onSendMessage }: Props = $props();

	let sidecar = $derived.by(() => {
		for (const message of messages.toReversed()) {
			for (const item of (message.items ?? []).toReversed()) {
				if (item.type === 'tool' && item.output && item.output?.content) {
					for (const output of item.output.content.toReversed()) {
						if (isUIResource(output) && output.resource._meta?.['ai.nanobot.meta/workspace']) {
							return output satisfies ChatMessageItemResource;
						}
					}
				}
			}
		}
		return null;
	});

	let key = $derived(sidecar?.resource?.text ?? sidecar?.resource?.blob ?? '');
</script>

{#if key}
	{#key key}
		<div class="workspace peer h-dvh w-3/4">
			<MessageItemUI item={sidecar} onSend={onSendMessage} />
		</div>
	{/key}
{/if}
