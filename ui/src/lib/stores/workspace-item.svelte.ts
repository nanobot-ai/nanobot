import type {
	WorkspaceItem,
	WorkspaceFile,
	FileTreeNode,
	TaskFlowchart,
	Resource,
	Resources,
	ResourceContents
} from '$lib/types';
import { UIPath } from '$lib/types';
import { SvelteSet } from 'svelte/reactivity';
import { SimpleClient } from '$lib/mcpclient';

// MIME types for different workspace resources
const TASK_MIME_TYPE = 'application/vnd.nanobot.task+json';
const AGENT_MIME_TYPE = 'application/vnd.nanobot.agent+json';
const CONVERSATION_MIME_TYPE = 'application/vnd.nanobot.conversation+json';
const FLOWCHART_MIME_TYPE = 'application/vnd.nanobot.flowchart+json';

/**
 * WorkspaceItemStore - Manages items (Tasks, Agents, Conversations, Files) for a single workspace
 *
 * Uses SimpleClient to communicate with workspace-scoped MCP endpoint.
 * All items are represented as MCP resources with workspace:// URIs.
 */
export class WorkspaceItemStore {
	workspaceId: string;
	items = $state<WorkspaceItem[]>([]);
	files = $state<WorkspaceFile[]>([]);
	taskFlowcharts = $state<TaskFlowchart[]>([]);
	expandedSections = $state<SvelteSet<string>>(new SvelteSet());
	expandedFilePaths = $state<SvelteSet<string>>(new SvelteSet());
	isLoading = $state(false);

	private client: SimpleClient;
	private unwatchResources?: () => void;

	constructor(workspaceId: string) {
		this.workspaceId = workspaceId;
		this.client = new SimpleClient({ path: `${UIPath}&workspace=${workspaceId}` });
		this.loadExpandedState();
	}

	/**
	 * Classify a resource into its appropriate type
	 * Returns the classified resource or null if it should be ignored
	 */
	private classifyResource(
		resource: Resource
	): { type: 'item'; data: WorkspaceItem } | { type: 'file'; data: WorkspaceFile } | null {
		// Only process resources with workspace:// prefix
		if (!resource.uri.startsWith('workspace://')) {
			return null;
		}

		switch (resource.mimeType) {
			case TASK_MIME_TYPE: {
				// Extract ID from URI (workspace://tasks/{id} where id can contain /)
				const id = resource.uri.replace('workspace://tasks/', '');
				return {
					type: 'item',
					data: {
						id,
						workspaceId: this.workspaceId,
						type: 'task',
						title: resource.name || 'Untitled Task',
						created: resource.annotations?.lastModified || new Date().toISOString(),
						status: 'active'
					}
				};
			}

			case AGENT_MIME_TYPE: {
				const id = resource.uri.replace('workspace://agents/', '');
				return {
					type: 'item',
					data: {
						id,
						workspaceId: this.workspaceId,
						type: 'agent',
						title: resource.name || 'Untitled Agent',
						created: resource.annotations?.lastModified || new Date().toISOString(),
						status: 'active'
					}
				};
			}

			case CONVERSATION_MIME_TYPE: {
				const id = resource.uri.replace('workspace://conversations/', '');
				return {
					type: 'item',
					data: {
						id,
						workspaceId: this.workspaceId,
						type: 'conversation',
						title: resource.name || 'Untitled Conversation',
						created: resource.annotations?.lastModified || new Date().toISOString(),
						status: 'active'
					}
				};
			}

			case FLOWCHART_MIME_TYPE:
				// Flowcharts are listed but need to be fetched separately to get full content
				// We'll load them on demand when getTaskFlowchart() is called
				return null;

			default:
				// Treat everything else as a file if it has workspace://files/ prefix
				if (resource.uri.startsWith('workspace://files/')) {
					const path = resource.uri.replace('workspace://files/', '');
					// Extract ID from the path (use the full path as ID since it can contain /)
					const id = path;
					return {
						type: 'file',
						data: {
							id,
							workspaceId: this.workspaceId,
							path,
							created: resource.annotations?.lastModified || new Date().toISOString(),
							size: resource.size || 0,
							mimeType: resource.mimeType
						}
					};
				}
				return null;
		}
	}

