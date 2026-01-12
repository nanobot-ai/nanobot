<script lang="ts">
	import { ChatService } from "$lib/chat.svelte";
	import { createRegistryStore, setRegistryContext } from "$lib/context/registry.svelte";
	import type { WorkspaceClient, WorkspaceFile } from "$lib/types";
	import type { Input, Task } from "./types";
    import * as mocks from '$lib/mocks';
	import { convertToTask } from "./utils";
	import Step from "./Step.svelte";
	import { SvelteMap } from "svelte/reactivity";
	import { WorkspaceService } from "$lib/workspace.svelte";
	import { onDestroy, onMount, untrack } from "svelte";
	import { mockTasks } from "$lib/mocks/stores/tasks.svelte";
	import StepRun from "../StepRun.svelte";
	import { fade } from "svelte/transition";

    type Props = {
        workspaceId: string;
        urlTaskId: string;
        runId: string;
    }
    
    let { workspaceId, urlTaskId, runId }: Props = $props();

    const workspaceService = new WorkspaceService();
    const registryStore = createRegistryStore();
    setRegistryContext(registryStore);

    let workspace = $state<WorkspaceClient | null>(null);
    let task = $state<Task | null>(null);
    let initialLoadComplete = $state(false);
    let prevRunId = $state('');
    let prevTaskId = $state('');

    let scrollContainer = $state<HTMLElement | null>(null);

    let runCreated = $state('');
    let runSession = new SvelteMap<string, { stepId: string, thread: ChatService, pending: boolean }>();
    let runArguments = $state<(Omit<Input, 'id'> & { value: string })[]>([]);
    let showDetails = $state(false);

    let name = $derived(task?.name || task?.steps[0].name || '');

    let summaryResults = $state<{ step: string, summary: string }[]>([]);
    let timeoutHandlers = $state<ReturnType<typeof setTimeout>[]>([]);
    let totalTime = $state(0);

    async function compileTask(idToUse: string, files: WorkspaceFile[]){
        if (!workspace) return;
        initialLoadComplete = false;
        task = null;

        if (mocks.taskIds.includes(idToUse)) {
            task = mocks.taskData[idToUse];
        } else {
            task = await convertToTask(workspace, files, idToUse);
        }
        prevTaskId = idToUse;
        
        initialLoadComplete = true;
        handleRun();
    }

    onMount(() => {
        const isMock = mocks.workspaceIds.includes(workspaceId);
        workspace = isMock ? mocks.workspaceInstances[workspaceId] : workspaceService.getWorkspace(workspaceId);
        registryStore.fetch();
    });

    onDestroy(() => {
        runSession.forEach((session) => session.thread.close());
        runSession.clear();
    });

    $effect(() => {
        const files = workspace?.files ?? [];
        
        if (initialLoadComplete && (prevRunId !== runId || prevTaskId !== urlTaskId)) {
            if (prevRunId !== runId) {
                handleRun();
            } else {
                compileTask(urlTaskId, files);
            }
        } else if (urlTaskId && workspace && urlTaskId && files.length > 0 && !initialLoadComplete) {
            compileTask(urlTaskId, files);
        }
    });

    function handleRun() {
        if (mocks.taskIds.includes(urlTaskId)) {
            // mock impl
            const match = mockTasks.current.tasks.find((task) => task.id === urlTaskId)?.runs.find((run) => run.id === runId);
            if (match) {
                runArguments = match.arguments ?? [];
                runCreated = match.created;
                if (match.stepSessions) {
                    for (const stepSession of match.stepSessions) {
                        const thread = new ChatService();
                        thread.setChatId(stepSession.threadId);
                        runSession.set(stepSession.stepId, { thread, stepId: stepSession.stepId, pending: false });
                    }

                    totalTime = Math.random() * 10000 + 1000;
                    summaryResults = [...mocks.summaryResults[urlTaskId]];
                } else {
                    showDetails = true;
                    runTask();
                }
            }
        } else {
            // TODO: actual impl
        }
        untrack(() => prevRunId = runId);
    }

    // TODO: change below to actually hit the run task endpoint once available
    async function runTask() {
        if (!task) return;
        
        const startTime = Date.now();
        let stepSessions = [];
        runSession.forEach((session) => session.thread.close());
        runSession.clear();
        // isStickToBottom = true; // Reset stick-to-bottom when starting a new run
        let priorSteps = '';
        let priorResponse = '';

        for (const step of task.steps) {
            const thread = workspace?.newSession ? await workspace?.newSession() : new ChatService();
            runSession.set(step.id, { thread: thread!, stepId: step.id, pending: true });
        }

        // showSidebarThread = true;
        for (const step of task.steps) {
            runSession.set(step.id, { ...runSession.get(step.id)!, pending: false });

            const message = `
You are a task runner. You are given a task and a list of steps to run. You are to run the task step by step. \n\n

You have the following arguments:
${JSON.stringify(runArguments)} \n\n

${priorSteps.length > 0 ? `
You have already run the following steps:
${priorSteps}
` : ''} \n\n

${priorResponse.length > 0 ? `
You have given the following response to the previous step(s):
${priorResponse}
` : ''} \n\n

You are currently running the following step: \n
Step name: ${step.name} \n
Step description: ${step.description} \n
Step content: ${step.content} \n\n
\n\n You have the following tools available to you:
${step.tools.join(', ')}
\n\n Do not indicate that you are simulating or mocking any data; act as if you are actually running the task.
            `;
            
            // Wait for this thread to complete before starting the next one
            const response = await runSession.get(step.id)?.thread?.sendMessage(message);
            if (response) {
                priorResponse += `${step.id} response: ${response.message}\n\n`;
                const threadId = runSession.get(step.id)?.thread?.chatId;
                if (threadId) {
                    stepSessions.push({ stepId: step.id, threadId });
                }
            }
            
            priorSteps += `Step ${step.id}: ${step.name} \n\n`;
        }

        totalTime = Date.now() - startTime;
        const finalHandler = setTimeout(() => {
            let summaryTimer = 0;
            for (const summaryResult of mocks.summaryResults[urlTaskId]) {
                const handler = setTimeout(() => {
                    summaryResults.push(summaryResult);
                }, summaryTimer);
                summaryTimer += 1000;
                timeoutHandlers.push(handler);
            }
        }, 1000);
        timeoutHandlers.push(finalHandler);

        mockTasks.updateRun(urlTaskId, runId, stepSessions); 
    }
