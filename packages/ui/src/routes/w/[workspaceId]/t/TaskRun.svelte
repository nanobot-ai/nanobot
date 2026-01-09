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

    let runSession = new SvelteMap<string, { stepId: string, thread: ChatService, pending: boolean }>();
    let runArguments = $state<(Omit<Input, 'id'> & { value: string })[]>([]);

    let name = $derived(task?.name || task?.steps[0].name || '');

    async function compileTask(idToUse: string, files: WorkspaceFile[]){
        console.log('compileTask', idToUse);
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
                runArguments = match.arguments;
                if (match.stepSessions) {
                    for (const stepSession of match.stepSessions) {
                        const thread = new ChatService();
                        thread.setChatId(stepSession.threadId);
                        runSession.set(stepSession.stepId, { thread, stepId: stepSession.stepId, pending: false });
                    }
                } else {
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
        console.log('runTask');

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

        mockTasks.updateRun(urlTaskId, runId, stepSessions); 
    }
</script>


{#if initialLoadComplete && task}
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
                            value={name} disabled
                        />
                    </div>
                </div>
            </div>
            <div>
                <div class="px-22 mb-6 flex flex-col gap-4">
                    <div class="w-full h-2"></div>
                    {#each runArguments as input (input.name)}
                        <div class="flex flex-col gap-2 bg-base-100 dark:bg-base-200 shadow-xs rounded-box p-4 pb-8 task-step relative">
                            <div class="flex flex-col gap-2 pr-12">
                                <label class="input w-full">
                                    <span class="label h-full font-semibold text-primary bg-primary/15 mr-0">$</span>
                                    <input type="text" class="font-semibold placeholder:font-normal" disabled value={input.name} />
                                </label>
                        
                                {#if input.description.length > 0}
                                    <input name="input-description" disabled class="input w-full placeholder:text-base-content/30" type="text" value={input.description} />
                                {/if}
                                {#if input.value.length > 0}
                                    <input name="input-value" disabled  class="input w-full placeholder:text-base-content/30" type="text" value={input.value} />
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
            </div>
            <div class="md:pr-22 flex flex-col gap-6">
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
                            <StepRun messages={runSession.get(step.id)?.thread?.messages?.slice(1) ?? []} pending={runSession.get(step.id)?.pending ?? false} />
                        </Step>
                    </div>
                {/each}
            </div>

            <div class="flex grow"></div>

            <!-- TODO: whether or not to add this back in -->
            <!-- {#if !showSidebarThread}
                <div in:fade={{ duration: 200 }} class="sticky bottom-0 right-0 self-end flex flex-col gap-4 z-10">
                    {#if showMessageInput}
                        <div class="bg-base-100 dark:bg-base-200 border border-base-300 rounded-selector w-sm md:w-2xl"
                            transition:fly={{ x: 100, duration: 200 }}
                        >
                            <MessageInput bind:this={toggleableMessageInput} 
                                onSend={async (message) => {
                                    isStickToBottom = true; // Reset stick-to-bottom when showing sidebar
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
            {/if} -->
        </div>

        <!-- TODO: whether or not to add this back in -->
        <!-- {#if showSidebarThread}
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
                                runHandlers.forEach((handler) => clearTimeout(handler));
                                runSession.forEach((session) => session.thread.close());
                                runHandlers = [];
                                runSession = [];
                            }

                            includeFilesInMessage = [];
                        }}
                    >
                        <X class="size-4" />
                    </button>
                </div>
                <div class="w-full flex-1 min-h-0 flex flex-col">
                    {#if runSession}
                        <div class="h-full overflow-y-auto px-4" 
                            bind:this={runSessionScrollContainer}
                            onscroll={handleRunSessionScroll}
                        >
                            {#each runSession as session (session.stepName)}
                            <details open class="collapse {session.thread.isLoading || session.pending ? '' : 'collapse-arrow'}">
                                <summary class="collapse-title font-semibold flex items-center gap-2 {session.thread.isLoading || session.pending ? 'text-base-content/35' : ''} "
                                    onmousedown={(e) => { 
                                        if (session.thread.isLoading) {
                                            e.preventDefault();
                                            e.stopPropagation();
                                        }
                                    }}
                                >
                                    {session.stepName}
                                    {#if session.pending}
                                        <Circle class="size-4" />
                                    {:else if session.thread.isLoading}
                                        <LoaderCircle class="size-4 animate-spin" />
                                    {:else}
                                        <CircleCheck class="size-4 text-primary" />
                                    {/if}
                                </summary>
                                <div class="collapse-content task-run-sidebar-content">
                                    {#if !session.pending}
                                        <div in:fade>
                                            <Messages inline messages={session.thread.messages.slice(1)} />
                                        </div>
                                    {/if}
                                </div>
                            </details>
                            {/each}
                        </div>
                    {:else if sharedChat.current}
                        {#key sharedChat.current.chatId}
                            <ThreadFromChat inline chat={sharedChat.current} files={includeFilesInMessage} />
                        {/key}
                    {/if}
                </div>
            </div>
        {/if} -->
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