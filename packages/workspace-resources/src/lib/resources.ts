import type { Resource } from "@modelcontextprotocol/sdk/types.js";
import type {
	ListResources,
	ListResourceTemplates,
	ReadResource,
} from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import { load } from "js-yaml";

/**
 * Parse YAML frontmatter from markdown content
 */
function parseYAMLFrontMatter(text: string): Record<string, unknown> {
	const match = text.match(/^---\n([\s\S]*?)\n---/);
	if (!match) return {};
	try {
		return (load(match[1]) as Record<string, unknown>) || {};
	} catch {
		return {};
	}
}

// Map file extensions to MIME types for common text/code files
const extensionToMimeType: Record<string, string> = {
	".md": "text/markdown",
	".markdown": "text/markdown",
	".txt": "text/plain",
	".json": "application/json",
	".yaml": "text/yaml",
	".yml": "text/yaml",
	".xml": "application/xml",
	".html": "text/html",
	".htm": "text/html",
	".css": "text/css",
	".js": "text/javascript",
	".ts": "text/typescript",
	".jsx": "text/javascript",
	".tsx": "text/typescript",
	".py": "text/x-python",
	".go": "text/x-go",
	".rs": "text/x-rust",
	".java": "text/x-java",
	".c": "text/x-c",
	".cpp": "text/x-c++",
	".h": "text/x-c",
	".hpp": "text/x-c++",
	".sh": "text/x-shellscript",
	".bash": "text/x-shellscript",
	".zsh": "text/x-shellscript",
	".svg": "image/svg+xml",
	".png": "image/png",
	".jpg": "image/jpeg",
	".jpeg": "image/jpeg",
	".gif": "image/gif",
	".webp": "image/webp",
	".pdf": "application/pdf",
};

function getMimeTypeFromPath(path: string): string {
	const ext = path.toLowerCase().match(/\.[^.]+$/)?.[0] || "";
	return extensionToMimeType[ext] || "application/octet-stream";
}

export const listResource: ListResources = async (ctx, { cursor }) => {
	const client = await ensureConnected(ctx.workspaceId);
	const entries = await client.listDir(".", {
		recursive: true,
		cursor,
		limit: 1000,
	});

	const files = entries.entries.filter((e) => e.isFile);

	// Process files and extract frontmatter for TASK.md files
	const resources = await Promise.all(
		files.map(async (entry): Promise<Resource> => {
			const resource: Resource = {
				uri: `workspace://${entry.name}`,
				name: entry.name,
				mimeType: getMimeTypeFromPath(entry.name),
			};

			// Extract frontmatter metadata for markdown files
			if (entry.name.endsWith(".md")) {
				try {
					const content = await client.readTextFile(`./${entry.name}`, {
						ignoreNotFound: true,
					});
					if (content) {
						const frontMatter = parseYAMLFrontMatter(content);
						if (Object.keys(frontMatter).length > 0) {
							resource._meta = {
								"ai.nanobot": {
									file: frontMatter,
								},
							};
						}
					}
				} catch {
					// Ignore errors reading frontmatter
				}
			}

			return resource;
		}),
	);

	return {
		cursor: entries.cursor,
		resources,
	};
};

export const listResourceTemplates: ListResourceTemplates = async () => {
	return {
		resourceTemplates: [
			{
				name: "workspace",
				title: "Workspace Files",
				uriTemplate: "workspace://{path*}",
			},
		],
	};
};

export const readResource: ReadResource = async (ctx, uri) => {
	if (!uri.startsWith("workspace://")) {
		throw new Error("invalid URI, must start with workspace://");
	}

	let path = uri.replace(/^workspace:\/\/+/, "");
	if (path === "" || path === ".") {
		path = ".";
	} else {
		path = `./${path}`;
	}

	const client = await ensureConnected(ctx.workspaceId);
	const content = await client.readTextFile(path, {
		binary: true,
	});

	// Get MIME type from file extension as fallback
	const fallbackMimeType = getMimeTypeFromPath(path);

	// Parse the data URI to extract mimeType and base64 content
	// Format: data:<mimeType>;base64,<content>
	const dataUriMatch = content.match(/^data:([^;]*);base64,(.+)$/);
	if (!dataUriMatch) {
		// Fallback if not a data URI (shouldn't happen with binary: true)
		return {
			uri: uri,
			mimeType: fallbackMimeType,
			blob: content,
		};
	}

	// Use the data URI mime type if present, otherwise use the fallback
	const mimeType = dataUriMatch[1] || fallbackMimeType;
	const base64Content = dataUriMatch[2];

	try {
		const textContent = Buffer.from(base64Content, "base64").toString("utf-8");
		// Verify it's valid UTF-8 by checking for invalid characters
		if (!textContent.includes("\uFFFD")) {
			return {
				uri: uri,
				mimeType: mimeType,
				text: textContent,
			};
		}
	} catch {
		// If decoding fails, fall through to return as blob
	}

	// Return as blob for binary content
	return {
		uri: uri,
		mimeType: mimeType,
		blob: base64Content,
	};
};
