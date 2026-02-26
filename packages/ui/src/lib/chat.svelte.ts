import { SvelteDate } from "svelte/reactivity";
import { getNotificationContext } from "./context/notifications.svelte";
import { SimpleClient } from "./mcpclient";
import {
	AgentResourcePrefix,
	ChatUIPath,
	ChatThreadResourcePrefix,
	UIPath,
} from "./types";
import type {
	Agent,
	Agents,
	Attachment,
	Chat,
	ChatMessage,
	ChatRequest,
	ChatResult,
	Elicitation,
	ElicitationResult,
	Event,
	Prompt,
	Prompts,
	Resource,
	Resources,
	ToolOutputItem,
	UploadedFile,
	UploadingFile,
} from "./types";

export interface CallToolResult {
	content?: ToolOutputItem[];
}

// Resource metadata can come from different servers with inconsistent typing.
// These helpers normalize unknown values before mapping them into UI types.
function asString(value: unknown): string | undefined {
	return typeof value === "string" && value.trim() ? value : undefined;
}

function asBoolean(value: unknown): boolean | undefined {
	return typeof value === "boolean" ? value : undefined;
}

function asStringArray(value: unknown): string[] | undefined {
	if (!Array.isArray(value)) {
		return undefined;
	}

	const values = value.filter((item): item is string => typeof item === "string");
	return values.length > 0 ? values : undefined;
}

function parseResourceID(uri: string, prefix: string): string | undefined {
	if (!uri.startsWith(prefix)) {
		return undefined;
	}
	const raw = uri.slice(prefix.length);
	if (!raw) {
		return undefined;
	}

	try {
		return decodeURIComponent(raw);
	} catch {
		return raw;
	}
}

// Convert agent resources from the meta MCP server into stable UI agent objects.
function mapAgentResource(resource: Resource): Agent {
	const meta = resource._meta ?? {};
	const id =
		asString(meta["id"]) ??
		parseResourceID(resource.uri, AgentResourcePrefix) ??
		resource.uri;

	return {
		id,
		name: asString(meta["name"]) ?? resource.title ?? resource.name ?? id,
		description: asString(meta["description"]) ?? resource.description,
		icon: asString(meta["icon"]),
		iconDark: asString(meta["iconDark"]),
		starterMessages: asStringArray(meta["starterMessages"]),
		current: asBoolean(meta["current"]) ?? false,
	};
}

// Convert chat thread resources into sidebar-friendly chat metadata.
function mapChatResource(resource: Resource): Chat {
	const meta = resource._meta ?? {};
	const id =
		asString(meta["id"]) ??
		parseResourceID(resource.uri, ChatThreadResourcePrefix) ??
		resource.uri;
	const created =
		asString(meta["created"]) ??
		asString(meta["sessionCreatedAt"]) ??
		new SvelteDate().toISOString();
	const visibility = asString(meta["visibility"]);
	const resourceName = asString(resource.name);
	const titleCandidate =
		asString(meta["title"]) ??
		resource.title ??
		resource.description ??
		(resourceName && resourceName !== id ? resourceName : undefined);

	return {
		id,
		title: titleCandidate ?? "Untitled",
		created,
		visibility: visibility === "public" || visibility === "private" ? visibility : undefined,
		readonly: asBoolean(meta["readonly"]),
		currentAgentId: asString(meta["currentAgentId"]),
		availableAgentIds: asStringArray(meta["availableAgentIds"]),
		workflowURIs: asStringArray(meta["workflowURIs"]),
	};
}

export class ChatAPI {
	private readonly baseUrl: string;
	private readonly fetcher: typeof fetch;
	private readonly chatClient: SimpleClient;
	private readonly metaClient: SimpleClient;

	constructor(
		baseUrl: string = "",
		opts?: {
			fetcher?: typeof fetch;
			sessionId?: string;
		},
	) {
		this.baseUrl = baseUrl;
		this.fetcher = opts?.fetcher || fetch;
		this.chatClient = new SimpleClient({
			baseUrl: baseUrl,
			path: ChatUIPath,
			fetcher: this.fetcher,
			sessionId: opts?.sessionId,
		});
		this.metaClient = new SimpleClient({
			baseUrl: baseUrl,
			path: UIPath,
			fetcher: this.fetcher,
		});
	}

