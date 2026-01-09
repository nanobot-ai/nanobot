<script lang="ts">
	import '$lib/../app.css';
    import { resolve } from '$app/paths';
	import DragDropList from '$lib/components/DragDropList.svelte';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	import { createRegistryStore, setRegistryContext } from '$lib/context/registry.svelte';
	import { ChevronDown, File, EllipsisVertical, GripVertical, MessageCircleMore, PencilLine, Play, Plus, ReceiptText, X, Bug } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { fade, fly, slide } from 'svelte/transition';
	import { afterNavigate, goto } from '$app/navigation';
	import { page } from '$app/state';
	import { WorkspaceService } from '$lib/workspace.svelte';
	import type { Attachment, WorkspaceClient, WorkspaceFile } from '$lib/types';
	import { getNotificationContext } from '$lib/context/notifications.svelte';
	import { type Input, type Task, type Step as StepType } from './types';
	import { compileOutputFiles, convertToTask, parseFrontmatterMarkdown, setupEmptyTask } from './utils';
	import StepActions from './StepActions.svelte';
	import TaskInputActions from './TaskInputActions.svelte';
    import TaskInput from './TaskInput.svelte';
	import Step from './Step.svelte';
	import RegistryToolSelector from './RegistryToolSelector.svelte';
	import { onDestroy, onMount, tick } from 'svelte';
	import ConfirmDelete from '$lib/components/ConfirmDelete.svelte';
    import * as mocks from '$lib/mocks';
	import { mockTasks } from '$lib/mocks/stores/tasks.svelte';
	import { setSharedChat, sharedChat } from '$lib/stores/chat.svelte';
	import MessageInput from '$lib/components/MessageInput.svelte';
	import type { ToolCallInfo } from '$lib/chat.svelte';
	import ThreadFromChat from '$lib/components/ThreadFromChat.svelte';
	
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
    let runMode = $state<'normal' | 'debug'>('normal');

    let confirmDeleteStep = $state<{ stepId: string, filename: string } | null>(null);
    let confirmDeleteStepModal = $state<ReturnType<typeof ConfirmDelete> | null>(null);

    const headerHeight = 72;
    let scrollContainer = $state<HTMLElement | null>(null);
    let argumentsList = $state<HTMLDivElement | null>(null);

    let stepBlockEditing = new SvelteMap<number | string, boolean>();
    let stepDescription = new SvelteMap<number | string, boolean>();

    let showSidebarThread = $state(true);
    let sidebarWidth = $state(520);
    let isResizing = $state(false);
    let toggleableMessageInput = $state<ReturnType<typeof MessageInput> | null>(null);

    let registryToolSelector = $state<ReturnType<typeof RegistryToolSelector> | null>(null);
    let currentAddingToolForStep = $state<StepType | null>(null);

    let showMessageInput = $state(false);
    let includeFilesInMessage = $state<Attachment[]>([]);

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
    
    layout.setIsSidebarCollapsed(true);

    /** Handle file modifications from chat to update task steps */
    async function handleFileModified(info: ToolCallInfo) {
        if (!task || !taskId || !workspace) return;
        
        const filePath = info.filePath;
        const taskPrefix = `.nanobot/tasks/${taskId}/`;
        const workspacePrefix = `/workspace/${taskPrefix}`;
        
        // Check if the modified file belongs to this task
        let relativePath = '';
        if (filePath.startsWith(workspacePrefix)) {
            relativePath = filePath.replace(`/workspace/`, '');
        } else if (filePath.startsWith(taskPrefix)) {
            relativePath = filePath;
        }
        
        if (!relativePath || !relativePath.startsWith(taskPrefix)) return;
        
        // Extract step identifier (e.g., "TASK.md" or "STEP_1.md")
        const stepFile = relativePath.replace(taskPrefix, '');
        
        try {
            // Read and parse just the modified file
            const fileContent = await workspace.readFile(relativePath);
            const parsed = await parseFrontmatterMarkdown(fileContent);
            
            // Find the step with matching id and update it
            const stepIndex = task.steps.findIndex(s => s.id === stepFile);
            if (stepIndex !== -1) {
                // Update the step in place
                task.steps[stepIndex] = {
                    ...task.steps[stepIndex],
                    name: parsed.name,
                    description: parsed.description,
                    content: parsed.content,
                    tools: parsed.tools
                };
                
                // If this is the TASK.md file, also update task-level properties
                if (stepFile === 'TASK.md') {
                    task.name = parsed.taskName;
                    task.description = parsed.taskDescription;
                    // Update inputs if provided
                    if (parsed.inputs.length > 0) {
                        task.inputs = parsed.inputs;
                    }
                }
            }
        } catch (error) {
            console.error('Failed to update step after file modification:', error);
        }
    }

    onMount(() => {
        const isMock = mocks.workspaceIds.includes(workspaceId);
        workspace = isMock ? mocks.workspaceInstances[workspaceId] : workspaceService.getWorkspace(workspaceId);
        registryStore.fetch();
    });
    
    onDestroy(() => {
        sharedChat.current?.close();
    });

    $effect(() => {
        if (workspace && !sharedChat.current) {
            initChat();
        }
    })

    async function initChat() {
        const isMock = mocks.workspaceIds.includes(workspaceId);
        const chat = isMock ? sharedChat.current : await workspace?.newSession({ editor: true });
        if (chat) {
            if (!isMock) {
                chat.setCallbacks({
                    onFileModified: handleFileModified,
                    onChatDone: () => {
                        workspace?.load();
                    }
                });
            }
            setSharedChat(chat);
        }
    }

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
                    goto(resolve(url.pathname + url.search as `/w/${string}/t/`), { replaceState: true, keepFocus: true });
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

        if (mocks.taskIds.includes(idToUse)) {
            task = mocks.taskData[idToUse];
        } else {
            task = await convertToTask(workspace, files, idToUse);
        }
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
        
        if (urlTaskId && workspace && urlTaskId !== taskId && files.length > 0) {
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
            submitRun();
        }
    }

    function startResize(e: MouseEvent) {
        e.preventDefault();
        isResizing = true;
        const startX = e.clientX;
        const startWidth = sidebarWidth;
        
        function onMouseMove(e: MouseEvent) {
            // Sidebar is on the left side, so dragging right increases width
            const delta = e.clientX - startX;
            sidebarWidth = Math.max(400, Math.min(startWidth + delta, window.innerWidth - 800));
        }
        
        function onMouseUp() {
            isResizing = false;
            document.removeEventListener('mousemove', onMouseMove);
            document.removeEventListener('mouseup', onMouseUp);
        }
        
        document.addEventListener('mousemove', onMouseMove);
        document.addEventListener('mouseup', onMouseUp);
    }

    // TODO: change below to actually hit the run task endpoint once available
    async function submitRun() {
        // navigate to run page
        const created = new Date().toISOString();
        const runId = crypto.randomUUID();

        if (mocks.taskIds.includes(taskId)) {
            const content = { id: runId, created, arguments: runFormData };
            mockTasks.addRun(taskId, content);
            goto(resolve(`/w/${workspaceId}/t?id=${taskId}&runId=${runId}`), { keepFocus: true });
        } else {
            // TODO: actual impl
        }
    }
