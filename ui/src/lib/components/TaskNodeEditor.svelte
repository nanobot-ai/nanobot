<script lang="ts">
	import { marked } from 'marked';
	import { workspaceStore } from '$lib/stores/workspaces.svelte';
	import type { TaskNode, TaskNodeInput } from '$lib/types';
	import { Wrench, Bot, ListTodo, X, Pencil, Trash, Plus } from '@lucide/svelte';

	interface Props {
		selectedNode: TaskNode;
		taskId: string;
		workspaceId: string;
		onNavigateToNode: (node: TaskNode) => void;
	}

	let { selectedNode, taskId, workspaceId, onNavigateToNode }: Props = $props();

	// Get item store for this workspace
	const itemStore = workspaceStore.getItemStore(workspaceId);

	// Get flowchart and derived data
	const flowchart = $derived(itemStore.getTaskFlowchart(taskId));
	const availableAgents = $derived(itemStore.getItems('agent'));
	const availableTasks = $derived(itemStore.getItems('task'));
	const availableTargetNodes = $derived(
		flowchart?.nodes.filter((n) => n.type !== 'start' && n.id !== selectedNode.id) ?? []
	);

	// Helper functions to get names from IDs
	function getAgentName(agentId: string): string {
		const agent = itemStore.items.find((item) => item.id === agentId && item.type === 'agent');
		return agent?.title || agentId;
	}

	function getTaskName(taskId: string): string {
		const task = itemStore.items.find((item) => item.id === taskId && item.type === 'task');
		return task?.title || taskId;
	}

	// Assignment management state
	let newTool = $state('');
	let newAgent = $state('');
	let newTask = $state('');
	let showAddTool = $state(false);
	let showAddAgent = $state(false);
	let showAddTask = $state(false);

	// Input editor state
	let showAddInput = $state(false);
	let editingInputId = $state<string | null>(null);
	let inputName = $state('');
	let inputDescription = $state('');
	let inputRequired = $state(true);

	// Edge editor state
	let showAddEdge = $state(false);
	let editingEdges = $state(false);
	let edgeType = $state<'simple' | 'decision'>('simple');
	let simpleTarget = $state('');
	let simpleTargetIsNew = $state(false);
	let newSimpleNodeLabel = $state('');
	let newSimpleNodeType = $state<'process' | 'decision' | 'end'>('process');
	let decisionCondition = $state('');
	let yesTarget = $state('');
	let yesTargetIsNew = $state(false);
	let newYesNodeLabel = $state('');
	let newYesNodeType = $state<'process' | 'decision' | 'end'>('process');
	let noTarget = $state('');
	let noTargetIsNew = $state(false);
	let newNoNodeLabel = $state('');
	let newNoNodeType = $state<'process' | 'decision' | 'end'>('process');

	// Assignment management functions
	async function addTool() {
		if (newTool.trim()) {
			await itemStore.addNodeAssignment(taskId, selectedNode.id, 'tools', newTool.trim());
			newTool = '';
			showAddTool = false;
		}
	}

	async function addAgent() {
		if (newAgent.trim()) {
			await itemStore.addNodeAssignment(taskId, selectedNode.id, 'agents', newAgent.trim());
			newAgent = '';
			showAddAgent = false;
		}
	}

	async function addTask() {
		if (newTask.trim()) {
			await itemStore.addNodeAssignment(taskId, selectedNode.id, 'tasks', newTask.trim());
			newTask = '';
			showAddTask = false;
		}
	}

	// Input management functions
	function startAddInput() {
		showAddInput = true;
		editingInputId = null;
		inputName = '';
		inputDescription = '';
		inputRequired = true;
	}

	function startEditInput(input: TaskNodeInput) {
		showAddInput = true;
		editingInputId = input.id;
		inputName = input.name;
		inputDescription = input.description;
		inputRequired = input.required;
	}

	function cancelInputForm() {
		showAddInput = false;
		editingInputId = null;
		inputName = '';
		inputDescription = '';
		inputRequired = true;
	}

	async function saveInput() {
		if (!inputName.trim()) return;

		if (editingInputId) {
			await itemStore.updateNodeInput(taskId, selectedNode.id, editingInputId, {
				name: inputName.trim(),
				description: inputDescription.trim(),
				required: inputRequired
			});
		} else {
			await itemStore.addNodeInput(
				taskId,
				selectedNode.id,
				inputName.trim(),
				inputDescription.trim(),
				inputRequired
			);
		}

		cancelInputForm();
	}

	// Edge management functions
	function startAddEdge() {
		showAddEdge = true;
		editingEdges = false;
		edgeType = 'simple';
		resetEdgeForm();
	}

	function startEditEdge() {
		if (!flowchart) return;
		const outgoingEdges = flowchart.edges.filter((e) => e.source === selectedNode.id);

		const hasDecisionEdges = outgoingEdges.some((e) => e.label === 'Yes' || e.label === 'No');

		if (hasDecisionEdges) {
			edgeType = 'decision';
			const yesEdge = outgoingEdges.find((e) => e.label === 'Yes');
			const noEdge = outgoingEdges.find((e) => e.label === 'No');

			yesTarget = yesEdge?.target || '';
			noTarget = noEdge?.target || '';
			yesTargetIsNew = false;
			noTargetIsNew = false;
		} else {
			edgeType = 'simple';
			simpleTarget = outgoingEdges[0]?.target || '';
			simpleTargetIsNew = false;
		}

		showAddEdge = true;
		editingEdges = true;
	}

	function resetEdgeForm() {
		simpleTarget = '';
		simpleTargetIsNew = false;
		newSimpleNodeLabel = '';
		newSimpleNodeType = 'process';
		decisionCondition = '';
		yesTarget = '';
		yesTargetIsNew = false;
		newYesNodeLabel = '';
		newYesNodeType = 'process';
		noTarget = '';
		noTargetIsNew = false;
		newNoNodeLabel = '';
		newNoNodeType = 'process';
	}

	function cancelAddEdge() {
		showAddEdge = false;
		editingEdges = false;
		resetEdgeForm();
	}

	async function saveEdge() {
		if (!flowchart) return;

		try {
			// If editing, delete existing edges first
			if (editingEdges) {
				const outgoingEdges = flowchart.edges.filter((e) => e.source === selectedNode.id);
				for (const edge of outgoingEdges) {
					await itemStore.deleteEdge(taskId, edge.id);
				}
			}

			if (edgeType === 'simple') {
				let targetNodeId = simpleTarget;

				// Create new node if needed
				if (simpleTargetIsNew && newSimpleNodeLabel.trim()) {
					targetNodeId = await itemStore.createNode(
						taskId,
						newSimpleNodeType,
						newSimpleNodeLabel.trim(),
						'# ' + newSimpleNodeLabel.trim(),
						selectedNode.id
					);
				}

				if (targetNodeId) {
					await itemStore.addEdge(taskId, selectedNode.id, targetNodeId);
				}
			} else {
				// Decision node
				let yesNodeId = yesTarget;
				let noNodeId = noTarget;

				// Create yes node if needed
				if (yesTargetIsNew && newYesNodeLabel.trim()) {
					yesNodeId = await itemStore.createNode(
						taskId,
						newYesNodeType,
						newYesNodeLabel.trim(),
						'# ' + newYesNodeLabel.trim(),
						selectedNode.id
					);
				}

				// Create no node if needed
				if (noTargetIsNew && newNoNodeLabel.trim()) {
					noNodeId = await itemStore.createNode(
						taskId,
						newNoNodeType,
						newNoNodeLabel.trim(),
						'# ' + newNoNodeLabel.trim(),
						selectedNode.id
					);
				}

				// Add edges with labels
				if (yesNodeId) {
					await itemStore.addEdge(taskId, selectedNode.id, yesNodeId, 'Yes');
				}
				if (noNodeId) {
					await itemStore.addEdge(taskId, selectedNode.id, noNodeId, 'No');
				}
			}

			// Reset form
			showAddEdge = false;
			editingEdges = false;
			resetEdgeForm();
		} catch (error) {
			console.error('Failed to save edge:', error);
		}
	}

	function renderMarkdown(content: string): string {
		return marked(content) as string;
	}
