import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { executeBashCommand } from "../lib/shell.js";

const schema = z.object({
	pattern: z
		.string()
		.describe("The regular expression pattern to search for in file contents"),
	path: z
		.string()
		.describe(
			"File or directory to search in (rg PATH). Defaults to current working directory.",
		)
		.optional(),
	glob: z
		.string()
		.describe(
			'Glob pattern to filter files (e.g. "*.js", "*.{ts,tsx}") - maps to rg --glob',
		)
		.optional(),
	output_mode: z
		.union([
			z.literal("content"),
			z.literal("files_with_matches"),
			z.literal("count"),
		])
		.describe(
			'Output mode: "content" shows matching lines (supports -A/-B/-C context, -n line numbers, head_limit), "files_with_matches" shows file paths (supports head_limit), "count" shows match counts (supports head_limit). Defaults to "files_with_matches".',
		)
		.optional()
		.default("files_with_matches"),
	"-B": z
		.number()
		.describe(
			'Number of lines to show before each match (rg -B). Requires output_mode: "content", ignored otherwise.',
		)
		.optional(),
	"-A": z
		.number()
		.describe(
			'Number of lines to show after each match (rg -A). Requires output_mode: "content", ignored otherwise.',
		)
		.optional(),
	"-C": z
		.number()
		.describe(
			'Number of lines to show before and after each match (rg -C). Requires output_mode: "content", ignored otherwise.',
		)
		.optional(),
	"-n": z
		.boolean()
		.describe(
			'Show line numbers in output (rg -n). Requires output_mode: "content", ignored otherwise. Defaults to true.',
		)
		.optional()
		.default(true),
	"-i": z
		.boolean()
		.describe("Case insensitive search (rg -i)")
		.optional()
		.default(false),
	type: z
		.string()
		.describe(
			"File type to search (rg --type). Common types: js, py, rust, go, java, etc. More efficient than include for standard file types.",
		)
		.optional(),
	head_limit: z
		.number()
		.describe(
			'Limit output to first N lines/entries, equivalent to "| head -N". Works across all output modes: content (limits output lines), files_with_matches (limits file paths), count (limits count entries). Defaults based on "cap" experiment value: 0 (unlimited), 20, or 100.',
		)
		.optional(),
	offset: z
		.number()
		.describe(
			'Skip first N lines/entries before applying head_limit, equivalent to "| tail -n +N | head -N". Works across all output modes. Defaults to 0.',
		)
		.optional()
		.default(0),
	multiline: z
		.boolean()
		.describe(
			"Enable multiline mode where . matches newlines and patterns can span lines (rg -U --multiline-dotall). Default: false.",
		)
		.optional()
		.default(false),
});

