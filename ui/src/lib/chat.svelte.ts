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
	type Prompt
} from './types';
import { getNotificationContext } from './context/notifications.svelte';

interface CallToolResult {
	Content?: ToolOutputItem[];
}

export class ChatAPI {
	private readonly baseUrl: string;
	private readonly fetcher: typeof fetch;

	constructor(
		baseUrl: string = '',
		opts?: {
			fetcher?: typeof fetch;
		}
	) {
		this.baseUrl = baseUrl;
		this.fetcher = opts?.fetcher || fetch;
	}

	async reply(id: string | number, result: unknown, opts?: { sessionId?: string }) {
		const resp = await this.fetcher(`${this.baseUrl}/mcp/ui`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				...(opts?.sessionId && { 'Mcp-Session-Id': opts.sessionId })
			},
			body: JSON.stringify({
				jsonrpc: '2.0',
				id,
				result
			})
		});

		// We expect a 204 No Content response
		if (resp.status == 204) {
			return;
		}

		if (!resp.ok) {
			const text = await resp.text();
			logError(`response: ${resp.status}: ${resp.statusText}: ${text}`);
			throw new Error(text);
		}

		try {
			// check for a protocol error
			const data = await resp.json();
			if (data.error?.message) {
				logError(data.error.message);
			}
		} catch (e) {
			console.debug('Error parsing JSON:', e);
		}
	}

	async exchange(method: string, params: unknown, opts?: { sessionId?: string }) {
		const resp = await this.fetcher(`${this.baseUrl}/mcp/ui`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				...(opts?.sessionId && { 'Mcp-Session-Id': opts.sessionId })
			},
			body: JSON.stringify({
				id: crypto.randomUUID(),
				jsonrpc: '2.0',
				method,
				params
			})
		});

		let toThrow = null;
		try {
			const data = await resp.json();
			if (data.error?.message) {
				logError(data.error.message);
				toThrow = new Error(data.error.message);
			}
			return data.result;
		} catch (e) {
			logError(e);
			toThrow = e;
		}
		throw toThrow;
	}

	private async callMCPTool<T>(
		name: string,
		opts?: {
			payload?: Record<string, unknown>;
			sessionId?: string;
			progressToken?: string;
			async?: boolean;
		}
	): Promise<T> {
		const resp = await this.fetcher(`${this.baseUrl}/mcp/ui`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				...(opts?.sessionId && { 'Mcp-Session-Id': opts.sessionId })
			},
			body: JSON.stringify({
				id: crypto.randomUUID(),
				jsonrpc: '2.0',
				method: 'tools/call',
				params: {
					name: name,
					arguments: opts?.payload || {},
					...(opts?.async && {
						_meta: {
							'ai.nanobot.async': true,
							progressToken: opts?.progressToken
						}
					})
				}
			})
		});

		const data = await resp.json();
		if (data.error?.message) {
			// Try to get notification context and show error
			try {
				const notifications = getNotificationContext();
				notifications.error('API Error', data.error.message);
			} catch {
				// If context is not available (e.g., during SSR), just log
				console.error('MCP Tool Error:', data.error.message);
			}
			throw new Error(data.error.message);
		}
		if (data.result?.content?.[0]?.structuredContent) {
			return data.result.content[0].structuredContent as T;
		}
		return {} as T;
	}

	async deleteThread(threadId: string): Promise<void> {
		await this.callMCPTool('delete_chat', {
			payload: {
				chatId: threadId
			}
		});
	}

	async renameThread(threadId: string, title: string): Promise<Chat> {
		return await this.callMCPTool<Chat>('update_chat', {
			payload: {
				chatId: threadId,
				title: title
			}
		});
	}

	async getThreads(): Promise<Chat[]> {
		return (
			await this.callMCPTool<{
				chats: Chat[];
			}>('list_chats')
		).chats;
	}

	async createThread(): Promise<Chat> {
		return await this.callMCPTool<Chat>('create_chat');
	}

	async sendMessage(request: ChatRequest): Promise<ChatResult> {
		await this.callMCPTool<CallToolResult>('chat_ui', {
			payload: {
				prompt: request.message
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
		}
	): () => void {
		const eventSource = new EventSource(`${this.baseUrl}/api/events/${threadId}`);
		eventSource.onmessage = (e) => {
			const data = JSON.parse(e.data);
			onEvent({
				type: 'message',
				message: data
			});
		};
		for (const type of opts?.events ?? []) {
			eventSource.addEventListener(type, (e) => {
				onEvent({
					id: e.lastEventId,
					type: type as
						| 'history-start'
						| 'history-end'
						| 'chat-in-progress'
						| 'chat-done'
						| 'elicitation/create'
						| 'error',
					data: JSON.parse(e.data)
				});
			});
		}
		eventSource.onerror = (e) => {
			onEvent({ type: 'error', error: String(e) });
			console.error('EventSource failed:', e);
			eventSource.close();
		};
		eventSource.onopen = () => {
			console.log('EventSource connected for thread:', threadId);
		};

		return () => eventSource.close();
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
export const chatApi = new ChatAPI();

export class ChatService {
	messages: ChatMessage[];
	prompts: Prompt[];
	elicitations: Elicitation[];
	isLoading: boolean;
	chatId: string;

	private api: ChatAPI;
	private closer = () => {};
	private history: ChatMessage[] | undefined;

	constructor(opts?: { api?: ChatAPI; chatId?: string }) {
		this.api = opts?.api || chatApi;
		this.messages = $state<ChatMessage[]>([]);
		this.history = $state<ChatMessage[]>();
		this.isLoading = $state(false);
		this.elicitations = $state<Elicitation[]>([]);
		this.prompts = $state<Prompt[]>([]);
		this.chatId = $state('');
		if (opts?.chatId) {
			this.setChatId(opts.chatId);
		}
	}

	close = () => {
		this.closer();
		this.setChatId('');
	};

	setChatId = async (chatId: string) => {
		if (chatId === this.chatId) {
			return;
		}
		this.chatId = chatId;
		this.messages = [];
		this.prompts = [];
		this.elicitations = [];
		this.history = undefined;
		this.isLoading = false;
		this.subscribe(chatId);

		if (chatId) {
			const prompts = await this.listPrompts();
			if (prompts.prompts) {
				this.prompts = prompts.prompts;
			}
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
				} else if (event.type == 'elicitation/create') {
					this.elicitations = [
						...this.elicitations,
						{
							id: event.id,
							...(event.data as object)
						} as Elicitation
					];
				}
				console.log('Received event:', event);
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
		console.log('replyToElicitation', elicitation.id, result);
		await this.api.reply(elicitation.id, result, {
			sessionId: this.chatId
		});
		this.elicitations = this.elicitations.filter((e) => e.id !== elicitation.id);
	};

	sendMessage = async (message: string) => {
		if (!message.trim() || this.isLoading) return;

		this.isLoading = true;

		if (!this.chatId) {
			// Create a new thread if it doesn'c exist
			const thread = await this.api.createThread();
			await this.setChatId(thread.id);
		}

		try {
			const response = await this.api.sendMessage({
				id: crypto.randomUUID(),
				threadId: this.chatId,
				message: message
			});

			this.messages = appendMessage(this.messages, response.message);
		} catch (error) {
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
		} finally {
			this.isLoading = false;
		}
	};
}

function now(): string {
	return new Date().toISOString();
}

function logError(error: unknown) {
	try {
		const notifications = getNotificationContext();
		notifications.error('API Error', error?.toString());
	} catch {
		// If context is not available (e.g., during SSR), just log
		console.error('MCP Tool Error:', error);
	}
	console.error('Error:', error);
}
