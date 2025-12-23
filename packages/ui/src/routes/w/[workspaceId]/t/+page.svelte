<script lang="ts">
	import '$lib/../app.css';
	import DragDropList from '$lib/components/DragDropList.svelte';
	import MarkdownEditor from '$lib/components/MarkdownEditor.svelte';
	import MessageInput from '$lib/components/MessageInput.svelte';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	import { EllipsisVertical, GripVertical, MessageCircleMore, Play, Plus, ReceiptText, Sparkles, ToolCase, Trash2, X } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { fade, fly, slide } from 'svelte/transition';
	import { createVariablePillPlugin } from '$lib/plugins/variablePillPlugin';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { WorkspaceService } from '$lib/workspace.svelte';
	import type { WorkspaceClient, WorkspaceFile } from '$lib/types';
	import { getNotificationContext } from '$lib/context/notifications.svelte';
	import type { Task } from './types';
	import { compileOutputFiles, convertToTask, setupEmptyTask } from './utils';

    let { data } = $props();
    let workspaceId = $derived(data.workspaceId);
    let searchParams = $derived(page.url.searchParams);
    let workspace = $state<WorkspaceClient | null>(null);

    let taskId = $state('');
    let task = $state<Task | null>(null);
    let loading = $state(false);
    let initialLoadComplete = $state(false);
    let saveTimeout: ReturnType<typeof setTimeout> | null = null;
    let saveAbortController: AbortController | null = null;
    let lastSavedTaskJson = '';
    
    const notifications = getNotificationContext();
    const layout = getLayoutContext();
    const workspaceService = new WorkspaceService();
    const variablePillPlugin = createVariablePillPlugin({
		onVariableAddition: (variable: string) => {
            console.log('variable added', variable);
        },
		onVariableDeletion: (variable: string) => {
            console.log('variable deleted', variable);
        },
	});

    $effect(() => {
        workspace = workspaceService.getWorkspace(workspaceId);
    })

    function debouncedSave() {
        if (saveTimeout) {
            clearTimeout(saveTimeout);
        }
        
        // Abort any ongoing save operation
        if (saveAbortController) {
            saveAbortController.abort();
        }
        
        const taskSnapshot = $state.snapshot(task);
        saveTimeout = setTimeout(async () => {
            if (!taskSnapshot) {
                console.error('task snapshot is null');
                return;
            }

            const url = new URL(page.url);
            if (!url.searchParams.get('id')) {
                taskId = crypto.randomUUID();
                url.searchParams.set('id', taskId);
                goto(url.toString(), { replaceState: true, keepFocus: true });
            }

            saveAbortController = new AbortController();
            const signal = saveAbortController.signal;

            const outputFiles = compileOutputFiles(taskSnapshot, taskId);
            for (const file of outputFiles) {
                // Check if this save operation was cancelled
                if (signal.aborted) {
                    console.log('save operation cancelled');
                    return;
                }
                
                const exists = workspace?.files.find((f) => f.name === file.id);
                try {
                    if (exists) {
                        console.log('update file', { file });
                        await workspace?.writeFile(file.id, file.data);
                    } else {
                        console.log('create file', { file });
                        await workspace?.createFile(file.id, file.data);
                    }
                } catch (error) {
                    // Only log errors if the operation wasn't cancelled
                    if (!signal.aborted) {
                        console.error('failed to save file', { file, error });
                    }
                }
            }
        }, 500);
    }

    $effect(() => {
        if (!task) return;
        
        // Serialize to track all nested changes (name, description, steps, and step properties)
        const taskJson = JSON.stringify(task);
        
        // Skip during loading or before initial load completes
        if (loading || !initialLoadComplete) {
            return;
        }
        
        // Only save if task actually changed from last saved state
        if (taskJson === lastSavedTaskJson) {
            return;
        }
        lastSavedTaskJson = taskJson;
        
        debouncedSave();
    });

    async function compileTask(idToUse: string, files: WorkspaceFile[]){
        if (!workspace) return;

        loading = true;
        task = await convertToTask(workspace, files, idToUse);
        taskId = idToUse;

        stepDescription.clear();
        for (const step of task.steps) {
            if (step.description) {
                stepDescription.set(step.id, true);
            }
        }

        // Mark this as the "saved" state so the effect doesn't trigger a save
        lastSavedTaskJson = JSON.stringify(task);
        loading = false;
        initialLoadComplete = true;
    }

    $effect(() => {
        const urlTaskId = searchParams.get('id') ?? '';
        const files = workspace?.files ?? [];
        
        if (urlTaskId && workspace && workspace.id === workspaceId && urlTaskId !== taskId && files.length > 0) {
            compileTask(urlTaskId, files);
        } else if (!urlTaskId && !initialLoadComplete) {
            // Set initial task without triggering save (only once)
            task = setupEmptyTask();
            // Mark this as the "saved" state so the effect doesn't trigger a save
            lastSavedTaskJson = JSON.stringify(task);
            initialLoadComplete = true;
        }
    })
    
    let scrollContainer = $state<HTMLElement | null>(null);

    let stepBlockEditing = new SvelteMap<number | string, boolean>();
    let stepDescription = new SvelteMap<number | string, boolean>();

    // TODO:
    let currentRun = $state<unknown | null>(null);
    let showCurrentRun = $state(false);

    let showMessageInput = $state(false);
    let showAlternateHeader = $state(false);
    
    function toggleStepBlockEditing(id: string, enabled: boolean) {
        stepBlockEditing.set(id, enabled);
    }

    function toggleStepDescription(id: string, enabled: boolean) {
        stepDescription.set(id, enabled);
    }

    function handleRun() {
        if (!task) return;
        if (task.steps.length === 0 || task.steps.every((step) => !step.content)) {
            notifications.error('Steps Required', 'Please add at least one step to the task before running it.');
            return;
        }

        // TODO:
        showCurrentRun = true;
    }
