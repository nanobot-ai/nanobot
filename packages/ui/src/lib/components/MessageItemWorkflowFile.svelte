<script lang="ts">
	import type { ChatMessageItemToolCall } from '$lib/types';
	import { parseToolFilePath } from '$lib/utils';
	import { FileIcon, FolderIcon } from '@lucide/svelte';
	interface Props {
		item: ChatMessageItemToolCall;
		onFileOpen?: (filename: string) => void;
	}

	let { item, onFileOpen }: Props = $props();

    const pending = $derived(item.hasMore);
    const name = $derived(item.arguments ? parseToolFilePath(item): item.name);
</script>

<div class="rounded-field text mt-3 mb-2 w-full border border-base-200 dark:border-base-300 bg-base-100 shadow-xs p-3">
    <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
            <FileIcon class="size-4" />
            
            {#if pending}
                <span class="text-sm animate-pulse">...</span>
            {:else}
                <span class="text-sm">{name}</span>
            {/if}
        </div>
        <button class="btn btn-sm tooltip" data-tip="Open" onclick={() => onFileOpen?.(name)}>
            <FolderIcon class="size-4" />
        </button>
    </div>
</div>