	#getChatClient(sessionId?: string) {
		if (sessionId) {
			// Session-scoped calls use an ephemeral client to avoid mutating shared client state.
			return new SimpleClient({
				baseUrl: this.baseUrl,
				path: ChatUIPath,
				fetcher: this.fetcher,
				sessionId,
			});
		}
		return this.chatClient;
	}

	#getMetaClient(sessionId?: string) {
		if (sessionId) {
			// Meta queries can also be session-scoped when we need per-session authorization.
			return new SimpleClient({
				baseUrl: this.baseUrl,
				path: UIPath,
				fetcher: this.fetcher,
				sessionId,
			});
		}
		return this.metaClient;
	}

	async #callMCPToolWithClient<T>(
		client: SimpleClient,
		name: string,
		opts?: {
			payload?: Record<string, unknown>;
			progressToken?: string;
			async?: boolean;
			abort?: AbortController;
			requestId?: string;
			parseResponse?: (data: CallToolResult) => T;
		},
	): Promise<{ result: T; requestId: string }> {
		const { result, requestId } = await client.exchange(
			"tools/call",
			{
				name: name,
				arguments: opts?.payload || {},
				...(opts?.async && {
					_meta: {
						"ai.nanobot.async": true,
						progressToken: opts?.progressToken,
					},
				}),
			},
			{ abort: opts?.abort, requestId: opts?.requestId },
		);

		// Most MCP tool calls return structuredContent, but some tools return raw payloads.
		let finalResult: T;
		if (opts?.parseResponse) {
			finalResult = opts.parseResponse(result as CallToolResult);
		} else if (
			result &&
			typeof result === "object" &&
			"structuredContent" in result
		) {
			finalResult = (result as { structuredContent: T }).structuredContent;
		} else {
			finalResult = result as T;
		}

		return { result: finalResult, requestId };
	}

	async reply(
		id: string | number,
		result: unknown,
		opts?: { sessionId?: string },
	) {
		// If sessionId is provided, create a new client instance with that session
		const client = this.#getChatClient(opts?.sessionId);
		await client.reply(id, result);
	}

	async exchange(
		method: string,
		params: unknown,
		opts?: { sessionId?: string },
	) {
		// If sessionId is provided, create a new client instance with that session
		const client = this.#getChatClient(opts?.sessionId);
		const { result } = await client.exchange(method, params);
		return result;
	}

	async callMCPTool<T>(
		name: string,
		opts?: {
			payload?: Record<string, unknown>;
			sessionId?: string;
			progressToken?: string;
			async?: boolean;
			abort?: AbortController;
			requestId?: string;
			parseResponse?: (data: CallToolResult) => T;
		},
	): Promise<{ result: T; requestId: string }> {
		// If sessionId is provided, create a new client instance with that session
		const client = this.#getChatClient(opts?.sessionId);

		try {
			return await this.#callMCPToolWithClient(client, name, opts);
		} catch (error) {
			// Try to get notification context and show error
			try {
				const notifications = getNotificationContext();
				const message = error instanceof Error ? error.message : String(error);
				notifications.error("API Error", message);
			} catch {
				// If context is not available (e.g., during SSR), just log
				console.error("MCP Tool Error:", error);
			}
			throw error;
		}
	}

	async capabilities() {
		const client = this.#getChatClient();
		const { initializeResult } = await client.getSessionDetails();
		return (
			initializeResult?.capabilities?.experimental?.["ai.nanobot"]?.session ??
			{}
		);
	}

	async deleteThread(threadId: string): Promise<void> {
		const client = this.#getChatClient(threadId);
		return client.deleteSession();
	}

	async renameThread(threadId: string, title: string): Promise<Chat> {
		const { result } = await this.#callMCPToolWithClient<Chat>(
			this.#getMetaClient(),
			"update_chat",
			{
				payload: {
					chatId: threadId,
					title: title,
				},
			},
		);
		return result;
	}

	async listAgents(_opts?: { sessionId?: string }): Promise<Agents> {
		// Agent state is global metadata, so this always comes from the meta endpoint.
		const result = await this.metaClient.listResources({
			prefix: AgentResourcePrefix,
		});

		return {
			agents: result.resources.map(mapAgentResource),
		};
	}

	async getThreads(): Promise<Chat[]> {
		// Thread metadata lives in resources and may be updated asynchronously.
		const result = await this.metaClient.listResources({
			prefix: ChatThreadResourcePrefix,
		});

		return result.resources
			.map(mapChatResource)
			.sort(
				(a, b) =>
					new Date(b.created).getTime() - new Date(a.created).getTime(),
			);
	}

	async getThread(threadId: string): Promise<Chat | undefined> {
		const threads = await this.getThreads();
		return threads.find((thread) => thread.id === threadId);
	}

	watchThreadListChanged(onListChanged: () => void): () => void {
		return this.metaClient.watchResourceListChanged(onListChanged);
	}

	watchThreadChanged(threadId: string, onChanged: () => void): () => void {
		return this.metaClient.watchResource(
			`${ChatThreadResourcePrefix}${encodeURIComponent(threadId)}`,
			() => onChanged(),
		);
	}

	async createThread(): Promise<Chat> {
		const client = this.#getChatClient("new");
		const { id } = await client.getSessionDetails();
		return {
			id,
			title: "New Chat",
			created: new SvelteDate().toISOString(),
		};
	}

	async createResource(
		name: string,
		mimeType: string,
		blob: string,
		opts?: {
			description?: string;
			sessionId?: string;
			abort?: AbortController;
		},
	): Promise<Attachment> {
		const { result } = await this.callMCPTool<Attachment>("create_resource", {
			payload: {
				blob,
				mimeType,
				name,
				...(opts?.description && { description: opts.description }),
			},
			sessionId: opts?.sessionId,
			abort: opts?.abort,
			parseResponse: (resp: CallToolResult) => {
				if (resp.content?.[0]?.type === "resource_link") {
					return {
						uri: resp.content[0].uri,
					};
				}
				return {
					uri: "",
				};
			},
		});
		return result;
	}

	async sendMessage(
		request: ChatRequest,
		toolName: string,
		requestId: string,
	): Promise<{ result: ChatResult; requestId: string }> {
		await this.callMCPTool<CallToolResult>(toolName, {
			requestId,
			payload: {
				prompt: request.message,
				attachments: request.attachments?.map((a) => {
					return {
						name: a.name,
						url: a.uri,
						mimeType: a.mimeType,
					};
				}),
			},
			sessionId: request.threadId,
			progressToken: request.id,
			async: true,
		});
		const message: ChatMessage = {
			id: request.id,
			role: "user",
			created: now(),
			items: [
				{
					id: request.id + "_0",
					type: "text",
					text: request.message,
				},
			],
		};
		return {
			result: { message },
			requestId,
		};
	}

	async cancelRequest(requestId: string, sessionId: string): Promise<void> {
		const client = this.#getChatClient(sessionId);
		await client.notify("notifications/cancelled", {
			requestId,
			reason: "User requested cancellation",
		});
	}

	subscribe(
		threadId: string,
		onEvent: (e: Event) => void,
		opts?: {
			events?: string[];
			batchInterval?: number;
		},
	): () => void {
		console.log("Subscribing to thread:", threadId);
		const eventSource = new EventSource(
			`${this.baseUrl}/api/events/${threadId}`,
		);

		// Batching setup
		const batchInterval = opts?.batchInterval ?? 200; // Default 200ms
		let eventBuffer: Event[] = [];
		let batchTimer: ReturnType<typeof setTimeout> | null = null;

		const flushBuffer = () => {
			if (eventBuffer.length === 0) return;

			// Process all buffered events at once
			const eventsToProcess = [...eventBuffer];
			eventBuffer = [];

			for (const event of eventsToProcess) {
				onEvent(event);
			}
		};

		const scheduleBatch = () => {
			if (batchTimer === null) {
				batchTimer = setTimeout(() => {
					flushBuffer();
					batchTimer = null;
				}, batchInterval);
			}
		};

		eventSource.onmessage = (e) => {
			const data = JSON.parse(e.data);
			eventBuffer.push({
				type: "message",
				message: data,
			});
			scheduleBatch();
		};

		for (const type of opts?.events ?? []) {
			eventSource.addEventListener(type, (e) => {
				const idInt = parseInt(e.lastEventId);
				const event: Event = {
					id: idInt || e.lastEventId,
					type: type as
						| "history-start"
						| "history-end"
						| "chat-in-progress"
						| "chat-done"
						| "elicitation/create"
						| "error",
					data: JSON.parse(e.data),
				};

				// Certain events should be processed immediately (not batched)
				if (
					type === "history-start" ||
					type === "history-end" ||
					type === "chat-done"
				) {
					// Flush any pending events first
					flushBuffer();
					if (batchTimer !== null) {
						clearTimeout(batchTimer);
						batchTimer = null;
					}
					// Then process this event immediately
					onEvent(event);
				} else {
					eventBuffer.push(event);
					scheduleBatch();
				}
			});
		}

		eventSource.onerror = (e) => {
			// Flush buffer before processing error
			flushBuffer();
			if (batchTimer !== null) {
				clearTimeout(batchTimer);
				batchTimer = null;
			}
			onEvent({ type: "error", error: String(e) });
			console.error("EventSource failed:", e);
			eventSource.close();
		};

		eventSource.onopen = () => {
			console.log("EventSource connected for thread:", threadId);
		};

		return () => {
			// Clean up: flush remaining events and clear timer
			flushBuffer();
			if (batchTimer !== null) {
				clearTimeout(batchTimer);
			}
			eventSource.close();
		};
	}
}

