<script lang="ts">
	import { page } from '$app/stores';
	import {
		Folder,
		FolderOpen,
		Plus,
		MoreVertical,
		Edit,
		Trash2,
		X,
		Check,
		ChevronRight,
		ListTodo,
		Bot,
		MessageSquare,
		FileText
	} from '@lucide/svelte';
	import { workspaceStore } from '$lib/stores/workspaces.svelte';
	import WorkspaceItemList from './WorkspaceItemList.svelte';
	import FileTree from './FileTree.svelte';
	import type { FileTreeNode } from '$lib/types';

	interface Props {
		onWorkspaceClick?: () => void;
	}

	let { onWorkspaceClick }: Props = $props();

	// Auto-expand workspace and section when viewing a task
	const currentWorkspaceId = $derived($page.params.workspaceId);
	const currentTaskId = $derived($page.params.taskId);

	$effect(() => {
		if (currentWorkspaceId && currentTaskId) {
			// Ensure the workspace is expanded
			if (!workspaceStore.isWorkspaceExpanded(currentWorkspaceId)) {
				workspaceStore.toggleWorkspace(currentWorkspaceId);
			}
			// Get the item store for this workspace and ensure the tasks section is expanded
			const itemStore = workspaceStore.getItemStore(currentWorkspaceId);
			if (!itemStore.isSectionExpanded('task')) {
				itemStore.toggleSection('task');
			}
		}
	});

	let editingWorkspaceId = $state<string | null>(null);
	let editName = $state('');
	let isCreatingWorkspace = $state(false);
	let newWorkspaceName = $state('');

	function toggleWorkspace(workspaceId: string) {
		workspaceStore.toggleWorkspace(workspaceId);
	}

	function toggleSection(workspaceId: string, sectionType: string) {
		const itemStore = workspaceStore.getItemStore(workspaceId);
		itemStore.toggleSection(sectionType);
	}

	function startRename(workspaceId: string, currentName: string) {
		editingWorkspaceId = workspaceId;
		editName = currentName || '';
	}

	async function saveRename() {
		if (editingWorkspaceId && editName.trim()) {
			try {
				await workspaceStore.updateWorkspace(editingWorkspaceId, { name: editName.trim() });
				editingWorkspaceId = null;
				editName = '';
			} catch (error) {
				console.error('Failed to rename workspace:', error);
			}
		}
	}

	function cancelRename() {
		editingWorkspaceId = null;
		editName = '';
	}

	async function handleDelete(workspaceId: string) {
		try {
			await workspaceStore.deleteWorkspace(workspaceId);
		} catch (error) {
			console.error('Failed to delete workspace:', error);
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			saveRename();
		} else if (e.key === 'Escape') {
			cancelRename();
		}
	}

	function startCreatingWorkspace() {
		isCreatingWorkspace = true;
		newWorkspaceName = '';
	}

	async function createWorkspace() {
		if (newWorkspaceName.trim()) {
			try {
				await workspaceStore.createWorkspace(newWorkspaceName.trim());
				isCreatingWorkspace = false;
				newWorkspaceName = '';
			} catch (error) {
				console.error('Failed to create workspace:', error);
			}
		}
	}

	function cancelCreateWorkspace() {
		isCreatingWorkspace = false;
		newWorkspaceName = '';
	}

	function handleNewWorkspaceKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			createWorkspace();
		} else if (e.key === 'Escape') {
			cancelCreateWorkspace();
		}
	}

	async function handleItemRename(workspaceId: string, itemId: string, newTitle: string) {
		try {
			const itemStore = workspaceStore.getItemStore(workspaceId);
			await itemStore.updateItem(itemId, { title: newTitle });
		} catch (error) {
			console.error('Failed to rename item:', error);
		}
	}

	async function handleItemDelete(workspaceId: string, itemId: string) {
		try {
			const itemStore = workspaceStore.getItemStore(workspaceId);
			await itemStore.deleteItem(itemId);
		} catch (error) {
			console.error('Failed to delete item:', error);
		}
	}

	function closeMobileSidebar() {
		onWorkspaceClick?.();
	}

	function handleFileClick(node: FileTreeNode) {
		// TODO: Open file for viewing/editing
		console.log('File clicked:', node);
		onWorkspaceClick?.();
	}
</script>

