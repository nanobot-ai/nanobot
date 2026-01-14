<script lang="ts">
	import '$lib/../app.css';
    import { resolve } from '$app/paths';
	import DragDropList from '$lib/components/DragDropList.svelte';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	import { createRegistryStore, setRegistryContext } from '$lib/context/registry.svelte';
	import { EllipsisVertical, GripVertical, PencilLine, Play, Plus, ReceiptText, X, MessageCircleMore, Square } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { fade, fly, slide } from 'svelte/transition';
	import { afterNavigate, goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Attachment, ChatMessage, WorkspaceClient, WorkspaceFile } from '$lib/types';
	import { getNotificationContext } from '$lib/context/notifications.svelte';
	import { type Input, type Task, type Step as StepType } from './types';
	import { compileOutputFiles, convertToTask, setupEmptyTask } from './utils';
	import StepActions from './StepActions.svelte';
	import TaskInputActions from './TaskInputActions.svelte';
    import TaskInput from './TaskInput.svelte';
	import Step from './Step.svelte';
	import RegistryToolSelector from './RegistryToolSelector.svelte';
	import { onMount } from 'svelte';
	import ConfirmDelete from '$lib/components/ConfirmDelete.svelte';
    import { sharedChat } from '$lib/stores/chat.svelte';
	import { ChatService } from '$lib/chat.svelte';
	import { renderMarkdown } from '$lib/markdown';
	import ThreadFromChat from '$lib/components/ThreadFromChat.svelte';
	import TaskRunInputs from './TaskRunInputs.svelte';
	import StepRun from '../StepRun.svelte';

    type Props = {
        workspace: WorkspaceClient;
        urlTaskId?: string;
    }
    let { workspace, urlTaskId }: Props = $props();

	const registryStore = createRegistryStore();
	setRegistryContext(registryStore);

    let taskId = $state('');
    let task = $state<Task | null>(null);
    let initialLoadComplete = $state(false);

    let showTaskTitle = $state(false);
    let showTaskDescription = $state(false);

    let inputsModal = $state<ReturnType<typeof TaskRunInputs> | null>(null);

    let confirmDeleteStep = $state<{ stepId: string, filename: string } | null>(null);
    let confirmDeleteStepModal = $state<ReturnType<typeof ConfirmDelete> | null>(null);

    const headerHeight = 72;
    let scrollContainer = $state<HTMLElement | null>(null);
    let argumentsList = $state<HTMLDivElement | null>(null);

    let stepBlockEditing = new SvelteMap<number | string, boolean>();
    let stepDescription = new SvelteMap<number | string, boolean>();

    let showSidebarThread = $state(true);
    let sidebarWidth = $state(325);
    let isResizing = $state(false);

    let registryToolSelector = $state<ReturnType<typeof RegistryToolSelector> | null>(null);
    let currentAddingToolForStep = $state<StepType | null>(null);

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

    let run = $state<ChatService | null>(null);
    let runSession = new SvelteMap<string, { stepId: string, messages: ChatMessage[], pending: boolean, completed: boolean }>();
    let running = $state(false);
    let completed = $state(false);
    
    // Extract summary text from the last message when workflow completes
    let runSummary = $derived.by(() => {
        if (!run || run.messages.length === 0) return '';
        const lastMessage = run.messages[run.messages.length - 1];
        if (!lastMessage.items) return '';
        // Find the text item in the last message
        const textItem = lastMessage.items.find(item => item.type === 'text' && 'text' in item);
        const text = textItem && 'text' in textItem ? textItem.text : '';
        return text ? renderMarkdown(text) : '';
    });

    const notifications = getNotificationContext();
    const layout = getLayoutContext();

    type StepSession = { stepId: string; messages: ChatMessage[]; pending: boolean; completed: boolean };
    type SessionData = Record<string, StepSession>;

    function isSummaryMessage(message: ChatMessage): boolean {
        const items = message?.items ?? [];
        return items.length > 0 && items.every(item => item.type === 'text');
    }

    function parseToolArgs(item: { arguments?: string }): Record<string, unknown> | null {
        try {
            return JSON.parse(item.arguments || '{}');
        } catch {
            return null;
        }
    }

    function getStepIdFromExecuteTask(args: Record<string, unknown>, steps: StepType[]): string {
        if (args.filename) {
            const step = steps.find(s => s.id === args.filename);
            return step?.id ?? '';
        }
        return steps[0]?.id ?? '';
    }

    function createStepSession(stepId: string): StepSession {
        return { stepId, messages: [], pending: true, completed: false };
    }

    function processExecuteTaskStep(
        item: { name?: string; arguments?: string },
        steps: StepType[],
        sessionData: SessionData,
        currentStepId: string
    ): string {
        const args = parseToolArgs(item);
        if (!args) return currentStepId;

        const stepId = getStepIdFromExecuteTask(args, steps);
        if (stepId && !sessionData[stepId]) {
            sessionData[stepId] = createStepSession(stepId);
        }
        return stepId || currentStepId;
    }

    function processTaskStepStatus(
        item: { name?: string; arguments?: string },
        sessionData: SessionData,
        activeStepId: string
    ): void {
        if (!activeStepId) return;

        const args = parseToolArgs(item);
        if (!args) return;

        const session = sessionData[activeStepId];
        if (session && (args.status === 'succeeded' || args.status === 'failed')) {
            session.pending = false;
            session.completed = true;
        }
    }

    function processMessage(
        message: ChatMessage,
        steps: StepType[],
        sessionData: SessionData,
        activeStepId: string,
        skipMessage: boolean
    ): string {
        const items = message.items ?? [];

        // Process tool calls in this message
        for (const item of items) {
            if (item.type === 'tool' && 'name' in item) {
                if (item.name === 'ExecuteTaskStep') {
                    activeStepId = processExecuteTaskStep(item, steps, sessionData, activeStepId);
                } else if (item.name === 'TaskStepStatus') {
                    processTaskStepStatus(item, sessionData, activeStepId);
                }
            }
        }

        if (!skipMessage && activeStepId && sessionData[activeStepId]) {
            const session = sessionData[activeStepId];
            if (!session.messages.some(m => m.id === message.id)) {
                session.messages.push(message);
            }
        }

        return activeStepId;
    }

    function buildSessionData(messages: ChatMessage[], steps: StepType[]): SessionData {
        const sessionData: SessionData = {};
        let activeStepId = '';

        const lastMessage = messages[messages.length - 1];
        const hasTrailingSummary = lastMessage && isSummaryMessage(lastMessage);

        for (let i = 0; i < messages.length; i++) {
            const isLastMessage = i === messages.length - 1;
            const skipMessage = isLastMessage && hasTrailingSummary;

            activeStepId = processMessage(messages[i], steps, sessionData, activeStepId, skipMessage);
        }

        return sessionData;
    }

    function areAllStepsCompleted(steps: StepType[], sessionData: SessionData): boolean {
        return steps.every(step => sessionData[step.id]?.completed);
    }

    $effect(() => {
        // Update runSession during a task run
        if (!run || !task || run.messages.length === 0) return;

        const sessionData = buildSessionData(run.messages, task.steps);

        for (const [stepId, data] of Object.entries(sessionData)) {
            runSession.set(stepId, data);
        }

        if (areAllStepsCompleted(task.steps, sessionData)) {
            completed = true;
        }
    })

    onMount(() => {
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

        if (task.inputs.length > 0) {
            inputsModal?.showModal();
        } else {
            submitRun();
        }
    }

    function cancelRun() {
        running = false;
        completed = false;
        runSession.clear();

        if (run) {
            run.close();
            run = null;
        }

        notifications.info('Workflow cancelled');
    }

    function startResize(e: MouseEvent) {
        e.preventDefault();
        isResizing = true;
        const startX = e.clientX;
        const startWidth = sidebarWidth;

        function onMouseMove(e: MouseEvent) {
            // Sidebar is on the right side, so dragging left increases width
            const delta = startX - e.clientX;
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

    async function submitRun(formData?: (Input & { value: string })[]) {
        if (!task || !workspace) return;
        // reset 
        runSession.clear();
        completed = false;

        running = true;
        run = await workspace.newSession();
        await run?.sendToolCall('ExecuteTaskStep', {taskName: taskId, arguments: formData});
    }
</script>

{#if initialLoadComplete && task}
    <div class="flex w-full h-dvh {isResizing ? 'cursor-col-resize' : ''}">
        <div class="
            flex flex-col grow p-4 pt-0 overflow-y-auto max-h-dvh transition-all duration-150
            {isResizing ? 'select-none' : ''}
        "
            bind:this={scrollContainer}
            onscroll={() => {
                showAlternateHeader = (scrollContainer?.scrollTop ?? 0) > 100;
            }}
        >
            <div class="sticky top-0 left-0 w-full bg-base-200 dark:bg-base-100 z-10 py-4">
                <div in:fade class="flex flex-col grow">
                    <div class="flex w-full items-center gap-4 {layout.isSidebarCollapsed ? 'pl-68' : ''}">
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
                                <button class="btn btn-primary w-48 {running ? 'tooltip tooltip-bottom' : ''}" data-tip="Cancel current run"
                                    onclick={() => {
                                        if (running && !completed) {
                                            cancelRun();
                                        } else {
                                            handleRun();
                                        }
                                    }}
                                >
                                    {#if running && !completed}
                                        <Square class="size-4" />
                                    {:else}
                                        Run <Play class="size-4" />
                                    {/if}
                                </button>
                                <!-- <div class="dropdown dropdown-end">
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
                                </div> -->
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
                            //TODO:
                        }}
                    >
                        {#if running}
                            <div in:fade={{ duration: 150 }}>
                                <StepRun
                                    messages={runSession.get(step.id)?.messages ?? []}
                                    pending={runSession.get(step.id)?.pending ?? false}
                                />
                            </div>
                        {/if}
                    </Step>
                {/snippet}
            </DragDropList>

            {#if completed}
                <div in:fade out:slide={{ axis: 'y', duration: 150 }} class="w-full flex flex-col justify-center items-center pl-22 pb-4 {showSidebarThread ? '' : 'md:pr-22'}">
                    <div class="w-full flex flex-col justify-center items-center border border-transparent dark:border-base-300 bg-base-100 dark:bg-base-200 shadow-xs rounded-field p-6 pb-8">
                        <h4 class="text-xl font-semibold">Workflow Completed</h4>
                        <p class="text text-base-content/50 text-center mt-1">
                            The workflow has completed successfully. Here are your summarized results:
                        </p>

                        <p class="text text-center mt-4">The workflow completed <b>{task.steps.length}</b> out of <b>{task.steps.length}</b> steps.</p>
                        <!-- <p class="text-sm text-center mt-1">It took a total time of <b>{(totalTime / 1000).toFixed(1)}s</b> to complete.</p>
                        <p class="text-sm text-center mt-1">A total of <b>{totalTokens}</b> tokens were used.</p> -->
                        {#if runSummary}
                            <div class="prose mt-4 text-left w-full max-w-none">
                                {@html runSummary}
                            </div>
                        {/if}
                    </div>
                </div>
            {/if}

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
                    <button class="float-right btn btn-lg btn-circle btn-primary self-end tooltip tooltip-left" onclick={async () => {
                        showSidebarThread = true;
                    }} data-tip="Show chat">
                        <MessageCircleMore class="size-6" />
                    </button>
                </div>
            {/if}
        </div>
        {#if showSidebarThread}
            <!-- Resize Handle -->
            <button
                type="button"
                class="w-1 bg-base-300 hover:bg-primary/50 cursor-col-resize transition-all duration-150 shrink-0 active:bg-primary border-none p-0"
                aria-label="Resize sidebar"
                onmousedown={startResize}
            ></button>
            <!-- Sidebar Thread -->
            <div
                transition:fly={{ x: 100, duration: 200 }}
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
                        <ThreadFromChat inline chat={sharedChat.current} files={includeFilesInMessage} agent={{ name: 'Workflow Assistant', icon: '/assets/obot-icon-blue.svg' }} />
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

<TaskRunInputs bind:this={inputsModal} onSubmit={submitRun} {task} additionalInputs={visibleInputs} />

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
        console.log(names);
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
