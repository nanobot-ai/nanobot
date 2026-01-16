<script lang="ts">
	import type { WorkspaceClient, WorkspaceFile } from "$lib/types";
	import { untrack } from "svelte";
	import { areAllStepsCompleted, buildRunSummary, buildSessionData, convertToTask } from "./utils";
	import type { Input, Task } from "./types";
	import { SvelteMap } from "svelte/reactivity";
    import TaskRunInputs from "./TaskRunInputs.svelte";
	import type { ChatService } from "$lib/chat.svelte";
	import TaskRunContent from "./TaskRunContent.svelte";
	import { LoaderCircle, Play, Square } from "@lucide/svelte";
	import Elicitation from "$lib/components/Elicitation.svelte";
	
    type Props = {
        workspace: WorkspaceClient;
        urlTaskId: string;
    }
    let { workspace, urlTaskId }: Props = $props();
    let task = $state<Task | null>(null);
    let initialLoadComplete = $state(false);
    let compiling = $state(false);
    let prevTaskId = $state<string | null>(null);

    let inputsModal = $state<ReturnType<typeof TaskRunInputs> | null>(null);
    let loading = $state(false);
    let cancelled =  $state(false);
    let completed = $state(false);
    let error = $state(false);

    let runArguments = $state<(Omit<Input, 'id'> & { value: string })[]>([]);
    let run = $state<ChatService | null>(null);

    let progressTimeout: ReturnType<typeof setTimeout> | null = null;
    let progress = $state(0);

    let totalTime = $state(0);
    let totalTokens = $state(0);
    let timeoutHandlers = $state<ReturnType<typeof setTimeout>[]>([]);

    const ongoingSteps = new SvelteMap<string, { loading: boolean, completed: boolean, oauth: string, totalTime?: number, tokens?: number, error?: boolean }>();
    let stepSummaries = $state<{ step: string, summary: string }[]>([]);
    let runSummary = $derived(run && !error ? buildRunSummary(run.messages) : '');

    async function compileTask(id: string, files: WorkspaceFile[]) {
        if (!workspace || !id || compiling) return;
        
        compiling = true;
        if (progressTimeout) {
            clearTimeout(progressTimeout);
        }

        initialLoadComplete = false;

        progress = 0;
        progressTimeout = setTimeout(() => {
            progress = 30;
            progressTimeout = setTimeout(() => {
                progress = 75;
            }, 3000);
        }, 1000);

        task = await convertToTask(workspace, files, id);
        clearTimeout(progressTimeout);
        progress = 100;

        await new Promise((resolve) => setTimeout(resolve, 500));
        initialLoadComplete = true;
        compiling = false;
    }

    $effect(() => {
        const files = workspace?.files ?? [];
        if (urlTaskId && files.length > 0 && prevTaskId !== urlTaskId) {
             if (untrack(() => !compiling)) {
                compileTask(urlTaskId, files);
                resetRun();
                prevTaskId = urlTaskId;
            }
        }
    });

    $effect(() => {
        // Update runSession during an ongoing task run
        if (!run || !task || run.messages.length === 0 || completed) return;

        const sessionData = buildSessionData(run.messages, task.steps);
        for (const [stepId, data] of Object.entries(sessionData)) {
            const existingStep = untrack(() => ongoingSteps.get(stepId));
            ongoingSteps.set(stepId, {
                loading: data.pending,
                completed: data.completed,
                oauth: '',
                totalTime: existingStep?.totalTime ?? Math.floor(Math.random() * 6000) + 1000,
                tokens: existingStep?.tokens ?? Math.floor(Math.random() * 4000) + 1000,
                error: false,
            });
        }

        if (areAllStepsCompleted(task.steps, sessionData)) {
            completed = true;
            loading = false;
        }
    })

    function resetRun(isCancel?: boolean) {
        runArguments = [];
        timeoutHandlers.forEach((handler) => clearTimeout(handler));
        timeoutHandlers = [];

        cancelled = isCancel ?? false;
        loading = false;
        completed = false;
        ongoingSteps.clear();
        stepSummaries = [];
        totalTime = 0;
        totalTokens = 0;
    }
    
    // TODO:
    async function handleRun(formData: (Input & { value: string })[]) {
        if (!task) return;
        if (loading) {
            resetRun(true);
            return;
        }
        
        resetRun();
        loading = true;
        runArguments = formData;

        run = await workspace?.newSession();
        run.setCallbacks({
            onChatStart: () => {
                workspace?.load();
            },
            onChatDone: () => {
                completed = true;
                loading = false;
                const sessionData = buildSessionData(run?.messages ?? [], task?.steps ?? []);
                error = !areAllStepsCompleted(task?.steps ?? [], sessionData);

                for (const step of task?.steps ?? []) {
                    const existingStep = ongoingSteps.get(step.id);
                    ongoingSteps.set(step.id, {
                        loading: false,
                        completed: sessionData[step.id]?.completed ?? false,
                        oauth: '',
                        totalTime: existingStep?.totalTime ?? Math.floor(Math.random() * 6000) + 1000,
                        tokens: existingStep?.tokens ?? Math.floor(Math.random() * 4000) + 1000,
                        error: sessionData[step.id]?.messages ? sessionData[step.id]?.messages.length === 0 : true,
                    });
                }
            }
        });
        await run?.sendToolCall('ExecuteTaskStep', {taskName: urlTaskId, arguments: formData});
    }
</script>

<TaskRunContent
    {task}
    {initialLoadComplete}
    loading={loading}
    {completed}
    {runArguments}
    {ongoingSteps}
    {stepSummaries}
    {runSummary}
    {error}
    {cancelled}
    {totalTime}
    {totalTokens}
    {progress}
    loadingText="Your task is currently running. Please wait a moment..."
>
    {#if cancelled}
        <button class="btn w-10 mt-4" disabled>
            <LoaderCircle class="size-4 animate-spin shrink-0" />
        </button>
    {:else}
        <button class="btn btn-primary transition-all mt-4 {loading ? 'w-10 tooltip' : 'w-48'}"  onclick={() => {
            if (completed || loading) {
                resetRun(true);
            } else {
                inputsModal?.showModal();
            }
        }} data-tip={loading ? 'Cancel run' : undefined}>
            {#if loading}
                <Square class="size-4 shrink-0" />
            {:else}
                Run 
                <Play class="size-4 shrink-0" />
            {/if}
        </button>
    {/if}
</TaskRunContent>

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

<TaskRunInputs bind:this={inputsModal} onSubmit={handleRun} {task} />