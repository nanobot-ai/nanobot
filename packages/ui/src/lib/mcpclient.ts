import { Client } from '@modelcontextprotocol/sdk/client';
import { StreamableHTTPClientTransport } from '@modelcontextprotocol/sdk/client/streamableHttp.js';
import { logError } from '$lib/notify';
import {
	type InitializationResult,
	type ResourceContents,
	type Resources,
	UIPath
} from '$lib/types';

interface StoredSession {
	sessionId: string;
	initializeResult?: InitializationResult;
}

/**
 * MCP client backed by the official SDK's Client + StreamableHTTPClientTransport.
 * Handles nanobot-specific concerns: localStorage sessions, _meta extensions,
 * reply() for elicitations, and 404 session retry.
 */
export class NanobotClient {
	readonly #url: string;
	readonly #rawUrl: string;
	readonly #fetcher: typeof fetch | undefined;
	readonly #externalSession: boolean;

	#client?: Client;
	#appClient?: Client;
	#transport?: StreamableHTTPClientTransport;
	#sessionId?: string;
	#initializeResult?: InitializationResult;
	#connectPromise?: Promise<void>;

	#sseSubscriptions = new Map<string, Set<(resource: ResourceContents) => void>>();

	constructor(opts?: {
		path?: string;
		baseUrl?: string;
		fetcher?: typeof fetch;
		workspaceId?: string;
		workspaceShared?: boolean;
		sessionId?: string;
	}) {
		const baseUrl = opts?.baseUrl || '';
		const path = opts?.path || UIPath;
		this.#url = `${baseUrl}${path}`;
		this.#fetcher = opts?.fetcher;
		if (opts?.workspaceId) {
			this.#url += `${this.#url.includes('?') ? '&' : '?'}workspace=${opts.workspaceId}`;
			if (opts.workspaceShared) {
				this.#url += `&shared=true`;
			}
		}
		this.#rawUrl = this.#url;

		// If sessionId provided in options, use it and mark as external
		if (opts?.sessionId) {
			this.#sessionId = opts.sessionId === 'new' ? undefined : opts.sessionId;
			this.#externalSession = true;
		} else {
			// Load session data from localStorage
			const stored = this.#getStoredSession();
			if (stored) {
				this.#sessionId = stored.sessionId;
				this.#initializeResult = stored.initializeResult;
			}
			this.#externalSession = false;
		}
	}

	/**
	 * Returns the underlying SDK Client instance, scoped to this session.
	 * Useful for passing to third-party libraries like @mcp-ui/client.
	 */
	get sdkClient(): Client | undefined {
		return this.#client;
	}

	/**
	 * Returns the underlying SDK Client after ensuring the connection is established.
	 * Use this when you need a fully-connected client (e.g., for @mcp-ui/client).
	 *
	 * When the main client was connected to an existing session, the SDK skips the
	 * initialize handshake (because sessionId is pre-set on the transport), leaving
	 * server capabilities undefined. In that case we create a dedicated "app client"
	 * whose transport does NOT have sessionId pre-set — forcing the SDK to run the
	 * full initialize handshake — but injects the Mcp-Session-Id header via a custom
	 * fetch so the server routes to the existing session.
	 */
	async ensureSDKClient(): Promise<Client | undefined> {
		await this.#ensureConnected();

		// If the main client completed a fresh handshake, it has capabilities — use it directly.
		if (this.#client?.getServerCapabilities()) {
			return this.#client;
		}

		// Return cached app client if already initialized.
		if (this.#appClient?.getServerCapabilities()) {
			return this.#appClient;
		}

		// For reconnected sessions the SDK skips initialization, so capabilities are null.
		// Create a dedicated Client whose transport does NOT have sessionId pre-set
		// (forcing the SDK to run the full initialize handshake), but injects the
		// Mcp-Session-Id header via a custom fetch so the server routes to the existing session.
		if (!this.#sessionId) return undefined;

		const sessionId = this.#sessionId;
		const baseFetch = this.#fetcher || fetch;

		const transport = new StreamableHTTPClientTransport(
			new URL(this.#url, window.location.origin),
			{
				fetch: ((url: string | URL, init?: RequestInit) => {
					const headers = new Headers(init?.headers);
					headers.set('mcp-session-id', sessionId);
					return baseFetch(url, { ...init, headers });
				}) as (url: string | URL, init?: RequestInit) => Promise<Response>
			}
		);

		const client = new Client(
			{ name: 'nanobot-ui', version: '0.0.1' },
			{ capabilities: {} }
		);

		await client.connect(transport);
		this.#appClient = client;
		return client;
	}

	async deleteSession(): Promise<void> {
		try {
			if (this.#transport) {
				await this.#transport.terminateSession();
			} else if (this.#sessionId) {
				// Fallback: manual DELETE if transport not yet created
				const fetcher = this.#fetcher || fetch;
				await fetcher(this.#url, {
					method: 'DELETE',
					headers: {
						'Mcp-Session-Id': this.#sessionId
					}
				});
			}
		} finally {
			this.#clearSession();
		}
	}

	#getStoredSession(): StoredSession | undefined {
		if (typeof window === 'undefined' || typeof localStorage === 'undefined') {
			return undefined;
		}
		const stored = localStorage.getItem(`mcp-session-${this.#rawUrl}`);
		if (!stored) {
			return undefined;
		}
		try {
			return JSON.parse(stored) as StoredSession;
		} catch (e) {
			console.error('[NanobotClient] Failed to parse stored session:', e);
			return undefined;
		}
	}