export function appendMessage(
	messages: ChatMessage[],
	newMessage: ChatMessage,
): ChatMessage[] {
	let found = false;
	if (newMessage.id) {
		messages = messages.map((oldMessage) => {
			if (oldMessage.id === newMessage.id) {
				found = true;
				return newMessage;
			}
			return oldMessage;
		});
	}
	if (!found) {
		messages = [...messages, newMessage];
	}
	return messages;
}

// Default instance
export const defaultChatApi = new ChatAPI();

export class ChatService {
	messages: ChatMessage[];
	prompts: Prompt[];
	resources: Resource[];
	agent: Agent;
	agents: Agent[];
	selectedAgentId: string;
	elicitations: Elicitation[];
	isLoading: boolean;
	chatId: string;
	uploadedFiles: UploadedFile[];
	uploadingFiles: UploadingFile[];

	private api: ChatAPI;
	private closer = () => {};
	private history: ChatMessage[] | undefined;
	private onChatDone: (() => void)[] = [];
	private currentRequestId: string | undefined;

	constructor(opts?: { api?: ChatAPI; chatId?: string }) {
		this.api = opts?.api || defaultChatApi;
		this.messages = $state<ChatMessage[]>([]);
		this.history = $state<ChatMessage[]>();
		this.isLoading = $state(false);
		this.elicitations = $state<Elicitation[]>([]);
		this.prompts = $state<Prompt[]>([]);
		this.resources = $state<Resource[]>([]);
		this.chatId = $state("");
		this.agent = $state<Agent>({ id: "" });
		this.agents = $state<Agent[]>([]);
		this.selectedAgentId = $state("");
		this.uploadedFiles = $state([]);
		this.uploadingFiles = $state([]);
		this.setChatId(opts?.chatId);
	}

