<script lang="ts">
	import '$lib/../app.css';
	import Thread from '$lib/components/Thread.svelte';
	import { ChatService } from '$lib/chat.svelte';
	import { onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	const chat = new ChatService();
	let doClose = true;

	$effect(() => {
		if (chat.chatId) {
			doClose = false;
			page.data.chat = chat;
			goto(`/c/${chat.chatId}`);
		}
	});

	onDestroy(() => {
		if (doClose) {
			chat.close();
		}
	});
</script>

<Thread
	messages={chat.messages}
	isLoading={chat.isLoading}
	onSendMessage={chat.sendMessage}
	prompts={chat.prompts}
	elicitations={chat.elicitations}
	onElicitationResult={chat.replyToElicitation}
/>
