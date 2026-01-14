import { randomUUID } from "node:crypto";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { WebStandardStreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/webStandardStreamableHttp.js";
import {
	type Implementation,
	isInitializeRequest,
} from "@modelcontextprotocol/sdk/types.js";
import type { RequestHandler } from "@remix-run/fetch-router";

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
	readonly #transports: Record<
		string,
		WebStandardStreamableHTTPServerTransport
	> = {};
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

	private newTransport = async (request: Request) => {
		const sessionId = randomUUID();
		const transport = new WebStandardStreamableHTTPServerTransport({
			sessionIdGenerator: () => sessionId,
			onsessioninitialized: (sessionId) => {
				this.#transports[sessionId] = transport;
			},
			onsessionclosed: (closedSessionId) => {
				delete this.#transports[closedSessionId];
			},
		});

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

		let transport: WebStandardStreamableHTTPServerTransport;
		let parsedBody: unknown;

		if (sessionId) {
			// Session ID provided - must exist in memory
			transport = this.#transports[sessionId];
			if (!transport) {
				console.error("[MCP] POST: Session not found:", sessionId);
				return Response.json(
					{
						jsonrpc: "2.0",
						error: {
							code: -32000,
							message: "Session not found",
						},
						id: null,
					},
					{
						status: 404,
					},
				);
			}
		} else {
			// No session ID - must be an initialize request
			parsedBody = await request.json();
			if (isInitializeRequest(parsedBody)) {
				transport = await this.newTransport(request);
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
						status: 400,
					},
				);
			}
		}

		const response = await transport.handleRequest(request, { parsedBody });
		return response;
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
		return await transport.handleRequest(request);
	};
}
