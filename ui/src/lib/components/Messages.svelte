<script lang="ts">
	import Message from './Message.svelte';
	import type { ChatMessage } from '$lib/types';

	interface Props {
		messages: ChatMessage[];
		onSend?: (message: string) => void;
	}

	let { messages, onSend }: Props = $props();
</script>

<div class="flex flex-col space-y-4">
	{#each messages as message (message.id)}
		<Message {message} {onSend} />
	{/each}
</div>