	close = () => {
		this.closer();
		this.setChatId("");
	};

	setChatId = async (chatId?: string) => {
		if (chatId === this.chatId) {
			return;
		}

		this.messages = [];
		this.prompts = [];
		this.resources = [];
		this.elicitations = [];
		this.history = undefined;
		this.isLoading = false;
		this.uploadedFiles = [];
		this.uploadingFiles = [];

		if (chatId) {
			this.chatId = chatId;
			this.subscribe(chatId);
		}

		this.listResources().then((r) => {
			if (r && r.resources) {
				this.resources = r.resources;
			}
		});

		this.listPrompts().then((prompts) => {
			if (prompts && prompts.prompts) {
				this.prompts = prompts.prompts;
			}
		});

		await this.reloadAgent();
	};

	private reloadAgent = async () => {
		const [agentsData, thread] = await Promise.all([
			this.api.listAgents(),
			this.chatId ? this.api.getThread(this.chatId) : Promise.resolve(undefined),
		]);

		let agents = agentsData.agents || [];
		if (thread?.availableAgentIds && thread.availableAgentIds.length > 0) {
			// Thread-scoped allowlists hide agents that are not part of this thread's config.
			const allowedAgents = new Set(thread.availableAgentIds);
			agents = agents.filter((agent) => allowedAgents.has(agent.id));
		}

		if (agents.length > 0) {
			this.agents = agents;
			this.agent =
				(thread?.currentAgentId && agents.find((a) => a.id === thread.currentAgentId)) ||
				agents.find((a) => a.current) ||
				agents[0];

			// Only reset selectedAgentId if:
			// 1. It's not set yet (empty string), OR
			// 2. The currently selected agent is no longer in the agents list
			const isSelectedAgentStillAvailable = agents.some(
				(a) => a.id === this.selectedAgentId,
			);

			if (!this.selectedAgentId || !isSelectedAgentStillAvailable) {
				this.selectedAgentId = this.agent.id || "";
			}
		}
	};

