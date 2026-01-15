<script lang="ts">
	import { fade, slide } from "svelte/transition";
	import type { Input, OngoingStep, Task } from "./types";
	import { Circle, CircleAlert, CircleCheck, ListTodo, LoaderCircle, Wrench } from "@lucide/svelte";
	import { getRegistryContext } from "$lib/context/registry.svelte";
	import type { SvelteMap } from "svelte/reactivity";
	import type { Snippet } from "svelte";

    interface Props {
        task?: Task | null;
        initialLoadComplete: boolean;
        appendTitle?: string;
        loading: boolean;
        runArguments: (Omit<Input, 'id'> & { value: string })[];
        ongoingSteps: SvelteMap<string, OngoingStep>;
        stepSummaries: { step: string, summary: string }[];
        runSummary: string;
        error?: boolean;
        cancelled?: boolean;
        totalTime: number;
        totalTokens: number;
        progress: number;
        children?: Snippet;
        loadingText?: string;
        completed: boolean;
    }
    
    let { 
        task, 
        initialLoadComplete, 
        appendTitle,
        loading, 
        runArguments, 
        ongoingSteps, 
        stepSummaries,
        runSummary,
        error,
        cancelled,
        totalTime,
        totalTokens,
        progress,
        children,
        loadingText,
        completed,
    }: Props = $props();
    const registry = getRegistryContext();
    
    let taskName = $derived(task?.name || task?.steps[0].name || '');
    let taskDescription = $derived(task?.description || task?.steps[0].description || '');
    let tools = $derived.by(() => {
        if ((task?.steps?.length || 0) === 0) return [];
        if (registry.loading || registry.servers.length === 0) return [];
        
        const tools = task?.steps.flatMap((step) => step.tools.map((tool) => registry.getServerByName(tool.name)).filter((tool) => tool !== undefined)) ?? [];
        return tools.filter((tool, index, self) => self.findIndex((t) => t.name === tool.name) === index);
    });

</script>

