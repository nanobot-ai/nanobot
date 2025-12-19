import "clsx";
import { S as SvelteSet, a as SimpleClient, U as UIPath, b as SvelteDate, W as WorkspaceMimeType } from "./mcpclient.js";
const TASK_MIME_TYPE = "application/vnd.nanobot.task+json";
const AGENT_MIME_TYPE = "application/vnd.nanobot.agent+json";
const CONVERSATION_MIME_TYPE = "application/vnd.nanobot.conversation+json";
const FLOWCHART_MIME_TYPE = "application/vnd.nanobot.flowchart+json";
class WorkspaceItemStore {
  workspaceId;
  items = [];
  files = [];
  taskFlowcharts = [];
  expandedSections = new SvelteSet();
  expandedFilePaths = new SvelteSet();
  isLoading = false;
  client;
  unwatchResources;
  constructor(workspaceId) {
    this.workspaceId = workspaceId;
    this.client = new SimpleClient({ path: `${UIPath}&workspace=${workspaceId}` });
    this.loadExpandedState();
  }
  /**
   * Classify a resource into its appropriate type
   * Returns the classified resource or null if it should be ignored
   */
  classifyResource(resource) {
    if (!resource.uri.startsWith("workspace://")) {
      return null;
    }
    const timestamp = resource.annotations?.lastModified || new SvelteDate().toISOString();
    switch (resource.mimeType) {
      case TASK_MIME_TYPE: {
        const id = resource.uri.replace("workspace://tasks/", "");
        return {
          type: "item",
          data: {
            id,
            workspaceId: this.workspaceId,
            type: "task",
            title: resource.name || "Untitled Task",
            created: timestamp,
            status: "active"
          }
        };
      }
      case AGENT_MIME_TYPE: {
        const id = resource.uri.replace("workspace://agents/", "");
        return {
          type: "item",
          data: {
            id,
            workspaceId: this.workspaceId,
            type: "agent",
            title: resource.name || "Untitled Agent",
            created: timestamp,
            status: "active"
          }
        };
      }
      case CONVERSATION_MIME_TYPE: {
        const id = resource.uri.replace("workspace://conversations/", "");
        return {
          type: "item",
          data: {
            id,
            workspaceId: this.workspaceId,
            type: "conversation",
            title: resource.name || "Untitled Conversation",
            created: timestamp,
            status: "active"
          }
        };
      }
      case FLOWCHART_MIME_TYPE:
        return null;
      default:
        if (resource.uri.startsWith("workspace://files/")) {
          const path = resource.uri.replace("workspace://files/", "");
          const id = path;
          return {
            type: "file",
            data: {
              id,
              workspaceId: this.workspaceId,
              path,
              created: timestamp,
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
  async updateSingleResource(resourceUpdate) {
    const uri = resourceUpdate.uri;
    const needsFullFetch = !uri.startsWith("workspace://");
    let fullResource;
    if (needsFullFetch) {
      const result = await this.client.listResources({ prefix: uri });
      const found = result.resources?.find((r) => r.uri === uri);
      if (!found) {
        return;
      }
      fullResource = found;
    } else {
      const timestamp = new SvelteDate().toISOString();
      fullResource = {
        uri: resourceUpdate.uri,
        mimeType: resourceUpdate.mimeType,
        name: "",
        // Will be set by classifyResource
        annotations: { lastModified: timestamp }
      };
    }
    const classified = this.classifyResource(fullResource);
    if (!classified) {
      return;
    }
    if (classified.type === "item") {
      const index = this.items.findIndex((i) => i.id === classified.data.id);
      if (index >= 0) {
        if (!classified.data.title || classified.data.title === "Untitled Task" || classified.data.title === "Untitled Agent" || classified.data.title === "Untitled Conversation") {
          classified.data.title = this.items[index].title;
        }
        this.items[index] = classified.data;
        this.items = [...this.items];
      } else {
        this.items = [...this.items, classified.data];
      }
    } else if (classified.type === "file") {
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
      const result = await this.client.listResources({ prefix: "workspace://" });
      if (!result.resources || result.resources.length === 0) {
        this.items = [];
        this.files = [];
        this.taskFlowcharts = [];
      } else {
        const items = [];
        const files = [];
        for (const resource of result.resources) {
          const classified = this.classifyResource(resource);
          if (classified) {
            if (classified.type === "item") {
              items.push(classified.data);
            } else if (classified.type === "file") {
              files.push(classified.data);
            }
          }
        }
        this.items = items;
        this.files = files;
        this.taskFlowcharts = [];
      }
      this.unwatchResources = this.client.watchResource("workspace://", (resource) => {
        this.updateSingleResource(resource);
      });
    } catch (error) {
      console.error("Failed to load workspace items:", error);
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
  getTaskFlowchart(taskId) {
    return this.taskFlowcharts.find((fc) => fc.taskId === taskId);
  }
  /**
   * Toggle node completion status
   */
  async toggleNodeCompletion(taskId, nodeId) {
    const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
    if (flowchart) {
      const node = flowchart.nodes.find((n) => n.id === nodeId);
      if (node && node.type !== "start" && node.type !== "end") {
        node.completed = !node.completed;
        await this.updateFlowchart(taskId, flowchart);
      }
    }
  }
  /**
   * Update flowchart resource
   */
  async updateFlowchart(taskId, flowchart) {
    await this.client.callMCPTool("update_resource", {
      payload: {
        uri: `workspace://flowcharts/${taskId}`,
        text: JSON.stringify(flowchart)
      }
    });
    this.taskFlowcharts = [...this.taskFlowcharts];
  }
  /**
   * Remove assignment from a node
   */
  async removeNodeAssignment(taskId, nodeId, type, value) {
    const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
    if (flowchart) {
      const node = flowchart.nodes.find((n) => n.id === nodeId);
      if (node && node[type]) {
        const array = node[type];
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
  async addNodeAssignment(taskId, nodeId, type, value) {
    const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
    if (flowchart) {
      const node = flowchart.nodes.find((n) => n.id === nodeId);
      if (node) {
        if (!node[type]) {
          node[type] = [];
        }
        if (!node[type].includes(value)) {
          node[type].push(value);
          await this.updateFlowchart(taskId, flowchart);
        }
      }
    }
  }
  /**
   * Add edge between nodes
   */
  async addEdge(taskId, sourceId, targetId, label) {
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
  async createNode(taskId, type, label, content, sourceNodeId) {
    const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
    if (flowchart) {
      const newNodeId = `node-${Date.now()}`;
      let position = { x: 250, y: 100 };
      if (sourceNodeId) {
        const sourceNode = flowchart.nodes.find((n) => n.id === sourceNodeId);
        if (sourceNode) {
          position = { x: sourceNode.position.x, y: sourceNode.position.y + 150 };
        }
      } else {
        const maxY = Math.max(...flowchart.nodes.map((n) => n.position.y));
        position = { x: 250, y: maxY + 150 };
      }
      const newNode = { id: newNodeId, type, label, content, position };
      flowchart.nodes.push(newNode);
      await this.updateFlowchart(taskId, flowchart);
      return newNodeId;
    }
    throw new Error("Task flowchart not found");
  }
  /**
   * Update edge
   */
  async updateEdge(taskId, edgeId, updates) {
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
  async deleteEdge(taskId, edgeId) {
    const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
    if (flowchart) {
      flowchart.edges = flowchart.edges.filter((e) => e.id !== edgeId);
      await this.updateFlowchart(taskId, flowchart);
    }
  }
  /**
   * Add input to a node
   */
  async addNodeInput(taskId, nodeId, name, description, required) {
    const flowchart = this.taskFlowcharts.find((fc) => fc.taskId === taskId);
    if (flowchart) {
      const node = flowchart.nodes.find((n) => n.id === nodeId);
      if (node) {
        if (!node.inputs) {
          node.inputs = [];
        }
        const newInput = { id: `input-${Date.now()}`, name, description, required };
        node.inputs.push(newInput);
        await this.updateFlowchart(taskId, flowchart);
      }
    }
  }
  /**
   * Update node input
   */
  async updateNodeInput(taskId, nodeId, inputId, updates) {
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
  async deleteNodeInput(taskId, nodeId, inputId) {
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
  getItems(type) {
    return this.items.filter((item) => !type || item.type === type);
  }
  /**
   * Get count of items by type
   */
  getItemCount(type) {
    return this.items.filter((item) => item.type === type).length;
  }
  /**
   * Get files
   */
  getFiles() {
    return this.files;
  }
  /**
   * Get file count
   */
  getFileCount() {
    return this.files.length;
  }
  /**
   * Build file tree from flat file list
   */
  buildFileTree() {
    const files = this.getFiles();
    const root = [];
    for (const file of files) {
      const parts = file.path.split("/");
      let currentLevel = root;
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i];
        const isLastPart = i === parts.length - 1;
        const fullPath = parts.slice(0, i + 1).join("/");
        let existing = currentLevel.find((node) => node.name === part);
        if (!existing) {
          const node = {
            name: part,
            path: fullPath,
            isDirectory: !isLastPart,
            file: isLastPart ? file : void 0
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
    const sortNodes = (nodes) => {
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
  toggleFilePath(path) {
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
  isFilePathExpanded(path) {
    return this.expandedFilePaths.has(path);
  }
  /**
   * Toggle section (tasks/agents/conversations) expanded state
   */
  toggleSection(sectionType) {
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
  isSectionExpanded(sectionType) {
    return this.expandedSections.has(sectionType);
  }
  /**
   * Create new item (task/agent/conversation)
   */
  async createItem(type, title) {
    const id = `${type}-${Date.now()}`;
    let mimeType;
    let uri;
    switch (type) {
      case "task":
        mimeType = TASK_MIME_TYPE;
        uri = `workspace://tasks/${id}`;
        break;
      case "agent":
        mimeType = AGENT_MIME_TYPE;
        uri = `workspace://agents/${id}`;
        break;
      case "conversation":
        mimeType = CONVERSATION_MIME_TYPE;
        uri = `workspace://conversations/${id}`;
        break;
    }
    await this.client.callMCPTool("create_resource", { payload: { uri, name: title, mimeType } });
    const timestamp = new SvelteDate().toISOString();
    const newItem = {
      id,
      workspaceId: this.workspaceId,
      type,
      title,
      created: timestamp,
      status: "active"
    };
    this.items = [...this.items, newItem];
    return newItem;
  }
  /**
   * Update item
   */
  async updateItem(itemId, data) {
    const index = this.items.findIndex((item2) => item2.id === itemId);
    if (index === -1) {
      throw new Error("Item not found");
    }
    const item = this.items[index];
    let uri;
    switch (item.type) {
      case "task":
        uri = `workspace://tasks/${itemId}`;
        break;
      case "agent":
        uri = `workspace://agents/${itemId}`;
        break;
      case "conversation":
        uri = `workspace://conversations/${itemId}`;
        break;
    }
    await this.client.callMCPTool("update_resource", { payload: { uri, ...data.title && { name: data.title } } });
    const updatedItem = { ...item, ...data };
    this.items = [
      ...this.items.slice(0, index),
      updatedItem,
      ...this.items.slice(index + 1)
    ];
    return updatedItem;
  }
  /**
   * Delete item
   */
  async deleteItem(itemId) {
    const item = this.items.find((i) => i.id === itemId);
    if (!item) {
      throw new Error("Item not found");
    }
    let uri;
    switch (item.type) {
      case "task":
        uri = `workspace://tasks/${itemId}`;
        break;
      case "agent":
        uri = `workspace://agents/${itemId}`;
        break;
      case "conversation":
        uri = `workspace://conversations/${itemId}`;
        break;
    }
    await this.client.callMCPTool("delete_resource", { payload: { uri } });
    this.items = this.items.filter((i) => i.id !== itemId);
  }
  /**
   * Close and cleanup resources for this store
   * Called when the workspace is deleted or the store is being removed from cache
   */
  close() {
    if (this.unwatchResources) {
      this.unwatchResources();
      this.unwatchResources = void 0;
    }
    this.saveExpandedState();
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
  loadExpandedState() {
    try {
      const sectionState = localStorage.getItem(`nanobot-expanded-sections-${this.workspaceId}`);
      if (sectionState) {
        const sections = JSON.parse(sectionState);
        sections.forEach((section) => this.expandedSections.add(section));
      }
      const filePathState = localStorage.getItem(`nanobot-expanded-file-paths-${this.workspaceId}`);
      if (filePathState) {
        const paths = JSON.parse(filePathState);
        paths.forEach((path) => this.expandedFilePaths.add(path));
      }
    } catch (error) {
      console.error("Failed to load expanded state:", error);
    }
  }
  /**
   * Save expanded state to localStorage
   */
  saveExpandedState() {
    try {
      localStorage.setItem(`nanobot-expanded-sections-${this.workspaceId}`, JSON.stringify([...this.expandedSections]));
      localStorage.setItem(`nanobot-expanded-file-paths-${this.workspaceId}`, JSON.stringify([...this.expandedFilePaths]));
    } catch (error) {
      console.error("Failed to save expanded state:", error);
    }
  }
}
class WorkspaceService {
  client;
  constructor(opts) {
    this.client = opts?.client || new SimpleClient();
  }
  /**
   * List all workspaces by querying resources with nanobot://workspaces/{id} URIs
   * Note: resources/list only returns metadata, not content
   */
  async listWorkspaces() {
    const result = await this.client.listResources({ prefix: "nanobot://workspaces/" });
    if (!result.resources) {
      return [];
    }
    const workspaces = [];
    for (const resource of result.resources) {
      if (resource.uri?.startsWith("nanobot://workspaces/") && resource.mimeType === WorkspaceMimeType) {
        try {
          const id = resource.uri.split("/").pop() || "";
          const workspace = {
            id,
            name: resource.name,
            created: resource.annotations?.lastModified || "",
            icons: resource.icons
          };
          const nanobotMeta = resource._meta?.["ai.nanobot"];
          if (nanobotMeta) {
            if (typeof nanobotMeta.order === "number") {
              workspace.order = nanobotMeta.order;
            }
            if (typeof nanobotMeta.color === "string") {
              workspace.color = nanobotMeta.color;
            }
          }
          workspaces.push(workspace);
        } catch (error) {
          console.error("Failed to parse workspace resource:", error);
        }
      }
    }
    return workspaces;
  }
  /**
   * Create a new workspace
   */
  async createWorkspace(workspace) {
    return await this.client.callMCPTool("create_workspace", { payload: workspace });
  }
  /**
   * Update an existing workspace
   */
  async updateWorkspace(workspaceId, updates) {
    return await this.client.callMCPTool("update_workspace", {
      payload: { uri: `nanobot://workspaces/${workspaceId}`, ...updates }
    });
  }
  /**
   * Delete a workspace
   */
  async deleteWorkspace(workspaceId) {
    await this.client.callMCPTool("delete_workspace", { payload: { uri: `nanobot://workspaces/${workspaceId}` } });
  }
}
new WorkspaceService();
class WorkspaceStore {
  workspaces = [];
  expandedWorkspaceIds = new SvelteSet();
  isLoading = false;
  // Cache of workspace item stores by workspace ID
  // Using regular Map (not SvelteMap) because we don't need this cache to be reactive
  // and we don't want mutations during derived computations
  itemStoreCache = /* @__PURE__ */ new Map();
  service;
  constructor(opts) {
    this.service = opts?.service || new WorkspaceService();
    this.loadExpandedState();
  }
  /**
   * Get or create a WorkspaceItemStore for a specific workspace
   * Uses untrack to allow store creation during template evaluation without triggering
   * state_unsafe_mutation errors
   */
  getItemStore(workspaceId) {
    let store = this.itemStoreCache.get(workspaceId);
    if (!store) {
      store = new WorkspaceItemStore(workspaceId);
      this.itemStoreCache.set(workspaceId, store);
    }
    return store;
  }
  /**
   * Clear cached item store for a workspace (useful when workspace is deleted)
   * Calls close() on the store to cleanup resources before removing from cache
   */
  clearItemStore(workspaceId) {
    const store = this.itemStoreCache.get(workspaceId);
    if (store) {
      store.close();
      this.itemStoreCache.delete(workspaceId);
    }
  }
  /**
   * Load workspaces from backend via WorkspaceService
   */
  async load() {
    this.isLoading = true;
    try {
      this.workspaces = await this.service.listWorkspaces();
    } catch (error) {
      console.error("Failed to load workspaces:", error);
      this.workspaces = [];
    } finally {
      this.isLoading = false;
    }
  }
  /**
   * Toggle workspace expanded state
   */
  toggleWorkspace(workspaceId) {
    if (this.expandedWorkspaceIds.has(workspaceId)) {
      this.expandedWorkspaceIds.delete(workspaceId);
    } else {
      this.expandedWorkspaceIds.add(workspaceId);
    }
    this.saveExpandedState();
  }
  /**
   * Check if workspace is expanded
   */
  isWorkspaceExpanded(workspaceId) {
    return this.expandedWorkspaceIds.has(workspaceId);
  }
  /**
   * Create new workspace via WorkspaceService
   */
  async createWorkspace(name) {
    const newWorkspace = await this.service.createWorkspace({ name, order: this.workspaces.length });
    this.workspaces = [...this.workspaces, newWorkspace];
    return newWorkspace;
  }
  /**
   * Update workspace via WorkspaceService
   */
  async updateWorkspace(workspaceId, data) {
    const index = this.workspaces.findIndex((w) => w.id === workspaceId);
    if (index === -1) {
      throw new Error("Workspace not found");
    }
    const { id, created, ...updates } = data;
    const updatedWorkspace = await this.service.updateWorkspace(workspaceId, updates);
    this.workspaces = [
      ...this.workspaces.slice(0, index),
      updatedWorkspace,
      ...this.workspaces.slice(index + 1)
    ];
    return updatedWorkspace;
  }
  /**
   * Delete workspace via WorkspaceService
   */
  async deleteWorkspace(workspaceId) {
    await this.service.deleteWorkspace(workspaceId);
    this.workspaces = this.workspaces.filter((w) => w.id !== workspaceId);
    this.clearItemStore(workspaceId);
    this.expandedWorkspaceIds.delete(workspaceId);
    this.saveExpandedState();
  }
  /**
   * Load expanded state from localStorage
   */
  loadExpandedState() {
    try {
      const workspaceState = localStorage.getItem("nanobot-expanded-workspaces");
      if (workspaceState) {
        const ids = JSON.parse(workspaceState);
        ids.forEach((id) => this.expandedWorkspaceIds.add(id));
      }
    } catch (error) {
      console.error("Failed to load expanded state:", error);
    }
  }
  /**
   * Save expanded state to localStorage
   */
  saveExpandedState() {
    try {
      localStorage.setItem("nanobot-expanded-workspaces", JSON.stringify([...this.expandedWorkspaceIds]));
    } catch (error) {
      console.error("Failed to save expanded state:", error);
    }
  }
}
const workspaceStore = new WorkspaceStore();
export {
  workspaceStore as w
};