	/**
	 * Update a single resource in the appropriate in-memory list
	 * Used for live updates from watchResource
	 */
	private async updateSingleResource(resourceUpdate: ResourceContents) {
		// Convert ResourceContents to a Resource-like object for classification
		// If we don't have the name field, we'll need to fetch the full resource
		const uri = resourceUpdate.uri;
		const needsFullFetch = !uri.startsWith('workspace://');

		let fullResource: Resource;

		if (needsFullFetch) {
			// Fetch full resource details
			const result = await this.client.listResources<Resources>({
				prefix: uri
			});
			const found = result.resources?.find((r) => r.uri === uri);
			if (!found) {
				return; // Resource not found or deleted
			}
			fullResource = found;
		} else {
			// We have enough info from the notification
			// Create a minimal Resource object for classification
			fullResource = {
				uri: resourceUpdate.uri,
				mimeType: resourceUpdate.mimeType,
				name: '', // Will be set by classifyResource
				annotations: {
					lastModified: new Date().toISOString()
				}
			} as Resource;
		}

		const classified = this.classifyResource(fullResource);
		if (!classified) {
			return;
		}

		if (classified.type === 'item') {
			const index = this.items.findIndex((i) => i.id === classified.data.id);
			if (index >= 0) {
				// Update existing item, preserving title if we don't have a new one
				if (
					!classified.data.title ||
					classified.data.title === 'Untitled Task' ||
					classified.data.title === 'Untitled Agent' ||
					classified.data.title === 'Untitled Conversation'
				) {
					classified.data.title = this.items[index].title;
				}
				this.items[index] = classified.data;
				this.items = [...this.items];
			} else {
				this.items = [...this.items, classified.data];
			}
		} else if (classified.type === 'file') {
			const index = this.files.findIndex((f) => f.id === classified.data.id);
			if (index >= 0) {
				this.files[index] = classified.data;
				this.files = [...this.files];
			} else {
				this.files = [...this.files, classified.data];
			}
		}
	}

	/**
	 * Load all resources for this workspace from the backend
	 */
	async load() {
		this.isLoading = true;

		try {
			// List all resources with workspace:// prefix
			const result = await this.client.listResources<Resources>({
				prefix: 'workspace://'
			});

			if (!result.resources || result.resources.length === 0) {
				this.items = [];
				this.files = [];
				this.taskFlowcharts = [];
			} else {
				// Classify all resources efficiently in a single pass
				const items: WorkspaceItem[] = [];
				const files: WorkspaceFile[] = [];

				for (const resource of result.resources) {
					const classified = this.classifyResource(resource);
					if (classified) {
						if (classified.type === 'item') {
							items.push(classified.data);
						} else if (classified.type === 'file') {
							files.push(classified.data);
						}
					}
				}

				// Update state once
				this.items = items;
				this.files = files;
				this.taskFlowcharts = [];
			}

			// Start watching for resource changes
			this.unwatchResources = this.client.watchResource('workspace://', (resource) => {
				this.updateSingleResource(resource);
			});
		} catch (error) {
			console.error('Failed to load workspace items:', error);
			this.items = [];
			this.files = [];
			this.taskFlowcharts = [];
		} finally {
			this.isLoading = false;
		}
	}

	/**
	 * Get flowchart for a task
	 */
	getTaskFlowchart(taskId: string): TaskFlowchart | undefined {
		return this.taskFlowcharts.find((fc) => fc.taskId === taskId);
	}

