import { spawn } from "node:child_process";
import { ensureConnected } from "@nanobot-ai/workspace-client";

/**
 * File extensions supported by markitdown conversion
 */
const MARKITDOWN_EXTENSIONS = new Set([
	".pdf",
	".docx",
	".pptx",
	".xlsx",
	".xls",
	".jpg",
	".jpeg",
	".png",
	".wav",
	".mp3",
	".html",
	".zip",
	".csv",
]);

/**
 * Options for reading a text file with optional conversion
 */
export interface ReadTextFileOptions {
	/**
	 * Line number to start reading from (0-based)
	 */
	line?: number;

	/**
	 * Maximum number of lines to read
	 */
	limit?: number;

	/**
	 * If true, convert file to markdown using markitdown CLI
	 * Only applies to supported file extensions (pdf, docx, xlsx, pptx, etc.)
	 */
	convert?: boolean;
}

/**
 * Check if a file extension is supported by markitdown
 */
function isMarkitdownSupported(filePath: string): boolean {
	const ext = filePath.toLowerCase().match(/\.[^.]+$/)?.[0];
	return ext ? MARKITDOWN_EXTENSIONS.has(ext) : false;
}

/**
 * Convert a file to markdown using markitdown CLI
 */
async function convertWithMarkitdown(filePath: string): Promise<string> {
	return new Promise((resolve, reject) => {
		const markitdown = spawn("markitdown", [filePath]);
		let stdout = "";
		let stderr = "";

		markitdown.stdout.on("data", (data) => {
			stdout += data.toString();
		});

		markitdown.stderr.on("data", (data) => {
			stderr += data.toString();
		});

		markitdown.on("close", (code) => {
			if (code !== 0) {
				reject(
					new Error(
						`markitdown failed with code ${code}: ${stderr || "Unknown error"}`,
					),
				);
			} else {
				resolve(stdout);
			}
		});

		markitdown.on("error", (err) => {
			reject(new Error(`Failed to spawn markitdown: ${err.message}`));
		});
	});
}

/**
 * Apply line offset and limit to text content
 */
function applyLineFiltering(
	content: string,
	options?: { line?: number; limit?: number },
): string {
	if (!options?.line && !options?.limit) {
		return content;
	}

	const lines = content.split("\n");
	const start = options.line ?? 0;
	const end = options.limit ? start + options.limit : lines.length;

	return lines.slice(start, end).join("\n");
}

/**
 * Enhanced readTextFile with optional markitdown conversion
 *
 * @param workspaceId - The workspace identifier
 * @param path - Path to the file to read
 * @param options - Optional reading options (line offset, limit, and convert flag)
 * @param url - Optional workspace server URL
 * @returns The file contents as a string
 *
 * @example
 * ```typescript
 * // Read a PDF file and convert to markdown
 * const markdown = await readTextFile(
 *   "workspace-1",
 *   "/path/to/document.pdf",
 *   { convert: true }
 * );
 *
 * // Read specific lines from a converted Excel file
 * const partial = await readTextFile(
 *   "workspace-1",
 *   "/path/to/spreadsheet.xlsx",
 *   { convert: true, line: 10, limit: 20 }
 * );
 * ```
 */
export async function readTextFile(
	workspaceId: string,
	path: string,
	options?: ReadTextFileOptions,
): Promise<string> {
	const client = await ensureConnected(workspaceId);

	// Check if conversion is requested and supported
	if (options?.convert && isMarkitdownSupported(path)) {
		try {
			// Convert file using markitdown
			const markdown = await convertWithMarkitdown(path);

			// Apply line filtering if specified
			return applyLineFiltering(markdown, options);
		} catch (error) {
			// If markitdown fails, fall back to regular file reading
			console.warn(
				`Markitdown conversion failed for ${path}, falling back to regular read:`,
				error,
			);
			return client.readTextFile(path, {
				line: options?.line,
				limit: options?.limit,
			});
		}
	}

	// Regular file reading without conversion
	return client.readTextFile(path, {
		line: options?.line,
		limit: options?.limit,
	});
}
