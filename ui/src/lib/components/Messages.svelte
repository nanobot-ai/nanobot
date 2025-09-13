<script lang="ts">
	import Message from './Message.svelte';
	import type { ChatMessage } from '$lib/types';

	interface Props {
		messages: ChatMessage[];
		onSend?: (message: string) => void;
	}

	let { messages, onSend }: Props = $props();
	let lastUserIDIndex = $derived(messages.findLastIndex((message) => message.role === 'user'));
	let lastUserMessageID = $derived(messages[lastUserIDIndex]?.id);
	let messagesBeforeLastUser = $derived(messages.slice(0, lastUserIDIndex));
	let messagesAfterIncludingLastUser = $derived(messages.slice(lastUserIDIndex));
</script>

<div class="flex flex-col space-y-4 pt-4">
	{#each messagesBeforeLastUser as message (message.id)}
		<Message {message} {onSend} />
	{/each}
	<div id="last" class="min-h-[calc(100vh-2rem)]" data-message-id={lastUserMessageID}>
		{#each messagesAfterIncludingLastUser as message (message.id)}
			<Message {message} {onSend} />
		{/each}
		{#if messages.length > 0}
			<div class="h-59"></div>
		{/if}
	</div>
</div>
