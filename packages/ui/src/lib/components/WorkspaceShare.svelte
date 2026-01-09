<script lang="ts">
    import type { Workspace } from "$lib/types";
	import { Search, X } from "@lucide/svelte";
    import * as mocks from '$lib/mocks';

    interface Props {
        sharingWorkspace?: Workspace | null;
        onCancel?: () => void;
        onShare?: () => void;
    }
    let { sharingWorkspace, onCancel, onShare }: Props = $props();
    let shareWorkspaceModal = $state<HTMLDialogElement | null>(null);
    let query = $state('');
    let selected = $state<typeof mocks.users[number][]>([]);

    export function showModal() {
        query = '';
        selected = [];
        shareWorkspaceModal?.showModal();
    }

    let results = $derived(mocks.users.filter(user => 
        (user.name.toLowerCase().includes(query.toLowerCase()) 
        || user.email.toLowerCase().includes(query.toLowerCase()))
        && !selected.some(s => s.id === user.id)
    ));
</script>


<dialog bind:this={shareWorkspaceModal} class="modal">
    <div class="modal-box overflow-y-visible">
        <h3 class="text-xl font-bold">Share {sharingWorkspace?.name}</h3>

        <div class="flex flex-col gap-2 mt-4 w-full">
            <div class="dropdown">
                <div class="input w-full">
                    <Search class="size-4" />
                    <input type="search" bind:value={query} class="grow" placeholder="Add by user..." />
                </div>
                {#if query.length > 0}
                    <ul tabindex="-1" class="w-full dropdown-content menu bg-base-100 rounded-box z-1 p-2 shadow-sm">
                        {#if results.length > 0}
                            {#each results as result (result.id)}
                                <li class="list-row flex w-full gap-0 px-0 items-center">
                                    <button class="w-full flex items-center gap-2"
                                        onclick={() => {
                                            selected.push(result);
                                            (document.activeElement as HTMLElement)?.blur();
                                        }}
                                    >
                                        <div class="avatar avatar-placeholder">
                                            <div class="bg-neutral text-neutral-content w-8 rounded-full">
                                                <span class="text-xs">{result.name.charAt(0)}</span>
                                            </div>
                                        </div>
                                        <div>
                                            <div class="font-medium">{result.name}</div>
                                            <div class="text-sm font-light text-base-content/50">{result.email}</div>
                                        </div>
                                    </button>
                                </li>
                            {/each}
                        {:else}
                            <li class="list-row flex w-full gap-0 px-0 items-center">
                                <div class="text-base-content/35 text-sm">No matching users for "{query}".</div>
                            </li>
                        {/if}
                    </ul>
                {/if}
              </div>

            <div class="flex w-full">
                <h4 class="text-base font-semibold mt-4 flex grow">People with access</h4>
                <div class="w-60 flex pb-0.5">
                    <div class="flex flex-1 text-xs font-semibold justify-center self-end">Read</div>
                    <div class="flex flex-1 text-xs font-semibold justify-center self-end">Write</div>
                    <div class="flex flex-1 text-xs font-semibold justify-center self-end">Execute Only</div>
                </div>
                <div class="w-8"></div>
            </div>
            <ul class="list">
                <li class="list-row flex w-full gap-0 px-0 items-center">
                    <div class="grow flex items-center gap-2">
                        <div class="avatar avatar-placeholder">
                            <div class="bg-neutral text-neutral-content w-8 rounded-full">
                                <span class="text-xs">M</span>
                            </div>
                        </div>
                        <div class="font-medium">Me <span class="text-base-content/50 text-xs">(Owner)</span></div>
                    </div>
                    <div class="w-60 flex items-center">
                        <div class="flex flex-1 justify-center">
                            <input type="checkbox" checked disabled class="checkbox rounded-field checkbox-sm" />
                        </div>
                        <div class="flex flex-1 justify-center">
                            <input type="checkbox" checked disabled class="checkbox rounded-field checkbox-sm" />
                        </div>
                        <div class="flex flex-1 justify-center">
                            <input type="checkbox" checked disabled class="checkbox rounded-field checkbox-sm" />
                        </div>
                    </div>
                    <div class="w-8"></div>
                </li>
                {#if selected.length > 0}
                    {#each selected as user (user.id)}
                        <li class="list-row flex w-full gap-0 px-0 items-center">
                            <div class="grow flex items-center gap-2">
                                <div class="avatar avatar-placeholder">
                                    <div class="bg-neutral text-neutral-content w-8 rounded-full">
                                        <span class="text-xs">{user.name.charAt(0)}</span>
                                    </div>
                                </div>
                                <div class="font-medium">{user.name}</div>
                            </div>
                            <div class="w-60 flex items-center">
                                <div class="flex flex-1 justify-center">
                                    <input type="checkbox" class="checkbox rounded-field checkbox-sm" />
                                </div>
                                <div class="flex flex-1 justify-center">
                                    <input type="checkbox" class="checkbox rounded-field checkbox-sm" />
                                </div>
                                <div class="flex flex-1 justify-center">
                                    <input type="checkbox" class="checkbox rounded-field checkbox-sm" />
                                </div>
                            </div>
                            <button class="btn btn-square btn-ghost btn-sm tooltip tooltip-left" data-tip="Remove"
                                onclick={() => {
                                    selected = selected.filter(s => s.id !== user.id);
                                }}
                            >
                                <X class="size-4" />
                            </button>
                        </li>
                    {/each}
                {/if}
            </ul>
        </div>
        
        <div class="modal-action">
            <button class="btn btn-ghost" onclick={() => {
                onCancel?.();
                shareWorkspaceModal?.close();
            }}>
                Cancel
            </button>
            <button class="btn btn-primary" onclick={() => {
                onShare?.();
                shareWorkspaceModal?.close();
            }}>
                Share
            </button>
        </div>
    </div>
    <form method="dialog" class="modal-backdrop">
        <button>close</button>
    </form>
</dialog>