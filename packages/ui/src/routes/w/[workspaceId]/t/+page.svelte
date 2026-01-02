<script lang="ts">
	import { page } from '$app/state';
	import TaskEditor from './TaskEditor.svelte';
	import TaskRunner from './TaskRunner.svelte';

    let { data } = $props();
    let workspaceId = $derived(data.workspaceId);
    let urlTaskId = $derived(page.url.searchParams.get('id') ?? '');

    const editor = false;
</script>

<svelte:head>
    <title>Nanobot | Tasks</title>
</svelte:head>

{#if editor}
    <TaskEditor {workspaceId} {urlTaskId} />
{:else}
    <TaskRunner {workspaceId} {urlTaskId} />
{/if}