	#storeSession(sessionId: string, initializeResult?: InitializationResult) {
		if (typeof window === 'undefined' || typeof localStorage === 'undefined') {
			return;
		}
		const session: StoredSession = {
			sessionId,
			initializeResult
		};
		localStorage.setItem(`mcp-session-${this.#rawUrl}`, JSON.stringify(session));
	}

	#clearSession() {
		if (typeof window === 'undefined' || typeof localStorage === 'undefined') {
			return;
		}
		localStorage.removeItem(`mcp-session-${this.#rawUrl}`);
		this.#sessionId = undefined;
		this.#initializeResult = undefined;
		this.#client = undefined;
		this.#appClient = undefined;
		this.#transport = undefined;
		this.#connectPromise = undefined;
		this.#sseSubscriptions.clear();
	}

	#createTransportAndClient(): { transport: StreamableHTTPClientTransport; client: Client } {
		const transportOpts: ConstructorParameters<typeof StreamableHTTPClientTransport>[1] = {};

		if (this.#sessionId) {
			transportOpts.sessionId = this.#sessionId;
		}

		if (this.#fetcher) {
			transportOpts.fetch = this.#fetcher as (
				url: string | URL,
				init?: RequestInit
			) => Promise<Response>;
		}

		const transport = new StreamableHTTPClientTransport(
			new URL(this.#url, window.location.origin),
			transportOpts
		);

		const client = new Client(
			{ name: 'nanobot-ui', version: '0.0.1' },
			{
				capabilities: {}
			}
		);

		return { transport, client };
	}

	async #ensureConnected(): Promise<void> {
		if (this.#connectPromise) {
			return this.#connectPromise;
		}

		// If we already have a connected client, nothing to do
		if (this.#client && this.#transport) {
			return;
		}

		this.#connectPromise = (async () => {
			try {
				const { transport, client } = this.#createTransportAndClient();
				this.#transport = transport;
				this.#client = client;

				// connect() performs the initialize/initialized handshake
				await client.connect(transport);

				// Extract session ID and capabilities from the connected transport/client
				const newSessionId = transport.sessionId;
				if (newSessionId) {
					this.#sessionId = newSessionId;
				}

				// Map server capabilities to our InitializationResult format
				const serverCapabilities = client.getServerCapabilities();
				if (serverCapabilities) {
					this.#initializeResult = {
						capabilities: serverCapabilities
					} as InitializationResult;
				}

				// Persist session to localStorage (unless external)
				if (!this.#externalSession && this.#sessionId) {
					this.#storeSession(this.#sessionId, this.#initializeResult);
				}
			} finally {
				this.#connectPromise = undefined;
			}
		})();

		return this.#connectPromise;
	}

	async #ensureSession(): Promise<string> {
		await this.#ensureConnected();
		if (!this.#sessionId) {
			throw new Error('Failed to establish session');
		}
		return this.#sessionId;
	}

	async getSessionDetails(): Promise<{
		id: string;
		initializeResult?: InitializationResult;
	}> {
		return {
			id: await this.#ensureSession(),
			initializeResult: this.#initializeResult
		};
	}

	async reply(id: string | number, result: unknown): Promise<void> {
		await this.#ensureConnected();

		if (!this.#transport) {
			throw new Error('Transport not available');
		}

		// Send a raw JSON-RPC response through the transport.
		// This is used for elicitation replies that arrive via separate SSE.
		const resp = await (this.#fetcher || fetch)(this.#url, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				...(this.#sessionId && { 'Mcp-Session-Id': this.#sessionId })
			},
			body: JSON.stringify({
				jsonrpc: '2.0',
				id,
				result
			})
		});

		// We expect a 204 No Content response or 202
		if (resp.status === 204 || resp.status === 202) {
			return;
		}

		if (!resp.ok) {
			const text = await resp.text();
			logError(`reply: ${resp.status}: ${resp.statusText}: ${text}`);
			throw new Error(text);
		}

		try {
			// check for a protocol error
			const data = (await resp.json()) as {
				error?: { code: number; message: string; data?: unknown };
			};
			if (data.error) {
				logError(data.error);
				throw Error(`${data.error.message}: ${JSON.stringify(data.error)}`);
			}
		} catch (e) {
			// If it's already an Error, rethrow it
			if (e instanceof Error && e.message !== 'Unexpected end of JSON input') {
				throw e;
			}
			// Otherwise ignore JSON parse errors
			console.debug('[NanobotClient] Error parsing JSON in reply:', e);
		}
	}

	async notify(method: string, params: unknown): Promise<void> {
		await this.#ensureConnected();

		if (!this.#client) {
			throw new Error('Client not available');
		}

		try {
			await this.#client.notification({
				method,
				params
			} as Parameters<Client['notification']>[0]);
		} catch (e) {
			logError(`notify: ${e instanceof Error ? e.message : String(e)}`);
			// Don't throw - notifications are fire-and-forget
		}
	}

	async exchange(
		method: string,
		params: unknown,
		opts?: { abort?: AbortController; requestId?: string }
	): Promise<{ result: unknown; requestId: string }> {
		const requestId = opts?.requestId ?? crypto.randomUUID();

		try {
			await this.#ensureConnected();
			const result = await this.#dispatchRequest(method, params, opts);
			return { result, requestId };
		} catch (e) {
			// Handle 404 - session expired or invalid
			if (this.#isSessionExpiredError(e)) {
				if (this.#externalSession) {
					throw new Error('Session not found (404). External sessions cannot be recreated.');
				}
				this.#clearSession();
				// Retry once with new session
				await this.#ensureConnected();
				const result = await this.#dispatchRequest(method, params, opts);
				return { result, requestId };
			}
			throw e;
		}
	}

	#isSessionExpiredError(e: unknown): boolean {
		if (e instanceof Error) {
			// The SDK transport throws errors with status info for HTTP errors
			const msg = e.message.toLowerCase();
			return msg.includes('404') || msg.includes('session');
		}
		return false;
	}

	async #dispatchRequest(
		method: string,
		params: unknown,
		opts?: { abort?: AbortController }
	): Promise<unknown> {
		if (!this.#client) {
			throw new Error('Client not connected');
		}

		const requestOptions = opts?.abort ? { signal: opts.abort.signal } : undefined;

		const p = (params || {}) as Record<string, unknown>;

		switch (method) {
			case 'tools/call':
				return await this.#client.callTool(
					p as Parameters<Client['callTool']>[0],
					undefined,
					requestOptions
				);
			case 'resources/list':
				return await this.#client.listResources(
					p as Parameters<Client['listResources']>[0],
					requestOptions
				);
			case 'resources/read':
				return await this.#client.readResource(
					p as Parameters<Client['readResource']>[0],
					requestOptions
				);
			case 'resources/subscribe':
				return await this.#client.subscribeResource(
					p as Parameters<Client['subscribeResource']>[0],
					requestOptions
				);
			case 'resources/unsubscribe':
				return await this.#client.unsubscribeResource(
					p as Parameters<Client['unsubscribeResource']>[0],
					requestOptions
				);
			case 'tools/list':
				return await this.#client.listTools(
					p as Parameters<Client['listTools']>[0],
					requestOptions
				);
			case 'prompts/list':
				return await this.#client.listPrompts(
					p as Parameters<Client['listPrompts']>[0],
					requestOptions
				);
			case 'prompts/get':
				return await this.#client.getPrompt(
					p as Parameters<Client['getPrompt']>[0],
					requestOptions
				);
			default:
				throw new Error(`Unsupported MCP method: ${method}`);
		}
	}

	async callMCPTool<T>(
		name: string,
		opts?: {
			payload?: unknown;
			progressToken?: string;
			async?: boolean;
			abort?: AbortController;
			requestId?: string;
		}
	): Promise<{ result: T; requestId: string }> {
		const { result, requestId } = await this.exchange(
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
			{ abort: opts?.abort, requestId: opts?.requestId }
		);

		// If the result has a structuredContent field, use that, otherwise return the full response
		let finalResult: T;
		if (result && typeof result === 'object' && 'structuredContent' in result) {
			finalResult = (result as { structuredContent: T }).structuredContent;
		} else {
			finalResult = result as T;
		}

		return { result: finalResult, requestId };
	}

	async listResources<T extends Resources = Resources>(opts?: {
		prefix?: string | string[];
		abort?: AbortController;
	}): Promise<T> {
		const prefixes = opts?.prefix
			? Array.isArray(opts.prefix)
				? opts.prefix
				: [opts.prefix]
			: undefined;

		const { result } = await this.exchange(
			'resources/list',
			{
				...(prefixes && {
					_meta: {
						'ai.nanobot': {
							prefix: prefixes.length === 1 ? prefixes[0] : prefixes
						}
					}
				})
			},
			{ abort: opts?.abort }
		);

		const typedResult = result as T;

		if (prefixes) {
			return {
				...typedResult,
				resources: typedResult.resources.filter(({ uri }) =>
					prefixes.some((prefix) => uri.startsWith(prefix))
				)
			};
		}

		return typedResult;
	}

	async readResource(
		uri: string,
		opts?: { abort?: AbortController }
	): Promise<{ contents: ResourceContents[] }> {
		const { result } = await this.exchange('resources/read', { uri }, { abort: opts?.abort });
		return result as { contents: ResourceContents[] };
	}

	/**
	 * Watch for resource changes with a given prefix.
	 * Uses the SDK Client's subscribeResource() + notification handler.
	 * Returns a cleanup function to stop watching.
	 */
	watchResource(prefix: string, callback: (resource: ResourceContents) => void): () => void {
		// Add callback to subscriptions
		if (!this.#sseSubscriptions.has(prefix)) {
			this.#sseSubscriptions.set(prefix, new Set());
		}
		this.#sseSubscriptions.get(prefix)!.add(callback);

		// Set up subscription
		this.#ensureConnected().then(async () => {
			if (!this.#client) return;

			// Register notification handler for resource updates (only once)
			this.#setupResourceNotificationHandler();

			// Subscribe to resource changes for this prefix
			try {
				await this.#client.subscribeResource({ uri: prefix });
				console.log(`[NanobotClient] Subscribed to resources with prefix: ${prefix}`);
			} catch (e) {
				console.error(`[NanobotClient] Failed to subscribe to ${prefix}:`, e);
			}
		});

		// Return cleanup function
		return () => {
			const callbacks = this.#sseSubscriptions.get(prefix);
			if (callbacks) {
				callbacks.delete(callback);
				if (callbacks.size === 0) {
					this.#sseSubscriptions.delete(prefix);
					// Unsubscribe from server
					if (this.#client) {
						this.#client.unsubscribeResource({ uri: prefix }).catch((e: unknown) => {
							console.error(`[NanobotClient] Failed to unsubscribe from ${prefix}:`, e);
						});
					}
				}
			}
			console.log(`[NanobotClient] Stopped watching resources with prefix: ${prefix}`);
		};
	}

	#resourceNotificationHandlerSet = false;

	#setupResourceNotificationHandler() {
		if (this.#resourceNotificationHandlerSet || !this.#client) return;
		this.#resourceNotificationHandlerSet = true;

		// The SDK Client emits notifications via setNotificationHandler.
		// We listen for resource/updated notifications.
		try {
			// Import the notification schema dynamically would be complex,
			// so we use the fallback notification handler instead.
			this.#client.fallbackNotificationHandler = async (notification) => {
				if (
					notification.method === 'notifications/resources/updated' &&
					notification.params &&
					typeof notification.params === 'object' &&
					'uri' in notification.params
				) {
					const uri = notification.params.uri as string;

					// Find all subscriptions that match this URI prefix
					for (const [prefix, callbacks] of this.#sseSubscriptions.entries()) {
						if (uri.startsWith(prefix)) {
							// Fetch the updated resource details
							const resource = await this.#fetchResourceDetails(uri);
							if (resource) {
								callbacks.forEach((cb) => cb(resource));
							}
						}
					}
				}
			};
		} catch (e) {
			console.error('[NanobotClient] Failed to set notification handler:', e);
		}
	}

	/**
	 * Fetch resource details for a given URI
	 */
	async #fetchResourceDetails(uri: string): Promise<ResourceContents | null> {
		try {
			const result = await this.readResource(uri);
			if (result.contents?.length) {
				return result.contents[0] as ResourceContents;
			}
		} catch (e) {
			logError(
				`[NanobotClient] Failed to fetch resource ${uri}: ${e instanceof Error ? e.message : String(e)}`
			);
		}
		return null;
	}
}

/** @deprecated Use NanobotClient instead */
export const SimpleClient = NanobotClient;
/** @deprecated Use NanobotClient instead */
export type SimpleClient = NanobotClient;
