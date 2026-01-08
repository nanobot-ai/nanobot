<script lang="ts">
	import { createRegistryStore, getRegistryContext, setRegistryContext } from "$lib/context/registry.svelte";
	import type { WorkspaceClient, WorkspaceFile } from "$lib/types";
	import { WorkspaceService } from "$lib/workspace.svelte";
	import { onDestroy, onMount } from "svelte";
	import { convertToTask } from "./utils";
	import type { Input, Task } from "./types";
	import { fade, fly, slide } from "svelte/transition";
	import { Circle, CircleCheck, ListTodo, LoaderCircle, Play, Square, Wrench } from "@lucide/svelte";
	import { setSharedChat, sharedChat } from "$lib/stores/chat.svelte";
	import { SvelteMap } from "svelte/reactivity";
	
    type Props = {
        workspaceId: string;
        urlTaskId: string;
    }
    let { workspaceId, urlTaskId }: Props = $props();
    
    const registryStore = createRegistryStore();
	setRegistryContext(registryStore);
    const registry = getRegistryContext();

    const workspaceService = new WorkspaceService();
    
    let workspace = $state<WorkspaceClient | null>(null);
    let task = $state<Task | null>(null);
    let initialLoadComplete = $state(false);

    let runFormData = $state<(Input & { value: string })[]>([]);
    let loading = $state(false);
    let disabled = $derived(runFormData.some((input) => input.value.trim().length === 0));
    let canceling =  $state(false);
    let completed = $state(false); // TODO: completed data to display?

    let progressTimeout: ReturnType<typeof setTimeout> | null = null;
    let progress = $state(0);

    let name = $derived(task?.name || task?.steps[0].name || '');
    let description = $derived((task?.description || task?.steps[0].description)?.trim() || '');
    let tools = $derived.by(() => {
        if ((task?.steps?.length || 0) === 0) return [];
        if (registry.loading || registry.servers.length === 0) return [];
        
        const tools = task?.steps.flatMap((step) => step.tools.map((toolName) => registry.getServerByName(toolName)).filter((tool) => tool !== undefined)) ?? [];
        return tools.filter((tool, index, self) => self.findIndex((t) => t.name === tool.name) === index);
    })

    let timeoutHandlers = $state<ReturnType<typeof setTimeout>[]>([]);

    const ongoingSteps = new SvelteMap<string, { loading: boolean, completed: boolean, oauth: string, totalTime?: number, tokens?: number }>();

    onMount(() => {
        if (urlTaskId && workspaceId) {
            workspace = workspaceService.getWorkspace(workspaceId);
            registryStore.fetch();
        }
    });

    async function initChat() {
        const chat = await workspace?.newSession({ editor: true });
        if (chat) {
            setSharedChat(chat);
        }
    }

    $effect(() => {
        if (workspace && !sharedChat.current) {
            initChat();
        }
    })
    
    onDestroy(() => {
        sharedChat.current?.close();
    });

    async function compileTask(id: string, files: WorkspaceFile[]) {
        if (!workspace || !id) return;
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
        runFormData = task.inputs.map((input) => ({
            ...input,
            value: input.default || '',
        }));
        clearTimeout(progressTimeout);
        progress = 100;

        await new Promise((resolve) => setTimeout(resolve, 1000));
        initialLoadComplete = true;
    }

    $effect(() => {
        const files = workspace?.files ?? [];
        if (urlTaskId && files.length > 0) {
            compileTask(urlTaskId, files);
        }
    });
    
    async function handleRun() {
        if (disabled || !task) return;
        if (loading) {
            canceling = true;
            timeoutHandlers.forEach((handler) => clearTimeout(handler));
            completed = true;
            loading = false;
            return;
        }
        
        canceling = false;
        loading = true;
        completed = false;
        timeoutHandlers = [];
        ongoingSteps.clear();

        let timeout = 0;
        for (const step of task.steps) {
            timeout += 1000;
            const handlerA = setTimeout(() => {
                ongoingSteps.set(step.id, { loading: true, completed: false, oauth: '' });
            }, timeout);
            
            const completeTime = Math.floor(Math.random() * 4000) + 1000;
            const tokens = Math.floor(Math.random() * 9000) + 1000;
            timeout += completeTime; // 1-5 seconds
            const handlerB = setTimeout(() => {
                ongoingSteps.set(step.id, { loading: false, completed: true, totalTime: completeTime, tokens, oauth: '' });
            }, timeout);
            timeoutHandlers.push(handlerA, handlerB);
        }
        const finalHandler = setTimeout(() => {
            loading = false;
            completed = true;
        }, timeout + 1000);
        timeoutHandlers.push(finalHandler);
    }

    function reset() {
        loading = false;
        completed = false;
        canceling = false;
        timeoutHandlers = [];
        ongoingSteps.clear();
    }