</script>

<svelte:head>
    <title>Nanobot | Tasks</title>
</svelte:head>

{#if initialLoadComplete && task}
    <div class="flex w-full h-dvh">
        <div class="
            flex flex-col grow p-4 pt-0 overflow-y-auto max-h-dvh transition-all duration-200 ease-in-out 
            {layout.isSidebarCollapsed ? 'mt-10' : ''}
        " 
            bind:this={scrollContainer}
            onscroll={() => {
                showAlternateHeader = (scrollContainer?.scrollTop ?? 0) > 100;
            }}
        >
            <div class="sticky top-0 left-0 w-full bg-base-200 dark:bg-base-100 z-10 py-4">
                <div in:fade class="flex flex-col grow">
                    <div class="flex w-full items-center gap-4">
                        {#if showAlternateHeader}
                            <p in:fade class="flex grow text-xl font-semibold">{task.name}</p>
                        {:else}
                            <input name="title" class="input input-ghost input-xl w-full placeholder:text-base-content/30 font-semibold" type="text" placeholder="Task title" 
                                bind:value={task.name}
                            />
                        {/if}
                        <button class="btn btn-primary w-48" onclick={handleRun}>
                            Run <Play class="size-4" /> 
                        </button>
                    </div>
                    {#if !showAlternateHeader}
                        <input out:slide={{ axis: 'y' }} name="description" class="input input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Task description"
                            bind:value={task.description}
                        />
                    {/if}
                </div>
            </div>
            <DragDropList bind:items={task.steps} scrollContainerEl={scrollContainer}
                class={showCurrentRun ? '' : 'md:pr-22'}
                classes={{
                    dropIndicator: 'mx-22 my-2 h-2',
                    item: 'pl-22'
                }}
            >
                {#snippet blockHandle({ startDrag, currentItem })}
                    <div class="flex items-center gap-2">
                        <button class="btn btn-ghost btn-square btn-sm" popoverTarget="add-to-task" style="anchor-name: --add-to-task-anchor;">
                            <Plus class="text-base-content/50" />
                        </button>
                        
                        <ul class="dropdown menu w-72 rounded-box bg-base-100 dark:bg-base-300 shadow-sm"
                            popover="auto" id="add-to-task" style="position-anchor: --add-to-task-anchor;">
                            <li>
                                <button class="justify-between"
                                    onclick={(e) => {
                                        const currentIndex = task!.steps.findIndex((step) => step.id === currentItem?.id);
                                        const newStep = {
                                            id: crypto.randomUUID(),
                                            name: '',
                                            description: '',
                                            content: ''
                                        };
                                        if (e.metaKey) {
                                            task!.steps.splice(currentIndex, 0, newStep);
                                        } else {
                                            task!.steps.splice(currentIndex + 1, 0, newStep);
                                        }

                                        (document.activeElement as HTMLElement)?.blur();
                                    }}
                                >
                                    <span>Add new step</span>
                                    <span class="text-[11px] text-base-content/50">
                                        click / <kbd class="kbd ">⌘</kbd> + click
                                    </span>
                                </button>
                            </li>
                            <li><button>Add a tool</button></li>
                        </ul>

                        <button class="btn btn-ghost btn-square cursor-grab btn-sm" onmousedown={startDrag}><GripVertical class="text-base-content/50" /></button>
                    </div>
                {/snippet}
                {#snippet children({ item: step })}
                    <div class="flex flex-col gap-2 bg-base-100 dark:bg-base-200 shadow-xs rounded-box p-4 pb-8 task-step relative">
                        <div class="absolute top-3 right-3 z-2">
                            {@render menu(step.id)}
                        </div>
                        
                        <div class="flex flex-col pr-12">
                            <input name="step-name" class="input input-ghost input-lg w-full font-semibold placeholder:text-base-content/30" type="text" placeholder="Step name" bind:value={step.name} />
                            {#if stepDescription.get(step.id) ?? false}
                                <input name="step-description" class="input text-[16px] input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Step description" bind:value={step.description} />
                            {/if}
                        </div>
                    
                        <MarkdownEditor 
                            value={step.content} 
                            blockEditEnabled={stepBlockEditing.get(step.id) ?? false} 
                            plugins={[variablePillPlugin]} 
                            onChange={(value) => {
                                step.content = value;
                            }}
                        />
                    </div>
                {/snippet}
            </DragDropList>

            <div class="flex items-center justify-center mt-4">
                <button class="btn btn-primary btn-square tooltip" data-tip="Add new step"
                    onclick={() => {
                        const newStep = {
                            id: crypto.randomUUID(),
                            name: '',
                            description: '',
                            content: ''
                        };
                        task!.steps.push(newStep);
                    }}
                >
                    <Plus />
                </button>
            </div>

            <div class="flex grow"></div>

            {#if !showCurrentRun}
                <div in:fade={{ duration: 200 }} class="sticky bottom-0 right-0 self-end flex flex-col gap-4 z-10">
                    {#if showMessageInput}
                        <div class="bg-base-100 dark:bg-base-200 border border-base-300 rounded-selector w-sm md:w-2xl"
                            transition:fly={{ x: 100, duration: 200 }}
                        >
                            <MessageInput />
                        </div>  
                    {/if}

                    <button class="float-right btn btn-lg btn-circle btn-primary self-end" onclick={() => showMessageInput = !showMessageInput}>
                        <MessageCircleMore class="size-6" />
                    </button>
                </div>
            {/if}
        </div>

        {#if showCurrentRun}
            <div transition:fly={{ x: 100, duration: 200 }} class="md:min-w-[520px] bg-base-100 h-dvh">
                <div class="w-full h-full flex flex-col max-h-dvh overflow-y-auto">
                    <div class="w-full flex justify-between items-center pr-4 bg-base-100">
                        <h4 class="text-lg font-semibold border-l-4 border-primary p-4 pr-0">{task.name} | Run {'{id}'}</h4>
                        <button class="btn btn-ghost btn-square btn-sm" onclick={() => showCurrentRun = false}>
                            <X class="size-4" />
                        </button>
                    </div>
                    <div class="flex grow p-4 pt-0">
                        Thread content here
                    </div>
                    <div class="sticky bottom-0 left-0 w-full">
                        <MessageInput />
                    </div>
                </div>
            </div>
        {/if}
    </div>
{:else}
    <div class="flex justify-center items-center p-12 grow">
        <span class="loading loading-spinner loading-xl"></span>
    </div>
{/if}

{#snippet menu(id: string)}
    <button class="btn btn-ghost btn-square btn-sm" popoverTarget={`step-${id}-action`} style={`anchor-name: --step-${id}-action-anchor;`}>
        <EllipsisVertical class="text-base-content/50" />
    </button>

    <ul class="dropdown flex flex-col gap-1 dropdown-end dropdown-bottom menu w-64 rounded-box bg-base-100 dark:bg-base-300 shadow-sm border border-base-300"
        popover="auto" id={`step-${id}-action`} style={`position-anchor: --step-${id}-action-anchor;`}>
        <li>
            <label for={`step-${id}-description`} class="flex gap-2 justify-between items-center">
                <span class="flex items-center gap-2">
                    <ReceiptText class="size-4" />
                    Description
                </span>
                <input type="checkbox" class="toggle toggle-sm" id={`step-${id}-description`} 
                    checked={stepDescription.get(id) ?? false}
                    onchange={(e) => toggleStepDescription(id, (e.target as HTMLInputElement)?.checked ?? false)}
                />
            </label>
        </li>
        <li>
            <label for={`step-${id}-block-editing`} class="flex gap-2 justify-between items-center">
                <span class="flex items-center gap-2">
                    <ToolCase class="size-4" />
                    Enable block editing
                </span>
                <input type="checkbox" class="toggle toggle-sm" id={`step-${id}-block-editing`} 
                    checked={stepBlockEditing.get(id) ?? false}
                    onchange={(e) => toggleStepBlockEditing(id, (e.target as HTMLInputElement)?.checked ?? false)}
                />
            </label>
        </li>
        <li>
            <button class="flex items-center gap-2">
                <Sparkles class="size-4" /> Improve with AI
            </button>
        </li>
        <li>
            <button class="flex items-center gap-2"
                onclick={() => {
                    task!.steps = task!.steps.filter((step) => step.id !== id);
                }}
            >
                <Trash2 class="size-4" /> Delete step
            </button>
        </li>
    </ul>
{/snippet}

<style>
    :root[data-theme=nanobotlight] {
        .task-step :global(.milkdown) {
            background: var(--color-base-100);
        }
    }

    :root[data-theme=nanobotdark] {
        .task-step :global(.milkdown) {
            background: var(--color-base-200);
        }
    }
</style>