<script lang="ts">
	import '$lib/../app.css';
	import { resolve } from '$app/paths';
	import DragDropList from '$lib/components/DragDropList.svelte';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	import { createRegistryStore, setRegistryContext } from '$lib/context/registry.svelte';
	import { GripVertical, Plus, MessageCircleMore } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { fade } from 'svelte/transition';
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
	import { ChatService } from '$lib/chat.svelte';
	import TaskRunInputs from './TaskRunInputs.svelte';
	import StepRun from './StepRun.svelte';
	import Elicitation from '$lib/components/Elicitation.svelte';
	import TaskEditorHeader from './TaskEditorHeader.svelte';
	import TaskEditorRunSummary from './TaskEditorRunSummary.svelte';
	import TaskEditorRunTimeline from './TaskEditorRunTimeline.svelte';
	import TaskEditorSidebar from './TaskEditorSidebar.svelte';

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
    let closeCurrentRunTimeline = $state(false);

    // auto-scrolling during a task run in the editor
    let stepElements = new SvelteMap<string, HTMLElement>();
    let shouldAutoScroll = $state(true);
    let isProgrammaticScroll = $state(false);
    let scrollTimeout: ReturnType<typeof setTimeout> | null = null;

    let mainContainerRef = $state<HTMLElement | null>(null);
    let runSummaryComponent = $state<ReturnType<typeof TaskEditorRunSummary> | null>(null);

    const notifications = getNotificationContext();
    const layout = getLayoutContext();

    function withProgrammaticScroll(action: () => void, delay = 500) {
        isProgrammaticScroll = true;
        action();
        setTimeout(() => { isProgrammaticScroll = false; }, delay);
    }

    function resetRunState() {
        closeCurrentRunTimeline = false;
        running = false;
        runSession.clear();
        completed = false;
        error = false;
        runSummary = '';
        currentRunStepId = null;
    }

    function markAsSaved() {
        lastSavedTaskJson = JSON.stringify(task);
        lastSavedVisibleInputsJson = JSON.stringify(visibleInputs);
    }

    function clearEditorMaps() {
        stepDescription.clear();
        inputDescription.clear();
        inputDefault.clear();
    }

    function addVisibleInput(input: Input) {
        visibleInputs.push(input);
    }

    function registerStepElement(node: HTMLElement, stepId: string) {
        stepElements.set(stepId, node);
        return { destroy: () => stepElements.delete(stepId) };
    }

    function scrollToStep(stepId: string) {
        const stepElement = stepElements.get(stepId);
        if (stepElement && scrollContainer) {
            withProgrammaticScroll(() => {
                stepElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
            });
        }
        shouldAutoScroll = stepId === currentRunStepId;
    }

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

        const messageCount = runSession.get(currentRunStepId)?.messages?.length ?? 0;
        void messageCount;

        const stepElement = stepElements.get(currentRunStepId);
        if (!stepElement || !scrollContainer) return;

        requestAnimationFrame(() => {
            if (!stepElement || !scrollContainer || !shouldAutoScroll) return;
            
            const containerRect = scrollContainer.getBoundingClientRect();
            const stepRect = stepElement.getBoundingClientRect();
            const stepBottomRelative = stepRect.bottom - containerRect.top;
            
            if (stepBottomRelative > containerRect.height - 20) {
                const scrollAmount = stepBottomRelative - containerRect.height + 40;
                withProgrammaticScroll(() => {
                    scrollContainer!.scrollBy({ top: scrollAmount, behavior: 'smooth' });
                });
            }
        });
    });

    // this is to reset an auto-scroll when a new run starts
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

        resetRunState();
        initialLoadComplete = false;
        task = null;
        task = await convertToTask(workspace, files, idToUse);
        taskId = idToUse;

        clearEditorMaps();
        for (const step of task.steps) {
            if (step.description) stepDescription.set(step.id, true);
        }

        visibleInputs = [];
        for (const input of task.inputs) {
            if (input.description.length || input.default?.length) {
                visibleInputs.push(input);
                if (input.description.length) inputDescription.set(input.id, true);
                if (input.default?.length) inputDefault.set(input.id, true);
            }
        }

        showTaskTitle = (task.name || '').length > 0;
        showTaskDescription = (task.description || '').length > 0;
        markAsSaved();
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
        visibleInputs = [];
        clearEditorMaps();
        markAsSaved();
        initialLoadComplete = true;
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
        resetRunState();
        run = null;
        notifications.info('Workflow cancelled');
    }

    function startResize(e: MouseEvent) {
        e.preventDefault();
        isResizing = true;
        const startX = e.clientX;
        const startWidth = sidebarWidth;

        function onMouseMove(e: MouseEvent) {
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
        requestAnimationFrame(() => {
            setTimeout(() => {
                const runSummaryElement = runSummaryComponent?.getElement();
                if (runSummaryElement && scrollContainer) {
                    const summaryRect = runSummaryElement.getBoundingClientRect();
                    const containerRect = scrollContainer.getBoundingClientRect();
                    const targetScroll = scrollContainer.scrollTop + summaryRect.bottom - containerRect.top - containerRect.height + 40;
                    withProgrammaticScroll(() => {
                        scrollContainer!.scrollTo({ top: targetScroll, behavior: 'smooth' });
                    });
                }
            }, 100);
        });
    }

    async function submitRun(formData?: (Input & { value: string })[]) {
        if (!task || !workspace) return;

        if (run) { run.close(); run = null; }
        resetRunState();

        for (const step of task.steps) {
            runSession.set(step.id, { stepId: step.id, messages: [], pending: true, completed: false, error: false });
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
    <div bind:this={mainContainerRef} class="flex w-full h-dvh relative {isResizing ? 'cursor-col-resize' : ''}">
        <div class="
            flex flex-col grow p-4 pt-0 overflow-y-auto max-h-dvh transition-all duration-150
            {isResizing ? 'select-none' : ''}
        "
            bind:this={scrollContainer}
            onscroll={handleScroll}
        >
            <TaskEditorHeader
                {task}
                {running}
                {completed}
                {showAlternateHeader}
                {showTaskTitle}
                {showTaskDescription}
                isSidebarCollapsed={layout.isSidebarCollapsed}
                onRun={handleRun}
                onCancel={cancelRun}
                onToggleTitle={(v) => showTaskTitle = v}
                onToggleDescription={(v) => showTaskDescription = v}
            />
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
                            <TaskInputActions task={task!} availableInputs={hiddenInputs} onAddInput={addVisibleInput} />
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
                                onAddInput={addVisibleInput}
                                onOpenSelectTool={() => { currentAddingToolForStep = currentItem; registryToolSelector?.showModal(); }}
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
                            class="border-2 {currentRunStepId === step.id && !completed 
                                ? 'border-primary' 
                                    : currentRunStepId === step.id && error 
                                        ? 'border-error' : 'border-transparent'}
                            "
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
                            onAddInput={addVisibleInput}
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
                <TaskEditorRunSummary
                    bind:this={runSummaryComponent}
                    {task}
                    {completed}
                    {error}
                    {runSummary}
                    {totalRunTime}
                    {totalRunTokens}
                    {showSidebarThread}
                />
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
            <TaskEditorSidebar
                width={sidebarWidth}
                {isResizing}
                files={includeFilesInMessage}
                onClose={() => {
                    showSidebarThread = false;
                    includeFilesInMessage = [];
                }}
                onStartResize={startResize}
            />
        {/if}

        {#if running && !closeCurrentRunTimeline}
            <TaskEditorRunTimeline
                {task}
                {currentRunStepId}
                {completed}
                {error}
                {runSummary}
                parentElement={mainContainerRef}
                onStepClick={scrollToStep}
                onSummaryClick={scrollToRunSummary}
                onClose={() => closeCurrentRunTimeline = true}
            />
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