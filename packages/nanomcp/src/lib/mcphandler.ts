import { randomUUID } from "node:crypto";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import {
	type Implementation,
	isInitializeRequest,
} from "@modelcontextprotocol/sdk/types.js";
import type { RequestHandler } from "@remix-run/fetch-router";
import { toFetchResponse, toReqRes } from "fetch-to-node";

export function createMcpHandler(
	serverInfo: { name: string; version: string },
	setupServer: SetupServer,
): RequestHandler {
	const middleware = new McpHandler(serverInfo, setupServer);
	return (x) => middleware.handle(x);
}

type SetupServer = (opts: {
	server: McpServer;
	request: Request;
	sessionId?: string;
}) => void | Promise<void>;

class McpHandler {
	readonly #transports: Record<string, StreamableHTTPServerTransport> = {};
	readonly #setupServer: SetupServer;
	readonly #serverInfo: Implementation;

	constructor(
		serverInfo: { name: string; version: string },
		setup: SetupServer,
	) {
		this.#setupServer = setup;
		this.#serverInfo = serverInfo;
	}

	handle: RequestHandler = ({ request }) => {
		if (request.method === "GET") {
			return this.getOrDelete(request);
		} else if (request.method === "DELETE") {
			return this.getOrDelete(request);
		} else if (request.method === "POST") {
			return this.post(request);
		}
		return new Response("Method not allowed", { status: 405 });
	};

	private newTransport = async (request: Request, sessionId?: string) => {
		let transport: StreamableHTTPServerTransport;
		if (sessionId) {
			transport = new StreamableHTTPServerTransport({
				sessionIdGenerator: undefined,
			});
			transport.sessionId = sessionId;
			this.#transports[sessionId] = transport;
			transport.onclose = () => {
				if (transport.sessionId) {
					delete this.#transports[transport.sessionId];
				}
			};
		} else {
			transport = new StreamableHTTPServerTransport({
				sessionIdGenerator: () => randomUUID(),
				onsessioninitialized: (sessionId) => {
					this.#transports[sessionId] = transport;
				},
			});
		}

		transport.onclose = () => {
			if (transport.sessionId) {
				delete this.#transports[transport.sessionId];
			}
		};

		const mcpServer = new McpServer(this.#serverInfo);
		await this.#setupServer({
			server: mcpServer,
			request,
			sessionId,
		});
		await mcpServer.connect(transport);
		return transport;
	};

	private post = async (request: Request) => {
		const sessionId = request.headers.get("mcp-session-id") || undefined;

		let transport: StreamableHTTPServerTransport;
		let parsedBody: unknown;
		const { req, res } = toReqRes(request);
		// read the full body into memory from the req (In

		if (sessionId && this.#transports[sessionId]) {
			transport = this.#transports[sessionId];
		} else {
			parsedBody = await request.json();
			if (isInitializeRequest(parsedBody) || validUUID(sessionId)) {
				transport = await this.newTransport(request, sessionId);
			} else {
				return Response.json(
					{
						jsonrpc: "2.0",
						error: {
							code: -32000,
							message: "Bad Request: No valid session ID provided",
						},
						id: null,
					},
					{
						status: 404,
					},
				);
			}
		}

		await transport.handleRequest(req, res, parsedBody);
		return await toFetchResponse(res);
	};

	private getOrDelete = async (request: Request) => {
		const sessionId = request.headers.get("mcp-session-id");
		const transport = sessionId && this.#transports[sessionId];

		if (!transport) {
			return new Response(
				`Invalid or missing session ID: ${sessionId ? `${sessionId.slice(0, 8)}...` : "none"}`,
				{ status: 404 },
			);
		}
		const { req, res } = toReqRes(request);
		await transport.handleRequest(req, res);
		return await toFetchResponse(res);
	};
}

function validUUID(id?: string): boolean {
	if (!id) return false;
	return /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(
		id,
	);
}
