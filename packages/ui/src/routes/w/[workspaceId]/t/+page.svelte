<script lang="ts">
	import { page } from '$app/state';
	import { ChatService } from '$lib/chat.svelte';
	import { setSharedChat } from '$lib/stores/chat.svelte';
	import { onDestroy } from 'svelte';
	import TaskEditor from './TaskEditor.svelte';
	import TaskRunner from './TaskRunner.svelte';
	import { WorkspaceService } from '$lib/workspace.svelte';
	import type { WorkspaceClient } from '$lib/types';

    let { data } = $props();
    let workspaceId = $derived(data.workspaceId);
    let urlTaskId = $derived(page.url.searchParams.get('id') ?? '');
    let runOnly = $derived(page.url.searchParams.get('run') === 'true');
    let runId = $derived(page.url.searchParams.get('runId') ?? '');

    const workspaceService = new WorkspaceService();
    let chat = $state<ChatService | null>(null);
    let workspace = $state<WorkspaceClient | null>(null);
    let chatInitialized = $state(false);

    $effect(() => {
        if (workspaceId) {
            workspace = workspaceService.getWorkspace(workspaceId);
        }
    })

    $effect(() => {
        if (workspace && !chatInitialized) {
            chatInitialized = true;
            chat = new ChatService();
            setSharedChat(chat);
        }
    })

    onDestroy(() => {
        chat?.close();
    });
</script>

{#if workspace && chat}
    {#if runId || runOnly}
        <TaskRunner {workspace} {urlTaskId} {runId} />
    {:else}
        <TaskEditor {workspace} {urlTaskId} />
    {/if}
{/if}

<svelte:head>
    <title>Nanobot | Workflows</title>
</svelte:head>