</script>

<div class="flex w-full h-dvh justify-center items-center flex-col relative">
    <div class="h-16 w-full flex px-4 items-center absolute top-0 left-0">
        {#if loading}
        <h2 in:fade class="text-xl font-semibold flex items-center gap-2">{name} <LoaderCircle class="size-4 animate-spin shrink-0" /></h2>
        {/if}
    </div>
    {#if initialLoadComplete && task}
        {#if completed}
            <div in:fade class="w-md flex flex-col justify-center items-center">
                <h4 class="text-xl font-semibold">{canceling ? 'Task Run Cancelled' : 'Task Run Completed'}</h4>
                <p class="text-sm text-base-content/50 text-center mt-1">
                    {#if canceling}
                        The task run has been cancelled. Would you like to run it again?
                    {:else}
                        The task has been completed successfully. 
                        Would you like to see specific details of the run or run it again?
                    {/if}
                </p>
                <div class="flex grow gap-2 w-full mt-4">
                    {#if !canceling}
                        <button class="btn flex-1">
                            View Details
                        </button>
                    {/if}
                    <button class="btn btn-primary flex-1" onclick={reset}>
                        Run Again
                    </button>
                </div>
            </div>
        {:else}
            <div class="md:w-4xl px-4 w-full flex flex-col justify-center items-center z-20">
                <div class="hero w-full bg-base-100 dark:bg-base-200 rounded-box shadow-xs dark:border-base-300 border-transparent border">
                    <div class="hero-content w-full grow flex-col md:flex-row">
                        <div class="pl-4 flex items-center gap-2">
                            <div class="rounded-full p-2 border-2 border-primary bg-primary/10 {loading ? 'animate-pulse' : ''} w-fit">
                                <ListTodo class="size-8 text-primary" />
                            </div>
                            <div class="w-xs">
                                <h4 class="mt-2 text-2xl font-semibold">{name}</h4>
                                {#if loading}
                                    <p in:fade class="font-light text-sm text-base-content/50">Your task is currently running. Please wait a moment...</p>
                                {:else}
                                    <div in:fade>
                                        {#if description.length > 0}
                                            <p class="text-xs text-base-content/50 mt-1">{description}</p>
                                        {/if}
                                        {#if tools.length > 0}
                                            <div class="flex flex-wrap gap-2 mt-2 mb-1">
                                                {#each tools as tool (tool.name)}
                                                    <div class="badge badge-sm badge-soft gap-1">
                                                        {#if tool.icons?.[0]?.src}
                                                            <img alt={tool.title} src={tool.icons[0].src} class="size-4" />
                                                        {:else}
                                                            <Wrench class="size-4" />
                                                        {/if}
                                                        {tool.title}
                                                    </div>
                                                {/each}
                                            </div>
                                        {/if}
                                    </div>
                                {/if}
                            </div>
                        </div>
                        <ul in:fade class="timeline timeline-vertical timeline-compact grow">
                            {#each task.steps as step, index (step.id)}
                                <li>
                                    {#if index > 0}
                                        <hr class="timeline-connector w-0.5 {ongoingSteps.get(task.steps[index - 1].id)?.completed ? 'completed' : ''}" />
                                    {/if}
                                    <div class="timeline-middle">
                                        {#if ongoingSteps.get(step.id)?.completed}
                                            <CircleCheck class="size-5 text-primary" />
                                        {:else if ongoingSteps.get(step.id)?.loading}
                                            <LoaderCircle class="size-5 animate-spin shrink-0 text-base-content/50" />
                                        {:else}
                                            <Circle class="size-5 text-base-content/50" />
                                        {/if}
                                    </div>
                                    <div class="timeline-end timeline-box border-0 shadow-none pl-1 py-2">
                                        {step.name} 
                                        {#if ongoingSteps.get(step.id)?.completed}
                                            <span in:fade class="text-xs text-base-content/35">({ongoingSteps.get(step.id)?.totalTime ? `${(ongoingSteps.get(step.id)!.totalTime! / 1000).toFixed(1)}s` : ''})</span>
                                        {/if}
                                        {#if ongoingSteps.get(step.id)?.tokens}
                                            <span in:fade class="text-xs italic text-base-content/35">{ongoingSteps.get(step.id)?.tokens ? `${ongoingSteps.get(step.id)!.tokens!} tokens` : ''}</span>
                                        {/if}
                                    </div>
                                    {#if index < task.steps.length - 1}
                                        <hr class="timeline-connector w-0.5 {ongoingSteps.get(step.id)?.completed ? 'completed' : ''}" />
                                    {/if}
                                </li>
                            {/each}
                        </ul>
                    </div>
                </div>
            </div>
            <div class="md:w-4xl p-4 w-full flex flex-col justify-center items-center z-20">
                {#if !loading}
                    <div class="w-full" out:slide={{ duration: 300 }}>
                        <div class="w-full flex flex-col justify-center items-center" out:fly={{ y: -100, duration: 200 }} >
                            {#if runFormData.length > 0}
                                <div class="p-4 flex flex-col gap-2 w-full border border-transparent dark:border-base-300 bg-base-100 dark:bg-base-200 shadow-xs rounded-field">
                                    <p class="text-xs text-primary">To get started, please fill out the following information:</p>
                                    <div class="flex flex-col gap-2">
                                        {#each runFormData as input (input.id)}
                                            <label class="input w-full validator">
                                                <span class="label h-full font-semibold text-primary bg-primary/15">{input.name}</span>
                                                <input type="text" bind:value={input.value} placeholder={input.description} required />
                                            </label>
                                        {/each}
                                    </div>
                                </div>
                            {/if}
                        </div>
                    </div>
                {/if}
                {#if canceling}
                    <button class="btn w-10 mt-4" disabled>
                        <LoaderCircle class="size-4 animate-spin shrink-0" />
                    </button>
                {:else}
                    <button class="btn btn-primary transition-all mt-4 {loading ? 'w-10 tooltip' : 'w-48'}"  onclick={handleRun} {disabled} data-tip={loading ? 'Cancel run' : undefined}>
                        {#if loading}
                            <Square class="size-4 shrink-0" />
                        {:else}
                            Run 
                            <Play class="size-4 shrink-0" />
                        {/if}
                    </button>
                {/if}

                {#if canceling}
                    <p class="text-sm text-base-content/25 mt-1">Cancelling current run...</p>
                {/if}
            </div>
        {/if}
    {:else}
        <div in:fade|global={{ duration: 300 }} class="radial-progress text-primary" style="--value:{progress};" aria-valuenow="{progress}" role="progressbar">{progress}%</div>
    {/if}
</div>

<style lang="postcss">
    :global(#thread-process #message-groups) {
        padding-top: 0;
        opacity: 0.15;
    }
    :global(#thread-process #message-groups .prose) {
        font-size: 0.75rem;
    }
    :global(#thread-process #message-groups > div) {
        min-height: unset !important;
    }
    :global(#thread-process #message-groups .h-59) {
        display: none;
    }

    /* Timeline connector fill animation */
    :global(.timeline-connector) {
        position: relative !important;
        background-color: color-mix(in oklch, var(--color-base-content) 50%, transparent);
        overflow: hidden !important;
    }

    :global(.timeline-connector::after) {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 0%;
        background-color: var(--color-primary);
        transition: height 0.4s ease-out;
    }

    :global(.timeline-connector.completed::after) {
        height: 100%;
    }
</style>