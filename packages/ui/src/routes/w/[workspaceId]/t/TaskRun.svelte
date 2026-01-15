<script lang="ts">
	import type { ChatMessage, WorkspaceClient, WorkspaceFile } from "$lib/types";
	import { untrack } from "svelte";
	import { buildRunSummary, buildSessionData, convertToTask } from "./utils";
	import type { Input, Task } from "./types";
	import { SvelteMap } from "svelte/reactivity";
    import type { ChatService } from "$lib/chat.svelte";
	import TaskRunContent from "./TaskRunContent.svelte";
	
    type Props = {
        workspace: WorkspaceClient;
        urlTaskId: string;
        runId: string;
    }
    let { workspace, urlTaskId, runId }: Props = $props();

    let task = $state<Task | null>(null);
    let initialLoadComplete = $state(false);
    let compiling = $state(false);

    let loading = $state(false);
    let error = $state(false);

    let runArguments = $state<(Omit<Input, 'id'> & { value: string })[]>([]);
    let run = $state<ChatService | null>(null);
    let compiledRunId = $state<string | null>(null);

    let progressTimeout: ReturnType<typeof setTimeout> | null = null;
    let progress = $state(0);

    let totalTime = $state(0);
    let totalTokens = $state(0);
    
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
        const currentRunId = runId; // track runId
        const taskReady = !!task; // track when task becomes available
        untrack(() => {
            if (!taskReady || !workspace) return;
            if (currentRunId) {
                loadRunSession();
            } else {
                stepSummaries = [];
            }
        });
    })

    function reset() {
        runArguments = [];
        loading = false;
        ongoingSteps.clear();
        stepSummaries = [];
        totalTime = 0;
        totalTokens = 0;
    }

    async function loadRunSession() {
        if (!task || !workspace) return;

        reset();
        run = null;
        compiledRunId = null;
        loading = true;
        run = await workspace.getSession(runId);
    }
    
    async function compileRun(messages: ChatMessage[]) {
        if (!task) return;
        const sessionData = buildSessionData(messages, task.steps);
        const sessionArgs: Record<string, typeof runArguments[number]> = {};
        for (const step of task.steps) {
            const data = sessionData[step.id];

            // look for arguments in task run messages
            for (const message of data?.messages ?? []) {
                const item = message.items?.[0];
                if (item?.type !== 'tool' || !('arguments' in item) || !item.arguments) continue;
                const parsed = JSON.parse(item.arguments);
                if (parsed.arguments?.[0]) {
                    sessionArgs[step.id] = parsed.arguments[0];
                    break;
                }
            }
            
            ongoingSteps.set(step.id, { 
                loading: false, 
                completed: true,
                oauth: '', 
                totalTime: Math.floor(Math.random() * 6000) + 1000, 
                tokens: Math.floor(Math.random() * 4000) + 1000,
                error: !data || data.messages.length === 0,
            });
        }

        runArguments = Object.values(sessionArgs);

        const ongoingStepsArr = Array.from(ongoingSteps.values());
        totalTokens = ongoingStepsArr.reduce((acc, step) => acc + (step.tokens ?? 0), 0);
        totalTime = ongoingStepsArr.reduce((acc, step) => acc + (step.totalTime ?? 0), 0);
        error = ongoingStepsArr.some((step) => step.error);

        loading = false;
    }

    $effect(() => {
        if (run && run.messages.length > 0 && task && compiledRunId !== runId) {
            compiledRunId = runId;
            compileRun(run.messages);
        }
    })

    $effect(() => {
        const files = workspace?.files ?? [];
        if (urlTaskId && files.length > 0) {
             if (untrack(() => !compiling)) {
                compileTask(urlTaskId, files);
            }
        }
    });
</script>

<TaskRunContent 
    {task}
    {initialLoadComplete} 
    appendTitle={run?.chatId ? `- ${run.chatId}` : ''}
    {loading} 
    {runArguments}
    {ongoingSteps}
    {stepSummaries}
    {runSummary}
    {error}
    {totalTime}
    {totalTokens}
    {progress}
    loadingText="Loading the previous task run now. Please wait a moment..."
    completed={!loading}
/>