<div class="w-full h-dvh flex flex-col relative">
    <div class="h-16 w-full flex px-4 items-center shrink-0 z-30 sticky top-0 left-0">
        <h2 in:fade class="text-xl font-semibold flex items-center gap-2">{taskName} {appendTitle} {#if loading}<LoaderCircle class="size-4 animate-spin shrink-0" />{/if}</h2>
    </div>
    <div class="w-full flex flex-col flex-1 items-center overflow-y-auto py-4">
    {#if initialLoadComplete && task}
        <div class="md:w-4xl px-4 w-full flex flex-col items-center z-20 my-auto">
            <div class="hero w-full bg-base-100 dark:bg-base-200 rounded-box shadow-xs dark:border-base-300 border-transparent border mb-4">
                <div class="hero-content w-full grow flex-col md:flex-row">
                    <div class="flex flex-col gap-4 grow max-w-sm">
                        <div class="pl-4 flex items-center gap-3">
                            <div class="rounded-full p-2 border-2 border-primary bg-primary/10 {loading ? 'animate-pulse' : ''} w-fit">
                                <ListTodo class="size-8 text-primary" />
                            </div>
                            <div class="w-xs">
                                <h3 class="text-2xl font-semibold">{taskName}</h3>
                                {#if loading}
                                    <p in:fade class="font-light text-sm text-base-content/50">
                                        {loadingText}
                                    </p>
                                {:else}
                                    <div in:fade>
                                        {#if taskDescription.length > 0}
                                            <p class="text-xs text-base-content/50 mt-1">{taskDescription}</p>
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
                        {#if runArguments.length > 0}
                            <div class="flex flex-col gap-4">
                                {#each runArguments as input (input.name)}
                                    <label class="input input-sm w-full border border-base-300">
                                        <span class="label h-full font-semibold text-primary bg-primary/15">{input.name}</span>
                                        <input disabled type="text" bind:value={input.value} />
                                    </label>
                                {/each}
                            </div>
                        {/if}
                    </div>
                    <ul in:fade class="timeline timeline-snap-icon timeline-vertical timeline-compact grow">
                        {#each task.steps as step, index (step.id)}
                            {@const stepStatus = ongoingSteps.get(step.id)}
                            <li>
                                {#if index > 0}
                                    <hr class="timeline-connector w-0.5 {ongoingSteps.get(task.steps[index - 1].id)?.error ? 'error' : ongoingSteps.get(task.steps[index - 1].id)?.completed ? 'completed' : ''}" />
                                {/if}
                                <div class="timeline-middle">
                                    {#if stepStatus?.error}
                                        <CircleAlert class="size-5 text-error/40" />
                                    {:else if stepStatus?.completed}
                                        <CircleCheck class="size-5 text-primary" />
                                    {:else if ongoingSteps.get(step.id)?.loading}
                                        <LoaderCircle class="size-5 animate-spin shrink-0 text-base-content/50" />
                                    {:else}
                                        <Circle class="size-5 text-base-content/50" />
                                    {/if}
                                </div>
                                <div class="timeline-end timeline-box border-0 shadow-none pl-1 py-2">
                                    <div>
                                        {step.name} 
                                        {#if stepStatus?.completed && !stepStatus?.error}
                                            <span in:fade class="text-xs text-base-content/35">({stepStatus?.totalTime ? `${(stepStatus.totalTime / 1000).toFixed(1)}s` : ''})</span>
                                            {#if stepStatus?.tokens}
                                                <span in:fade class="text-xs italic text-base-content/35">{stepStatus?.tokens ? `${stepStatus.tokens} tokens` : ''}</span>
                                            {/if}
                                        {/if}
                                    </div>
                                    {#if stepSummaries[index]?.summary}
                                        <p in:slide={{ axis: 'y', duration: 150 }} class="text-sm text-base-content/50 mt-2">{stepSummaries[index].summary}</p>
                                    {/if}
                                </div>
                                {#if index < task.steps.length - 1}
                                    <hr class="timeline-connector w-0.5 {stepStatus?.error ? 'error' : stepStatus?.completed ? 'completed' : ''}" />
                                {/if}
                            </li>
                        {/each}
                    </ul>
                </div>
            </div>

            {#if completed}
                <div in:fade out:slide={{ axis: 'y', duration: 150 }} class="w-full flex flex-col justify-center items-center py-4">
                    <div class="w-full flex flex-col justify-center items-center border border-transparent dark:border-base-300 bg-base-100 dark:bg-base-200 shadow-xs rounded-field p-6 pb-8">
                        <h4 class="text-xl font-semibold">{error ? 'Workflow Failed' : cancelled ? 'Workflow Cancelled' : 'Workflow Completed'}</h4>
                        <p class="text-sm text-base-content/50 text-center mt-1">
                            {#if error}
                                The workflow did not complete successfully. Please try again or contact the workflow owner.
                            {:else if cancelled}
                                The workflow has been cancelled. Would you like to run it again?
                            {:else}
                                The workflow has completed successfully. Here are your summarized results:
                            {/if}
                        </p>

                        <p class="text-sm text-center mt-4">The workflow completed <b>{task.steps.length}</b> out of <b>{task.steps.length}</b> steps.</p> 
                        <p class="text-sm text-center mt-1">It took a total time of <b>{(totalTime / 1000).toFixed(1)}s</b> to complete.</p>
                        <p class="text-sm text-center mt-1">A total of <b>{totalTokens}</b> tokens were used.</p>
                        
                        {#if runSummary}
                            <div transition:fade class="prose-sm mt-4 text-left w-full max-w-none">
                                {@html runSummary}
                            </div>
                        {/if}
                    </div>
                </div>
            {/if}
            {@render children?.()}
        </div>
    {:else}
        <div in:fade|global={{ duration: 300 }} class="radial-progress text-primary my-auto" style="--value:{progress};" aria-valuenow="{progress}" role="progressbar">{progress}%</div>
    {/if}
    </div>
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

    :global(.timeline-connector.error::after) {
        background-color: color-mix(in oklch, var(--color-error) 40%, var(--color-base-100));
    }

    :global(.timeline-connector.completed::after, .timeline-connector.error::after) {
        height: 100%;
    }
</style>