<script lang="ts">
	import { ChevronRight, File, Folder, FolderOpen } from '@lucide/svelte';
	import type { FileTreeNode } from '$lib/types';
	import type { WorkspaceItemStore } from '$lib/stores/workspace-item.svelte';
	import Self from './FileTree.svelte';

	interface Props {
		nodes: FileTreeNode[];
		itemStore: WorkspaceItemStore;
		onFileClick?: (node: FileTreeNode) => void;
	}

	let { nodes, itemStore, onFileClick }: Props = $props();

	function toggleNode(node: FileTreeNode) {
		if (node.isDirectory) {
			itemStore.toggleFilePath(node.path);
		} else if (onFileClick) {
			onFileClick(node);
		}
	}
</script>

{#each nodes as node (node.path)}
	<div>
		{#if node.isDirectory}
			{@const isExpanded = itemStore.isFilePathExpanded(node.path)}
			<!-- Directory -->
			<button
				class="flex w-full items-center gap-1 py-1 pr-2 pl-6 text-left text-sm hover:bg-base-100"
				onclick={() => toggleNode(node)}
			>
				<ChevronRight
					class="h-3 w-3 flex-shrink-0 transition-transform {isExpanded ? 'rotate-90' : ''}"
				/>
				{#if isExpanded}
					<FolderOpen class="h-3.5 w-3.5 flex-shrink-0 text-warning" />
				{:else}
					<Folder class="h-3.5 w-3.5 flex-shrink-0 text-warning" />
				{/if}
				<span class="truncate">{node.name}</span>
			</button>

			<!-- Children (recursive) -->
			{#if isExpanded && node.children}
				<div class="pl-4">
					<Self nodes={node.children} {itemStore} {onFileClick} />
				</div>
			{/if}
		{:else}
			<!-- File -->
			<button
				class="flex w-full items-center gap-1 py-1 pr-2 pl-9 text-left text-sm hover:bg-base-100"
				onclick={() => toggleNode(node)}
			>
				<File class="h-3.5 w-3.5 flex-shrink-0" />
				<span class="truncate">{node.name}</span>
			</button>
		{/if}
	</div>
{/each}
