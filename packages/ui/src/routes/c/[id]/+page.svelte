<script lang="ts">
	import '$lib/../app.css';
	import { page } from '$app/state';
	import { ChatService } from '$lib/chat.svelte';
	import { onDestroy } from 'svelte';
	import { getNotificationContext } from '$lib/context/notifications.svelte';
	import { sharedChat, setSharedChat, clearSharedChat } from '$lib/stores/chat.svelte';
	import Workspace from '$lib/components/Workspace.svelte';
	import ThreadFromChat from "$lib/components/ThreadFromChat.svelte";

	// Reuse shared chat if available (e.g., from / page), otherwise create new
	const chat = sharedChat.current && sharedChat.current.chatId === page.params.id ? sharedChat.current : new ChatService();
	
	// Ensure the chat is always shared (for direct navigation to /c/[id])
	if (!chat.chatId) {
		setSharedChat(chat);
	}
	
	const notification = getNotificationContext();

	$effect(() => {
		if (!page.params.id) return;
		chat.setChatId(page.params.id).catch((e) => {
			console.error('Error setting chat ID:', e);
			notification.error(e.message);
		});
	});

	onDestroy(() => {
		clearSharedChat();
		chat.close();
	});
</script>

<svelte:head>
	{#if chat.agent?.name}
		<title>{chat.agent.name}</title>
	{:else}
		<title>Nanobot</title>
	{/if}
</svelte:head>

<div class="grid grid-cols-1 md:flex md:flex-row md:grow">
	<Workspace messages={chat.messages} onSendMessage={chat.sendMessage} />
	<ThreadFromChat {chat}/>
</div>
