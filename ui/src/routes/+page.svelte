<script lang="ts">
	import '$lib/../app.css';
	import Thread from '$lib/components/Thread.svelte';
	import { ChatService } from '$lib/chat.svelte';
	import { onDestroy, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';

	const chat = new ChatService();
	let doClose = true;

	$effect(() => {
		if (chat.chatId) {
			doClose = false;
			page.data.chat = chat;
			goto(resolve(`/c/${chat.chatId}`));
		}
	});

	onMount(() => {
		if (window.location.search.includes('new')) {
			chat.newChat();
		}
	});

	onDestroy(() => {
		if (doClose) {
			chat.close();
		}
	});
</script>

<svelte:head>
	{#if chat.agent?.name}
		<title>{chat.agent.name}</title>
	{:else}
		<title>Nanobot</title>
	{/if}
</svelte:head>

<Thread
	messages={chat.messages}
	isLoading={chat.isLoading}
	onSendMessage={chat.sendMessage}
	prompts={chat.prompts}
	elicitations={chat.elicitations}
	agent={chat.agent}
	onElicitationResult={chat.replyToElicitation}
/>
