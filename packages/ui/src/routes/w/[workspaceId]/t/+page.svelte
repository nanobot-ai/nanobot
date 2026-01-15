<script lang="ts">
	import { page } from '$app/state';
	import { ChatService } from '$lib/chat.svelte';
	import { setSharedChat } from '$lib/stores/chat.svelte';
	import { onDestroy, onMount } from 'svelte';
	import TaskEditor from './TaskEditor.svelte';
	import TaskRunner from './TaskRunner.svelte';
	import type { WorkspaceClient } from '$lib/types';
	import TaskRun from './TaskRun.svelte';
	import { createRegistryStore, setRegistryContext } from '$lib/context/registry.svelte';
	import { getWorkspaceService } from '$lib/stores/workspace.svelte';

    let { data } = $props();
    let workspaceId = $derived(data.workspaceId);
    let urlTaskId = $derived(page.url.searchParams.get('id') ?? '');
    let runOnly = $derived(page.url.searchParams.get('run') === 'true');
    let runId = $derived(page.url.searchParams.get('runId') ?? '');

    const workspaceService = getWorkspaceService();
    let chat = $state<ChatService | null>(null);
    let workspace = $state<WorkspaceClient | null>(null);
    let chatInitialized = $state(false);

    const registryStore = createRegistryStore();
	setRegistryContext(registryStore);

    onMount(() => {
        registryStore.fetch();
    })

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
    {#if runId}
        <TaskRun {workspace} {urlTaskId} {runId} />
    {:else if runOnly}
        <TaskRunner {workspace} {urlTaskId} />
    {:else}
        <TaskEditor {workspace} {urlTaskId} />
    {/if}
{/if}

<svelte:head>
    <title>Nanobot | Workflows</title>
</svelte:head>