<script lang="ts">
	import '$lib/../app.css';
    import { resolve } from '$app/paths';
	import DragDropList from '$lib/components/DragDropList.svelte';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	import { createRegistryStore, setRegistryContext } from '$lib/context/registry.svelte';
	import { EllipsisVertical, GripVertical, PencilLine, Play, Plus, ReceiptText, X, MessageCircleMore, Square, CircleCheck, LoaderCircle, Circle, CircleAlert } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { fade, fly, slide } from 'svelte/transition';
	import { afterNavigate, goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Attachment, WorkspaceClient, WorkspaceFile } from '$lib/types';
	import { getNotificationContext } from '$lib/context/notifications.svelte';
	import { type Input, type Task, type Step as StepType, type StepSession } from './types';
	import { areAllStepsCompleted, buildRunSummary, buildSessionData, compileOutputFiles, convertToTask, setupEmptyTask } from './utils';
	import StepActions from './StepActions.svelte';
	import TaskInputActions from './TaskInputActions.svelte';
    import TaskInput from './TaskInput.svelte';
	import Step from './Step.svelte';
	import RegistryToolSelector from './RegistryToolSelector.svelte';
	import { onMount } from 'svelte';
	import ConfirmDelete from '$lib/components/ConfirmDelete.svelte';
    import { sharedChat } from '$lib/stores/chat.svelte';
	import { ChatService } from '$lib/chat.svelte';
	import ThreadFromChat from '$lib/components/ThreadFromChat.svelte';
	import TaskRunInputs from './TaskRunInputs.svelte';
	import StepRun from './StepRun.svelte';
	import Elicitation from '$lib/components/Elicitation.svelte';

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
    let runSession = new SvelteMap<string, StepSession>();
    let running = $state(false);
    let completed = $state(false);
    let error = $state(false);
    
    let runSummary = $state('');
    let currentRunStepId = $state<string | null>(null);
    let totalRunTime = $state(0);
    let totalRunTokens = $state(0);

    // Auto-scroll state for keeping current running step in view
    let stepElements = new SvelteMap<string, HTMLElement>();
    let shouldAutoScroll = $state(true);
    let isProgrammaticScroll = $state(false); // Flag to ignore programmatic scroll events
    let scrollTimeout: ReturnType<typeof setTimeout> | null = null;

    // Svelte action to register step elements for auto-scroll
    function registerStepElement(node: HTMLElement, stepId: string) {
        stepElements.set(stepId, node);
        return {
            destroy() {
                stepElements.delete(stepId);
            }
        };
    }

    // Reference to run summary element for scrolling on completion
    let runSummaryElement = $state<HTMLElement | null>(null);

    const notifications = getNotificationContext();
    const layout = getLayoutContext();

    $effect(() => {
        // Update runSession during a task run
        if (!run || !task || run.messages.length === 0 || completed) return;

        let lastStepWithMessages: string | null = null;
            
        const sessionData = buildSessionData(run.messages, task.steps);
        for (const [stepId, data] of Object.entries(sessionData)) {
            if (sessionData[stepId]?.messages?.length > 0) {
                lastStepWithMessages = stepId;
            }
            runSession.set(stepId, data);
        }
        currentRunStepId = lastStepWithMessages;
    })

    // Check if user is near the bottom of the scroll container
    function isNearBottom(): boolean {
        if (!scrollContainer) return true;
        const threshold = 100; // pixels from bottom to consider "at bottom"
        const { scrollTop, scrollHeight, clientHeight } = scrollContainer;
        return scrollHeight - scrollTop - clientHeight < threshold;
    }

    // Handle scroll events to detect user manual scrolling
    function handleScroll() {
        showAlternateHeader = (scrollContainer?.scrollTop ?? 0) > 100;

        // Ignore scroll events caused by programmatic scrolling
        if (isProgrammaticScroll) return;

        // Clear existing timeout
        if (scrollTimeout) {
            clearTimeout(scrollTimeout);
        }

        // Debounce to detect when scroll stops, then check position
        scrollTimeout = setTimeout(() => {
            // User scrolled manually, check if they're near the bottom
            if (isNearBottom()) {
                shouldAutoScroll = true;
            } else {
                shouldAutoScroll = false;
            }
        }, 150);
    }

    // Auto-scroll effect for current running step
    $effect(() => {
        if (!running || completed || !currentRunStepId || !shouldAutoScroll) return;

        // Track message count changes to trigger scroll on new content
        const currentSession = runSession.get(currentRunStepId);
        const messageCount = currentSession?.messages?.length ?? 0;
        // Access messageCount to make this effect reactive to message changes
        void messageCount;

        const stepElement = stepElements.get(currentRunStepId);
        if (!stepElement || !scrollContainer) return;

        // Use requestAnimationFrame to ensure DOM has updated
        requestAnimationFrame(() => {
            if (!stepElement || !scrollContainer || !shouldAutoScroll) return;
            
            const containerRect = scrollContainer.getBoundingClientRect();
            const stepRect = stepElement.getBoundingClientRect();
            
            // Calculate where the bottom of the step is relative to the container
            const stepBottomRelativeToContainer = stepRect.bottom - containerRect.top;
            const containerVisibleHeight = containerRect.height;
            
            // If the step's bottom is below the visible area, scroll to show it
            if (stepBottomRelativeToContainer > containerVisibleHeight - 20) {
                const scrollAmount = stepBottomRelativeToContainer - containerVisibleHeight + 40;
                
                // Mark as programmatic scroll to ignore scroll events
                isProgrammaticScroll = true;
                scrollContainer.scrollBy({
                    top: scrollAmount,
                    behavior: 'smooth'
                });
                // Clear flag after scroll animation completes
                setTimeout(() => {
                    isProgrammaticScroll = false;
                }, 500);
            }
        });
    });

    // Reset auto-scroll when a new run starts
    $effect(() => {
        if (running && !completed) {
            shouldAutoScroll = true;
        }
    });

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

        // reset if there was a run
        running = false;
        runSession.clear();
        completed = false;
        
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

    function scrollToRunSummary() {
        // Wait for DOM to update with run summary, then scroll to bottom
        requestAnimationFrame(() => {
            setTimeout(() => {
                if (runSummaryElement && scrollContainer) {
                    const summaryRect = runSummaryElement.getBoundingClientRect();
                    const containerRect = scrollContainer.getBoundingClientRect();
                    const summaryBottomRelativeToContainer = summaryRect.bottom - containerRect.top;
                    
                    // Mark as programmatic scroll to ignore scroll events
                    isProgrammaticScroll = true;
                    // Scroll to show the bottom of the summary with some padding
                    const targetScroll = scrollContainer.scrollTop + summaryBottomRelativeToContainer - containerRect.height + 40;
                    scrollContainer.scrollTo({
                        top: targetScroll,
                        behavior: 'smooth'
                    });
                    // Clear flag after scroll animation completes
                    setTimeout(() => {
                        isProgrammaticScroll = false;
                    }, 500);
                }
            }, 100); // Small delay to ensure the summary element is rendered
        });
    }

    async function submitRun(formData?: (Input & { value: string })[]) {
        if (!task || !workspace) return;

        if (run) {
            run.close();
            run = null;
        }
        // reset all run state
        runSession.clear();
        completed = false;
        error = false;
        runSummary = '';
        currentRunStepId = null;

        for (const step of task.steps) {
            runSession.set(step.id, {
                stepId: step.id,
                messages: [],
                pending: true,
                completed: false,
                error: false,
            });
        }

        const initialTime = Date.now();
        running = true;
        run = await workspace.newSession();
        run.setCallbacks({
            onChatStart: () => {
                workspace?.load();
            },
            onChatDone: () => {
                completed = true;
                const sessionData = buildSessionData(run?.messages ?? [], task?.steps ?? []);
                error = !areAllStepsCompleted(task?.steps ?? [], sessionData);

                for (const step of task?.steps ?? []) {
                    runSession.set(step.id, {
                        stepId: step.id,
                        messages: sessionData[step.id]?.messages ?? [],
                        pending: sessionData[step.id]?.pending ?? false,
                        completed: sessionData[step.id]?.completed ?? false,
                        error: sessionData[step.id]?.messages ? sessionData[step.id]?.messages.length === 0 : true,
                    });
                }

                totalRunTime = Date.now() - initialTime;
                totalRunTokens = Math.floor(Math.random() * 7000) + 1000;

                if (!error) {
                    runSummary = buildRunSummary(run?.messages || []);
                    if (shouldAutoScroll) {
                        scrollToRunSummary();
                    }
                }
            }
        });
        await run?.sendToolCall('ExecuteTaskStep', {taskName: taskId, arguments: formData});
    }
