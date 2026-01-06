<script lang="ts">
	import '$lib/../app.css';
    import { resolve } from '$app/paths';
	import DragDropList from '$lib/components/DragDropList.svelte';
	import MessageInput from '$lib/components/MessageInput.svelte';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	import { createRegistryStore, setRegistryContext } from '$lib/context/registry.svelte';
	import { EllipsisVertical, GripVertical, MessageCircleMore, PencilLine, Play, Plus, ReceiptText, X } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { fade, fly, slide } from 'svelte/transition';
	import { afterNavigate, goto } from '$app/navigation';
	import { page } from '$app/state';
	import { WorkspaceService } from '$lib/workspace.svelte';
	import type { WorkspaceClient, WorkspaceFile } from '$lib/types';
	import { getNotificationContext } from '$lib/context/notifications.svelte';
	import { type Input, type Task, type Step as StepType } from './types';
	import { compileOutputFiles, convertToTask, setupEmptyTask } from './utils';
	import StepActions from './StepActions.svelte';
	import TaskInputActions from './TaskInputActions.svelte';
    import TaskInput from './TaskInput.svelte';
	import Step from './Step.svelte';
	import RegistryToolSelector from './RegistryToolSelector.svelte';
	import { onMount, tick } from 'svelte';
	import { ChatService } from '$lib/chat.svelte';
	import ThreadFromChat from '$lib/components/ThreadFromChat.svelte';
	import ConfirmDelete from '$lib/components/ConfirmDelete.svelte';

    type Props = {
        workspaceId: string;
        urlTaskId?: string;
    }
    let { workspaceId, urlTaskId }: Props = $props();

	const registryStore = createRegistryStore();
	setRegistryContext(registryStore);

    let workspace = $state<WorkspaceClient | null>(null);
    let taskId = $state('');
    let task = $state<Task | null>(null);
    let initialLoadComplete = $state(false);

    let showTaskTitle = $state(false);
    let showTaskDescription = $state(false);

    let inputsModal = $state<HTMLDialogElement | null>(null);
    let runFormData = $state<(Input & { value: string })[]>([]);

    let confirmDeleteStep = $state<{ stepId: string, filename: string } | null>(null);
    let confirmDeleteStepModal = $state<ReturnType<typeof ConfirmDelete> | null>(null);

    const headerHeight = 72;
    let scrollContainer = $state<HTMLElement | null>(null);
    let argumentsList = $state<HTMLDivElement | null>(null);

    let stepBlockEditing = new SvelteMap<number | string, boolean>();
    let stepDescription = new SvelteMap<number | string, boolean>();

    let currentRun = $state<unknown | null>(null);
    let showSidebarThread = $state(false);
    let toggleableMessageInput = $state<ReturnType<typeof MessageInput> | null>(null);

    let registryToolSelector = $state<ReturnType<typeof RegistryToolSelector> | null>(null);
    let currentAddingToolForStep = $state<StepType | null>(null);

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
    let chat = $state<ChatService>();
    let runSession = $state<ChatService>();

    onMount(() => {
        workspace = workspaceService.getWorkspace(workspaceId);
        registryStore.fetch();
    });

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
                    goto(resolve(url.toString() as `/w/${string}/t/`), { replaceState: true, keepFocus: true });
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
                        await workspace?.writeFile(file.id, file.data);
                    } else {
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
    });

    function initNewTask() {
        task = setupEmptyTask();
        taskId = '';
        stepDescription.clear();
        lastSavedTaskJson = JSON.stringify(task);
        lastSavedVisibleInputsJson = JSON.stringify(visibleInputs);
        initialLoadComplete = true;
        visibleInputs = [];
        inputDescription.clear();
        inputDefault.clear();
    }

    onMount(() => {
        if (!urlTaskId) {
            initNewTask();
        }
    })

    afterNavigate(() => {
        if (!urlTaskId) {
            initNewTask();
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
            // no inputs provided, just run task
            showSidebarThread = true;
        }
    }

    async function submitRun() {
        // TODO: change below to actually hit the run task endpoint once available
        runSession = await workspace?.newSession();
        runSession?.sendMessage(`
Use the files under the ".nanobot/tasks/${taskId}" directory for context to help you simulate the task run.
These are the following inputs to simulate the task run with: \n\n
${JSON.stringify(runFormData)}

\n\n Do not indicate that you are simulating; act as if you are actually running the task.
`)
        showSidebarThread = true;
    }
</script>

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
            <div bind:this={argumentsList}>
                <DragDropList 
                    bind:items={visibleInputs} 
                    scrollContainerEl={scrollContainer}
                    class={showSidebarThread ? '' : 'md:pr-22'}
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
                                <button class="btn btn-ghost btn-square cursor-grab btn-sm tooltip tooltip-right" data-tip="Drag to reorder" onmousedown={startDrag}>
                                    <GripVertical class="text-base-content/50" />
                                </button>
                            {/if}
                        </div>
                    {/snippet}
                    {#snippet children({ item: input })}
                        <TaskInput 
                            {taskId}
                            task={task!} 
                            {input} 
                            {inputDescription}
                            {inputDefault}
                            onHideInput={(id) => {
                                visibleInputs = visibleInputs.filter((input) => input.id !== id);
                            }}
                            onDeleteInput={(id) => {
                                task!.inputs = task!.inputs.filter((input) => input.id !== id);
                                visibleInputs = visibleInputs.filter((input) => input.id !== id);
                            }}
                            onToggleInputDescription={(id, value) => inputDescription.set(id, value)}
                            onToggleInputDefault={(id, value) => inputDefault.set(id, value)}
                            onSuggestImprovement={async (content) => {
                                if (!chat) {
                                    chat = await workspace?.newSession({ editor: true });
                                }
                                chat?.sendMessage(content);
                                showSidebarThread = true;
                            }}
                        />
                    {/snippet}
                </DragDropList>
            </div>
            {/if}
            <DragDropList bind:items={task.steps} scrollContainerEl={scrollContainer}
                class="{visibleInputs.length > 0 ? '-mt-6' : ''} {showSidebarThread ? '' : 'md:pr-22'}"
                classes={{
                    dropIndicator: 'mx-22 my-2 h-2',
                    item: 'pl-22'
                }}
                offset={(argumentsList?.getBoundingClientRect().height ?? 0) - (headerHeight/2)}
            >
                {#snippet blockHandle({ startDrag, currentItem })}
                    <div class="flex items-center gap-2">
                        {#if currentItem}
                            <StepActions 
                                task={task!} 
                                item={currentItem} 
                                availableInputs={hiddenInputs} 
                                onAddInput={(input) => {
                                    visibleInputs.push(input);
                                }} 
                                onOpenSelectTool={() => {
                                    currentAddingToolForStep = currentItem;
                                    registryToolSelector?.showModal();
                                }}
                            />
                            <button class="btn btn-ghost btn-square cursor-grab btn-sm tooltip tooltip-right" onmousedown={startDrag} data-tip="Drag to reorder">
                                <GripVertical class="text-base-content/50" />
                            </button>
                        {/if}
                    </div>
                {/snippet}
                {#snippet children({ item: step })}
                    <Step 
                        taskId={taskId}
                        task={task!}
                        {step}
                        {stepDescription}
                        {stepBlockEditing}
                        onToggleStepDescription={(id, value) => stepDescription.set(id, value)}
                        onToggleStepBlockEditing={(id, value) => stepBlockEditing.set(id, value)}
                        onUpdateStep={(id, updates) => {
                            const idx = task!.steps.findIndex(s => s.id === id);
                            if (idx !== -1) Object.assign(task!.steps[idx], updates);
                        }}
                        onAddInput={(input) => visibleInputs.push(input)}
                        onAddTaskInput={(input) => task!.inputs.push(input)}
                        onRemoveTaskInput={(inputName) => {
                            task!.inputs = task!.inputs.filter((input) => input.name !== inputName);
                        }}
                        onDeleteStep={(stepId, filename) => {
                            confirmDeleteStep = {
                                stepId,
                                filename,
                            };
                            confirmDeleteStepModal?.showModal();
                        }}
                        onSuggestImprovement={async (content) => {
                            if (!chat) {
                                chat = await workspace?.newSession({ editor: true });
                            }
                            chat?.sendMessage(content);
                            showSidebarThread = true;
                        }}
                        {visibleInputs}
                        onUpdateVisibleInputs={(inputs) => visibleInputs = inputs}
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
                            content: '',
                            tools: [],
                        };
                        task!.steps.push(newStep);
                    }}
                >
                    <Plus />
                </button>
            </div>

            <div class="flex grow"></div>

            {#if !showSidebarThread}
                <div in:fade={{ duration: 200 }} class="sticky bottom-0 right-0 self-end flex flex-col gap-4 z-10">
                    {#if showMessageInput}
                        <div class="bg-base-100 dark:bg-base-200 border border-base-300 rounded-selector w-sm md:w-2xl"
                            transition:fly={{ x: 100, duration: 200 }}
                        >
                            <MessageInput bind:this={toggleableMessageInput} 
                                onSend={async (message) => {
                                    showSidebarThread = true;
                                    showMessageInput = false;

                                    if (!chat) {
                                        chat = await workspace?.newSession({ editor: true });
                                    }
                                    return chat?.sendMessage(message);
                                }} 
                            />
                        </div>  
                    {/if}

                    <button class="float-right btn btn-lg btn-circle btn-primary self-end" onclick={async () => {
                        showMessageInput = !showMessageInput;
                        await tick();
                        toggleableMessageInput?.focus();
                    }}>
                        <MessageCircleMore class="size-6" />
                    </button>
                </div>
            {/if}
        </div>

        {#if showSidebarThread}
            <div transition:fly={{ x: 100, duration: 200 }} class="md:min-w-[520px] bg-base-100 h-dvh flex flex-col">
                <div class="w-full flex justify-between items-center p-4 bg-base-100 shrink-0">
                    {#if currentRun}
                        <h4 class="text-lg font-semibold border-l-4 border-primary">{task.name} | Run {'{id}'}</h4>
                    {:else}
                        <div class="w-full"></div>
                    {/if}
                    <button class="btn btn-ghost btn-square btn-sm tooltip tooltip-left" data-tip="Close" 
                        onclick={() => {
                            showSidebarThread = false;
                            if (runSession) {
                                runSession.close();
                                runSession = undefined;
                            }
                        }}
                    >
                        <X class="size-4" />
                    </button>
                </div>
                <div class="w-full flex-1 min-h-0 flex flex-col">
                    {#if runSession}
                        {#key runSession.chatId}
                            <ThreadFromChat inline chat={runSession} />
                        {/key}
                    {:else if chat}
                        {#key chat.chatId}
                            <ThreadFromChat inline {chat} />
                        {/key}
                    {/if}
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
  <div class="modal-box bg-base-200 dark:bg-base-100 p-0 border border-transparent dark:border-base-300">
    <form method="dialog">
        <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">
            <X class="size-4" />
        </button>
      </form>
    <h4 class="text-lg font-semibold p-4 bg-base-100 dark:bg-base-200">Run Task</h4>
    <div class="p-4 flex flex-col gap-2">
        {#each runFormData as input (input.id)}
            <label class="input w-full">
                <span class="label h-full font-semibold text-primary bg-primary/15">{input.name}</span>
                <input type="text" bind:value={input.value} />
            </label>
        {/each}
    </div>
    <div class="modal-action mt-0 px-4 py-2 bg-base-100 dark:bg-base-200">
        <form method="dialog">
            <button class="btn btn-ghost" onclick={() => inputsModal?.close()}>Cancel</button>
            <button class="btn btn-primary" onclick={submitRun}>Run</button>
        </form>
    </div>
  </div>
  <form method="dialog" class="modal-backdrop">
    <button>close</button>
  </form>
</dialog>

<ConfirmDelete 
    bind:this={confirmDeleteStepModal}
    title="Delete this step?"
    message="This step will be permanently deleted and cannot be recovered."
    onConfirm={() => {
        if (!confirmDeleteStep) return;
        task!.steps = task!.steps.filter((s) => s.id !== confirmDeleteStep?.stepId);
        workspace?.deleteFile(confirmDeleteStep?.filename ?? '');
    }}
/>

<RegistryToolSelector bind:this={registryToolSelector} 
    omit={currentAddingToolForStep?.tools ?? []}
    onToolsSelect={(names) => {
        if (!currentAddingToolForStep) return;
        const stepIndex = task?.steps.findIndex((step) => step.id === currentAddingToolForStep?.id);
        if (stepIndex === undefined) return;
        task!.steps[stepIndex].tools.push(...names);
    }} 
/>