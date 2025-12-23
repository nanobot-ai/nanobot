import {
	type Event,
	type Chat,
	type ChatResult,
	type ChatRequest,
	type ChatMessage,
	type ToolOutputItem,
	type Elicitation,
	type ElicitationResult,
	type Prompts,
	type Prompt,
	type Agent,
	type Agents,
	type Attachment,
	type UploadedFile,
	type UploadingFile,
	type Resource,
	type Resources
} from './types';
import { getNotificationContext } from './context/notifications.svelte';
import { SimpleClient } from './mcpclient';
import { SvelteDate } from 'svelte/reactivity';

export interface CallToolResult {
	content?: ToolOutputItem[];
}

export class ChatAPI {
	private readonly baseUrl: string;
	private readonly mcpClient: SimpleClient;

	constructor(
		baseUrl: string = '',
		opts?: {
			fetcher?: typeof fetch;
			sessionId?: string;
		}
	) {
		this.baseUrl = baseUrl;
		this.mcpClient = new SimpleClient({
			baseUrl: baseUrl,
			fetcher: opts?.fetcher,
			sessionId: opts?.sessionId
		});
	}

	#getClient(sessionId?: string) {
		if (sessionId) {
			return new SimpleClient({
				baseUrl: this.baseUrl,
				sessionId
			});
		}
		return this.mcpClient;
	}

	async reply(id: string | number, result: unknown, opts?: { sessionId?: string }) {
		// If sessionId is provided, create a new client instance with that session
		const client = this.#getClient(opts?.sessionId);
		await client.reply(id, result);
	}

	async exchange(method: string, params: unknown, opts?: { sessionId?: string }) {
		// If sessionId is provided, create a new client instance with that session
		const client = this.#getClient(opts?.sessionId);
		return await client.exchange(method, params);
	}

	async callMCPTool<T>(
		name: string,
		opts?: {
			payload?: Record<string, unknown>;
			sessionId?: string;
			progressToken?: string;
			async?: boolean;
			abort?: AbortController;
			parseResponse?: (data: CallToolResult) => T;
		}
	): Promise<T> {
		// If sessionId is provided, create a new client instance with that session
		const client = this.#getClient(opts?.sessionId);

		try {
			// Get the raw result from exchange to support parseResponse
			const result = await client.exchange(
				'tools/call',
				{
					name: name,
					arguments: opts?.payload || {},
					...(opts?.async && {
						_meta: {
							'ai.nanobot.async': true,
							progressToken: opts?.progressToken
						}
					})
				},
				{ abort: opts?.abort }
			);

			if (opts?.parseResponse) {
				return opts.parseResponse(result as CallToolResult);
			}

			// Handle structured content
			if (result && typeof result === 'object' && 'structuredContent' in result) {
				return (result as { structuredContent: T }).structuredContent;
			}

			return result as T;
		} catch (error) {
			// Try to get notification context and show error
			try {
				const notifications = getNotificationContext();
				const message = error instanceof Error ? error.message : String(error);
				notifications.error('API Error', message);
			} catch {
				// If context is not available (e.g., during SSR), just log
				console.error('MCP Tool Error:', error);
			}
			throw error;
		}
	}

	async capabilities() {
		const client = this.#getClient();
		const { initializeResult } = await client.getSessionDetails();
		return initializeResult?.capabilities?.experimental?.['ai.nanobot']?.session ?? {};
	}

	async deleteThread(threadId: string): Promise<void> {
		const client = this.#getClient(threadId);
		return client.deleteSession();
	}

	async renameThread(threadId: string, title: string): Promise<Chat> {
		return await this.callMCPTool<Chat>('update_chat', {
			payload: {
				chatId: threadId,
				title: title
			}
		});
	}

	async listAgents(opts?: { sessionId?: string }): Promise<Agents> {
		return await this.callMCPTool<Agents>('list_agents', opts);
	}

	async getThreads(): Promise<Chat[]> {
		return (
			await this.callMCPTool<{
				chats: Chat[];
			}>('list_chats')
		).chats;
	}

	async createThread(): Promise<Chat> {
		const client = this.#getClient('new');
		const { id } = await client.getSessionDetails();
		return {
			id,
			title: 'New Chat',
			created: new SvelteDate().toISOString()
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
		}
	): Promise<Attachment> {
		return await this.callMCPTool<Attachment>('create_resource', {
			payload: {
				blob,
				mimeType,
				name,
				...(opts?.description && { description: opts.description })
			},
			sessionId: opts?.sessionId,
			abort: opts?.abort,
			parseResponse: (resp: CallToolResult) => {
				if (resp.content?.[0]?.type === 'resource_link') {
					return {
						uri: resp.content[0].uri
					};
				}
				return {
					uri: ''
				};
			}
		});
	}

	async sendMessage(request: ChatRequest): Promise<ChatResult> {
		await this.callMCPTool<CallToolResult>('chat', {
			payload: {
				prompt: request.message,
				attachments: request.attachments?.map((a) => {
					return {
						name: a.name,
						url: a.uri,
						mimeType: a.mimeType
					};
				})
			},
			sessionId: request.threadId,
			progressToken: request.id,
			async: true
		});
		const message: ChatMessage = {
			id: request.id,
			role: 'user',
			created: now(),
			items: [
				{
					id: request.id + '_0',
					type: 'text',
					text: request.message
				}
			]
		};
		return {
			message
		};
	}

	subscribe(
		threadId: string,
		onEvent: (e: Event) => void,
		opts?: {
			events?: string[];
			batchInterval?: number;
		}
	): () => void {
		console.log('Subscribing to thread:', threadId);
		const eventSource = new EventSource(`${this.baseUrl}/api/events/${threadId}`);

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
				type: 'message',
				message: data
			});
			scheduleBatch();
		};

		for (const type of opts?.events ?? []) {
			eventSource.addEventListener(type, (e) => {
				const idInt = parseInt(e.lastEventId);
				const event: Event = {
					id: idInt || e.lastEventId,
					type: type as
						| 'history-start'
						| 'history-end'
						| 'chat-in-progress'
						| 'chat-done'
						| 'elicitation/create'
						| 'error',
					data: JSON.parse(e.data)
				};

				// Certain events should be processed immediately (not batched)
				if (type === 'history-start' || type === 'history-end' || type === 'chat-done') {
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
			onEvent({ type: 'error', error: String(e) });
			console.error('EventSource failed:', e);
			eventSource.close();
		};

		eventSource.onopen = () => {
			console.log('EventSource connected for thread:', threadId);
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

export function appendMessage(messages: ChatMessage[], newMessage: ChatMessage): ChatMessage[] {
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
	elicitations: Elicitation[];
	isLoading: boolean;
	chatId: string;
	uploadedFiles: UploadedFile[];
	uploadingFiles: UploadingFile[];

	private api: ChatAPI;
	private closer = () => {};
	private history: ChatMessage[] | undefined;
	private onChatDone: (() => void)[] = [];

	constructor(opts?: { api?: ChatAPI; chatId?: string }) {
		this.api = opts?.api || defaultChatApi;
		this.messages = $state<ChatMessage[]>([]);
		this.history = $state<ChatMessage[]>();
		this.isLoading = $state(false);
		this.elicitations = $state<Elicitation[]>([]);
		this.prompts = $state<Prompt[]>([]);
		this.resources = $state<Resource[]>([]);
		this.chatId = $state('');
		this.agent = $state<Agent>({});
		this.uploadedFiles = $state([]);
		this.uploadingFiles = $state([]);
		this.setChatId(opts?.chatId);
	}

	close = () => {
		this.closer();
		this.setChatId('');
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
		const agents = await this.api.listAgents({ sessionId: this.chatId });
		if (agents.agents?.length > 0) {
			this.agent = agents.agents[0];
		}
	};

	listPrompts = async () => {
		return (await this.api.exchange(
			'prompts/list',
			{},
			{
				sessionId: this.chatId
			}
		)) as Prompts;
	};

	listResources = async () => {
		return (await this.api.exchange(
			'resources/list',
			{},
			{
				sessionId: this.chatId
			}
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
				if (event.type == 'message' && event.message?.id) {
					if (this.history) {
						this.history = appendMessage(this.history, event.message);
					} else {
						this.messages = appendMessage(this.messages, event.message);
					}
				} else if (event.type == 'history-start') {
					this.history = [];
				} else if (event.type == 'history-end') {
					this.messages = this.history || [];
					this.history = undefined;
				} else if (event.type == 'chat-in-progress') {
					this.isLoading = true;
				} else if (event.type == 'chat-done') {
					this.isLoading = false;
					for (const waiting of this.onChatDone) {
						waiting();
					}
					this.onChatDone = [];
				} else if (event.type == 'elicitation/create') {
					this.elicitations = [
						...this.elicitations,
						{
							id: event.id,
							...(event.data as object)
						} as Elicitation
					];
				}
				console.debug('Received event:', event);
			},
			{
				events: [
					'history-start',
					'history-end',
					'chat-in-progress',
					'chat-done',
					'elicitation/create'
				]
			}
		);
	}

	replyToElicitation = async (elicitation: Elicitation, result: ElicitationResult) => {
		await this.api.reply(elicitation.id, result, {
			sessionId: this.chatId
		});
		this.elicitations = this.elicitations.filter((e) => e.id !== elicitation.id);
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

		try {
			const response = await this.api.sendMessage({
				id: crypto.randomUUID(),
				threadId: this.chatId,
				message: message,
				attachments: [...this.uploadedFiles, ...(attachments || [])]
			});
			this.uploadedFiles = [];

			this.messages = appendMessage(this.messages, response.message);
			return new Promise<ChatResult | void>((resolve) => {
				this.onChatDone.push(() => {
					this.isLoading = false;
					const i = this.messages.findIndex((m) => m.id === response.message.id);
					if (i !== -1 && i <= this.messages.length) {
						resolve({
							message: this.messages[i + 1]
						});
					} else {
						resolve();
					}
				});
			});
		} catch (error) {
			this.isLoading = false;
			this.messages = appendMessage(this.messages, {
				id: crypto.randomUUID(),
				role: 'assistant',
				created: now(),
				items: [
					{
						id: crypto.randomUUID(),
						type: 'text',
						text: `Sorry, I couldn't send your message. Please try again. Error: ${error}`
					}
				]
			});
		}
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
		}
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
			controller
		});

		try {
			const result = await this.doUploadFile(file, controller);
			this.uploadedFiles.push({
				file,
				uri: result.uri,
				id: fileId,
				mimeType: result.mimeType
			});
			return result;
		} finally {
			this.uploadingFiles = this.uploadingFiles.filter((f) => f.id !== fileId);
		}
	};

	private doUploadFile = async (file: File, controller: AbortController): Promise<Attachment> => {
		// convert file to base64 string
		const reader = new FileReader();
		reader.readAsDataURL(file);
		await new Promise((resolve, reject) => {
			reader.onloadend = resolve;
			reader.onerror = reject;
		});
		const base64 = (reader.result as string).split(',')[1];

		if (!this.chatId) {
			throw new Error('Chat ID not set');
		}

		return await this.api.createResource(file.name, file.type, base64, {
			description: file.name,
			sessionId: this.chatId,
			abort: controller
		});
	};
}

function now(): string {
	return new Date().toISOString();
}