	/**
	 * Toggle node completion status
	 */
	async toggleNodeCompletion(taskId: string, nodeId: string): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const node = flowchart.nodes.find((n) => n.id === nodeId);
			if (node && node.type !== 'start' && node.type !== 'end') {
				node.completed = !node.completed;
				// Update the resource
				await this.updateFlowchart(taskId, flowchart);
			}
		}
	}

	/**
	 * Update flowchart resource
	 */
	private async updateFlowchart(taskId: string, flowchart: TaskFlowchart): Promise<void> {
		await this.client.callMCPTool('update_resource', {
			payload: {
				uri: `workspace://flowcharts/${taskId}`,
				text: JSON.stringify(flowchart)
			}
		});
		// Trigger reactivity
		this.taskFlowcharts = [...this.taskFlowcharts];
	}

	/**
	 * Remove assignment from a node
	 */
	async removeNodeAssignment(
		taskId: string,
		nodeId: string,
		type: 'tools' | 'agents' | 'tasks',
		value: string
	): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const node = flowchart.nodes.find((n) => n.id === nodeId);
			if (node && node[type]) {
				const array = node[type]!;
				const index = array.indexOf(value);
				if (index > -1) {
					array.splice(index, 1);
					await this.updateFlowchart(taskId, flowchart);
				}
			}
		}
	}

	/**
	 * Add assignment to a node
	 */
	async addNodeAssignment(
		taskId: string,
		nodeId: string,
		type: 'tools' | 'agents' | 'tasks',
		value: string
	): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const node = flowchart.nodes.find((n) => n.id === nodeId);
			if (node) {
				if (!node[type]) {
					node[type] = [];
				}
				if (!node[type]!.includes(value)) {
					node[type]!.push(value);
					await this.updateFlowchart(taskId, flowchart);
				}
			}
		}
	}

	/**
	 * Add edge between nodes
	 */
	async addEdge(taskId: string, sourceId: string, targetId: string, label?: string): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const newEdge = {
				id: `e${Date.now()}`,
				source: sourceId,
				target: targetId,
				label
			};
			flowchart.edges.push(newEdge);
			await this.updateFlowchart(taskId, flowchart);
		}
	}

	/**
	 * Create new node
	 */
	async createNode(
		taskId: string,
		type: 'start' | 'process' | 'decision' | 'end',
		label: string,
		content: string,
		sourceNodeId?: string
	): Promise<string> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const newNodeId = `node-${Date.now()}`;

			// Calculate position: below source node if provided, else at bottom
			let position = { x: 250, y: 100 };
			if (sourceNodeId) {
				const sourceNode = flowchart.nodes.find((n) => n.id === sourceNodeId);
				if (sourceNode) {
					position = { x: sourceNode.position.x, y: sourceNode.position.y + 150 };
				}
			} else {
				// Place at bottom
				const maxY = Math.max(...flowchart.nodes.map((n) => n.position.y));
				position = { x: 250, y: maxY + 150 };
			}

			const newNode = {
				id: newNodeId,
				type,
				label,
				content,
				position
			};

			flowchart.nodes.push(newNode);
			await this.updateFlowchart(taskId, flowchart);
			return newNodeId;
		}
		throw new Error('Task flowchart not found');
	}

	/**
	 * Update edge
	 */
	async updateEdge(
		taskId: string,
		edgeId: string,
		updates: { source?: string; target?: string; label?: string }
	): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const edge = flowchart.edges.find((e) => e.id === edgeId);
			if (edge) {
				Object.assign(edge, updates);
				await this.updateFlowchart(taskId, flowchart);
			}
		}
	}

	/**
	 * Delete edge
	 */
	async deleteEdge(taskId: string, edgeId: string): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			flowchart.edges = flowchart.edges.filter((e) => e.id !== edgeId);
			await this.updateFlowchart(taskId, flowchart);
		}
	}

	/**
	 * Add input to a node
	 */
	async addNodeInput(
		taskId: string,
		nodeId: string,
		name: string,
		description: string,
		required: boolean
	): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const node = flowchart.nodes.find((n) => n.id === nodeId);
			if (node) {
				if (!node.inputs) {
					node.inputs = [];
				}
				const newInput = {
					id: `input-${Date.now()}`,
					name,
					description,
					required
				};
				node.inputs.push(newInput);
				await this.updateFlowchart(taskId, flowchart);
			}
		}
	}

	/**
	 * Update node input
	 */
	async updateNodeInput(
		taskId: string,
		nodeId: string,
		inputId: string,
		updates: { name?: string; description?: string; required?: boolean }
	): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const node = flowchart.nodes.find((n) => n.id === nodeId);
			if (node && node.inputs) {
				const input = node.inputs.find((i) => i.id === inputId);
				if (input) {
					Object.assign(input, updates);
					await this.updateFlowchart(taskId, flowchart);
				}
			}
		}
	}

	/**
	 * Delete node input
	 */
	async deleteNodeInput(taskId: string, nodeId: string, inputId: string): Promise<void> {
		const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
		if (flowchart) {
			const node = flowchart.nodes.find((n) => n.id === nodeId);
			if (node && node.inputs) {
				node.inputs = node.inputs.filter((i) => i.id !== inputId);
				await this.updateFlowchart(taskId, flowchart);
			}
		}
	}

	/**
	 * Get items by type
	 */
	getItems(type?: 'task' | 'agent' | 'conversation'): WorkspaceItem[] {
		return this.items.filter((item) => !type || item.type === type);
	}

	/**
	 * Get count of items by type
	 */
	getItemCount(type: 'task' | 'agent' | 'conversation'): number {
		return this.items.filter((item) => item.type === type).length;
	}

	/**
	 * Get files
	 */
	getFiles(): WorkspaceFile[] {
		return this.files;
	}

	/**
	 * Get file count
	 */
	getFileCount(): number {
		return this.files.length;
	}

	/**
	 * Build file tree from flat file list
	 */
	buildFileTree(): FileTreeNode[] {
		const files = this.getFiles();
		const root: FileTreeNode[] = [];

		for (const file of files) {
			const parts = file.path.split('/');
			let currentLevel = root;

			for (let i = 0; i < parts.length; i++) {
				const part = parts[i];
				const isLastPart = i === parts.length - 1;
				const fullPath = parts.slice(0, i + 1).join('/');

				let existing = currentLevel.find((node) => node.name === part);

				if (!existing) {
					const node: FileTreeNode = {
						name: part,
						path: fullPath,
						isDirectory: !isLastPart,
						file: isLastPart ? file : undefined
					};

					if (!isLastPart) {
						node.children = [];
					}

					currentLevel.push(node);
					existing = node;
				}

				if (!isLastPart && existing.children) {
					currentLevel = existing.children;
				}
			}
		}

		// Sort: directories first, then files, alphabetically
		const sortNodes = (nodes: FileTreeNode[]) => {
			nodes.sort((a, b) => {
				if (a.isDirectory && !b.isDirectory) return -1;
				if (!a.isDirectory && b.isDirectory) return 1;
				return a.name.localeCompare(b.name);
			});
			nodes.forEach((node) => {
				if (node.children) {
					sortNodes(node.children);
				}
			});
		};

		sortNodes(root);
		return root;
	}

	/**
	 * Toggle file path expanded state
	 */
	toggleFilePath(path: string) {
		if (this.expandedFilePaths.has(path)) {
			this.expandedFilePaths.delete(path);
		} else {
			this.expandedFilePaths.add(path);
		}
		this.saveExpandedState();
	}

	/**
	 * Check if file path is expanded
	 */
	isFilePathExpanded(path: string): boolean {
		return this.expandedFilePaths.has(path);
	}

	/**
	 * Toggle section (tasks/agents/conversations) expanded state
	 */
	toggleSection(sectionType: string) {
		if (this.expandedSections.has(sectionType)) {
			this.expandedSections.delete(sectionType);
		} else {
			this.expandedSections.add(sectionType);
		}
		this.saveExpandedState();
	}

	/**
	 * Check if section is expanded
	 */
	isSectionExpanded(sectionType: string): boolean {
		return this.expandedSections.has(sectionType);
	}

	/**
	 * Create new item (task/agent/conversation)
	 */
	async createItem(type: 'task' | 'agent' | 'conversation', title: string): Promise<WorkspaceItem> {
		const id = `${type}-${Date.now()}`;
		let mimeType: string;
		let uri: string;

		switch (type) {
			case 'task':
				mimeType = TASK_MIME_TYPE;
				uri = `workspace://tasks/${id}`;
				break;
			case 'agent':
				mimeType = AGENT_MIME_TYPE;
				uri = `workspace://agents/${id}`;
				break;
			case 'conversation':
				mimeType = CONVERSATION_MIME_TYPE;
				uri = `workspace://conversations/${id}`;
				break;
		}

		await this.client.callMCPTool('create_resource', {
			payload: {
				uri,
				name: title,
				mimeType
			}
		});

		const newItem: WorkspaceItem = {
			id,
			workspaceId: this.workspaceId,
			type,
			title,
			created: new Date().toISOString(),
			status: 'active'
		};

		this.items = [...this.items, newItem];
		return newItem;
	}

	/**
	 * Update item
	 */
	async updateItem(itemId: string, data: Partial<WorkspaceItem>): Promise<WorkspaceItem> {
		const index = this.items.findIndex((item) => item.id === itemId);
		if (index === -1) {
			throw new Error('Item not found');
		}

		const item = this.items[index];
		let uri: string;

		switch (item.type) {
			case 'task':
				uri = `workspace://tasks/${itemId}`;
				break;
			case 'agent':
				uri = `workspace://agents/${itemId}`;
				break;
			case 'conversation':
				uri = `workspace://conversations/${itemId}`;
				break;
		}

		await this.client.callMCPTool('update_resource', {
			payload: {
				uri,
				...(data.title && { name: data.title })
			}
		});

		const updatedItem = { ...item, ...data };
		this.items = [...this.items.slice(0, index), updatedItem, ...this.items.slice(index + 1)];
		return updatedItem;
	}

	/**
	 * Delete item
	 */
	async deleteItem(itemId: string): Promise<void> {
		const item = this.items.find((i) => i.id === itemId);
		if (!item) {
			throw new Error('Item not found');
		}

		let uri: string;
		switch (item.type) {
			case 'task':
				uri = `workspace://tasks/${itemId}`;
				break;
			case 'agent':
				uri = `workspace://agents/${itemId}`;
				break;
			case 'conversation':
				uri = `workspace://conversations/${itemId}`;
				break;
		}

		await this.client.callMCPTool('delete_resource', {
			payload: { uri }
		});

		this.items = this.items.filter((i) => i.id !== itemId);
	}

	/**
	 * Close and cleanup resources for this store
	 * Called when the workspace is deleted or the store is being removed from cache
	 */
	close() {
		// Stop watching for resource changes
		if (this.unwatchResources) {
			this.unwatchResources();
			this.unwatchResources = undefined;
		}

		// Save any pending state changes
		this.saveExpandedState();

		// Clear all reactive state
		this.items = [];
		this.files = [];
		this.taskFlowcharts = [];
		this.expandedSections.clear();
		this.expandedFilePaths.clear();
		this.isLoading = false;
	}

	/**
	 * Load expanded state from localStorage
	 */
	private loadExpandedState() {
		try {
			const sectionState = localStorage.getItem(`nanobot-expanded-sections-${this.workspaceId}`);
			if (sectionState) {
				const sections = JSON.parse(sectionState) as string[];
				sections.forEach((section) => this.expandedSections.add(section));
			}

			const filePathState = localStorage.getItem(`nanobot-expanded-file-paths-${this.workspaceId}`);
			if (filePathState) {
				const paths = JSON.parse(filePathState) as string[];
				paths.forEach((path) => this.expandedFilePaths.add(path));
			}
		} catch (error) {
			console.error('Failed to load expanded state:', error);
		}
	}

	/**
	 * Save expanded state to localStorage
	 */
	private saveExpandedState() {
		try {
			localStorage.setItem(
				`nanobot-expanded-sections-${this.workspaceId}`,
				JSON.stringify([...this.expandedSections])
			);

			localStorage.setItem(
				`nanobot-expanded-file-paths-${this.workspaceId}`,
				JSON.stringify([...this.expandedFilePaths])
			);
		} catch (error) {
			console.error('Failed to save expanded state:', error);
		}
	}
}
