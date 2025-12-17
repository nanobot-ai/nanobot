import type { Resource } from "@modelcontextprotocol/sdk/types.js";
import type {
	ListResources,
	ListResourceTemplates,
	ReadResource,
} from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";

export const listResource: ListResources = async (ctx, { cursor }) => {
	const client = await ensureConnected(ctx.workspaceId);
	const entries = await client.listDir(".", {
		recursive: true,
		cursor,
		limit: 1000,
	});
	return {
		cursor: entries.cursor,
		resources: entries.entries
			.filter((e) => e.isFile)
			.map(
				(entry) =>
					({
						uri: `workspace://${entry.name}`,
						name: entry.name,
						mimeType: "application/octet-stream",
					}) satisfies Resource,
			),
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

	// Parse the data URI to extract mimeType and base64 content
	// Format: data:<mimeType>;base64,<content>
	const dataUriMatch = content.match(/^data:([^;]+);base64,(.+)$/);
	if (!dataUriMatch) {
		// Fallback if not a data URI (shouldn't happen with binary: true)
		return {
			uri: uri,
			blob: content,
		};
	}

	const mimeType = dataUriMatch[1];
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
