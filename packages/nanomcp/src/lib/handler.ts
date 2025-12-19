import type { McpServer as McpSdkServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { RequestHandlerExtra } from "@modelcontextprotocol/sdk/shared/protocol.js";
import {
	type Implementation,
	ListResourcesRequestSchema,
	type ListResourcesResult,
	ListResourceTemplatesRequestSchema,
	type ListResourceTemplatesResult,
	ReadResourceRequestSchema,
	type ReadResourceResult,
	type ServerNotification,
	type ServerRequest,
} from "@modelcontextprotocol/sdk/types.js";
import type { RequestHandler } from "@remix-run/fetch-router";
import { createMcpHandler } from "./mcphandler.ts";
import type { AnyTool, Config, Context } from "./types.ts";

export class McpRequestHandler {
	readonly #routeConfig: Config;
	readonly #handler: RequestHandler;

	constructor(serverInfo: Implementation, config: Config) {
		this.#routeConfig = config;

		this.#handler = createMcpHandler(
			serverInfo,
			({ server, request, sessionId }) => {
				return this.#setupServer(server, request, sessionId);
			},
		);
	}

	handler: RequestHandler = async (ctx) => {
		return this.#handler(ctx);
	};

	async #checkEnabled(ctx: Context, tool: AnyTool) {
		try {
			return await (tool.enabled ?? (() => true))(ctx);
		} catch (e: unknown) {
			console.log(
				`Error checking if tool is enabled: ${e instanceof Error ? e.message : String(e)}`,
			);
			return false;
		}
	}

	async #getDescription(ctx: Context, toolName: string, tool: AnyTool) {
		if (typeof tool.description === "string") {
			return tool.description;
		}
		try {
			return await (tool.description ?? (() => undefined))(ctx);
		} catch (e: unknown) {
			console.log(
				`Error getting description for tool: ${e instanceof Error ? e.message : String(e)}`,
			);
			return `Tool: ${tool.title ?? toolName}`;
		}
	}

	async #setupResources(baseCtx: Context, server: McpSdkServer) {
		const cfg = this.#routeConfig.resources;
		if (!cfg || (!cfg.list && !cfg.read)) {
			return;
		}
		server.server.registerCapabilities({
			resources: {},
		});

		if (cfg.list) {
			server.server.setRequestHandler(
				ListResourcesRequestSchema,
				async (request, ctx): Promise<ListResourcesResult> => {
					try {
						if (!cfg.list) {
							return { resources: [] };
						}
						return cfg.list(this.#toContext(baseCtx, server, ctx), {
							cursor: request.params?.cursor,
						});
					} catch (e: unknown) {
						console.log(
							`Uncaught exception listing resources ${e instanceof Error ? e.message : String(e)}`,
						);
						throw e;
					}
				},
			);
		}

		if (cfg.read) {
			server.server.setRequestHandler(
				ReadResourceRequestSchema,
				async (request, ctx): Promise<ReadResourceResult> => {
					try {
						const content = await cfg.read?.(
							this.#toContext(baseCtx, server, ctx),
							request.params.uri,
						);
						if (!content) {
							return { contents: [] };
						}
						return {
							contents: [content],
						};
					} catch (e: unknown) {
						console.log(
							`Uncaught exception reading resource ${e instanceof Error ? e.message : String(e)}`,
						);
						throw e;
					}
				},
			);
		}

		if (cfg.templates) {
			server.server.setRequestHandler(
				ListResourceTemplatesRequestSchema,
				async (_, ctx): Promise<ListResourceTemplatesResult> => {
					return (
						cfg.templates?.(this.#toContext(baseCtx, server, ctx)) ?? {
							resourceTemplates: [],
						}
					);
				},
			);
		}
	}

	#toContext(
		baseCtx: Context,
		server: McpSdkServer,
		extra: RequestHandlerExtra<ServerRequest, ServerNotification>,
	): Context {
		return {
			...baseCtx,
			auth: extra.authInfo,
			elicit: async (message, schema) => {
				return await server.server.elicitInput({
					message,
					requestedSchema: schema,
				});
			},
			sample: (...args) => {
				return server.server.createMessage(...args);
			},
		};
	}

	async #setupTools(baseCtx: Context, server: McpSdkServer) {
		// Register each tool
		for (const [name, tool] of Object.entries(this.#routeConfig.tools ?? {})) {
			const enabled = await this.#checkEnabled(baseCtx, tool);
			if (!enabled) {
				continue;
			}

			const description = await this.#getDescription(baseCtx, name, tool);

			server.registerTool(
				name,
				{
					description,
					inputSchema: tool.inputSchema,
					outputSchema: tool.outputSchema,
				},
				// biome-ignore lint/suspicious/noExplicitAny: Bridging AnyTool (type-erased) to MCP SDK callback, both use any
				(args: any, ctx) => {
					try {
						return tool.handler(args, this.#toContext(baseCtx, server, ctx));
					} catch (e: unknown) {
						console.log(
							`Uncaught exception ${e instanceof Error ? e.message : String(e)}`,
						);
						throw e;
					}
				},
			);
		}
	}

	async #setupServer(
		server: McpSdkServer,
		req: Request,
		sessionIdRaw: string | undefined,
	) {
		const sessionId = sessionIdRaw ?? crypto.randomUUID();
		const baseCtx: Context = {
			sessionId,
			workspaceId: req.headers.get("x-nanobot-workspace-id") || sessionId,
			signal: req.signal,
		};

		console.log(`New session: ${sessionId}`);
		if (req.headers.get("x-nanobot-workspace-id")) {
			console.log(
				`Using workspace ID from header: ${req.headers.get("x-nanobot-workspace-id")}`,
			);
		}

		await this.#setupTools(baseCtx, server);
		await this.#setupResources(baseCtx, server);
	}
}