</script>

<div class="space-y-4">
	<!-- Node Header -->
	<div class="flex items-start justify-between">
		<div>
			<h2 class="text-lg font-semibold">{selectedNode.label}</h2>
			<span class="badge badge-sm {selectedNode.completed ? 'badge-success' : 'badge-ghost'}">
				{selectedNode.type}
			</span>
		</div>
		{#if selectedNode.type !== 'start' && selectedNode.type !== 'end'}
			<button
				class="btn btn-sm {selectedNode.completed ? 'btn-success' : 'btn-outline'}"
				onclick={() => itemStore.toggleNodeCompletion(taskId, selectedNode.id)}
			>
				{selectedNode.completed ? '✓ Done' : 'Mark Done'}
			</button>
		{/if}
	</div>

	<!-- Section 1: For Start nodes show Inputs, for others show Using... -->
	{#if selectedNode.type === 'start'}
		<!-- Inputs Section for Start Nodes -->
		<div class="rounded-lg border border-base-300 bg-base-100 p-4">
			<h3 class="mb-3 text-sm font-semibold text-base-content/70">Inputs</h3>

			{#if selectedNode.inputs && selectedNode.inputs.length > 0}
				<div class="space-y-2">
					{#each selectedNode.inputs as input (input.name)}
						<div class="bg-base-50 rounded border border-base-200 p-3">
							<div class="mb-1 flex items-start justify-between gap-2">
								<div class="flex-1">
									<div class="flex items-center gap-2">
										<span class="text-sm font-medium">{input.name}</span>
										<span class="badge badge-xs {input.required ? 'badge-error' : 'badge-ghost'}">
											{input.required ? 'Required' : 'Optional'}
										</span>
									</div>
									{#if input.description}
										<p class="mt-1 text-xs text-base-content/60">
											{input.description}
										</p>
									{/if}
								</div>
								<div class="flex gap-1">
									<button
										onclick={() => startEditInput(input)}
										class="btn btn-square btn-ghost btn-xs"
										aria-label="Edit input"
									>
										<Pencil class="h-3 w-3" />
									</button>
									<button
										onclick={() => itemStore.deleteNodeInput(taskId, selectedNode.id, input.id)}
										class="btn btn-square btn-ghost btn-xs hover:text-error"
										aria-label="Delete input"
									>
										<Trash class="h-3 w-3" />
									</button>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{:else}
				<p class="text-sm text-base-content/60">No inputs defined</p>
			{/if}

			{#if !showAddInput}
				<button class="btn mt-3 btn-block gap-1 btn-outline btn-sm" onclick={startAddInput}>
					<Plus class="h-3 w-3" />
					Add Input
				</button>
			{:else}
				<!-- Input Form -->
				<div class="bg-base-50 mt-3 space-y-3 rounded-lg border border-base-200 p-3">
					<div>
						<label for="input-name" class="label">
							<span class="label-text text-xs">Name</span>
						</label>
						<input
							id="input-name"
							type="text"
							bind:value={inputName}
							placeholder="e.g., Project Name"
							class="input-bordered input input-sm w-full"
						/>
					</div>

					<div>
						<label for="input-description" class="label">
							<span class="label-text text-xs">Description</span>
						</label>
						<textarea
							id="input-description"
							bind:value={inputDescription}
							placeholder="Describe what this input is for..."
							class="textarea-bordered textarea w-full textarea-sm"
							rows="2"
						></textarea>
					</div>

					<div class="flex items-center gap-2">
						<input
							type="checkbox"
							bind:checked={inputRequired}
							class="checkbox checkbox-sm"
							id="input-required"
						/>
						<label for="input-required" class="label-text cursor-pointer text-xs"> Required </label>
					</div>

					<div class="flex gap-2">
						<button class="btn flex-1 btn-ghost btn-sm" onclick={cancelInputForm}> Cancel </button>
						<button
							class="btn flex-1 btn-sm btn-primary"
							onclick={saveInput}
							disabled={!inputName.trim()}
						>
							{editingInputId ? 'Save Changes' : 'Add Input'}
						</button>
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<!-- Using... Section for Non-Start Nodes -->
		<div class="rounded-lg border border-base-300 bg-base-100 p-4">
			<h3 class="mb-3 text-sm font-semibold text-base-content/70">Using...</h3>

			<!-- Tools -->
			{#if selectedNode.tools && selectedNode.tools.length > 0}
				<div>
					<div class="mb-1 flex items-center gap-2 text-xs font-medium text-base-content/60">
						<Wrench class="h-3.5 w-3.5" />
						<span>Tools</span>
					</div>
					<div class="flex flex-wrap gap-1">
						{#each selectedNode.tools as tool (tool)}
							<span class="badge gap-1 badge-sm badge-warning">
								{tool}
								<button
									onclick={() =>
										itemStore.removeNodeAssignment(taskId, selectedNode.id, 'tools', tool)}
									class="hover:text-error"
									aria-label="Remove tool"
								>
									<X class="h-3 w-3" />
								</button>
							</span>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Agents -->
			{#if selectedNode.agents && selectedNode.agents.length > 0}
				<div>
					<div class="mb-1 flex items-center gap-2 text-xs font-medium text-base-content/60">
						<Bot class="h-3.5 w-3.5" />
						<span>Agents</span>
					</div>
					<div class="flex flex-wrap gap-1">
						{#each selectedNode.agents as agent (agent)}
							<span class="badge gap-1 badge-sm badge-info">
								{getAgentName(agent)}
								<button
									onclick={() =>
										itemStore.removeNodeAssignment(taskId, selectedNode.id, 'agents', agent)}
									class="hover:text-error"
									aria-label="Remove agent"
								>
									<X class="h-3 w-3" />
								</button>
							</span>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Tasks -->
			{#if selectedNode.tasks && selectedNode.tasks.length > 0}
				<div>
					<div class="mb-1 flex items-center gap-2 text-xs font-medium text-base-content/60">
						<ListTodo class="h-3.5 w-3.5" />
						<span>Tasks</span>
					</div>
					<div class="flex flex-wrap gap-1">
						{#each selectedNode.tasks as task (task)}
							<span class="badge gap-1 badge-sm badge-secondary">
								{getTaskName(task)}
								<button
									onclick={() =>
										itemStore.removeNodeAssignment(taskId, selectedNode.id, 'tasks', task)}
									class="hover:text-error"
									aria-label="Remove task"
								>
									<X class="h-3 w-3" />
								</button>
							</span>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Add new assignments -->
			<div class="flex flex-wrap gap-2 border-t border-base-300 pt-3">
				<!-- Add Tool Button/Form -->
				{#if !showAddTool}
					<button
						onclick={() => (showAddTool = true)}
						class="btn gap-1 btn-outline btn-xs btn-warning"
					>
						<Wrench class="h-3 w-3" />
						Add Tool
					</button>
				{:else}
					<div class="flex w-full gap-2">
						<input
							type="text"
							bind:value={newTool}
							placeholder="Tool name..."
							class="input-bordered input input-xs flex-1"
							onkeydown={(e) => {
								if (e.key === 'Enter') addTool();
								if (e.key === 'Escape') {
									showAddTool = false;
									newTool = '';
								}
							}}
						/>
						<button onclick={addTool} class="btn btn-xs btn-warning" disabled={!newTool.trim()}>
							Add
						</button>
						<button
							onclick={() => {
								showAddTool = false;
								newTool = '';
							}}
							class="btn btn-ghost btn-xs"
						>
							<X class="h-3 w-3" />
						</button>
					</div>
				{/if}

				<!-- Add Agent Button/Dropdown -->
				{#if !showAddAgent}
					<button
						onclick={() => (showAddAgent = true)}
						class="btn gap-1 btn-outline btn-xs btn-info"
					>
						<Bot class="h-3 w-3" />
						Add Agent
					</button>
				{:else}
					<div class="flex w-full gap-2">
						<select
							bind:value={newAgent}
							class="select-bordered select flex-1 select-xs"
							onchange={addAgent}
						>
							<option value="">Select agent...</option>
							{#each availableAgents as agent (agent.id)}
								<option value={agent.id}>{agent.title}</option>
							{/each}
						</select>
						<button
							onclick={() => {
								showAddAgent = false;
								newAgent = '';
							}}
							class="btn btn-ghost btn-xs"
						>
							<X class="h-3 w-3" />
						</button>
					</div>
				{/if}

				<!-- Add Task Button/Dropdown -->
				{#if !showAddTask}
					<button
						onclick={() => (showAddTask = true)}
						class="btn gap-1 btn-outline btn-xs btn-secondary"
					>
						<ListTodo class="h-3 w-3" />
						Add Task
					</button>
				{:else}
					<div class="flex w-full gap-2">
						<select
							bind:value={newTask}
							class="select-bordered select flex-1 select-xs"
							onchange={addTask}
						>
							<option value="">Select task...</option>
							{#each availableTasks as task (task.id)}
								<option value={task.id}>{task.title}</option>
							{/each}
						</select>
						<button
							onclick={() => {
								showAddTask = false;
								newTask = '';
							}}
							class="btn btn-ghost btn-xs"
						>
							<X class="h-3 w-3" />
						</button>
					</div>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Section 2: Do the following... -->
	<div class="rounded-lg border border-base-300 bg-base-100 p-4">
		<h3 class="mb-3 text-sm font-semibold text-base-content/70">Do the following...</h3>
		<div class="prose prose-sm max-w-none">
			{@html renderMarkdown(selectedNode.content)}
		</div>
	</div>

	<!-- Section 3: Then... (Next Steps) -->
	<div class="rounded-lg border border-base-300 bg-base-100 p-4">
		<h3 class="mb-3 text-sm font-semibold text-base-content/70">Then...</h3>
		{#if !showAddEdge && flowchart && flowchart.edges.filter((e) => e.source === selectedNode.id).length > 0}
			{@const outgoingEdges = flowchart.edges.filter((e) => e.source === selectedNode.id)}
			<!-- View Mode: Show existing edges -->
			<div class="space-y-2">
				{#each outgoingEdges as edge (edge.id)}
					{@const targetNode = flowchart.nodes.find((n) => n.id === edge.target)}
					{#if targetNode}
						<button
							class="btn btn-block justify-start btn-sm"
							onclick={() => onNavigateToNode(targetNode)}
						>
							<span class="font-normal text-base-content/60">{edge.label || 'Go to'}:</span>
							<span class="font-medium">{targetNode.label}</span>
						</button>
					{/if}
				{/each}
			</div>

			<!-- Edit/Delete buttons -->
			<div class="mt-3 flex gap-2">
				<button class="btn flex-1 btn-outline btn-sm" onclick={startEditEdge}>
					<Pencil class="h-3 w-3" />
					Edit
				</button>
				<button
					class="btn flex-1 btn-outline btn-sm btn-error"
					onclick={async () => {
						if (!flowchart) return;
						const edges = flowchart.edges.filter((e) => e.source === selectedNode.id);
						for (const edge of edges) {
							await itemStore.deleteEdge(taskId, edge.id);
						}
					}}
				>
					<Trash class="h-3 w-3" />
					Delete
				</button>
			</div>
		{:else if !showAddEdge}
			<!-- No edges exist -->
			<p class="mb-3 text-sm text-base-content/60">No next steps defined</p>
			<button class="btn btn-block gap-1 btn-outline btn-sm" onclick={startAddEdge}>
				<span>+</span>
				Add Next Step
			</button>
		{/if}

		{#if showAddEdge}
			<!-- Edge Editor Form -->
			<div class="bg-base-50 mt-4 space-y-3 rounded-lg border border-base-200 p-3">
				<!-- Edge Type Selector -->
				<div class="flex gap-2">
					<button
						class="btn flex-1 btn-sm {edgeType === 'simple' ? 'btn-primary' : 'btn-outline'}"
						onclick={() => (edgeType = 'simple')}
					>
						Simple Next
					</button>
					<button
						class="btn flex-1 btn-sm {edgeType === 'decision' ? 'btn-primary' : 'btn-outline'}"
						onclick={() => (edgeType = 'decision')}
					>
						Decision
					</button>
				</div>

				{#if edgeType === 'simple'}
					<!-- Simple Next Step -->
					<div class="space-y-2">
						<label for="simple-target-select" class="label">
							<span class="label-text">Next node:</span>
						</label>
						<select
							id="simple-target-select"
							bind:value={simpleTarget}
							onchange={() => (simpleTargetIsNew = simpleTarget === 'new')}
							class="select-bordered select w-full select-sm"
						>
							<option value="">Select a node...</option>
							{#each availableTargetNodes as node (node.id)}
								<option value={node.id}>{node.label}</option>
							{/each}
							<option value="new">+ Create New Node</option>
						</select>

						{#if simpleTargetIsNew}
							<div class="space-y-2 rounded border border-base-300 bg-base-100 p-2">
								<input
									type="text"
									bind:value={newSimpleNodeLabel}
									placeholder="New node label..."
									class="input-bordered input input-sm w-full"
								/>
								<select
									bind:value={newSimpleNodeType}
									class="select-bordered select w-full select-sm"
								>
									<option value="process">Process</option>
									<option value="decision">Decision</option>
									<option value="end">End</option>
								</select>
							</div>
						{/if}
					</div>
				{:else}
					<!-- Decision Branches -->
					<div class="space-y-3">
						<div>
							<label for="decision-condition-input" class="label">
								<span class="label-text text-xs font-medium">Condition/Question:</span>
							</label>
							<input
								id="decision-condition-input"
								type="text"
								bind:value={decisionCondition}
								placeholder="e.g., Design approved?"
								class="input-bordered input input-sm w-full"
							/>
						</div>

						<!-- Yes/True Branch -->
						<div class="space-y-2 rounded border border-success/30 bg-success/5 p-2">
							<label for="yes-target-select" class="label py-0">
								<span class="label-text text-xs font-medium text-success">Yes/True →</span>
							</label>
							<select
								id="yes-target-select"
								bind:value={yesTarget}
								onchange={() => (yesTargetIsNew = yesTarget === 'new')}
								class="select-bordered select w-full select-sm"
							>
								<option value="">Select a node...</option>
								{#each availableTargetNodes as node (node.id)}
									<option value={node.id}>{node.label}</option>
								{/each}
								<option value="new">+ Create New Node</option>
							</select>

							{#if yesTargetIsNew}
								<div class="space-y-1">
									<input
										type="text"
										bind:value={newYesNodeLabel}
										placeholder="New node label..."
										class="input-bordered input input-xs w-full"
									/>
									<select
										bind:value={newYesNodeType}
										class="select-bordered select w-full select-xs"
									>
										<option value="process">Process</option>
										<option value="decision">Decision</option>
										<option value="end">End</option>
									</select>
								</div>
							{/if}
						</div>

						<!-- No/False Branch -->
						<div class="space-y-2 rounded border border-error/30 bg-error/5 p-2">
							<label for="no-target-select" class="label py-0">
								<span class="label-text text-xs font-medium text-error">No/False →</span>
							</label>
							<select
								id="no-target-select"
								bind:value={noTarget}
								onchange={() => (noTargetIsNew = noTarget === 'new')}
								class="select-bordered select w-full select-sm"
							>
								<option value="">Select a node...</option>
								{#each availableTargetNodes as node (node.id)}
									<option value={node.id}>{node.label}</option>
								{/each}
								<option value="new">+ Create New Node</option>
							</select>

							{#if noTargetIsNew}
								<div class="space-y-1">
									<input
										type="text"
										bind:value={newNoNodeLabel}
										placeholder="New node label..."
										class="input-bordered input input-xs w-full"
									/>
									<select
										bind:value={newNoNodeType}
										class="select-bordered select w-full select-xs"
									>
										<option value="process">Process</option>
										<option value="decision">Decision</option>
										<option value="end">End</option>
									</select>
								</div>
							{/if}
						</div>
					</div>
				{/if}

				<!-- Action Buttons -->
				<div class="flex gap-2">
					<button class="btn flex-1 btn-ghost btn-sm" onclick={cancelAddEdge}> Cancel </button>
					<button
						class="btn flex-1 btn-sm btn-primary"
						onclick={saveEdge}
						disabled={edgeType === 'simple'
							? !simpleTarget || (simpleTargetIsNew && !newSimpleNodeLabel.trim())
							: (!yesTarget && !noTarget) ||
								(yesTargetIsNew && !newYesNodeLabel.trim()) ||
								(noTargetIsNew && !newNoNodeLabel.trim())}
					>
						{editingEdges ? 'Save Changes' : 'Save'}
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
