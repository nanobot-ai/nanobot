<script lang="ts">
	import '$lib/../app.css';
	import { page } from '$app/state';
	import Thread from '$lib/components/Thread.svelte';
	import { ChatService } from '$lib/chat.svelte';
	import { onDestroy } from 'svelte';
	import { getNotificationContext } from '$lib/context/notifications.svelte';
	import Workspace from '$lib/components/Workspace.svelte';

	// The existing chat might have been set by / so don't recreate it because that will
	// loose the event stream.
	const chat = page.data.chat || new ChatService();
	const notification = getNotificationContext();

	$effect(() => {
		if (!page.params.id) return;
		chat.setChatId(page.params.id).catch((e) => {
			console.error('Error setting chat ID:', e);
			notification.error(e.message);
		});
	});

	onDestroy(() => {
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

<div class="grid grid-cols-1 md:flex md:flex-row">
	<Workspace messages={chat.messages} onSendMessage={chat.sendMessage} />

	<Thread
		messages={chat.messages}
		isLoading={chat.isLoading}
		onFileUpload={chat.uploadFile}
		onSendMessage={chat.sendMessage}
		cancelUpload={chat.cancelUpload}
		prompts={chat.prompts}
		resources={chat.resources}
		elicitations={chat.elicitations}
		agent={chat.agent}
		uploadingFiles={chat.uploadingFiles}
		uploadedFiles={chat.uploadedFiles}
		onElicitationResult={chat.replyToElicitation}
	/>
</div>
