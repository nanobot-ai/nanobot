<script lang="ts">
	import { WorkspaceInstance, WorkspaceService } from "$lib/workspace.svelte";
	import { Bot, ChevronDown, ChevronRight, CircleX, Edit, Folder, FolderOpen, GripVertical, ListTodo, MessageSquare, MoreVertical, PaintBucket, Plus, Save, Trash2 } from "@lucide/svelte";
	import { onMount, tick } from "svelte";
	import type { Component } from "svelte";
	import DragDropList from "./DragDropList.svelte";
	import { SvelteMap } from "svelte/reactivity";
	import { resolve } from '$app/paths';
	import { goto } from "$app/navigation";

    interface Props {
        inverse?: boolean;
        scrollContainerEl?: HTMLElement | null;
    }

    type WorkspaceManifest = {
        name: string;
        color: string;
        order: number;
    }

    type Workspace = {
        id: string;
        created: string;
    } & WorkspaceManifest;

    let { inverse, scrollContainerEl }: Props = $props();

    let loading = $state(false);
    let error = $state<string | null>(null);

    let loadingWorkspace = new SvelteMap<string, boolean>();
    let workspaceData = new SvelteMap<string, WorkspaceInstance>();

    let newWorkspace = $state<WorkspaceManifest | null>(null);
    let newWorkspaceEl = $state<HTMLInputElement | null>(null);
    const workspaceService = new WorkspaceService();

    let editingWorkspace = $state<Workspace | null>(null);
    let editingWorkspaceEl = $state<HTMLInputElement | null>(null);

    onMount(() => {
        loadWorkspaces();
    });

    async function loadWorkspaces() {
        loading = true;
        error = null;
        try {
            await workspaceService.load();
        } catch (e) {
            error = e instanceof Error ? e.message : String(e);
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
            error = 'Workspace name is required';
            return;
        }

        loading = true;
        error = null;
        try {
            await workspaceService.createWorkspace({
                name: newWorkspace.name,
                color: newWorkspace.color,
                order: newWorkspace.order
            });

            newWorkspace = null;
        } catch (e) {
            error = e instanceof Error ? e.message : String(e);
        } finally {
            loading = false;
        }
    }
    
    async function saveEdit() {
        if (!editingWorkspace?.name.trim()) {
            error = 'Workspace name is required';
            return;
        }

        const workspace = workspaceService.workspaces.find((w) => w.id === editingWorkspace?.id);
        if (!workspace) {
            error = 'Workspace not found';
            return;
        }

        loading = true;
        error = null;
        try {
            await workspaceService.updateWorkspace({
                ...workspace,
                name: editingWorkspace.name,
                color: editingWorkspace.color,
                order: editingWorkspace.order
            });
            editingWorkspace = null;
        } catch (e) {
            error = e instanceof Error ? e.message : String(e);
        } finally {
            loading = false;
        }
    }

    async function createTask(e: MouseEvent, workspaceId: string) {
        const url = new URL(window.location.origin + resolve(`/w/${workspaceId}/t`));
        url.search = '';
        if (e.metaKey) {
            window.open(url, '_blank');
        } else {
            goto(url.pathname, { replaceState: false, invalidateAll: true });
        }
    }
</script>