export default createTool({
	title: "Grep",
	description:
		'A powerful search tool built on ripgrep\n\n  Usage:\n  - ALWAYS use Grep for search tasks. NEVER invoke `grep` or `rg` as a Bash command. The Grep tool has been optimized for correct permissions and access.\n  - Supports full regex syntax (e.g., "log.*Error", "function\\s+\\w+")\n  - Filter files with glob parameter (e.g., "*.js", "**/*.tsx") or type parameter (e.g., "js", "py", "rust")\n  - Output modes: "content" shows matching lines, "files_with_matches" shows only file paths (default), "count" shows match counts\n  - Use Task tool for open-ended searches requiring multiple rounds\n  - Pattern syntax: Uses ripgrep (not grep) - literal braces need escaping (use `interface\\{\\}` to find `interface{}` in Go code)\n  - Multiline matching: By default patterns match within single lines only. For cross-line patterns like `struct \\{[\\s\\S]*?field`, use `multiline: true`\n',
	messages: {
		invoking: "Searching content",
		invoked: "Content search complete",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const {
			pattern,
			path,
			glob,
			output_mode,
			"-B": beforeContext,
			"-A": afterContext,
			"-C": context,
			"-n": showLineNumbers,
			"-i": caseInsensitive,
			type,
			head_limit,
			offset,
			multiline,
		} = args;

		try {
			const client = await ensureConnected(ctx.workspaceId);

			// Build ripgrep command with JSON output for structured parsing
			const rgArgs: string[] = ["rg", "--json"];

			// Add pattern (properly escaped)
			rgArgs.push(JSON.stringify(pattern));

			// Context (only in content mode)
			if (output_mode === "content") {
				if (context !== undefined) {
					rgArgs.push(`-C ${context}`);
				} else {
					if (beforeContext !== undefined) {
						rgArgs.push(`-B ${beforeContext}`);
					}
					if (afterContext !== undefined) {
						rgArgs.push(`-A ${afterContext}`);
					}
				}
			}

			// Case insensitive
			if (caseInsensitive) {
				rgArgs.push("-i");
			}

			// Multiline
			if (multiline) {
				rgArgs.push("-U --multiline-dotall");
			}

			// File type
			if (type) {
				rgArgs.push(`--type ${type}`);
			}

			// Glob pattern
			if (glob) {
				rgArgs.push(`--glob ${JSON.stringify(glob)}`);
			}

			// Path
			if (path) {
				rgArgs.push(JSON.stringify(path));
			}

			// Build full command
			const command = `${rgArgs.join(" ")} || true`;

			const result = await executeBashCommand(client, command);

			if (!result.output.trim()) {
				return toolResult.text("No matches found");
			}

			// Parse JSON output and format based on output_mode
			const lines = result.output.trim().split("\n");
			const matches: Array<{
				file: string;
				line?: number;
				text?: string;
				count?: number;
			}> = [];

			let currentFile = "";

			for (const line of lines) {
				try {
					const parsed = JSON.parse(line);

					if (parsed.type === "begin") {
						currentFile = parsed.data?.path?.text || "";
					} else if (parsed.type === "match") {
						const lineNumber = parsed.data?.line_number;
						const lineText = parsed.data?.lines?.text;

						if (output_mode === "content") {
							matches.push({
								file: currentFile,
								line: lineNumber,
								text: lineText,
							});
						} else if (output_mode === "files_with_matches") {
							// Add file only once
							if (!matches.some((m) => m.file === currentFile)) {
								matches.push({ file: currentFile });
							}
						}
					} else if (parsed.type === "summary") {
						// For count mode, we can get stats from summary
						if (output_mode === "count") {
							const stats = parsed.data?.stats;
							if (stats) {
								// Count mode shows file:count format
								// We'll accumulate during match processing instead
							}
						}
					}
				} catch (_parseError) {}
			}

			// Apply offset and limit
			let filteredMatches = matches;
			if (offset && offset > 0) {
				filteredMatches = filteredMatches.slice(offset);
			}
			if (head_limit !== undefined && head_limit > 0) {
				filteredMatches = filteredMatches.slice(0, head_limit);
			}

			if (filteredMatches.length === 0) {
				return toolResult.text("No matches found");
			}

			// Format output based on mode
			let output: string;

			if (output_mode === "content") {
				// Format: file:line:text
				output = filteredMatches
					.map((m) => {
						if (showLineNumbers && m.line) {
							return `${m.file}:${m.line}:${m.text}`;
						}
						return `${m.file}:${m.text}`;
					})
					.join("\n");
			} else if (output_mode === "files_with_matches") {
				// Format: just file paths
				output = filteredMatches.map((m) => m.file).join("\n");
			} else if (output_mode === "count") {
				// Format: file:count
				// Count matches per file
				const countMap = new Map<string, number>();
				for (const match of matches) {
					countMap.set(match.file, (countMap.get(match.file) || 0) + 1);
				}
				output = Array.from(countMap.entries())
					.map(([file, count]) => `${file}:${count}`)
					.join("\n");
			} else {
				output = filteredMatches.map((m) => m.file).join("\n");
			}

			return toolResult.text(output);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error searching content: ${errorMessage}`);
		}
	},
});