</script>

{#if initialLoadComplete && task}
    <div class="flex w-full h-dvh relative {isResizing ? 'cursor-col-resize' : ''}">
        <div class="
            flex flex-col grow p-4 pt-0 overflow-y-auto max-h-dvh transition-all duration-150
            {isResizing ? 'select-none' : ''}
        "
            bind:this={scrollContainer}
            onscroll={handleScroll}
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
                    <div use:registerStepElement={step.id}>
                        <Step
                            class={currentRunStepId === step.id && !completed 
                                ? 'border-primary border-2' 
                                    : currentRunStepId === step.id && error 
                                        ? 'border-error border-2' : ''
                            }
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
                                        error={runSession.get(step.id)?.error ?? false}
                                    />
                                </div>
                            {/if}
                        </Step>
                    </div>
                {/snippet}
            </DragDropList>

            
            {#if running}
                <div bind:this={runSummaryElement} in:fade out:slide={{ axis: 'y', duration: 150 }} class="w-full flex flex-col justify-center items-center pl-22 pb-4 {showSidebarThread ? '' : 'md:pr-22'}">
                    <div class="w-full flex flex-col justify-center items-center border border-transparent dark:border-base-300 bg-base-100 dark:bg-base-200 shadow-xs rounded-field p-6 pb-8">
                        {#if completed && !error}
                            
                                <h4 class="text-xl font-semibold">Workflow Completed</h4>
                                <p class="text text-base-content/50 text-center mt-1">
                                    The workflow has completed successfully. Here are your summarized results:
                                </p>

                                {#if runSummary}
                                    <div class="prose mt-4 text-left w-full max-w-none">
                                        {@html runSummary}
                                    </div>
                                    <div class="divider"></div>
                                {/if}

                                <p class="text-sm text-center">The workflow completed <b>{task.steps.length}</b> out of <b>{task.steps.length}</b> steps.</p>
                                <p class="text-sm text-center mt-1">It took a total time of <b>{(totalRunTime / 1000).toFixed(1)}s</b> to complete.</p>
                                <p class="text-sm text-center mt-1">A total of <b>{totalRunTokens}</b> tokens were used.</p>
                            
                        {:else if error}
                            <h4 class="text-xl font-semibold">Workflow Failed</h4>
                            <p class="text-sm text-base-content/50 text-center mt-1">
                                The workflow did not complete successfully. Please try again.
                            </p>
                        {:else}
                            <div class="skeleton skeleton-text">
                                The workflow is running...
                            </div>
                        {/if}
                    </div>
                </div>
            {/if}

            <div class="flex items-center justify-center">
                <button class="btn btn-primary btn-square tooltip {showSidebarThread ? 'translate-x-11' : ''}" data-tip="Add new step"
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

        {#if running}
        <div class="absolute bottom-4 left-4" in:fade>
            <ul in:fade class="timeline timeline-snap-icon timeline-vertical timeline-compact grow">
                {#if task}
                    {#each task.steps as step, index (step.id)}
                        {@const isBeforeCurrentStep = index < task.steps.findIndex(s => s.id === currentRunStepId)}
                        <li>
                            {#if index > 0}
                                <hr class="timeline-connector w-0.5 {isBeforeCurrentStep ? 'completed' : ''}" />
                            {/if}
                            <div class="timeline-middle">
                                {#if isBeforeCurrentStep || (currentRunStepId === step.id && completed)}
                                    <CircleCheck class="size-5 text-primary" />
                                {:else if currentRunStepId === step.id && !error && !completed}
                                    <LoaderCircle class="size-5 animate-spin shrink-0 text-base-content/50" />
                                {:else if currentRunStepId === step.id && error}
                                    <CircleAlert class="size-5 text-error/50" />
                                {:else}
                                    <Circle class="size-5 text-base-content/50" />
                                {/if}   
                            </div>
                            <button class="timeline-end timeline-box py-2 cursor-pointer hover:bg-base-200" onclick={() => {
                                const stepElement = stepElements.get(step.id);
                                if (stepElement && scrollContainer) {
                                    // Mark as programmatic scroll
                                    isProgrammaticScroll = true;
                                    
                                    // Scroll the step into view
                                    stepElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
                                    
                                    // Clear programmatic flag after animation
                                    setTimeout(() => {
                                        isProgrammaticScroll = false;
                                    }, 500);
                                }
                                
                                // Manage auto-scroll based on which step was clicked
                                if (step.id === currentRunStepId) {
                                    // Clicked on current step - resume auto-scrolling
                                    shouldAutoScroll = true;
                                } else {
                                    // Clicked on a different step - stop auto-scrolling
                                    shouldAutoScroll = false;
                                }
                            }}>
                                <div>
                                    {step.name} 
                                </div>
                            </button>
                            <hr class="timeline-connector w-0.5 {isBeforeCurrentStep ? 'completed' : ''}" />
                        </li>
                    {/each}
                    <li>
                        <hr class="timeline-connector w-0.5 {runSummary && completed ? 'completed' : ''}" />
                        <div class="timeline-middle">
                            {#if completed && runSummary}
                                <CircleCheck class="size-5 text-primary" />
                            {:else}
                                <Circle class="size-5 text-base-content/50" />
                            {/if}
                        </div>
                        <button class="timeline-end timeline-box py-2 {runSummary ? 'cursor-pointer hover:bg-base-200' : 'cursor-default opacity-50'}" onclick={() => {
                            scrollToRunSummary();
                        }}>
                            <div>
                                Run Summary
                            </div>
                        </button>
                    </li>
                {/if}
            </ul>
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

{#if run && run.elicitations && run.elicitations.length > 0}
    {#key run.elicitations[0].id}
        <Elicitation
            elicitation={run.elicitations[0]}
            open
            onresult={(result) => {
                run?.replyToElicitation(run.elicitations[0], result);
            }}
        />
    {/key}
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
    omit={currentAddingToolForStep?.tools.map(tool => tool.name) ?? []}
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
