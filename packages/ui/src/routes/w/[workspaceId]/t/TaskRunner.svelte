<script lang="ts">
	import { createRegistryStore, getRegistryContext, setRegistryContext } from "$lib/context/registry.svelte";
	import type { WorkspaceClient, WorkspaceFile } from "$lib/types";
	import { WorkspaceService } from "$lib/workspace.svelte";
	import { onDestroy, onMount, untrack } from "svelte";
	import { convertToTask } from "./utils";
	import type { Input, Task } from "./types";
	import { fade, fly, slide } from "svelte/transition";
	import { LoaderCircle, Play, Square, Wrench } from "@lucide/svelte";
	import { ChatService } from "$lib/chat.svelte";
	import Messages from "$lib/components/Messages.svelte";

    type Props = {
        workspaceId: string;
        urlTaskId: string;
    }
    let { workspaceId, urlTaskId }: Props = $props();
    
    const registryStore = createRegistryStore();
	setRegistryContext(registryStore);
    const registry = getRegistryContext();

    const workspaceService = new WorkspaceService();
    const chat = new ChatService();
    
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

    onMount(() => {
        if (urlTaskId && workspaceId) {
            workspace = workspaceService.getWorkspace(workspaceId);
            registryStore.fetch();
        }
    });
    
    onDestroy(() => {
        chat.close();
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

    $effect(() => {
        if (!chat.isLoading && loading) {
            untrack(() => loading = false);
            completed = true;
        }
    });
    
    async function handleRun() {
        if (disabled) return;
        if (loading) {
            canceling = true;
            chat.close();
            return;
        }
        
        // TODO: change below to hit running task with arguments once available
        chat.sendMessage('Write a very long story about the history of the universe. It should be a 1 minute read at least.');
        setTimeout(() => {
            loading = true
        }, 300);
    }
</script>

<div class="flex w-full h-dvh justify-center items-center flex-col relative">
    <div class="h-16 w-full flex p-4 items-center absolute top-0 left-0">
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
                    <button class="btn btn-primary flex-1">
                        Run Again
                    </button>
                </div>
            </div>
        {:else}
            <div class="md:w-xl w-full flex flex-col grow justify-center items-center z-20">
                {#if !loading}
                    <div class="w-full" out:slide={{ duration: 300 }}>
                        <div class="w-full flex flex-col justify-center items-center" out:fly={{ y: -100, duration: 200 }} >
                            <h2 class="text-xl font-semibold">{name}</h2>
                            {#if description.length > 0}
                                <p class="text-xs text-base-content/50 mt-1">{description}</p>
                            {/if}
                            {#if tools.length > 0}
                                <div class="flex flex-wrap gap-2 mt-2 mb-1">
                                    {#each tools as tool}
                                        <div class="badge badge-sm badge-outline badge-primary">
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
                            {#if runFormData.length > 0}
                                <div class="mt-4 p-4 flex flex-col gap-2 w-full border border-transparent dark:border-base-300 bg-base-100 dark:bg-base-200 shadow-xs rounded-field">
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
                    <button class="btn btn-primary transition-all mt-4 {loading ? 'w-10' : 'w-48'}"  onclick={handleRun} {disabled}>
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

        {#if loading && !canceling}
            <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 h-48 overflow-hidden flex flex-col justify-end md:w-2xl w-full" id="thread-process">
                <div class="absolute inset-x-0 top-0 h-16 bg-linear-to-b from-base-200 dark:from-base-100 to-transparent z-10 pointer-events-none"></div>
                <Messages messages={chat.messages} onSend={chat.sendMessage} isLoading={chat.isLoading} agent={chat.agent} />
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
</style>