</script>

{#if initialLoadComplete && task}
    <div class="flex w-full h-dvh {isResizing ? 'cursor-col-resize' : ''}">
        {#if showSidebarThread}
            <!-- Sidebar Thread -->
            <div 
                transition:fly={{ x: -100, duration: 200 }} 
                class="border-l border-l-base-300 bg-base-100 h-dvh flex flex-col shrink-0 {isResizing ? 'select-none' : ''}"
                style="width: {sidebarWidth}px; min-width: 400px;"
            >
                <div class="w-full flex justify-between items-center p-4 bg-base-100 shrink-0">
                    <div class="w-full"></div>
                    <button class="btn btn-ghost btn-square btn-sm tooltip tooltip-left" data-tip="Close" 
                        onclick={() => {
                            showSidebarThread = false;
                            includeFilesInMessage = [];
                        }}
                    >
                        <X class="size-4" />
                    </button>
                </div>
                <div class="w-full flex-1 min-h-0 flex flex-col">
                    {#if sharedChat.current}
                        {#key sharedChat.current.chatId}
                            <ThreadFromChat inline chat={sharedChat.current} files={includeFilesInMessage} />
                        {/key}
                    {/if}
                </div>
            </div>
            <!-- Resize Handle -->
            <button 
                type="button"
                class="w-1 bg-base-300 hover:bg-primary/50 cursor-col-resize transition-all duration-150 shrink-0 active:bg-primary border-none p-0"
                aria-label="Resize sidebar"
                onmousedown={startResize}
            ></button>
        {/if}
        <div class="
            flex flex-col grow p-4 pt-0 overflow-y-auto max-h-dvh transition-all duration-200 ease-in-out 
            {layout.isSidebarCollapsed && !showSidebarThread ? 'mt-13' : ''}
            {isResizing ? 'select-none' : ''}
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
                            <input name="title" class="input input-ghost input-lg w-full placeholder:text-base-content/30 font-semibold" type="text" placeholder="Workflow title" 
                                bind:value={task.name}
                            />
                        {:else}
                            <div class="w-full"></div>
                        {/if}
                        <div class="flex shrink-0 items-center gap-2">
                            <div class="flex">
                                <button class="btn btn-primary rounded-r-none w-48" onclick={handleRun}>
                                    {runMode === 'normal' ? 'Run' : 'Debug'} <Play class="size-4" /> 
                                </button>
                                <div class="dropdown dropdown-end">
                                    <div tabindex="0" role="button" class="btn rounded-l-none btn-primary btn-square border-l-white">
                                        <ChevronDown class="size-4" />
                                    </div>
                                    <ul tabindex="-1" class="menu dropdown-content bg-base-100 rounded-box z-1 w-64 p-2 shadow-sm">
                                        <li>
                                            <button class="flex flex-col gap-0 items-start"
                                                onclick={() => {
                                                    runMode = 'normal';
                                                    (document.activeElement as HTMLElement)?.blur();
                                                }}
                                            >
                                                <span class="flex items-center gap-2 font-medium"><Play class="size-3" /> Run</span>
                                                <span class="text-xs text-base-content/50">Perform a standard workflow run, runs all steps at once and outputs a summarized result.</span>
                                            </button>
                                        </li>
                                        <li>
                                            <button class="flex flex-col gap-0 items-start"
                                                onclick={() => {
                                                    runMode = 'debug';
                                                    (document.activeElement as HTMLElement)?.blur();
                                                }}
                                            >
                                                <span class="flex items-center gap-2 font-medium"><Bug class="size-3" /> Debug</span>
                                                <span class="text-xs text-base-content/50">Run the workflow in debug mode, runs step by step and outputs any errors that occur.</span>
                                            </button>
                                        </li>
                                    </ul>
                                </div>
                            </div>
                            <button class="btn btn-ghost btn-square" popoverTarget="task-actions" style="anchor-name: --task-actions-anchor;">
                                <EllipsisVertical class="text-base-content/50" />
                            </button>
                        
                            <ul class="dropdown flex flex-col gap-1 dropdown-end dropdown-bottom menu w-64 rounded-box bg-base-100 dark:bg-base-300 shadow-sm border border-base-300"
                                popover="auto" id="task-actions" style="position-anchor: --task-actions-anchor;">
                                <li>
                                    <label for="task-title" class="flex gap-2 justify-between items-center">
                                        <span class="flex items-center gap-2">
                                            <PencilLine class="size-4" />
                                            Workflow title
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
                                            Workflow description
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
                        <input out:slide={{ axis: 'y' }} name="description" class="input input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Workflow description"
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
                        showDescription={stepDescription.get(step.id) ?? false}
                        showBlockEditing={stepBlockEditing.get(step.id) ?? false}
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
                        {visibleInputs}
                        onUpdateVisibleInputs={(inputs) => visibleInputs = inputs}
                        onSuggestImprovement={async (file) => {
                            if (!includeFilesInMessage.some((f) => f.uri === file.uri)) {
                                includeFilesInMessage.push(file);
                            }
                            showMessageInput = true;
                        }}
                    />
                {/snippet}
            </DragDropList>

            <div class="flex items-center justify-center">
                <button class="btn btn-primary btn-square tooltip" data-tip="Add new step"
                    onclick={() => {
                        const newStep = {
                            id: `STEP_${task!.steps.length}.md`,
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
                                    // isStickToBottom = true; // Reset stick-to-bottom when showing sidebar
                                    showSidebarThread = true;
                                    showMessageInput = false;

                                    return sharedChat.current?.sendMessage(message, includeFilesInMessage.length > 0 ? includeFilesInMessage : undefined);
                                }} 
                            >
                                {#snippet customActions()}
                                    {#if includeFilesInMessage.length > 0}
                                        {#each includeFilesInMessage as file (file.uri)}
                                            <div class="badge badge-sm badge-primary gap-1 group">
                                                <button class="hidden group-hover:block cursor-pointer text-white/50 hover:text-white transition-colors" 
                                                    onclick={() => {
                                                        includeFilesInMessage = includeFilesInMessage.filter((f) => f.uri !== file.uri);
                                                    }} 
                                                >
                                                    <X class="size-3" />
                                                </button>
                                                <File class="size-3 block group-hover:hidden" />
                                                {file.name}
                                            </div>
                                        {/each}
                                    {/if}
                                {/snippet}
                            </MessageInput>
                        </div>  
                    {/if}

                    <button class="float-right btn btn-lg btn-circle btn-primary self-end tooltip tooltip-left" onclick={async () => {
                        showMessageInput = !showMessageInput;
                        await tick();
                        toggleableMessageInput?.focus();
                    }} data-tip={showMessageInput ? 'Hide chat' : 'Show chat'}>
                        <MessageCircleMore class="size-6" />
                    </button>
                </div>
            {/if}
        </div>
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
        confirmDeleteStep = null;
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

<style lang="postcss">
    :global(.task-run-sidebar-content .prose) {
        font-size: var(--text-sm);
    }
</style>