<div class="flex h-full flex-col">
	<!-- Header -->
	<div class="flex px-2 pt-2 items-center justify-between">
		<h2 class="font-semibold text-base-content/60">Workspaces</h2>
        <button class="btn btn-square btn-sm btn-ghost tooltip tooltip-left" data-tip="Create new workspace"
            onclick={setupNewWorkspace}
        >
            <Plus class="size-4" />
        </button>
	</div>

    {#if loading && workspaceService.workspaces.length === 0}
        <div class="flex justify-center items-center p-12">
            <span class="loading loading-spinner loading-sm"></span>
        </div>
    {:else}
        {#if newWorkspace}
            <div class="group flex py-1 items-center justify-between px-2 {inverse ? 'hover:bg-base-200' : 'hover:bg-base-100'}">
                <input
                    type="text"
                    class="input input-bordered input-sm flex grow mr-2"
                    bind:value={newWorkspace.name}
                    bind:this={newWorkspaceEl}
                    placeholder="Add workspace name..."
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
        <DragDropList bind:items={workspaceService.workspaces} {scrollContainerEl}
            class="menu menu-sm w-full p-0"
            classes={{
                dropIndicator: 'mx-2 my-0.5 h-0.5',
                item: 'relative group overflow-visible',
                itemsContainer: 'w-full',
            }}
            as="ul"
            childrenAs="li"
        >
            {#snippet children({ item: workspace })}
                <details class="workspace-details flex flex-col w-full overflow-visible">
                    <summary class="flex px-2 items-center justify-between rounded-none list-none [&::-webkit-details-marker]:hidden overflow-visible {inverse ? 'hover:bg-base-200' : 'hover:bg-base-100'}" onclick={(e) => e.preventDefault()}>
                        <div class="flex grow items-center gap-2">
                            <button class="chevron-icon shrink-0 btn btn-square btn-ghost btn-xs" 
                                onmousedown={(e) => e.stopPropagation()} 
                                onclick={async (e) => {
                                    const details = e.currentTarget.closest('details');
                                    if (details) details.open = !details.open;
                                    if (details && details.open) {
                                        loadingWorkspace.set(workspace.id, true);
                                        workspaceData.set(workspace.id, workspaceService.getWorkspace(workspace.id) as WorkspaceInstance);
                                        workspaceData.get(workspace.id)?.load();
                                        loadingWorkspace.set(workspace.id, false);
                                    }
                                }}
                            >
                                <ChevronRight class="size-4 chevron-closed" />
                                <ChevronDown class="size-4 chevron-open" />
                            </button>
                            {#if editingWorkspace?.id !== workspace.id}
                                <Folder class="size-4 shrink-0 folder-closed" style="color: {workspace.color};" />
                                <FolderOpen class="size-4 shrink-0 folder-open" style="color: {workspace.color};" />
                            {/if}
                            {#if editingWorkspace && editingWorkspace?.id === workspace.id}
                                <input
                                    type="text"
                                    class="input input-bordered input-sm flex grow mr-2"
                                    bind:value={editingWorkspace.name}
                                    bind:this={editingWorkspaceEl}
                                />
                            {:else}
                                <h3 class="truncate text-sm font-medium relative z-20 flex grow">{workspace.name || 'Untitled'}</h3>
                            {/if}
                        </div>
                        <div class="shrink-0 flex items-center opacity-0 transition-opacity group-hover:opacity-100 relative z-30">
                            {#if editingWorkspace?.id !== workspace.id}
                                <div class="dropdown dropdown-end" 
                                    role="presentation"
                                    onmousedown={(e) => e.stopPropagation()} 
                                >
                                    <div 
                                        tabindex="0" 
                                        role="button" 
                                        class="btn btn-square btn-ghost btn-sm" 
                                    >
                                        <MoreVertical class="h-4 w-4" />
                                    </div>
                                    <ul
                                        class="dropdown-content menu w-48 rounded-box border border-base-300 bg-base-100 dark:bg-base-300 p-2 shadow"
                                    >
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
                                                <Edit class="size-4" />
                                                Rename
                                            </button>
                                        </li>
                                        <li>
                                            <button 
                                            onmousedown={(e) => e.stopPropagation()} 
                                            onclick={(e) => {
                                                // TODO:
                                            }}
                                            class="text-sm"
                                        >
                                                <PaintBucket class="size-4" />
                                                Change color
                                            </button>
                                        </li>
                                        <li>
                                            <button 
                                                onmousedown={(e) => e.stopPropagation()} 
                                                onclick={(e) => {
                                                    // TODO:
                                                }} 
                                                class="text-sm text-error"
                                            >
                                                <Trash2 class="size-4" />
                                                Delete
                                            </button>
                                        </li>
                                    </ul>
                                </div>
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
                                        onclick={saveEdit}
                                    >
                                        <Save class="size-4" />
                                    </button>
                                </div>
                            {/if}
                        </div>
                    </summary>
                    <div onmousedown={(e) => e.stopPropagation()} role="presentation">
                        {#if loadingWorkspace.get(workspace.id)}
                            <div class="flex justify-center items-center p-2">
                                <span class="loading loading-spinner loading-sm"></span>
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
                        
                            <ul>
                                {@render workspaceChild(workspace.id, 'Tasks', ListTodo, Object.keys(tasks), (e) => createTask(e, workspace.id))}
                                {@render workspaceChild(workspace.id, 'Agents', Bot, [])}
                                {@render workspaceChild(workspace.id, 'Conversations', MessageSquare, [])}
                            </ul>
                        {/if}
                    </div>
                </details>
            {/snippet}
        </DragDropList>
    {/if}
</div>

{#snippet workspaceChild(workspaceId: string, title: string, Icon: Component, items: string[], onCreate?: (e: MouseEvent) => void)}
<li class="flex grow">
    <details class="workspace-details w-full">
        <summary class="flex rounded-r-none px-2 items-center justify-between gap-2 [&::-webkit-details-marker]:hidden overflow-visible {inverse ? 'hover:bg-base-200' : 'hover:bg-base-100'}">
            <div class="flex items-center gap-2">
                <span class="chevron-icon shrink-0">
                    <ChevronRight class="size-4 chevron-closed" />
                    <ChevronDown class="size-4 chevron-open" />
                </span>
                <Icon class="size-4" />
                <h3 class="text-sm font-medium">{title}</h3>
            </div>
            <div class="flex items-center gap-2">
                <div class="badge badge-sm badge-ghost">{items.length}</div>
                {#if onCreate}
                    <button class="btn btn-square btn-ghost btn-sm" onclick={onCreate}>
                        <Plus class="size-4" />
                    </button>
                {:else}
                    <div class="size-8"></div>
                {/if}
            </div>
        </summary>
        <ul>
            {#if items.length === 0}
                <li class="w-full">
                    <p class="p-2 italic text-base-content/30 flex hover:bg-transparent cursor-default">No {title.toLowerCase()}. Click <Plus class="size-2.5 inline-flex" /> to get started.</p>
                </li>
            {:else}
                {#each items as item, index (index)}
                    <li class="w-full">
                        {#if title === 'Tasks'}
                            <a href={resolve(`/w/${workspaceId}/t?id=${item}`)} class="block p-2 w-full overflow-hidden rounded-r-none truncate {inverse ? 'hover:bg-base-200' : 'hover:bg-base-100'}">{item}</a>
                        {:else}
                            <p class="p-2 w-full overflow-hidden truncate">{item}</p>
                        {/if}
                    </li>
                {/each}
            {/if}
        </ul>
    </details>
</li>
{/snippet}

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