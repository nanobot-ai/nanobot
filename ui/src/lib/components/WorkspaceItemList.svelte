<script lang="ts">
	import { page } from '$app/state';
	import { MoreVertical, Edit, Trash2, X, Check } from '@lucide/svelte';
	import type { WorkspaceItem } from '$lib/types';

	interface Props {
		items: WorkspaceItem[];
		workspaceId: string;
		onRename: (itemId: string, newTitle: string) => void;
		onDelete: (itemId: string) => void;
		onItemClick?: () => void;
	}

	let { items, workspaceId, onRename, onDelete, onItemClick }: Props = $props();

	// Get current route params to highlight active item
	const currentTaskId = $derived(page.params.taskId);

	let editingItemId = $state<string | null>(null);
	let editTitle = $state('');

	function formatTime(timestamp: string): string {
		const now = new Date();
		const diff = now.getTime() - new Date(timestamp).getTime();
		const minutes = Math.floor(diff / (1000 * 60));
		const hours = Math.floor(diff / (1000 * 60 * 60));
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));

		if (minutes < 1) return 'now';
		if (minutes < 60) return `${minutes}m`;
		if (hours < 24) return `${hours}h`;
		return `${days}d`;
	}

	function startRename(itemId: string, currentTitle: string) {
		editingItemId = itemId;
		editTitle = currentTitle || '';
	}

	function saveRename() {
		if (editingItemId && editTitle.trim()) {
			onRename(editingItemId, editTitle.trim());
			editingItemId = null;
			editTitle = '';
		}
	}

	function cancelRename() {
		editingItemId = null;
		editTitle = '';
	}

	function handleDelete(itemId: string) {
		onDelete(itemId);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			saveRename();
		} else if (e.key === 'Escape') {
			cancelRename();
		}
	}

	function getItemUrl(item: WorkspaceItem): string {
		// Get URL for the item based on type
		if (item.type === 'task') {
			return `/w/${workspaceId}/t/${item.id}`;
		} else {
			// TODO: Add routes for agents and conversations
			return '#';
		}
	}
</script>

{#if items.length === 0}
	<div class="px-3 py-2 text-xs text-base-content/40 italic">No items</div>
{:else}
	{#each items as item (item.id)}
		{@const isActive = item.type === 'task' && item.id === currentTaskId}
		{@const itemUrl = getItemUrl(item)}
		<div
			class="group flex items-center border-b border-base-200 {isActive
				? 'bg-primary/10'
				: 'hover:bg-base-100'}"
		>
			<!-- Item title area (clickable) -->
			<a
				href={itemUrl}
				class="flex-1 truncate py-2 pr-3 pl-6 text-left transition-colors focus:outline-none {isActive
					? 'font-semibold'
					: ''}"
				onclick={() => onItemClick?.()}
			>
				<div class="flex items-center justify-between gap-2">
					<div class="flex min-w-0 flex-1 items-center gap-2">
						{#if editingItemId === item.id}
							<input
								type="text"
								bind:value={editTitle}
								onkeydown={handleKeydown}
								class="input input-sm min-w-0 flex-1"
								onclick={(e) => e.stopPropagation()}
								onfocus={(e) => (e.target as HTMLInputElement).select()}
							/>
						{:else}
							<span class="truncate text-sm">{item.title || 'Untitled'}</span>
							{#if item.status === 'completed'}
								<span class="badge badge-xs badge-success">Done</span>
							{/if}
						{/if}
					</div>
					{#if editingItemId !== item.id}
						<span class="flex-shrink-0 text-xs text-base-content/50">
							{formatTime(item.created)}
						</span>
					{/if}
				</div>
			</a>

			<!-- Save/Cancel buttons for editing -->
			{#if editingItemId === item.id}
				<div class="flex items-center gap-1 px-2">
					<button class="btn btn-ghost btn-xs" onclick={cancelRename} aria-label="Cancel editing">
						<X class="h-3 w-3" />
					</button>
					<button
						class="btn text-success btn-ghost btn-xs hover:bg-success/20"
						onclick={saveRename}
						aria-label="Save changes"
					>
						<Check class="h-3 w-3" />
					</button>
				</div>
			{/if}

			{#if editingItemId !== item.id}
				<!-- Dropdown menu - only show on hover -->
				<div class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100">
					<div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm">
						<MoreVertical class="h-4 w-4" />
					</div>
					<ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow">
						<li>
							<button onclick={() => startRename(item.id, item.title)} class="text-sm">
								<Edit class="h-4 w-4" />
								Rename
							</button>
						</li>
						<li>
							<button onclick={() => handleDelete(item.id)} class="text-sm text-error">
								<Trash2 class="h-4 w-4" />
								Delete
							</button>
						</li>
					</ul>
				</div>
			{/if}
		</div>
	{/each}
{/if}
