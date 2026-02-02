<script lang="ts">
	import type { ChatService } from '$lib/chat.svelte';
	import type { ResourceContents } from '$lib/types';
	import { X } from '@lucide/svelte';
	import { linear } from 'svelte/easing';
	import { slide } from 'svelte/transition';
	import MarkdownEditor from './MarkdownEditor.svelte';

	interface Props {
		filename: string;
		chat: ChatService;
        onClose: () => void;
	}

	let { filename, chat, onClose }: Props = $props();

	let resource = $state<ResourceContents | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	$effect(() => {
		// Reset state when filename changes
		resource = null;
		loading = true;
		error = null;

		// Initial fetch of the resource
        const match = chat.resources.find((r) => r.name === filename);
        if (!match) {
            loading = false;
            return;
        }

		chat
			.readResource(match.uri)
			.then((result) => {
				if (result.contents?.length) {
					resource = result.contents[0];
				}
				loading = false;
			})
			.catch((e) => {
				error = e instanceof Error ? e.message : String(e);
				loading = false;
			});

		// Subscribe to live updates
		const cleanup = chat.watchResource(filename, (updatedResource) => {
			resource = updatedResource;
		});

		// Cleanup subscription when component unmounts or filename changes
		return cleanup;
	});

	// Derive the content to display
	let content = $derived(resource?.text ?? '');
	let mimeType = $derived(resource?.mimeType ?? 'text/plain');
</script>

<div
	class="min-w-[500px] h-dvh bg-base-200 flex flex-col"
	in:slide={{ axis: 'x', easing: linear, duration: 50 }}
>
    <div class="flex gap-2 items-center px-4 py-2 border-b border-base-300">
        <div class="flex grow items-center justify-between">
            <span class="text-sm font-medium truncate">{filename}</span>
            {#if mimeType}
                <span class="text-xs text-base-content/60">{mimeType}</span>
            {/if}
        </div>
        <button class="btn btn-sm btn-square tooltip tooltip-left" data-tip="Close" onclick={onClose}>
            <X class="size-4" />
        </button>
    </div>

	<div class="flex-1 overflow-auto p-4 pt-0">
		{#if loading}
			<div class="flex items-center justify-center h-full">
				<span class="loading loading-spinner loading-md"></span>
			</div>
		{:else if error}
			<div class="alert alert-error">
				<span>Failed to load resource: {error}</span>
			</div>
		{:else if resource?.blob}
			<!-- Binary content - show as image if possible -->
			{#if mimeType.startsWith('image/')}
				<img
					src="data:{mimeType};base64,{resource.blob}"
					alt={filename}
					class="max-w-full h-auto"
				/>
			{:else}
				<div class="text-base-content/60">Binary content ({mimeType})</div>
			{/if}
		{:else if content}
			<MarkdownEditor value={content} />
		{:else}
			<div class="text-base-content/60 italic">The contents of this file are empty.</div>
		{/if}
	</div>
</div>