	selectAgent = (agentId: string) => {
		this.selectedAgentId = agentId;
		// Keep this.agent in sync with the selectedAgentId so the UI
		// (which may rely on chat.agent) reflects the newly selected agent.
		const selectedAgent = this.agents?.find((a) => a.id === agentId);
		if (selectedAgent) {
			this.agent = selectedAgent;
		}
	};

	listPrompts = async () => {
		return (await this.api.exchange(
			"prompts/list",
			{},
			{
				sessionId: this.chatId,
			},
		)) as Prompts;
	};

	listResources = async () => {
		return (await this.api.exchange(
			"resources/list",
			{},
			{
				sessionId: this.chatId,
			},
		)) as Resources;
	};

	private subscribe(chatId: string) {
		this.closer();
		if (!chatId) {
			return;
		}
		this.closer = this.api.subscribe(
			chatId,
			(event) => {
				if (event.type == "message" && event.message?.id) {
					if (this.history) {
						this.history = appendMessage(this.history, event.message);
					} else {
						this.messages = appendMessage(this.messages, event.message);
					}
				} else if (event.type == "history-start") {
					this.history = [];
				} else if (event.type == "history-end") {
					this.messages = this.history || [];
					this.history = undefined;
				} else if (event.type == "chat-in-progress") {
					this.isLoading = true;
				} else if (event.type == "chat-done") {
					this.isLoading = false;
					for (const waiting of this.onChatDone) {
						waiting();
					}
					this.onChatDone = [];
				} else if (event.type == "elicitation/create") {
					this.elicitations = [
						...this.elicitations,
						{
							id: event.id,
							...(event.data as object),
						} as Elicitation,
					];
				}
				console.debug("Received event:", event);
			},
			{
				events: [
					"history-start",
					"history-end",
					"chat-in-progress",
					"chat-done",
					"elicitation/create",
				],
			},
		);
	}

	replyToElicitation = async (
		elicitation: Elicitation,
		result: ElicitationResult,
	) => {
		await this.api.reply(elicitation.id, result, {
			sessionId: this.chatId,
		});
		this.elicitations = this.elicitations.filter(
			(e) => e.id !== elicitation.id,
		);
	};

	newChat = async () => {
		const thread = await this.api.createThread();
		await this.setChatId(thread.id);
	};

