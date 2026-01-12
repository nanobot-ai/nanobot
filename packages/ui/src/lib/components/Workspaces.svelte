<script lang="ts">
	import { WorkspaceInstance, WorkspaceService } from "$lib/workspace.svelte";
	import { ChevronDown, ChevronRight, CircleX, Copy, EllipsisVertical, FileText, Folder, FolderOpen, ListTodo, PaintBucket, PencilLine, Play, Plus, Save, Share, Trash2 } from "@lucide/svelte";
	import { onMount, tick } from "svelte";
	import type { Component } from "svelte";
	import DragDropList from "./DragDropList.svelte";
	import { SvelteMap, SvelteSet } from "svelte/reactivity";
	import { resolve } from '$app/paths';
	import { goto } from "$app/navigation";
	import type { WorkspaceFile } from "$lib/types";
	import ConfirmDelete from "./ConfirmDelete.svelte";
	import { getNotificationContext } from "$lib/context/notifications.svelte";
	import { page } from "$app/state";
    import * as mocks from '$lib/mocks';
	import WorkspaceShare from "./WorkspaceShare.svelte";
	import { mockTasks } from "$lib/mocks/stores/tasks.svelte";
	import { browser } from '$app/environment';

    interface Props {
        inverse?: boolean;
        scrollContainerEl?: HTMLElement | null;
        selectedTaskId?: string | null;
        selectedRunId?: string | null;
    }

    type WorkspaceManifest = {
        name: string;
        color?: string;
        order?: number;
    }

    type Workspace = {
        id: string;
        created: string;
    } & WorkspaceManifest;

    let { inverse, scrollContainerEl, selectedTaskId, selectedRunId }: Props = $props();

    let loading = $state(false);

    let loadingWorkspace = new SvelteMap<string, boolean>();
    let workspaceData = new SvelteMap<string, WorkspaceInstance>();

    const workspaceService = new WorkspaceService();
    const notifications = getNotificationContext();

    // Persist expanded state across route changes
    const EXPANDED_STATE_KEY = 'workspaces-expanded-state';
    let expandedSections = new SvelteSet<string>();

    function loadExpandedState() {
        if (!browser) return;
        try {
            const saved = localStorage.getItem(EXPANDED_STATE_KEY);
            if (saved) {
                const parsed = JSON.parse(saved) as string[];
                expandedSections.clear();
                parsed.forEach(key => expandedSections.add(key));
            }
        } catch {
            // Ignore parse errors
        }
    }

    function saveExpandedState() {
        if (!browser) return;
        try {
            localStorage.setItem(EXPANDED_STATE_KEY, JSON.stringify([...expandedSections]));
        } catch {
            // Ignore storage errors
        }
    }

    function isExpanded(key: string): boolean {
        return expandedSections.has(key);
    }

    function toggleExpanded(key: string, open: boolean) {
        if (open) {
            expandedSections.add(key);
        } else {
            expandedSections.delete(key);
        }
        saveExpandedState();
    }
    
    function handleToggle(key: string) {
        return (e: Event) => {
            const details = e.currentTarget as HTMLDetailsElement;
            toggleExpanded(key, details.open);
        };
    }
    
    let newWorkspace = $state<WorkspaceManifest | null>(null);
    let newWorkspaceEl = $state<HTMLInputElement | null>(null);

    let confirmDeleteWorkspaceId = $state<string | null>(null);
    let confirmDeleteWorkspaceModal = $state<ReturnType<typeof ConfirmDelete> | null>(null);
    let confirmDeleteTask = $state<{ taskId: string, workspaceId: string } | null>(null);
    let confirmDeleteTaskModal = $state<ReturnType<typeof ConfirmDelete> | null>(null);

    let sharingWorkspace = $state<Workspace | null>(null);
    let shareWorkspaceModal = $state<ReturnType<typeof WorkspaceShare> | null>(null);

    let editingWorkspace = $state<Workspace | null>(null);
    let editingWorkspaceEl = $state<HTMLInputElement | null>(null);

    let selectedColor = $state<string>('');
    let summaryPointerDownTime = 0;

    let workspaces = $derived([...workspaceService.workspaces, mocks.workspace]);
    let sharedWorkspaces = $state<Workspace[]>(mocks.sharedWorkspaces);
    let workspacePermissions = $state(mocks.workspacePermissions);

    let taskRuns = $derived(mockTasks.current.tasks.reduce<Record<string, { runs: { id: string; created: string }[] }>>((acc, task) => {
        if (!acc[task.id]) {
            acc[task.id] = { 
                runs: task.runs.map((run) => ({ id: run.id, created: run.created })) 
            };
        }
        return acc;
    }, {}))

    onMount(() => {
        loadExpandedState();
        loadWorkspaces();
        // Restore data for expanded workspaces
        restoreExpandedWorkspaces();
    });
    
    async function restoreExpandedWorkspaces() {
        // Load data for any workspaces that were previously expanded
        for (const key of expandedSections) {
            if (key.startsWith('workspace:') && key.split(':').length === 2) {
                const workspaceId = key.split(':')[1];
                if (!workspaceData.has(workspaceId)) {
                    loadingWorkspace.set(workspaceId, true);
                    if (mocks.workspaceIds.includes(workspaceId)) {
                        workspaceData.set(workspaceId, mocks.workspaceInstances[workspaceId]);
                        loadingWorkspace.set(workspaceId, false);
                    } else {
                        workspaceData.set(workspaceId, workspaceService.getWorkspace(workspaceId) as WorkspaceInstance);
                        try {
                            await workspaceData.get(workspaceId)?.load();
                        } catch (err) {
                            console.error(err);
                        } finally {
                            loadingWorkspace.set(workspaceId, false);
                        }
                    }
                }
            }
        }
    }

    async function loadWorkspaces() {
        loading = true;
        try {
            await workspaceService.load();
        } catch (e) {
            const error = e instanceof Error ? e.message : String(e);
            notifications.error('Error loading workspaces', error);
        } finally {
            loading = false;
        }
    }

    async function setupNewWorkspace() {
        editingWorkspace = null;
        newWorkspace = {
            name: '',
            color: '#3b82f6',
            order: 0,
        };

        await tick();
        newWorkspaceEl?.focus();
    }

    async function createWorkspace() {
        if (!newWorkspace?.name.trim()) {
            notifications.error('Workspace name is required', 'Please enter a name for the workspace before saving.');
            return;
        }

        loading = true;
        try {
            await workspaceService.createWorkspace({
                name: newWorkspace.name,
                color: newWorkspace.color,
                order: newWorkspace.order
            });

            newWorkspace = null;
        } catch (e) {
            const error = e instanceof Error ? e.message : String(e);
            notifications.error('Error creating workspace', error);
        } finally {
            loading = false;
        }
    }
    
    async function saveEditName() {
        if (!editingWorkspace?.name.trim()) {
            notifications.error('Workspace name is required', 'Please enter a name for the workspace before saving.');
            return;
        }

        const workspace = workspaces.find((w) => w.id === editingWorkspace?.id);
        if (!workspace) {
            notifications.error('Workspace not found', 'The workspace you are trying to edit does not exist.');
            return;
        }

        loading = true;
        try {
            await workspaceService.updateWorkspace({
                ...workspace,
                name: editingWorkspace.name,
                color: editingWorkspace.color,
                order: editingWorkspace.order
            });
            editingWorkspace = null;
        } catch (e) {
            const error = e instanceof Error ? e.message : String(e);
            notifications.error('Error updating workspace', error);
        } finally {
            loading = false;
        }
    }

    async function createTask(e: MouseEvent, workspaceId: string) {
        const url = resolve(`/w/${workspaceId}/t`);
        if (e.metaKey) {
            window.open(url, '_blank');
        } else {    
            page.url.search = '';
            goto(url, { replaceState: false, invalidateAll: true });
        }
    }

    async function handleLoadWorkspace(e: MouseEvent & { currentTarget: EventTarget & HTMLElement }, workspaceId: string) {
        e.preventDefault();
        e.stopPropagation();

        const key = `workspace:${workspaceId}`;
        const willOpen = !isExpanded(key);
        toggleExpanded(key, willOpen);
        
        if (willOpen) {
            loadingWorkspace.set(workspaceId, true);
            if (mocks.workspaceIds.includes(workspaceId)) {
                workspaceData.set(workspaceId, mocks.workspaceInstances[workspaceId]);
                loadingWorkspace.set(workspaceId, false);
            } else {
                workspaceData.set(workspaceId, workspaceService.getWorkspace(workspaceId) as WorkspaceInstance);
                try {
                    await workspaceData.get(workspaceId)?.load();
                } catch (err) {
                    console.error(err);
                    // TODO: handle error, temp disabled cause of mock shared workspaces
                    // const error = e instanceof Error ? e.message : JSON.stringify(e);
                    // notifications.error('Error loading workspace', error);
                } finally {
                    loadingWorkspace.set(workspaceId, false);
                }
            }
        }
    }

    const initialColorOptions = [
        '#380067',
        '#4f7ef3',
        '#2ddcec',
        '#ff4044',
        '#fdcc11',
        '#06eaa7',
        '#ff7240',
        '#840032',
        '#4f772d'
    ]
