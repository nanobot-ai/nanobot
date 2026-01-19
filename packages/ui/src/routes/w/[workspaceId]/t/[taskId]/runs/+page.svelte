<script lang="ts">
	import { getWorkspaceService } from '$lib/stores/workspace.svelte';
	import type { Session, WorkspaceClient } from '$lib/types.js';
	import { onMount } from 'svelte';
    import { resolve } from '$app/paths';
	import { Trash2 } from '@lucide/svelte';
	import ConfirmDelete from '$lib/components/ConfirmDelete.svelte';
	import { goto } from '$app/navigation';

    let { data } = $props();
    let workspaceId = $derived(data.workspaceId);
    let taskId = $derived(data.taskId);
    let taskName = $state('');

    const workspaceService = getWorkspaceService();
    let workspace = $state<WorkspaceClient | null>(null);
    let taskRuns = $state<Session[]>([]);

    let confirmDeleteRunModal = $state<ReturnType<typeof ConfirmDelete> | null>(null);
    let confirmDeleteRun = $state<Session | null>(null);

    onMount(() => {
        workspace = workspaceService.getWorkspace(workspaceId);
    })

    $effect(() => {
        if (workspace) {
            const task = workspace.files.find((file) => file.name === `.nanobot/tasks/${taskId}/TASK.md`);
            taskName = (task?.file?.task_name || task?.file?.name || taskId) as string;
            taskRuns = workspace.sessions
                .filter((session) => session.parentTaskName === taskId)
                .sort((a, b) => new Date(b.createdAt ?? '').getTime() - new Date(a.createdAt ?? '').getTime());
        }
    })
</script>

<div class="flex flex-col gap-4 min-h-dvh w-full p-4 overflow-y-auto">
    <h2 class="text-2xl font-semibold">{taskName} <span class="text-base-content/50 text-lg font-extralight">Workflow Runs</span></h2>
    <table class="table bg-base-100 dark:bg-base-200 rounded-box">
        <!-- head -->
        <thead>
            <tr>
                <th>Created</th>
                <th>Id</th>
                <th></th>
            </tr>
        </thead>
        <tbody>
            {#each taskRuns as taskRun (taskRun.id)}
                <tr class="cursor-pointer hover:bg-base-200/50 dark:hover:bg-base-300/50" 
                    onclick={() => {
                        goto(resolve(`/w/${workspaceId}/t?id=${taskId}&runId=${taskRun.id}`));
                    }}
                >
                    <td>{taskRun.createdAt}</td>
                    <td>{taskRun.id}</td>
                    <td class="flex justify-end">
                        <button class="btn btn-square btn-ghost btn-sm tooltip tooltip-left" data-tip="Delete run"
                            onclick={(e) => {
                                e.stopPropagation();
                                confirmDeleteRun = taskRun;
                                confirmDeleteRunModal?.showModal();
                            }}
                        >
                            <Trash2 class="size-4" />
                        </button>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</div>

<ConfirmDelete
    bind:this={confirmDeleteRunModal}
    title="Delete run"
    message="This run will be permanently deleted and cannot be recovered."
    onConfirm={() => {
        if (!confirmDeleteRun) return;
        workspace?.deleteSession(confirmDeleteRun.id);
        confirmDeleteRun = null;
    }}
/>

<svelte:head>
    <title>Nanobot | {taskName} Runs</title>
</svelte:head>