	sendMessage = async (message: string, attachments?: Attachment[]) => {
		if (!message.trim() || this.isLoading) return;

		this.isLoading = true;

		if (!this.chatId) {
			await this.newChat();
		}

		// Determine which tool to call based on selected or current agent
		const effectiveAgentId = this.selectedAgentId || this.agent?.id;
		if (!effectiveAgentId) {
			this.isLoading = false;
			throw new Error(
				"No agent selected or available for sending chat messages.",
			);
		}
		const toolName = `chat-with-${effectiveAgentId}`;

		try {
			// Store the request ID before the exchange so cancellation works immediately
			const requestId = crypto.randomUUID();
			this.currentRequestId = requestId;

			const { result } = await this.api.sendMessage(
				{
					id: crypto.randomUUID(),
					threadId: this.chatId,
					message: message,
					attachments: [...this.uploadedFiles, ...(attachments || [])],
				},
				toolName,
				requestId,
			);
			this.uploadedFiles = [];

			this.messages = appendMessage(this.messages, result.message);
			return new Promise<ChatResult | void>((resolve) => {
				this.onChatDone.push(() => {
					this.isLoading = false;
					this.currentRequestId = undefined;
					const i = this.messages.findIndex((m) => m.id === result.message.id);
					if (i !== -1 && i <= this.messages.length) {
						resolve({
							message: this.messages[i + 1],
						});
					} else {
						resolve();
					}
				});
			});
		} catch (error) {
			this.isLoading = false;
			this.currentRequestId = undefined;
			this.messages = appendMessage(this.messages, {
				id: crypto.randomUUID(),
				role: "assistant",
				created: now(),
				items: [
					{
						id: crypto.randomUUID(),
						type: "text",
						text: `Sorry, I couldn't send your message. Please try again. Error: ${error}`,
					},
				],
			});
		}
	};

	cancelChat = async () => {
		if (!this.currentRequestId || !this.chatId) return;

		const requestId = this.currentRequestId;
		this.currentRequestId = undefined;
		this.isLoading = false;

		// Fire all onChatDone callbacks
		for (const waiting of this.onChatDone) {
			waiting();
		}
		this.onChatDone = [];

		// Send the cancellation notification
		await this.api.cancelRequest(requestId, this.chatId);
	};

	cancelUpload = (fileId: string) => {
		this.uploadingFiles = this.uploadingFiles.filter((f) => {
			if (f.id !== fileId) {
				return true;
			}
			if (f.controller) {
				f.controller.abort();
			}
			return false;
		});
		this.uploadedFiles = this.uploadedFiles.filter((f) => f.id !== fileId);
	};

	uploadFile = async (
		file: File,
		opts?: {
			controller?: AbortController;
		},
	): Promise<Attachment> => {
		// Create thread if it doesn't exist
		if (!this.chatId) {
			const thread = await this.api.createThread();
			await this.setChatId(thread.id);
		}

		const fileId = crypto.randomUUID();
		const controller = opts?.controller || new AbortController();

		this.uploadingFiles.push({
			file,
			id: fileId,
			controller,
		});

		try {
			const result = await this.doUploadFile(file, controller);
			this.uploadedFiles.push({
				file,
				uri: result.uri,
				id: fileId,
				mimeType: result.mimeType,
			});
			return result;
		} finally {
			this.uploadingFiles = this.uploadingFiles.filter((f) => f.id !== fileId);
		}
	};

	private doUploadFile = async (
		file: File,
		controller: AbortController,
	): Promise<Attachment> => {
		// convert file to base64 string
		const reader = new FileReader();
		reader.readAsDataURL(file);
		await new Promise((resolve, reject) => {
			reader.onloadend = resolve;
			reader.onerror = reject;
		});
		const base64 = (reader.result as string).split(",")[1];

		if (!this.chatId) {
			throw new Error("Chat ID not set");
		}

		return await this.api.createResource(file.name, file.type, base64, {
			description: file.name,
			sessionId: this.chatId,
			abort: controller,
		});
	};
}

function now(): string {
	return new Date().toISOString();
}
