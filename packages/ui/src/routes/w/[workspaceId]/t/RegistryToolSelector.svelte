<script lang="ts">
	import { getNotificationContext } from "$lib/context/notifications.svelte";
	import { getRegistryContext } from "$lib/context/registry.svelte";
	import { SvelteMap } from "svelte/reactivity";
	import { Check, Plus, Search, ServerIcon } from "@lucide/svelte";
	import type { Server } from "$lib/types";

    interface Props {
        onToolsSelect: (tools: string[]) => void;
        omit: string[];
    }

    let { onToolsSelect, omit }: Props = $props();

    let dialog = $state<HTMLDialogElement | null>(null);
    let query = $state('');
    
    let selected = new SvelteMap<string, Server>();

    const notificationContext = getNotificationContext();
    const registry = getRegistryContext();

    let results = $derived.by(() => {
        const withoutOmitted = registry.servers.filter(server => !omit.includes(server.name));
        if (query) {
            return withoutOmitted.filter(server => server.title.toLowerCase().includes(query.toLowerCase()));
        } else {
            return withoutOmitted;
        }
    });

    export async function showModal() {
        selected.clear();
        query = '';

        dialog?.showModal();
        
        try {
            await registry.fetch();
        } catch (error) {
            notificationContext.error('Failed to get registry', error instanceof Error ? error.message : String(error));
        }
    }
</script>

<dialog bind:this={dialog} class="modal">
    <div class="modal-box bg-base-100 dark:bg-base-200 w-full md:w-7xl h-full max-w-full max-h-full md:max-w-[calc(100vw-2rem)] md:max-h-[calc(100vh-2rem)] overflow-y-auto">
        <form method="dialog">
            <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
          </form>
        <div class="flex flex-col h-full">
            <h3 class="text-lg font-bold">Select tools to add</h3>
            <div class="flex flex-col grow overflow-y-auto max-h-[calc(100vh-8.75rem)] md:max-h-[calc(100vh-10.75rem)]">
                {#if !registry.loading}
                    <div class="px-1 py-4">
                        <label class="input w-full">
                            <Search class="opacity-50 h-[1em]" />
                            <input name="search-tools" type="search" required placeholder="Search tools..." bind:value={query} />
                        </label>
                    </div>
                {/if}
                {#if registry.loading}
                    {@render loadingSkeleton()}
                {:else if results.length > 0}
                    <ul class="list grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {#each results as server}
                            <button class="list-row text-left cursor-pointer hover:bg-base-200/50 active:bg-base-200 dark:hover:bg-base-300/50 dark:active:bg-base-300 {selected.has(server.name) ? 'bg-primary/10' : ''}"
                                onclick={() => {
                                    if (selected.has(server.name)) {
                                        selected.delete(server.name);
                                    } else {
                                        selected.set(server.name, server);
                                    }
                                }}
                            >
                                <div>
                                    {#if server.icons?.[0]?.src}
                                        <img src={server.icons[0].src} alt={server.name} class="size-6 rounded-field" />
                                    {:else}
                                        <ServerIcon class="size-6 text-base-content/50" />
                                    {/if}
                                </div>
                                <div class="list-col-grow">
                                    <h4 class="text-sm font-semibold flex items-center gap-2">
                                        {server.title}
                                        {#if server.version !== 'latest'}
                                            <span class="badge badge-xs badge-outline text-base-content/50">{server.version}</span>
                                        {/if}
                                    </h4>
                                    <p class="text-xs text-base-content/50 line-clamp-2">{server.description}</p>
                                </div>
                                <div>
                                    {#if selected.has(server.name)}
                                        <Check class="size-4 text-primary" />
                                    {:else}
                                        <Plus class="size-4" />
                                    {/if}
                                </div>
                                
                            </button>
                        {/each}
                    </ul>
                {:else if query}
                    <p class="py-4 text-base-content/50 text-sm">No tools found matching "{query}".</p>
                {:else}
                    <p class="py-4 text-base-content/50 text-sm">There are no tools currently available.</p>
                {/if}
            </div>
            <div class="modal-action items-center justify-between">
                <div class="text-base-content/50 text-sm flex items-center gap-2">
                    {#if selected.size > 0}
                        {selected.size} selected

                        <button class="btn btn-ghost" onclick={() => {
                            selected.clear();
                        }}>
                            Clear All
                        </button>
                    {/if}
                </div>   
                <div>
                    <button class="btn btn-ghost" onclick={() => {
                        dialog?.close();
                    }}>
                        Cancel
                    </button>
                    <button class="btn btn-primary" onclick={() => {
                        onToolsSelect(Array.from(selected.values()).map(server => server.name));
                        dialog?.close();
                    }}>
                        Add
                    </button>
                </div>
            </div>
        </div>
    </div>
    <form method="dialog" class="modal-backdrop">
        <button>close</button>
    </form>
</dialog>

{#snippet loadingSkeleton()}
<div class="skeleton h-10 w-full my-4"></div>
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 py-4">
{#each Array(9).fill(null) as _, index (index)}
    <div class="flex items-center gap-4">
        <div class="skeleton size-6 shrink-0 rounded-field"></div>
        <div class="flex flex-col gap-4 grow">
            <div class="skeleton h-4"></div>
            <div class="skeleton h-4"></div>
        </div>
    </div>
{/each}
</div>
{/snippet}