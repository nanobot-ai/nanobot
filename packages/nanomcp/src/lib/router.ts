import http from "node:http";
import process from "node:process";
import {
	createRouter,
	type RouteMap,
	type Router,
	route,
} from "@remix-run/fetch-router";
import { logger } from "@remix-run/logger-middleware";
import { createRequestListener } from "@remix-run/node-fetch-server";
import { cors } from "./cors";
import { McpRequestHandler } from "./handler.ts";
import type { Config } from "./types.ts";

export class Server {
	readonly router: Router;
	readonly routes: RouteMap;
	readonly config: Config;
	readonly mcp: McpRequestHandler;

	constructor(info: { name: string; version: string }, config: Config) {
		this.mcp = new McpRequestHandler(info, config);
		this.config = config;

		const routes = route({
			mcp: "/mcp",
		});
		this.router = createRouter({
			middleware: [
				cors({
					exposedHeaders: ["Mcp-Session-Id"],
				}),
				...(process.env.NODE_ENV === "development" ? [logger()] : []),
			],
		});
		this.router.map(routes.mcp, this.mcp.handler);
		this.routes = routes;
	}

	fetch: typeof this.router.fetch = (...args) => {
		return this.router.fetch(...args);
	};

	requestListener(): http.RequestListener {
		return createRequestListener(async (request) => {
			try {
				return await this.fetch(request);
			} catch (error) {
				console.error(error);
				return new Response("Internal Server Error", { status: 500 });
			}
		});
	}

	serve = async (port?: number) => {
		port = port || parseInt(process.env.PORT || "9010", 10);
		const server = http.createServer(this.requestListener());

		server.listen(port, () => {
			console.log(`Server is running on http://localhost:${port}`);
		});
	};
}
