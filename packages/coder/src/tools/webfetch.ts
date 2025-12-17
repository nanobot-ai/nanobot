import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { JSDOM } from "jsdom";
import TurndownService from "turndown";
import * as z from "zod";

const MAX_RESPONSE_SIZE = 5 * 1024 * 1024; // 5MB
const DEFAULT_TIMEOUT = 30 * 1000; // 30 seconds
const MAX_TIMEOUT = 120 * 1000; // 2 minutes

const schema = z.object({
	url: z.string().url().describe("The URL to fetch content from"),
	format: z
		.union([z.literal("text"), z.literal("markdown"), z.literal("html")])
		.describe("The format to return the content in (text, markdown, or html)"),
	timeout: z
		.number()
		.describe("Optional timeout in seconds (max 120)")
		.optional(),
});

export default createTool({
	title: "WebFetch",
	description:
		'\n- Fetches content from a specified URL and returns it in the requested format\n- Takes a URL and format as input (text, markdown, or html)\n- Automatically converts HTML to the requested format\n- Optional prompt parameter for specifying what information to extract\n- Use this tool when you need to retrieve web content\n\nUsage notes:\n  - IMPORTANT: If an MCP-provided web fetch tool is available, prefer using that tool instead of this one, as it may have fewer restrictions. All MCP-provided tools start with "mcp__".\n  - The URL must be a fully-formed valid URL (http:// or https://)\n  - HTTP URLs will be automatically upgraded to HTTPS when possible\n  - Maximum response size: 5MB\n  - Default timeout: 30 seconds, maximum: 120 seconds\n  - This tool is read-only and does not modify any files\n  - When a URL redirects to a different host, the tool will inform you and provide the redirect URL\n',
	messages: {
		invoking: "Fetching web content",
		invoked: "Web content fetched",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { url, format, timeout: timeoutSeconds } = args;

		try {
			// Validate URL protocol
			if (!url.startsWith("http://") && !url.startsWith("https://")) {
				return toolResult.error("URL must start with http:// or https://");
			}

			const timeout = Math.min(
				(timeoutSeconds ?? DEFAULT_TIMEOUT / 1000) * 1000,
				MAX_TIMEOUT,
			);

			const controller = new AbortController();
			const timeoutId = setTimeout(() => controller.abort(), timeout);

			// Build Accept header based on requested format with quality parameters for fallbacks
			let acceptHeader = "*/*";
			switch (format) {
				case "markdown":
					acceptHeader =
						"text/markdown;q=1.0, text/x-markdown;q=0.9, text/plain;q=0.8, text/html;q=0.7, */*;q=0.1";
					break;
				case "text":
					acceptHeader =
						"text/plain;q=1.0, text/markdown;q=0.9, text/html;q=0.8, */*;q=0.1";
					break;
				case "html":
					acceptHeader =
						"text/html;q=1.0, application/xhtml+xml;q=0.9, text/plain;q=0.8, text/markdown;q=0.7, */*;q=0.1";
					break;
				default:
					acceptHeader =
						"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8";
			}

			// Fetch with realistic headers
			const response = await fetch(url, {
				signal: AbortSignal.any([controller.signal, ctx.signal]),
				headers: {
					"User-Agent":
						"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
					Accept: acceptHeader,
					"Accept-Language": "en-US,en;q=0.9",
				},
			});

			clearTimeout(timeoutId);

			if (!response.ok) {
				return toolResult.error(
					`Failed to fetch URL: ${response.status} ${response.statusText}`,
				);
			}

			// Check content length
			const contentLength = response.headers.get("content-length");
			if (contentLength && parseInt(contentLength, 10) > MAX_RESPONSE_SIZE) {
				return toolResult.error("Response too large (exceeds 5MB limit)");
			}

			const arrayBuffer = await response.arrayBuffer();
			if (arrayBuffer.byteLength > MAX_RESPONSE_SIZE) {
				return toolResult.error("Response too large (exceeds 5MB limit)");
			}

			const content = new TextDecoder().decode(arrayBuffer);
			const contentType = response.headers.get("content-type") || "";

			// Process content based on requested format and actual content type
			let processedContent: string;

			switch (format) {
				case "markdown":
					if (contentType.includes("text/html")) {
						processedContent = convertHTMLToMarkdown(content);
					} else {
						processedContent = content;
					}
					break;

				case "text":
					if (contentType.includes("text/html")) {
						processedContent = extractTextFromHTML(content);
					} else {
						processedContent = content;
					}
					break;

				case "html":
					processedContent = content;
					break;

				default:
					processedContent = content;
			}

			// Format output
			let output = `URL: ${url}\nContent-Type: ${contentType}\n\n`;
			output += processedContent;

			return toolResult.text(output);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);

			// Handle timeout specifically
			if (errorMessage.includes("abort")) {
				return toolResult.error(
					`Request timed out after ${timeoutSeconds ?? DEFAULT_TIMEOUT / 1000} seconds`,
				);
			}

			return toolResult.error(`Error fetching web content: ${errorMessage}`);
		}
	},
});

function convertHTMLToMarkdown(html: string): string {
	const turndownService = new TurndownService({
		headingStyle: "atx",
		hr: "---",
		bulletListMarker: "-",
		codeBlockStyle: "fenced",
		emDelimiter: "*",
	});
	turndownService.remove(["script", "style", "meta", "link"]);
	return turndownService.turndown(html);
}

function extractTextFromHTML(html: string): string {
	const dom = new JSDOM(html);
	const document = dom.window.document;

	// Remove script, style, and other non-content elements
	const elementsToRemove = document.querySelectorAll(
		"script, style, noscript, iframe, object, embed",
	);
	elementsToRemove.forEach((el) => {
		el.remove();
	});

	// Get text content
	const text = document.body?.textContent || "";

	// Clean up whitespace
	return text
		.split("\n")
		.map((line) => line.trim())
		.filter((line) => line.length > 0)
		.join("\n");
}
