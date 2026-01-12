<script lang="ts">
	import { page } from '$app/state';
	import { ChatService } from '$lib/chat.svelte';
	import { setSharedChat } from '$lib/stores/chat.svelte';
	import { onDestroy } from 'svelte';
	import TaskEditor from './TaskEditor.svelte';
	import TaskRunner from './TaskRunner.svelte';

    let { data } = $props();
    let workspaceId = $derived(data.workspaceId);
    let urlTaskId = $derived(page.url.searchParams.get('id') ?? '');
    let runOnly = $derived(page.url.searchParams.get('run') === 'true');
    let runId = $derived(page.url.searchParams.get('runId') ?? '');

    const chat = new ChatService();
    setSharedChat(chat);

    onDestroy(() => {
        chat.close();
    });
</script>

{#if runId || runOnly}
    <TaskRunner {workspaceId} {urlTaskId} {runId} />
{:else}
    <TaskEditor {workspaceId} {urlTaskId} />
{/if}

<svelte:head>
    <title>Nanobot | Workflows</title>
</svelte:head>