<div class="flex h-full flex-col">
	<!-- Header -->
	<div class="flex flex-shrink-0 items-center justify-between p-2">
		<h2 class="font-semibold text-base-content/60">Workspaces</h2>
		<button
			class="btn btn-ghost btn-xs"
			onclick={startCreatingWorkspace}
			aria-label="Create new workspace"
		>
			<Plus class="h-4 w-4" />
		</button>
	</div>

	<!-- Workspace list -->
	<div class="flex-1 overflow-y-auto">
		{#if workspaceStore.isLoading}
			<!-- Skeleton UI when loading -->
			{#each Array(3).fill(null) as _, index (index)}
				<div class="border-b border-base-200 p-3">
					<div class="h-5 w-48 skeleton"></div>
				</div>
			{/each}
		{:else}
			<!-- New workspace input -->
			{#if isCreatingWorkspace}
				<div class="flex items-center gap-2 border-b border-base-200 bg-base-100 p-3">
					<input
						type="text"
						bind:value={newWorkspaceName}
						onkeydown={handleNewWorkspaceKeydown}
						placeholder="Workspace name..."
						class="input input-sm flex-1"
					/>
					<button class="btn btn-ghost btn-xs" onclick={cancelCreateWorkspace} aria-label="Cancel">
						<X class="h-3 w-3" />
					</button>
					<button
						class="btn text-success btn-ghost btn-xs hover:bg-success/20"
						onclick={createWorkspace}
						aria-label="Create"
					>
						<Check class="h-3 w-3" />
					</button>
				</div>
			{/if}

			<!-- Workspaces -->
			{#each workspaceStore.workspaces as workspace (workspace.id)}
				<div class="border-b border-base-200">
					<!-- Workspace header -->
					<div class="group flex items-center hover:bg-base-100">
						<!-- Toggle button -->
						<button
							class="btn px-2 btn-ghost btn-xs"
							onclick={() => toggleWorkspace(workspace.id)}
							aria-label={workspaceStore.isWorkspaceExpanded(workspace.id)
								? 'Collapse workspace'
								: 'Expand workspace'}
						>
							<ChevronRight
								class="h-3 w-3 transition-transform {workspaceStore.isWorkspaceExpanded(
									workspace.id
								)
									? 'rotate-90'
									: ''}"
							/>
						</button>

						<!-- Workspace name area -->
						<button
							class="flex-1 truncate py-2 text-left transition-colors focus:outline-none"
							onclick={() => toggleWorkspace(workspace.id)}
						>
							<div class="flex items-center gap-2">
								{#if workspaceStore.isWorkspaceExpanded(workspace.id)}
									<FolderOpen class="h-4 w-4" style="color: {workspace.color || '#888'}" />
								{:else}
									<Folder class="h-4 w-4" style="color: {workspace.color || '#888'}" />
								{/if}

								{#if editingWorkspaceId === workspace.id}
									<input
										type="text"
										bind:value={editName}
										onkeydown={handleKeydown}
										class="input input-sm min-w-0 flex-1"
										onclick={(e) => e.stopPropagation()}
										onfocus={(e) => (e.target as HTMLInputElement).select()}
									/>
								{:else}
									<span class="truncate text-sm font-medium">{workspace.name}</span>
								{/if}
							</div>
						</button>

						<!-- Save/Cancel buttons for editing -->
						{#if editingWorkspaceId === workspace.id}
							<div class="flex items-center gap-1 px-2">
								<button
									class="btn btn-ghost btn-xs"
									onclick={cancelRename}
									aria-label="Cancel editing"
								>
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

						{#if editingWorkspaceId !== workspace.id}
							<!-- Dropdown menu -->
							<div
								class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100"
							>
								<div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm">
									<MoreVertical class="h-4 w-4" />
								</div>
								<ul
									class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"
								>
									<li>
										<button
											onclick={() => startRename(workspace.id, workspace.name)}
											class="text-sm"
										>
											<Edit class="h-4 w-4" />
											Rename
										</button>
									</li>
									<li>
										<button onclick={() => handleDelete(workspace.id)} class="text-sm text-error">
											<Trash2 class="h-4 w-4" />
											Delete
										</button>
									</li>
								</ul>
							</div>
						{/if}
					</div>

					<!-- Workspace content (expandable) -->
					{#if workspaceStore.isWorkspaceExpanded(workspace.id)}
						{@const itemStore = workspaceStore.getItemStore(workspace.id)}
						{@const taskCount = itemStore.getItemCount('task')}
						{@const tasksExpanded = itemStore.isSectionExpanded('task')}
						{@const agentCount = itemStore.getItemCount('agent')}
						{@const agentsExpanded = itemStore.isSectionExpanded('agent')}
						{@const conversationCount = itemStore.getItemCount('conversation')}
						{@const conversationsExpanded = itemStore.isSectionExpanded('conversation')}
						{@const fileCount = itemStore.getFileCount()}
						{@const filesExpanded = itemStore.isSectionExpanded('files')}
						{@const fileTree = itemStore.buildFileTree()}
						<div class="bg-base-50">
							<!-- Tasks Section -->
							<div class="border-t border-base-200/50">
								<button
									class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"
									onclick={() => toggleSection(workspace.id, 'task')}
								>
									<ChevronRight
										class="h-3 w-3 transition-transform {tasksExpanded ? 'rotate-90' : ''}"
									/>
									<ListTodo class="h-3.5 w-3.5" />
									<span class="flex-1">Tasks</span>
									{#if taskCount > 0}
										<span class="badge badge-xs">{taskCount}</span>
									{/if}
								</button>
								{#if tasksExpanded}
									<WorkspaceItemList
										items={itemStore.getItems('task')}
										workspaceId={workspace.id}
										onRename={(itemId, title) => handleItemRename(workspace.id, itemId, title)}
										onDelete={(itemId) => handleItemDelete(workspace.id, itemId)}
										onItemClick={closeMobileSidebar}
									/>
								{/if}
							</div>

							<!-- Agents Section -->
							<div class="border-t border-base-200/50">
								<button
									class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"
									onclick={() => toggleSection(workspace.id, 'agent')}
								>
									<ChevronRight
										class="h-3 w-3 transition-transform {agentsExpanded ? 'rotate-90' : ''}"
									/>
									<Bot class="h-3.5 w-3.5" />
									<span class="flex-1">Agents</span>
									{#if agentCount > 0}
										<span class="badge badge-xs">{agentCount}</span>
									{/if}
								</button>
								{#if agentsExpanded}
									<WorkspaceItemList
										items={itemStore.getItems('agent')}
										workspaceId={workspace.id}
										onRename={(itemId, title) => handleItemRename(workspace.id, itemId, title)}
										onDelete={(itemId) => handleItemDelete(workspace.id, itemId)}
										onItemClick={closeMobileSidebar}
									/>
								{/if}
							</div>

							<!-- Conversations Section -->
							<div class="border-t border-base-200/50">
								<button
									class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"
									onclick={() => toggleSection(workspace.id, 'conversation')}
								>
									<ChevronRight
										class="h-3 w-3 transition-transform {conversationsExpanded ? 'rotate-90' : ''}"
									/>
									<MessageSquare class="h-3.5 w-3.5" />
									<span class="flex-1">Conversations</span>
									{#if conversationCount > 0}
										<span class="badge badge-xs">{conversationCount}</span>
									{/if}
								</button>
								{#if conversationsExpanded}
									<WorkspaceItemList
										items={itemStore.getItems('conversation')}
										workspaceId={workspace.id}
										onRename={(itemId, title) => handleItemRename(workspace.id, itemId, title)}
										onDelete={(itemId) => handleItemDelete(workspace.id, itemId)}
										onItemClick={closeMobileSidebar}
									/>
								{/if}
							</div>

							<!-- Files Section -->
							<div class="border-t border-base-200/50">
								<button
									class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"
									onclick={() => toggleSection(workspace.id, 'files')}
								>
									<ChevronRight
										class="h-3 w-3 transition-transform {filesExpanded ? 'rotate-90' : ''}"
									/>
									<FileText class="h-3.5 w-3.5" />
									<span class="flex-1">Files</span>
									{#if fileCount > 0}
										<span class="badge badge-xs">{fileCount}</span>
									{/if}
								</button>
								{#if filesExpanded}
									<div class="max-h-64 overflow-y-auto">
										<FileTree nodes={fileTree} {itemStore} onFileClick={handleFileClick} />
									</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{/each}

			{#if workspaceStore.workspaces.length === 0 && !isCreatingWorkspace}
				<div class="p-4 text-center text-sm text-base-content/40">
					No workspaces yet. Click the + button to create one.
				</div>
			{/if}
		{/if}
	</div>
</div>
