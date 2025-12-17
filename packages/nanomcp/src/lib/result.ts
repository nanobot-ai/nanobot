import type { CallToolResult } from "@modelcontextprotocol/sdk/types.js";

export namespace toolResult {
	export function structured<T extends Record<string, unknown>>(
		msg: string,
		obj: T,
	): CallToolResult;
	export function structured<T extends Record<string, unknown>>(
		obj: T,
	): CallToolResult;
	export function structured<T extends Record<string, unknown> | undefined>(
		first: T | string,
		second?: T,
	): CallToolResult {
		const msg: string | undefined =
			typeof first === "string" ? first : undefined;
		const obj: T = typeof first === "string" ? (second as T) : first;
		const text = msg === undefined ? JSON.stringify(obj) : msg;
		return {
			content: [
				{
					type: "text",
					text,
				},
			],
			structuredContent: obj,
		};
	}

	export function text(...text: string[]): CallToolResult {
		return {
			content: text.map((t) => ({ type: "text", text: t })),
		};
	}

	export function error(
		text: string,
		structured?: Record<string, unknown>,
	): CallToolResult & { isError: true } {
		return {
			isError: true,
			content: [
				{
					type: "text",
					text,
				},
			],
			structuredContent: structured,
		};
	}
}
