<script lang="ts">
	import { X } from '@lucide/svelte';
	import { fly } from 'svelte/transition';
	import ThreadFromChat from '$lib/components/ThreadFromChat.svelte';
	import { sharedChat } from '$lib/stores/chat.svelte';
	import type { Attachment } from '$lib/types';

	type Props = {
		width: number;
		isResizing: boolean;
		files: Attachment[];
		onClose: () => void;
		onStartResize: (e: MouseEvent) => void;
	};

	let { width, isResizing, files, onClose, onStartResize }: Props = $props();
</script>

<!-- Resize Handle -->
<button
	type="button"
	class="w-1 bg-base-300 hover:bg-primary/50 cursor-col-resize transition-all duration-150 shrink-0 active:bg-primary border-none p-0"
	aria-label="Resize sidebar"
	onmousedown={onStartResize}
></button>

<!-- Sidebar Thread -->
<div
	transition:fly={{ x: 100, duration: 200 }}
	class="border-l border-l-base-300 bg-base-100 h-dvh flex flex-col shrink-0 {isResizing
		? 'select-none'
		: ''}"
	style="width: {width}px; min-width: 400px;"
>
	<div class="w-full flex justify-between items-center p-4 bg-base-100 shrink-0">
		<div class="w-full"></div>
		<button
			class="btn btn-ghost btn-square btn-sm tooltip tooltip-left"
			data-tip="Close"
			onclick={onClose}
		>
			<X class="size-4" />
		</button>
	</div>
	<div class="w-full flex-1 min-h-0 flex flex-col">
		{#if sharedChat.current}
			<ThreadFromChat
				inline
				chat={sharedChat.current}
				{files}
				agent={{ name: 'Workflow Assistant', icon: '/assets/obot-icon-blue.svg' }}
			/>
		{/if}
	</div>
</div>
