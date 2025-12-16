<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { workspaceStore } from '$lib/stores/workspaces.svelte';
	import type { TaskNode } from '$lib/types';
	import TaskFlowchart from '$lib/components/TaskFlowchart.svelte';
	import TaskNodeEditor from '$lib/components/TaskNodeEditor.svelte';

	const workspaceId = $page.params.workspaceId ?? '';
	const taskId = $page.params.taskId ?? '';

	// Get the item store for this workspace
	const itemStore = workspaceStore.getItemStore(workspaceId);

	// Get the task, workspace, and flowchart
	const task = $derived(itemStore.items.find((item) => item.id === taskId && item.type === 'task'));
	const workspace = $derived(workspaceStore.workspaces.find((w) => w.id === workspaceId));
	const flowchart = $derived(taskId ? itemStore.getTaskFlowchart(taskId) : undefined);

	let selectedNode = $state<TaskNode | null>(null);

	function handleNodeClick(node: TaskNode) {
		selectedNode = node;
	}
</script>

<div class="flex h-full flex-col bg-base-100">
	{#if task && workspace && flowchart}
		<!-- Header spanning full width -->
		<div class="flex-shrink-0 bg-base-200 p-4">
			<div class="mb-2 text-xs text-base-content/60">
				<button onclick={() => goto(resolve('/'))} class="hover:text-base-content"
					>Workspaces</button
				>
				<span class="mx-1">/</span>
				<span>{workspace.name}</span>
				<span class="mx-1">/</span>
				<span>Tasks</span>
			</div>
			<h1 class="text-xl font-bold">{task.title}</h1>
		</div>

		<!-- Content area with main panel and flowchart -->
		<div class="flex flex-1 overflow-hidden">
			<!-- Main Content Panel (Left) -->
			<div class="flex flex-1 flex-col overflow-hidden border-r border-base-200 bg-base-200">
				<!-- Panel Content -->
				<div class="flex-1 overflow-y-auto p-4">
					{#if selectedNode && flowchart && taskId && workspaceId}
						<TaskNodeEditor
							{selectedNode}
							{taskId}
							{workspaceId}
							onNavigateToNode={handleNodeClick}
						/>
					{:else}
						<div class="flex h-full items-center justify-center text-center">
							<div class="text-base-content/60">
								<p class="mb-2 text-lg font-medium">Select a Node</p>
								<p class="text-sm">Click on any node in the flowchart to view its details</p>
							</div>
						</div>
					{/if}
				</div>
			</div>

			<!-- Flowchart Area (Right) - Side Reference -->
			<div class="flex-shrink-0 overflow-auto bg-base-100">
				<div class="mx-auto" style="width: fit-content;">
					<TaskFlowchart
						{flowchart}
						selectedNodeId={selectedNode?.id}
						onNodeClick={handleNodeClick}
					/>
				</div>
			</div>
		</div>
	{:else}
		<!-- Not Found -->
		<div class="flex h-full w-full items-center justify-center">
			<div class="text-center">
				<h2 class="mb-2 text-2xl font-bold">Task Not Found</h2>
				<p class="mb-4 text-base-content/60">
					The task you're looking for doesn't exist or has been deleted.
				</p>
				<button onclick={() => goto(resolve('/'))} class="btn btn-primary">Back to Home</button>
			</div>
		</div>
	{/if}
</div>
