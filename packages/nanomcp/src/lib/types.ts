import type { AuthInfo } from "@modelcontextprotocol/sdk/server/auth/types.js";
import type {
	BlobResourceContents,
	CallToolResult,
	CreateMessageRequestParams,
	CreateMessageResult,
	ElicitRequestFormParams,
	ElicitResult,
	ListResourcesResult,
	ListResourceTemplatesResult,
	TextResourceContents,
} from "@modelcontextprotocol/sdk/types.js";
import type { ZodType, z } from "zod";

type RequestedSchema = ElicitRequestFormParams["requestedSchema"];

export type Context = {
	sessionId?: string;
	auth?: AuthInfo;
	threadId?: string;
	workspaceId: string;
	signal: AbortSignal;
	elicit?: (message: string, schema: RequestedSchema) => Promise<ElicitResult>;
	sample?: (params: CreateMessageRequestParams) => Promise<CreateMessageResult>;
};

// Base type for heterogeneous tool collections (type-erased)
export type AnyTool = Tool<ZodType, ZodType>;

// Strongly-typed generic tool (for tool definitions)
export type Tool<
	InputSchema extends ZodType = ZodType,
	OutputSchema extends ZodType | undefined = ZodType | undefined,
> = {
	title?: string;
	description?: string | DescriptionCallback;
	enabled?: EnabledCallback;
	inputSchema: InputSchema;
	outputSchema?: OutputSchema;
	messages?: {
		invoking?: string;
		invoked?: string;
	};
	handler: ToolCallback<z.infer<InputSchema>>;
};

export type ToolCallback<T> = (
	args: T,
	ctx: Context,
) => CallToolResult | Promise<CallToolResult>;

export type EnabledCallback = (ctx: Context) => boolean | Promise<boolean>;
export type DescriptionCallback = (ctx: Context) => string | Promise<string>;

export const createTool = <
	InputSchema extends ZodType,
	OutputSchema extends ZodType,
>(
	config: Tool<InputSchema, OutputSchema>,
) => config;

export type Config = {
	tools?: Record<string, AnyTool>;
	resources?: {
		list?: ListResources;
		read?: ReadResource;
		templates?: ListResourceTemplates;
	};
};

export type ListResources = (
	ctx: Context,
	opts: {
		cursor?: string;
	},
) => Promise<ListResourcesResult>;
export type ReadResource = (
	ctx: Context,
	uri: string,
) => Promise<TextResourceContents | BlobResourceContents>;

export type ListResourceTemplates = (
	ctx: Context,
) => Promise<ListResourceTemplatesResult>;

export const defineConfig = (config: Config) => config;

/**
 * Merge multiple config objects, checking for conflicts
 * @param config - The config objects
 * @returns The merged config
 * @throws Error if there are conflicting tool names or resource handlers
 */
export function mergeConfig(...config: Config[]): Config {
	return config.reduce((acc, curr) => _mergeConfig(acc, curr), {});
}

function _mergeConfig(config1: Config, config2: Config): Config {
	const merged: Config = {};

	// Merge tools
	if (config1.tools || config2.tools) {
		merged.tools = {};

		// Check for tool conflicts
		const tools1Keys = new Set(Object.keys(config1.tools ?? {}));
		const tools2Keys = new Set(Object.keys(config2.tools ?? {}));

		const conflicts: string[] = [];
		for (const key of tools2Keys) {
			if (tools1Keys.has(key)) {
				conflicts.push(key);
			}
		}

		if (conflicts.length > 0) {
			throw new Error(
				`Tool name conflicts found: ${conflicts.join(", ")}. Cannot merge configs with duplicate tool names.`,
			);
		}

		// Merge tools (no conflicts)
		merged.tools = {
			...(config1.tools ?? {}),
			...(config2.tools ?? {}),
		};
	}

	// Merge resources
	if (config1.resources || config2.resources) {
		merged.resources = {};

		// Check for resource handler conflicts
		const hasListConflict = config1.resources?.list && config2.resources?.list;
		const hasReadConflict = config1.resources?.read && config2.resources?.read;
		const hasTemplatesConflict =
			config1.resources?.templates && config2.resources?.templates;

		const resourceConflicts: string[] = [];
		if (hasListConflict) resourceConflicts.push("list");
		if (hasReadConflict) resourceConflicts.push("read");
		if (hasTemplatesConflict) resourceConflicts.push("templates");

		if (resourceConflicts.length > 0) {
			throw new Error(
				`Resource handler conflicts found: ${resourceConflicts.join(", ")}. Cannot merge configs with duplicate resource handlers.`,
			);
		}

		// Merge resources (no conflicts)
		merged.resources = {
			list: config1.resources?.list ?? config2.resources?.list,
			read: config1.resources?.read ?? config2.resources?.read,
			templates: config1.resources?.templates ?? config2.resources?.templates,
		};
	}

	return merged;
}