</script>


{#if initialLoadComplete && task}
    {@const runCreatedDate = new Date(runCreated).toLocaleString().replace(',', '')}
    <div class="flex w-full h-dvh">
        <div class="
            flex flex-col grow p-4 pt-0 overflow-y-auto max-h-dvh transition-all duration-200 ease-in-out 
        " 
            bind:this={scrollContainer}
        >
            <div class="sticky top-0 left-0 w-full bg-base-200 dark:bg-base-100 z-10 py-4">
                <div class="flex flex-col grow">
                    <div class="flex w-full items-center gap-4">
                        <input name="title" class="input input-ghost input-lg w-full font-semibold disabled:text-base-content" type="text" placeholder="Workflow title" 
                            value={`${name} - ${runCreatedDate}`} disabled
                        />
                    </div>
                </div>
            </div>
            {#if showDetails}
                <div in:fade>
                    <div class="px-22 mb-6 flex flex-col gap-4">
                        <div class="w-full h-2"></div>
                        {#each runArguments as input (input.name)}
                            <div class="flex flex-col gap-2 bg-base-100 dark:bg-base-200 shadow-xs rounded-box p-4 task-step relative">
                                <div class="flex flex-col gap-2">
                                    <label class="input w-full">
                                        <span class="label h-full font-semibold text-primary bg-primary/15 mr-0">${input.name}</span>
                                        <input type="text" class="font-semibold placeholder:font-normal" disabled value={input.value} />
                                    </label>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>
                <div in:fade class="md:pr-22 flex flex-col gap-6">
                    {#each task.steps as step (step.id)}
                        <div class="pl-22">
                            <Step 
                                taskId={urlTaskId}
                                task={task!}
                                {step}
                                showDescription
                                showBlockEditing={false}
                                readonly
                            >
                                <StepRun 
                                    messages={runSession.get(step.id)?.thread?.messages?.slice(1) ?? []} 
                                    pending={runSession.get(step.id)?.pending ?? false} 
                                    chatLoading={runSession.get(step.id)?.thread?.isLoading ?? false}
                                />
                            </Step>
                        </div>
                    {/each}
                </div>
            {/if}
            
            {#if summaryResults.length > 0}
                <div class="md:px-22 mb-12 flex flex-col justify-center {showDetails ? 'mt-6' : ''}">
                    <div class="p-6 pb-12 w-full flex flex-col justify-center items-center border border-transparent dark:border-base-300 bg-base-100 dark:bg-base-200 shadow-xs rounded-field">
                        <h4 class="text-xl font-semibold">Workflow Completed</h4>
                        <p class="text-sm text-base-content/50 text-center mt-1">
                            The workflow has completed successfully. Here are your summarized results:
                        </p>

                        <p class="text-sm text-base-content/50 text-center mt-1">Total time: {(totalTime / 1000).toFixed(1)}s, Total tokens: {(Math.floor(Math.random() * 10000) + 1000)}</p>

                        <div class="w-xl flex flex-col gap-4 mt-6">
                            {#each summaryResults as result (result.step)}
                                <div in:fade class="flex flex-col">
                                    <h4 class="font-semibold">{result.step}</h4>
                                    <p class="text-sm text-base-content/50">
                                        {result.summary}
                                    </p>
                                </div>
                            {/each}
                        </div>

                        {#if !showDetails}
                        <button class="mt-8 btn w-xs self-center" onclick={() => showDetails = true}>
                            View details
                        </button>
                        {/if}
                    </div>
                </div>
            {/if}

            <div class="flex grow"></div>
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

<style lang="postcss">
    :global(.step-agent .prose) {
        font-size: var(--text-sm);
    }
</style>