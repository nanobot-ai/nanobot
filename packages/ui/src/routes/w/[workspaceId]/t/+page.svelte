<script lang="ts">
	import { page } from '$app/state';
	import TaskEditor from './TaskEditor.svelte';
	import TaskRunner from './TaskRunner.svelte';

    let { data } = $props();
    let workspaceId = $derived(data.workspaceId);
    let urlTaskId = $derived(page.url.searchParams.get('id') ?? '');
    let runOnly = $derived(page.url.searchParams.get('run') === 'true');
</script>

<svelte:head>
    <title>Nanobot | Workflows</title>
</svelte:head>

{#if runOnly}
    <TaskRunner {workspaceId} {urlTaskId} />
{:else}
    <TaskEditor {workspaceId} {urlTaskId} />
{/if}