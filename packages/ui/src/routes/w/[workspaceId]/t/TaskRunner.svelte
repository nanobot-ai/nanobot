<script lang="ts">
	import { createRegistryStore, setRegistryContext } from "$lib/context/registry.svelte";
	import type { WorkspaceClient, WorkspaceFile } from "$lib/types";
	import { WorkspaceService } from "$lib/workspace.svelte";
	import { onMount } from "svelte";
	import { convertToTask } from "./utils";
	import type { Input, Task } from "./types";
	import { fade } from "svelte/transition";
	import { ListCheck, Play } from "@lucide/svelte";

    type Props = {
        workspaceId: string;
        urlTaskId: string;
    }
    let { workspaceId, urlTaskId }: Props = $props();
    
    const registryStore = createRegistryStore();
	setRegistryContext(registryStore);

    const workspaceService = new WorkspaceService();
    
    let workspace = $state<WorkspaceClient | null>(null);
    let task = $state<Task | null>(null);
    let initialLoadComplete = $state(false);

    let runFormData = $state<(Input & { value: string })[]>([]);
    let loading = $state(false);
    let disabled = $derived(loading || runFormData.some((input) => input.value.trim().length === 0));

    let progressTimeout: ReturnType<typeof setTimeout> | null = null;
    let progress = $state(0);

    onMount(() => {
        if (urlTaskId && workspaceId) {
            workspace = workspaceService.getWorkspace(workspaceId);
            registryStore.fetch();
        }
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
    })

    function handleRun() {
        if (disabled) return;

        // TODO: run task
    }
</script>

<div class="flex w-full h-dvh justify-center items-center flex-col">
    {#if initialLoadComplete && task}
        {@const description = (task.description || task.steps[0].description)?.trim()}    
        <div class="md:w-xl w-full flex flex-col items-center justify-center">
            <ListCheck class="size-8 text-base-content/20 mb-2" />
            <h2 class="text-xl font-semibold">{task.name || task.steps[0].name}</h2>
            {#if description.length > 0}
                <p class="text-sm text-base-content/50 mt-1">{description}</p>
            {/if}
            {#if runFormData.length > 0}
                <div class="p-4 flex flex-col gap-2 w-full">
                    <div>
                        {#each runFormData as input (input.id)}
                            <label class="input w-full validator">
                                <span class="label h-full font-semibold text-primary bg-primary/15">{input.name}</span>
                                <input type="text" bind:value={input.value} placeholder={input.description} required />
                            </label>
                        {/each}
                    </div>
                </div>
            {/if}
            <button class="btn btn-primary w-48" onclick={handleRun} {disabled}>
                Run Task
                <Play class="size-4" />
            </button>
        </div>
    {:else}
        <div in:fade|global={{ duration: 300 }} class="radial-progress text-primary" style="--value:{progress};" aria-valuenow="{progress}" role="progressbar">{progress}%</div>
    {/if}
</div>