<script lang="ts">
	import Message from './Message.svelte';
	import type { ChatMessage } from '$lib/types';

	interface Props {
		messages: ChatMessage[];
		onSend?: (message: string) => void;
		isLoading?: boolean;
	}

	let { messages, onSend, isLoading = false }: Props = $props();
	let lastUserIDIndex = $derived(messages.findLastIndex((message) => message.role === 'user'));
	let lastUserMessageID = $derived(messages[lastUserIDIndex]?.id);
	let messagesBeforeLastUser = $derived(messages.slice(0, lastUserIDIndex));
	let messagesAfterIncludingLastUser = $derived(messages.slice(lastUserIDIndex));

	// Check if any messages have content (text items)
	let hasMessageContent = $derived(
		messagesAfterIncludingLastUser.some(
			(message) =>
				message.role === 'assistant' &&
				message.items &&
				message.items.some(
					(item) =>
						item.type === 'tool' ||
						(item.type === 'text' && item.text && item.text.trim().length > 0)
				)
		)
	);

	// Show loading indicator when loading and no content has been printed yet
	let showLoadingIndicator = $derived(isLoading && !hasMessageContent);
</script>

<div class="flex flex-col space-y-4 pt-4">
	{#each messagesBeforeLastUser as message (message.id)}
		<Message {message} {onSend} />
	{/each}
	<div id="last" class="min-h-[calc(100vh-2rem)]" data-message-id={lastUserMessageID}>
		{#each messagesAfterIncludingLastUser as message (message.id)}
			<Message {message} {onSend} />
		{/each}
		{#if showLoadingIndicator}
			<div class="flex w-full items-start gap-3">
				<div class="flex min-w-0 flex-1 flex-col items-start">
					<div class="flex items-center justify-center p-8">
						<span class="loading loading-lg loading-spinner text-base-content/30"></span>
					</div>
				</div>
			</div>
		{/if}
		{#if messages.length > 0}
			<div class="h-59"></div>
		{/if}
	</div>
</div>