</script>

<div class="flex flex-col">
    {@render myWorkspaces()}
    {@render sharedWorkspacesContent()}
</div>

{#snippet workspaceTitle(workspace: Workspace)}
    <div class="flex grow items-center gap-2">
        <button class="chevron-icon shrink-0 btn btn-square btn-ghost btn-xs" 
            onmousedown={(e) => e.stopPropagation()} 
            onclick={(e) => {
                handleLoadWorkspace(e, workspace.id);
            }}
        >
            <ChevronRight class="size-4 chevron-closed" />
            <ChevronDown class="size-4 chevron-open" />
        </button>
        {#if editingWorkspace?.id !== workspace.id}
            <Folder class="size-4 shrink-0 folder-closed" style="color: {workspace.color || '#000'}; fill: color-mix(in oklab, {workspace.color || '#000'} 50%, var(--color-base-100))" />
            <FolderOpen class="size-4 shrink-0 folder-open" style="color: {workspace.color || '#000'}; fill: color-mix(in oklab, {workspace.color || '#000'} 50%, var(--color-base-100))" />
        {/if}
        {#if editingWorkspace && editingWorkspace?.id === workspace.id}
            <input
                type="text"
                class="input input-bordered input-sm flex grow mr-2"
                bind:value={editingWorkspace.name}
                bind:this={editingWorkspaceEl}
                onkeydown={(e) => {
                    if (e.key === 'Enter') {
                        saveEditName();
                    } else if (e.key === 'Escape') {
                        editingWorkspace = null;
                    }
                }}
            />
        {:else}
            <h3 class="truncate text-sm font-medium relative z-20 flex grow">{workspace.name || 'Untitled'}</h3>
        {/if}
    </div>
{/snippet}

{#snippet myWorkspaces()}
    <div class="flex px-2 pt-2 items-center justify-between">
        <h2 class="font-semibold text-base-content/60">My Workspaces</h2>
        <button class="btn btn-square btn-sm btn-ghost tooltip tooltip-left" data-tip="Create new workspace"
            onclick={setupNewWorkspace}
        >
            <Plus class="size-4" />
        </button>
    </div>

    {#if loading && workspaces.length === 0}
        <div class="flex justify-center items-center p-12">
            <span class="loading loading-spinner loading-sm"></span>
        </div>
    {:else}
        {#if newWorkspace}
            <div class="group flex py-1 items-center justify-between px-2 {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'}">
                <input
                    type="text"
                    class="input input-bordered input-sm flex grow mr-2"
                    bind:value={newWorkspace.name}
                    bind:this={newWorkspaceEl}
                    placeholder="Add workspace name..."
                    onkeydown={(e) => {
                        if (e.key === 'Enter') {
                            createWorkspace();
                        } else if (e.key === 'Escape') {
                            newWorkspace = null;
                        }
                    }}
                />

                <div class="flex items-center">
                    <button class="btn btn-square btn-ghost btn-sm" 
                        onclick={() => {
                            newWorkspace = null;
                        }}
                    >
                        <CircleX class="size-4" />
                    </button>
                    <button class="btn btn-square btn-ghost btn-sm btn-primary"
                        onclick={createWorkspace}
                    >
                        <Save class="size-4" />
                    </button>
                </div>
            </div>
        {/if}
        <DragDropList bind:items={workspaces} {scrollContainerEl}
            class="menu menu-sm w-full p-0"
            classes={{
                dropIndicator: 'mx-2 my-0.5 h-0.5',
                item: 'relative group overflow-visible',
                itemsContainer: 'w-full',
            }}
            as="ul"
            childrenAs="li"
            useLongPress
        >
            {#snippet children({ item: workspace })}
                <details class="workspace-details flex flex-col w-full overflow-visible" open={isExpanded(`workspace:${workspace.id}`)}>
                    <summary 
                        class="flex px-2 items-center justify-between rounded-none list-none [&::-webkit-details-marker]:hidden overflow-visible {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'}" 
                        onpointerdown={() => { summaryPointerDownTime = Date.now(); }}
                        onclick={async (e) => { 
                            if (Date.now() - summaryPointerDownTime > 300) {
                                e.preventDefault();
                            } else {
                                handleLoadWorkspace(e, workspace.id);
                            }
                        }}
                    >
                        {@render workspaceTitle(workspace)}
                        <div class="shrink-0 flex items-center opacity-0 transition-opacity group-hover:opacity-100 relative z-30">
                            {#if editingWorkspace?.id !== workspace.id}
                                {@render workspaceMenuActions(workspace, ['read', 'write', 'execute'])}
                            {:else}
                                <div class="flex items-center">
                                    <button class="btn btn-square btn-ghost btn-sm" 
                                        onmousedown={(e) => e.stopPropagation()} 
                                        onclick={() => {
                                            editingWorkspace = null;
                                        }}
                                    >
                                        <CircleX class="size-4" />
                                    </button>
                                    <button class="btn btn-square btn-ghost btn-sm btn-primary"
                                        onmousedown={(e) => e.stopPropagation()} 
                                        onclick={saveEditName}
                                    >
                                        <Save class="size-4" />
                                    </button>
                                </div>
                            {/if}
                        </div>
                    </summary>
                    {@render workspaceContent(workspace, ['read', 'write', 'execute'])}
                </details>
            {/snippet}
        </DragDropList>
    {/if}
{/snippet}

{#snippet sharedWorkspacesContent()}
    <div class="flex p-2 items-center justify-between mt-2">
        <h2 class="font-semibold text-base-content/60">Shared With Me</h2>
    </div>

    <ul class="menu menu-sm w-full p-0">
        {#each sharedWorkspaces as workspace (workspace.id)}
        <li class="group">
            <details class="workspace-details flex flex-col w-full overflow-visible" open={isExpanded(`workspace:${workspace.id}`)}>
                <summary class="flex px-2 items-center justify-between rounded-none list-none [&::-webkit-details-marker]:hidden overflow-visible {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'}"
                    onclick={(e) => {
                        handleLoadWorkspace(e, workspace.id);
                    }}
                >
                    {@render workspaceTitle(workspace)}
                    <div class="shrink-0 flex items-center opacity-0 transition-opacity group-hover:opacity-100 relative z-30">
                        {@render workspaceMenuActions(workspace, workspacePermissions[workspace.id] ?? [], true)}
                    </div>
                </summary>
                {@render workspaceContent(workspace, workspacePermissions[workspace.id] ?? [])}
            </details>
        </li>
        {/each}
    </ul>
{/snippet}

{#snippet workspaceContent(workspace: Workspace, permissions: string[])}
<div onmousedown={(e) => e.stopPropagation()} role="presentation">
    {#if loadingWorkspace.get(workspace.id)}
        <div class="flex flex-col gap-1 w-full p-2 pl-8">
            <div class="skeleton w-full h-7 rounded-field"></div>
            <div class="skeleton w-full h-7 rounded-field"></div>
        </div>
    {:else}
        {@const workspaceInstance = workspaceData.get(workspace.id)}
        {@const tasks = (workspaceInstance?.files ?? [])
            .filter((f: { name: string }) => f.name.startsWith('.nanobot/tasks/'))
            .reduce<Record<string, boolean>>((acc, f: { name: string }) => {
                const taskId = f.name.split('/')[2];
                acc[taskId] = true;
                return acc;
            }, {})
        }
        {@const files = (workspaceInstance?.files ?? []).filter((f: { name: string }) => !f.name.startsWith('.nanobot/tasks/'))}
        <!-- {@const conversations = workspaceInstance?.sessions ?? []} -->
        <ul>
            {@render tasksSection(workspace.id, tasks, permissions)}
            <!-- {@render conversationsSection(workspace.id, conversations)} -->
            {@render filesSection(workspace.id, files, permissions)}
        </ul>
    {/if}
</div>
{/snippet}

{#snippet empty(title: string, hasCreate?: boolean)}
    <li class="w-full">
        <p class="p-2 italic text-base-content/30 flex hover:bg-transparent cursor-default">
            No {title.toLowerCase()}. 
            {#if hasCreate}
                Click <Plus class="size-2.5 inline-flex" /> to get started.
            {/if}
        </p>
    </li>
{/snippet}

{#snippet sectionTitle(title: string, Icon: Component, items: unknown[], onCreate?: (e: MouseEvent) => void, createLabel?: string)}
    <summary class="flex rounded-r-none px-2 items-center justify-between gap-2 [&::-webkit-details-marker]:hidden overflow-visible {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'}">
        <div class="flex items-center gap-2">
            <span class="chevron-icon shrink-0">
                <ChevronRight class="size-4 chevron-closed" />
                <ChevronDown class="size-4 chevron-open" />
            </span>
            <Icon class="size-4" />
            <h3 class="text-sm font-medium">{title}</h3>
        </div>
        <div class="flex items-center gap-2">
            <div class="size-8 flex items-center justify-center">
                {#if items.length > 0}
                    <div class="badge badge-sm badge-ghost">{items.length}</div>
                {/if}
            </div>
            {#if onCreate}
                <button class="btn btn-square btn-ghost btn-sm tooltip tooltip-left" onclick={onCreate} data-tip={createLabel ?? 'Create new...'}>
                    <Plus class="size-4" />
                </button>
            {/if}
        </div>
    </summary>
{/snippet}

{#snippet tasksSection(workspaceId: string, tasks: Record<string, boolean>, permissions: string[])}
{@const items = Object.keys(tasks)}
{@const title =  'Workflows'}
<li class="flex grow">
    <details class="workspace-details w-full" open={isExpanded(`workspace:${workspaceId}:tasks`)}
        ontoggle={handleToggle(`workspace:${workspaceId}:tasks`)}>
        {@render sectionTitle(title, ListTodo, items, permissions.includes('write') ? (e) => createTask(e, workspaceId) :undefined, 'Create new workflow')}
        <ul>
            {#if items.length === 0}
                {@render empty(title, permissions.includes('write'))}
            {:else}
                {#each items as item, index (index)}
                {@const name = mocks.taskData[item]?.name ?? item}           
                    <li class="flex grow">
                        <details class="workspace-details w-full" open={isExpanded(`workspace:${workspaceId}:task:${item}`)}
                            ontoggle={handleToggle(`workspace:${workspaceId}:task:${item}`)}>
                            <summary class="flex rounded-r-none px-2 items-center justify-between gap-2 [&::-webkit-details-marker]:hidden overflow-visible 
                                {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'} 
                                {selectedTaskId && !selectedRunId  && selectedTaskId === item ? 'bg-base-200 dark:bg-base-100' : ''}
                            ">
                                <div class="flex items-center gap-2">
                                    <span class="chevron-icon shrink-0">
                                        <ChevronRight class="size-4 chevron-closed" />
                                        <ChevronDown class="size-4 chevron-open" />
                                    </span>

                                    {#if permissions.includes('write') || permissions.includes('read')}
                                        <a href={resolve(`/w/${workspaceId}/t?id=${item}`)} class="flex min-h-8 grow overflow-hidden rounded-r-none truncate hover:bg-transparent items-center">{name}</a>
                                    {:else}
                                        <a href={resolve(`/w/${workspaceId}/t?id=${item}&run=true`)} class="flex min-h-8 grow overflow-hidden rounded-r-none truncate hover:bg-transparent items-center">{name}</a>
                                    {/if}
                                </div>
                                <div class="flex items-center gap-2">
                                    {#if permissions.includes('write') || permissions.includes('read')}
                                        <button class="btn btn-square btn-ghost btn-sm tooltip tooltip-left mr-1" popovertarget="popover-task-actions-{item}" style="anchor-name:--task-actions-anchor-{item}"
                                            data-tip="Edit workflow"
                                        >
                                            <EllipsisVertical class="size-4 shrink-0" />
                                        </button>
                                        <ul 
                                            id="popover-task-actions-{item}"
                                            class="dropdown menu min-w-36 rounded-box bg-base-100 shadow-sm"
                                            popover style="position-anchor:--task-actions-anchor-{item}"
                                        >
                                            <li>
                                                <a 
                                                    href={resolve(`/w/${workspaceId}/t?id=${item}&run=true`)}
                                                    class="text-sm {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'}"
                                                >
                                                    <Play class="size-4" /> Run
                                                </a>
                                            </li> 
                                            {#if permissions.includes('write')}
                                                <li>
                                                    <button 
                                                        onmousedown={(e) => e.stopPropagation()} 
                                                        onclick={() => {
                                                            confirmDeleteTask = {
                                                                taskId: item,
                                                                workspaceId,
                                                            };
                                                            confirmDeleteTaskModal?.showModal();
                                                        }} 
                                                        class="menu-alert"
                                                    >
                                                        <Trash2 class="size-4" /> Delete
                                                    </button>
                                                </li>
                                            {/if}
                                        </ul>
                                    {/if}
                                </div>
                            </summary>
                            <ul>
                                {#if taskRuns[item]?.runs?.length > 0}
                                    {@const runOnly = !permissions.includes('write') && !permissions.includes('read') && permissions.includes('execute')}
                                    {#each (taskRuns[item]?.runs ?? []) as run (run.id)}
                                        <li>
                                            <div class="flex items-center gap-2 rounded-r-none
                                                {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'}
                                                {selectedRunId && selectedRunId === run.id ? 'bg-base-200 dark:bg-base-100' : ''}
                                            ">
                                                <a
                                                    href={resolve(`/w/${workspaceId}/t?id=${item}&runId=${run.id}${runOnly ? '&run=true' : ''}`)}
                                                    class="block h-full p-2 w-full overflow-hidden truncate"
                                                >
                                                    {new Date(run.created).toLocaleString().replace(',', '')}
                                                </a>
                                                <button class="btn btn-square btn-ghost btn-sm mr-1" popovertarget="popover-task-run-actions-{item}-{run.id}" style="anchor-name:--task-run-actions-anchor-{item}-{run.id}">
                                                    <EllipsisVertical class="size-4 shrink-0" />
                                                </button>
                                                <ul 
                                                    id="popover-task-run-actions-{item}-{run.id}"
                                                    class="dropdown menu min-w-36 rounded-box bg-base-100 shadow-sm"
                                                    popover style="position-anchor:--task-run-actions-anchor-{item}-{run.id}"
                                                >
                                                    <li>
                                                        <button 
                                                            onmousedown={(e) => e.stopPropagation()} 
                                                            onclick={() => mockTasks.deleteRun(item, run.id)} 
                                                            class="menu-alert"
                                                        >
                                                            <Trash2 class="size-4" /> Delete
                                                        </button>
                                                    </li>
                                                </ul>
                                            </div>
                                        </li>
                                    {/each}
                                {:else}
                                    <li>
                                        <p class="p-2 italic text-base-content/30 text-xs">No runs.</p>
                                    </li>
                                {/if}
                            </ul>
                        </details>
                    </li>
                {/each}
            {/if}
        </ul>
    </details>
</li>
{/snippet}

{#snippet workspaceMenuActions(workspace: Workspace, permissions: string[], isShared?: boolean)}
{@const canWrite = permissions.includes('write')}
<button class="btn btn-ghost btn-square btn-sm tooltip tooltip-left" popoverTarget="workspace-actions-{workspace.id}" style="anchor-name: --workspace-actions-anchor-{workspace.id};"
    onmousedown={(e) => e.stopPropagation()}
    onclick={(e) => e.stopPropagation()}
    data-tip="Edit workspace"
>
    <EllipsisVertical class="size-4" />
</button>
<ul class="dropdown menu w-48 rounded-box bg-base-100 dark:bg-base-300 shadow-sm overflow-visible"
    popover="auto" id="workspace-actions-{workspace.id}" style="position-anchor: --workspace-actions-anchor-{workspace.id};">
    {#if canWrite}
        <li>
            <button 
                onmousedown={(e) => e.stopPropagation()} 
                onclick={async (e) => {
                    editingWorkspace = workspace;
                    e.currentTarget.blur();

                    await tick();
                    editingWorkspaceEl?.focus();
                }} 
                class="text-sm"
            >
                <PencilLine class="size-4" />
                Rename
            </button>
        </li>
        <li class="group/submenu relative" 
            role="presentation"
            onmousedown={(e) => e.stopPropagation()}
            onmouseleave={(e) => {
                if (!e.currentTarget.contains(document.activeElement)) {
                    selectedColor = '';
                }
            }}
            onfocusout={(e) => {
                if (!e.currentTarget.contains(e.relatedTarget as Node)) {
                    selectedColor = '';
                }
            }}
        >
            <div class="flex justify-between items-center">
                <div class="flex items-center gap-2 text-sm">
                    <PaintBucket class="size-4" />
                    Change color
                </div>
                <ChevronRight class="size-3" />
            </div>
            <ul class="ml-0 menu -translate-y-2 bg-base-100 dark:bg-base-300 rounded-box shadow-md absolute left-full top-0 w-52 invisible opacity-0 group-hover/submenu:visible group-hover/submenu:opacity-100 group-focus-within/submenu:visible group-focus-within/submenu:opacity-100 transition-opacity z-50 before:hidden grid grid-cols-3 gap-0.5">
                {#each initialColorOptions as color (color)}
                    <li>
                        <button class="text-sm justify-center flex border {color === selectedColor ? 'bg-base-300 border-primary' : 'border-transparent '}" 
                            onclick={(_e) => {
                                selectedColor = color;
                            }} aria-label="Change color to {color}"
                        >
                            <div class="w-8 h-4 rounded-input" style="background-color: {color};"></div>
                        </button>
                    </li>
                {/each}
                <li class="col-span-3 relative">
                    <button class="btn btn-sm btn-ghost z-10 pointer-events-none border {selectedColor && !initialColorOptions.includes(selectedColor) ? 'bg-base-300 border-primary' : 'border-transparent'}">
                        {#if selectedColor === '' || initialColorOptions.includes(selectedColor)}
                            Custom color
                        {:else}
                            <div class="w-full h-4 rounded-input" style="background-color: {selectedColor};"></div>
                        {/if}
                    </button>
                    <input type="color" class="w-full absolute top-0 left-0 h-full" onmousedown={(e) => e.stopPropagation()} onclick={(e) => e.stopPropagation()} bind:value={selectedColor} />
                </li>
                <li class="col-span-3 mt-2">
                    <button class="btn button-soft btn-sm"
                        disabled={!selectedColor}
                        onclick={() => {
                            loading = true;
                            try {
                                workspaceService.updateWorkspace({
                                    ...workspace,
                                    color: selectedColor,
                                });
                            } catch (e) {
                                const error = e instanceof Error ? e.message : String(e);
                                notifications.error('Error updating workspace', error);
                            } finally {
                                selectedColor = '';
                                loading = false;
                            }
                        }}
                    >
                        Apply
                    </button>
                </li>
            </ul>
        </li>
    {/if}
    {#if permissions.includes('write')}
        <li>
            <button 
                onmousedown={(e) => e.stopPropagation()} 
                onclick={() => {
                    sharingWorkspace = workspace;
                    shareWorkspaceModal?.showModal();
                }} 
                class="text-sm"
            >
                <Share class="size-4" /> Share
            </button>
        </li>
    {/if}
    <li>
        <button 
            onmousedown={(e) => e.stopPropagation()} 
            onclick={async (e) => {
                e.stopPropagation();
                workspaces = [...workspaces, {
                    ...workspace,
                    id: `copy-${workspace.id}`,
                    name: `Copy of ${workspace.name}`,
                    created: new Date().toISOString(),
                }];
                document.getElementById(`workspace-actions-${workspace.id}`)?.hidePopover();
            }} 
            class="text-sm"
        >
            <Copy class="size-4" />
            Make a copy
        </button>
    </li>
    {#if canWrite && !isShared}
        <li>
            <button 
                onmousedown={(e) => e.stopPropagation()} 
                onclick={() => {
                    confirmDeleteWorkspaceId = workspace.id;
                    confirmDeleteWorkspaceModal?.showModal();
                }} 
                class="menu-alert"
            >
                <Trash2 class="size-4" /> Delete
            </button>
        </li>
    {/if}
</ul>
{/snippet}

<!-- {#snippet conversationsSection(_workspaceId: string, conversations: Session[])}
<li class="flex grow">
    <details class="workspace-details w-full">
        {@render sectionTitle('Conversations', MessageSquare, conversations)}
        <ul>
            {#if conversations.length === 0}
                {@render empty('Conversations')}
            {:else}
                {#each conversations as conversation, index (index)}
                    <li class="w-full">
                        <a href={resolve(`/c/${conversation.id}`)} class="block p-2 w-full overflow-hidden rounded-r-none truncate {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'}">{conversation.title}</a>
                    </li>
                {/each}
            {/if}
        </ul>
    </details>
</li>
{/snippet} -->

{#snippet filesSection(_workspaceId: string, files: WorkspaceFile[], permissions: string[])}
<li class="flex grow">
    <details class="workspace-details w-full" open={isExpanded(`workspace:${_workspaceId}:files`)}
        ontoggle={handleToggle(`workspace:${_workspaceId}:files`)}>
        {@render sectionTitle('Files', FileText, files, permissions.includes('write') ? undefined : undefined, 'Create new file')}
        <ul>
            {#if files.length === 0}
                {@render empty('Files')}
            {:else}
                {#each files as file, index (index)}
                    <li class="w-full">
                        <button class="block p-2 w-full overflow-hidden rounded-r-none truncate {inverse ? 'hover:bg-base-200 dark:hover:bg-base-100' : 'hover:bg-base-100'}">{file.name}</button>
                    </li>
                {/each}
            {/if}
        </ul>
    </details>
</li>
{/snippet}

<ConfirmDelete
    bind:this={confirmDeleteWorkspaceModal}
    title="Delete this workspace?"
    message="This workspace will be permanently deleted and cannot be recovered."
    onConfirm={() => {
        if (!confirmDeleteWorkspaceId) return;


        if (mocks.workspaceIds.includes(confirmDeleteWorkspaceId) || confirmDeleteWorkspaceId.startsWith('copy-')) {
            workspaces = workspaces.filter((w) => w.id !== confirmDeleteWorkspaceId);            
        } else {
            workspaceService.deleteWorkspace(confirmDeleteWorkspaceId);
        }
        confirmDeleteWorkspaceModal?.close();
    }}
/>

<ConfirmDelete 
    bind:this={confirmDeleteTaskModal}
    title="Delete this task?"
    message="This task will be permanently deleted and cannot be recovered."
    onConfirm={() => {
        if (!confirmDeleteTask) return;

        if (mocks.taskIds.includes(confirmDeleteTask.taskId)) {
            const refWorkspace = workspaceData.get(confirmDeleteTask.workspaceId);
            if (refWorkspace) {
                // Create a new object to trigger Svelte reactivity
                workspaceData.set(confirmDeleteTask.workspaceId, {
                    ...refWorkspace,
                    files: refWorkspace.files.filter((f) => !f.name.startsWith(`.nanobot/tasks/${confirmDeleteTask?.taskId}`))
                } as unknown as WorkspaceInstance);
            }
        } else {
            const workspace = workspaceService.getWorkspace(confirmDeleteTask.workspaceId);
            const allTaskFiles = workspace.files.filter((f) => f.name.startsWith(`.nanobot/tasks/${confirmDeleteTask?.taskId}`));
            for (const file of allTaskFiles) {
                workspace.deleteFile(file.name);
            }
        }
        confirmDeleteTaskModal?.close();
    }}
/>

<WorkspaceShare
    bind:this={shareWorkspaceModal}
    sharingWorkspace={sharingWorkspace}
    onCancel={() => {
        sharingWorkspace = null;
    }}
    onShare={() => {
        sharingWorkspace = null;
    }}
/>

<style>
    /* Hide daisyUI's default menu marker */
    .workspace-details > summary::after {
        display: none !important;
    }
    
    .chevron-icon {
        display: flex;
    }
    
    .chevron-icon :global(.chevron-open) {
        display: none;
    }
    
    .chevron-icon :global(.chevron-closed) {
        display: block;
    }
    
    .workspace-details[open] > summary .chevron-icon :global(.chevron-open) {
        display: block;
    }
    
    .workspace-details[open] > summary .chevron-icon :global(.chevron-closed) {
        display: none;
    }
    
    .workspace-details :global(.folder-open) {
        display: none;
    }
    
    .workspace-details :global(.folder-closed) {
        display: block;
    }
    
    .workspace-details[open] > summary :global(.folder-open) {
        display: block;
    }
    
    .workspace-details[open] > summary :global(.folder-closed) {
        display: none;
    }
</style>