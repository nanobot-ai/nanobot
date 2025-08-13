<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import nanobotLogo from '$lib/assets/nanobot.svg';
	import Threads from '$lib/components/Threads.svelte';
	import Notifications from '$lib/components/Notifications.svelte';
	import { chatApi } from '$lib/chat.svelte';
	import { notifications } from '$lib/stores/notifications.svelte';
	import { setNotificationContext } from '$lib/context/notifications.svelte';
	import { onMount } from 'svelte';
	import type { Chat } from '$lib/types';

	let { children } = $props();

	let threads = $state<Chat[]>([]);
	let isLoading = $state(true);

	// Set notification context for global access
	setNotificationContext(notifications);

	onMount(async () => {
		threads = await chatApi.getThreads();
		isLoading = false;
	});

	async function handleRenameThread(threadId: string, newTitle: string) {
		try {
			await chatApi.renameThread(threadId, newTitle);
			const threadIndex = threads.findIndex((t) => t.id === threadId);
			if (threadIndex !== -1) {
				threads[threadIndex].title = newTitle;
			}
			notifications.success('Thread Renamed', `Successfully renamed to "${newTitle}"`);
		} catch (error) {
			notifications.error('Rename Failed', 'Unable to rename the thread. Please try again.');
			console.error('Failed to rename thread:', error);
		}
	}

	async function handleDeleteThread(threadId: string) {
		try {
			await chatApi.deleteThread(threadId);
			const threadToDelete = threads.find((t) => t.id === threadId);
			threads = threads.filter((t) => t.id !== threadId);
			notifications.success('Thread Deleted', `Deleted "${threadToDelete?.title || 'thread'}"`);
		} catch (error) {
			notifications.error('Delete Failed', 'Unable to delete the thread. Please try again.');
			console.error('Failed to delete thread:', error);
		}
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div id="sidebar" class="drawer lg:drawer-open">
	<input id="my-drawer" type="checkbox" class="drawer-toggle" />
	<div class="drawer-content h-dvh">
		<!--		<label for="my-drawer" class="btn btn-primary drawer-button">Open drawer</label>-->
		<!-- Chat area - takes remaining space -->
		{@render children?.()}
	</div>

	<div class="drawer-side">
		<label for="my-drawer" aria-label="close sidebar" class="drawer-overlay"></label>
		<div class="min-h-full w-80 bg-base-200 p-4">
			<!-- Logo at the top of sidebar -->
			<a href="/" class="flex items-center gap-2 text-xl font-bold hover:opacity-80">
				<img src={nanobotLogo} alt="Nanobot" class="h-12" />
			</a>
			<Threads {threads} onRename={handleRenameThread} onDelete={handleDeleteThread} {isLoading} />
		</div>
	</div>
</div>

<!-- Notifications -->
<Notifications />
