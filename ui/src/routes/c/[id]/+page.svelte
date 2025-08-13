<script lang="ts">
	import '$lib/../app.css';
	import { page } from '$app/state';
	import Thread from '$lib/components/Thread.svelte';
	import { ChatService } from '$lib/chat.svelte';
	import { onDestroy } from 'svelte';

	// The existing chat might have been set by / so don't recreate it because that will
	// loose the event stream.
	const chat = page.data.chat || new ChatService();

	$effect(() => {
		if (!page.params.id) return;
		chat.setChatId(page.params.id);
	});

	onDestroy(() => {
		chat.close();
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
