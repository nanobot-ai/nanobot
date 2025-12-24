<script lang="ts">
	import '$lib/../app.css';
	import DragDropList from '$lib/components/DragDropList.svelte';
	import MessageInput from '$lib/components/MessageInput.svelte';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	import { EllipsisVertical, GripVertical, MessageCircleMore, PencilLine, Play, Plus, ReceiptText, X } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { fade, fly, slide } from 'svelte/transition';
	import { afterNavigate, goto } from '$app/navigation';
	import { page } from '$app/state';
	import { WorkspaceService } from '$lib/workspace.svelte';
	import type { WorkspaceClient, WorkspaceFile } from '$lib/types';
	import { getNotificationContext } from '$lib/context/notifications.svelte';
	import { type Input, type Task } from './types';
	import { compileOutputFiles, convertToTask, setupEmptyTask } from './utils';
	import StepActions from './StepActions.svelte';
	import TaskInputActions from './TaskInputActions.svelte';
    import TaskInput from './TaskInput.svelte';
	import Step from './Step.svelte';

    let { data } = $props();
    let workspaceId = $derived(data.workspaceId);
    let urlTaskId = $derived(page.url.searchParams.get('id') ?? '');

    let workspace = $state<WorkspaceClient | null>(null);
    let taskId = $state('');
    let task = $state<Task | null>(null);
    let initialLoadComplete = $state(false);

    let showTaskTitle = $state(false);
    let showTaskDescription = $state(false);

    let inputsModal = $state<HTMLDialogElement | null>(null);
    let runFormData = $state<(Input & { value: string })[]>([]);

    let scrollContainer = $state<HTMLElement | null>(null);

    let stepBlockEditing = new SvelteMap<number | string, boolean>();
    let stepDescription = new SvelteMap<number | string, boolean>();

    // TODO:
    let currentRun = $state<unknown | null>(null);
    let showCurrentRun = $state(false);

    let showMessageInput = $state(false);
    let showAlternateHeader = $state(false);
    let lastSavedTaskJson = '';
    let lastSavedVisibleInputsJson = '';

    let saveTimeout: ReturnType<typeof setTimeout> | null = null;
    let saveAbortController: AbortController | null = null;

    let visibleInputs = $state<Input[]>([]);
    let inputDescription = new SvelteMap<string, boolean>();
    let inputDefault = new SvelteMap<string, boolean>();
    let hiddenInputs = $derived(task?.inputs.filter((input) => !visibleInputs.some((visibleInput) => visibleInput.name === input.name)) ?? []);

    const notifications = getNotificationContext();
    const layout = getLayoutContext();
    const workspaceService = new WorkspaceService();

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
        const visibleInputsSnapshot = $state.snapshot(visibleInputs);
        saveTimeout = setTimeout(async () => {
            if (!taskSnapshot) {
                console.error('task snapshot is empty');
                return;
            }

            const url = new URL(page.url);
            if (!url.searchParams.get('id')) {
                const nameToUse = taskSnapshot.name.trim() || taskSnapshot.steps[0].name.trim();
                // use task name to make id -- all lowercase, alphanumeric only, no special characters, underscores instead of spaces
                taskId = nameToUse.toLowerCase().replace(/ /g, '_').replace(/[^a-z0-9_]/g, '');
                if (taskId) {
                    url.searchParams.set('id', taskId);
                    goto(url.toString(), { replaceState: true, keepFocus: true });
                } else {
                    console.info('skipping save, initial task name or step name required');
                    return;
                }
            }

            saveAbortController = new AbortController();
            const signal = saveAbortController.signal;

            const outputFiles = compileOutputFiles(taskSnapshot, visibleInputsSnapshot, taskId);
            for (const file of outputFiles) {
                // Check if this save operation was cancelled
                if (signal.aborted) {
                    console.info('save operation cancelled');
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
        }, 1000);
    }

    $effect(() => {
        if (!task) return;
        
        // Serialize to track all nested changes (name, description, steps, and step properties)
        const taskJson = JSON.stringify(task);
        const visibleInputsJson = JSON.stringify(visibleInputs);
        
        // Skip during loading or before initial load completes
        if (!initialLoadComplete) {
            return;
        }
        
        // Only save if task actually changed from last saved state
        if (taskJson === lastSavedTaskJson && visibleInputsJson === lastSavedVisibleInputsJson) {
            return;
        }
        lastSavedTaskJson = taskJson;
        lastSavedVisibleInputsJson = visibleInputsJson;
        debouncedSave();
    });

    async function compileTask(idToUse: string, files: WorkspaceFile[]){
        if (!workspace) return;
        initialLoadComplete = false;
        task = null;

        task = await convertToTask(workspace, files, idToUse);
        taskId = idToUse;

        stepDescription.clear();
        for (const step of task.steps) {
            if (step.description) {
                stepDescription.set(step.id, true);
            }
        }

        // set up visible inputs
        visibleInputs = [];
        inputDescription.clear();
        inputDefault.clear();
        for (const input of task.inputs) {
            if (input.description.length || input.default?.length) {
                visibleInputs.push(input);
                if (input.description.length) {
                    inputDescription.set(input.id, true);
                }
                if (input.default?.length) {
                    inputDefault.set(input.id, true);
                }
            }
        }

        showTaskTitle = (task.name || '').length > 0;
        showTaskDescription = (task.description || '').length > 0;

        // Mark this as the "saved" state so the effect doesn't trigger a save
        lastSavedTaskJson = JSON.stringify(task);
        lastSavedVisibleInputsJson = JSON.stringify(visibleInputs);
        initialLoadComplete = true;
    }

    $effect(() => {
        const files = workspace?.files ?? [];
        
        if (urlTaskId && workspace && workspace.id === workspaceId && urlTaskId !== taskId && files.length > 0) {
            compileTask(urlTaskId, files);
        }
    })


    afterNavigate(() => {
       if (!urlTaskId) {
            task = setupEmptyTask();
            taskId = '';
            stepDescription.clear();
            lastSavedTaskJson = JSON.stringify(task);
            lastSavedVisibleInputsJson = JSON.stringify(visibleInputs);
            initialLoadComplete = true;
       } 
    });

    function handleRun() {
        if (!task) return;
        if (task.steps.length === 0 || task.steps.every((step) => !step.content)) {
            notifications.error('Steps Required', 'Please add at least one step to the task before running it.');
            return;
        }

        const visibleInputMapping = new Map(visibleInputs.map((input) => [input.id, input]));
        if (task.inputs.length > 0) {
            runFormData = task.inputs.map((input) => ({
                ...input,
                ...(visibleInputMapping.get(input.id) ?? {}),
                value: input.default || visibleInputMapping.get(input.id)?.default || '',
            }));
            inputsModal?.showModal();
        } else {
            // do run without inputs
            showCurrentRun = true;
        }
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
                        {:else if showTaskTitle}
                            <input name="title" class="input input-ghost input-lg w-full placeholder:text-base-content/30 font-semibold" type="text" placeholder="Task title" 
                                bind:value={task.name}
                            />
                        {:else}
                            <div class="w-full"></div>
                        {/if}
                        <div class="flex shrink-0 items-center gap-2">
                            <button class="btn btn-primary w-48" onclick={handleRun}>
                                Run <Play class="size-4" /> 
                            </button>
                            <button class="btn btn-ghost btn-square" popoverTarget="task-actions" style="anchor-name: --task-actions-anchor;">
                                <EllipsisVertical class="text-base-content/50" />
                            </button>
                        
                            <ul class="dropdown flex flex-col gap-1 dropdown-end dropdown-bottom menu w-64 rounded-box bg-base-100 dark:bg-base-300 shadow-sm border border-base-300"
                                popover="auto" id="task-actions" style="position-anchor: --task-actions-anchor;">
                                <li>
                                    <label for="task-title" class="flex gap-2 justify-between items-center">
                                        <span class="flex items-center gap-2">
                                            <PencilLine class="size-4" />
                                            Task title
                                        </span>
                                        <input type="checkbox" class="toggle toggle-sm" id="task-title" 
                                            checked={showTaskTitle}
                                            onchange={(e) => showTaskTitle = (e.target as HTMLInputElement)?.checked ?? false}
                                        />
                                    </label>
                                </li>
                                <li>
                                    <label for="task-description" class="flex gap-2 justify-between items-center">
                                        <span class="flex items-center gap-2">
                                            <ReceiptText class="size-4" />
                                            Task description
                                        </span>
                                        <input type="checkbox" class="toggle toggle-sm" id="task-description" 
                                            checked={showTaskDescription}
                                            onchange={(e) => showTaskDescription = (e.target as HTMLInputElement)?.checked ?? false}
                                        />
                                    </label>
                                </li>
                            </ul>
                        </div>
                    </div>
                    {#if !showAlternateHeader && showTaskDescription}
                        <input out:slide={{ axis: 'y' }} name="description" class="input input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Task description"
                            bind:value={task.description}
                        />
                    {/if}
                </div>
            </div>
            {#if visibleInputs.length > 0}
                <DragDropList bind:items={visibleInputs} scrollContainerEl={scrollContainer}
                    class={showCurrentRun ? '' : 'md:pr-22'}
                    classes={{
                        dropIndicator: 'mx-22 my-2 h-2',
                        item: 'pl-22'
                    }}
                >
                    {#snippet blockHandle({ startDrag })}
                        <div class="flex items-center gap-2">
                            <TaskInputActions task={task!} availableInputs={hiddenInputs} 
                                onAddInput={(input) => {
                                    visibleInputs.push(input);
                                }}
                            />
                            {#if visibleInputs.length > 1}
                                <button class="btn btn-ghost btn-square cursor-grab btn-sm" onmousedown={startDrag}><GripVertical class="text-base-content/50" /></button>
                            {/if}
                        </div>
                    {/snippet}
                    {#snippet children({ item: input })}
                        <TaskInput 
                            id={taskId} 
                            task={task!} 
                            {input} 
                            {inputDescription}
                            {inputDefault}
                            onHideInput={(id) => {
                                visibleInputs = visibleInputs.filter((input) => input.id !== id);
                            }}
                            onToggleInputDescription={(id, value) => inputDescription.set(id, value)}
                            onToggleInputDefault={(id, value) => inputDefault.set(id, value)}
                            {visibleInputs}
                        />
                    {/snippet}
                </DragDropList>
            {/if}
            <DragDropList bind:items={task.steps} scrollContainerEl={scrollContainer}
                class="{visibleInputs.length > 0 ? '-mt-6' : ''} {showCurrentRun ? '' : 'md:pr-22'}"
                classes={{
                    dropIndicator: 'mx-22 my-2 h-2',
                    item: 'pl-22'
                }}
            >
                {#snippet blockHandle({ startDrag, currentItem })}
                    <div class="flex items-center gap-2">
                        <StepActions 
                            task={task!} 
                            item={currentItem} 
                            availableInputs={hiddenInputs} 
                            onAddInput={(input) => {
                                visibleInputs.push(input);
                            }} 
                        />
                        <button class="btn btn-ghost btn-square cursor-grab btn-sm" onmousedown={startDrag}><GripVertical class="text-base-content/50" /></button>
                    </div>
                {/snippet}
                {#snippet children({ item: step })}
                    <Step 
                        id={taskId} 
                        task={task!} 
                        {step}
                        {stepDescription}
                        {stepBlockEditing}
                        onToggleStepDescription={(id, value) => stepDescription.set(id, value)}
                        onToggleStepBlockEditing={(id, value) => stepBlockEditing.set(id, value)}
                        onAddInput={(input) => visibleInputs.push(input)}
                        onDeleteStep={(filename) => workspace?.deleteFile(filename)}
                        {visibleInputs}
                    />
                {/snippet}
            </DragDropList>

            <div class="flex items-center justify-center">
                <button class="btn btn-primary btn-square tooltip" data-tip="Add new step"
                    onclick={() => {
                        const newStep = {
                            id: task!.steps.length.toString(),
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
    <div class="w-full flex flex-col gap-8">
        <div class="flex w-full items-center justify-between pt-4 px-4">
            <div></div>
            <div class="flex items-center gap-2">
                <div class="skeleton rounded-field w-48 h-10"></div>
                <div class="skeleton rounded-field size-10"></div>
            </div>
        </div>
        <div class="flex flex-col gap-4 mx-22 bg-base-100/30 dark:bg-base-200/70 rounded-box p-4">
            <div class="skeleton h-10 w-full"></div>
            <div class="skeleton h-10 w-full"></div>
            <div class="skeleton h-32 w-full"></div>
        </div>
    </div>
{/if}

<dialog bind:this={inputsModal} class="modal">
  <div class="modal-box bg-base-200 dark:bg-base-100 p-0">
    <h4 class="text-lg font-semibold p-4 py-2 bg-base-100 dark:bg-base-200">Run Task</h4>
    <div class="p-4 flex flex-col gap-2">
        {#each runFormData as input (input.id)}
            <label class="input w-full">
                <span class="label h-full font-semibold text-primary bg-primary/15">{input.name}</span>
                <input type="text" bind:value={input.value} />
            </label>
        {/each}
    </div>
    <div class="modal-action px-4 py-2 bg-base-100 dark:bg-base-200">
        <form method="dialog">
            <button class="btn btn-ghost">Cancel</button>
            <button class="btn btn-primary">Run</button>
        </form>
    </div>
  </div>
  <form method="dialog" class="modal-backdrop">
    <button>close</button>
  </form>
</dialog>