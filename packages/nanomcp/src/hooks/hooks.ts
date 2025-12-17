/**
 * TypeScript definitions for Nanobot Agent Hooks
 * Generated from pkg/types/hooks.go and related types
 */

import { z } from "zod";

// ============================================================================
// Helper Types (defined first due to dependencies)
// ============================================================================

export const AgentReasoningSchema = z.object({
	effort: z.string().optional(),
	summary: z.string().optional(),
});

export type AgentReasoning = z.infer<typeof AgentReasoningSchema>;

export const DynamicInstructionsSchema = z.object({
	// Can be either a simple string (instructions) or an object with MCPServer/Prompt
	instructions: z.string().optional(),
	mcpServer: z.string().optional(),
	prompt: z.string().optional(),
	args: z.record(z.string()).optional(),
});

export type DynamicInstructions = z.infer<typeof DynamicInstructionsSchema>;

export const FieldSchema: z.ZodType<{
	description?: string;
	fields?: Record<string, unknown>;
	required?: boolean;
}> = z.lazy(() =>
	z.object({
		description: z.string().optional(),
		fields: z.record(FieldSchema).optional(),
		required: z.boolean().optional(),
	}),
);

export type Field = z.infer<typeof FieldSchema>;

export const OutputSchemaSchema = z.object({
	name: z.string().optional(),
	description: z.string().optional(),
	schema: z.any().optional(), // json.RawMessage in Go
	strict: z.boolean().optional(),
	fields: z.record(FieldSchema).optional(),
});

export type OutputSchema = z.infer<typeof OutputSchemaSchema>;

export const AgentSchema = z.object({
	name: z.string().optional(),
	shortName: z.string().optional(),
	description: z.string().optional(),
	icon: z.string().optional(),
	iconDark: z.string().optional(),
	starterMessages: z.array(z.string()).optional(),
	instructions: z.union([DynamicInstructionsSchema, z.string()]).optional(),
	model: z.string().optional(),
	before: z.array(z.string()).optional(),
	after: z.array(z.string()).optional(),
	mcpServers: z.array(z.string()).optional(),
	workspaceId: z.string().optional(),
	workspaceBaseUri: z.string().optional(),
	tools: z.array(z.string()).optional(),
	agents: z.array(z.string()).optional(),
	flows: z.array(z.string()).optional(),
	prompts: z.array(z.string()).optional(),
	resources: z.array(z.string()).optional(),
	reasoning: AgentReasoningSchema.optional(),
	threadName: z.string().optional(),
	chat: z.boolean().optional(),
	toolExtensions: z.record(z.record(z.any())).optional(),
	toolChoice: z.string().optional(),
	temperature: z.number().optional(),
	topP: z.number().optional(),
	output: OutputSchemaSchema.optional(),
	truncation: z.string().optional(),
	maxTokens: z.number().optional(),
	mimeTypes: z.array(z.string()).optional(),
	hooks: z.any().optional(), // mcp.Hooks type - keeping as any for flexibility

	// Selection criteria fields
	aliases: z.array(z.string()).optional(),
	cost: z.number().optional(),
	speed: z.number().optional(),
	intelligence: z.number().optional(),
});

export type Agent = z.infer<typeof AgentSchema>;

export const ContentSchema = z.object({
	type: z.enum(["text", "image", "audio", "resource"]),
	text: z.string().optional(),
	data: z.string().optional(),
	mimeType: z.string().optional(),
	uri: z.string().optional(),
	annotations: z.any().optional(),
});

export type Content = z.infer<typeof ContentSchema>;

export const SummaryTextSchema = z.object({
	text: z.string().optional(),
});

export type SummaryText = z.infer<typeof SummaryTextSchema>;

export const ReasoningSchema = z.object({
	encryptedContent: z.string().optional(),
	summary: z.array(SummaryTextSchema).optional(),
});

export type Reasoning = z.infer<typeof ReasoningSchema>;

export const ToolCallSchema = z.object({
	arguments: z.string().optional(),
	callID: z.string().optional(),
	name: z.string().optional(),
	target: z.string().optional(),
	targetType: z.string().optional(),
});

export type ToolCall = z.infer<typeof ToolCallSchema>;

export const CallResultSchema = z.object({
	content: z.array(ContentSchema).optional(),
	isError: z.boolean().optional(),
	chatResponse: z.boolean().optional(),
	agent: z.string().optional(),
	model: z.string().optional(),
	stopReason: z.string().optional(),
	structuredContent: z.any().optional(),
});

export type CallResult = z.infer<typeof CallResultSchema>;

export const ToolCallResultSchema = z.object({
	outputRole: z.string().optional(),
	callID: z.string().optional(),
	output: CallResultSchema.optional(),
});

export type ToolCallResult = z.infer<typeof ToolCallResultSchema>;

export const CompletionItemSchema = z.object({
	id: z.string().optional(),
	partial: z.boolean().optional(),
	hasMore: z.boolean().optional(),
	content: ContentSchema.optional(),
	toolCall: ToolCallSchema.optional(),
	toolCallResult: ToolCallResultSchema.optional(),
	reasoning: ReasoningSchema.optional(),
});

export type CompletionItem = z.infer<typeof CompletionItemSchema>;

export const MessageSchema = z.object({
	id: z.string().optional(),
	created: z.string().optional(), // ISO date string (time.Time in Go)
	role: z.string().optional(),
	items: z.array(CompletionItemSchema).optional(),
	hasMore: z.boolean().optional(),
});

export type Message = z.infer<typeof MessageSchema>;

export const ToolUseDefinitionSchema = z.object({
	name: z.string().optional(),
	parameters: z.any().optional(), // json.RawMessage in Go
	description: z.string().optional(),
	attributes: z.record(z.any()).optional(), // Not included in JSON (json:"-")
});

export type ToolUseDefinition = z.infer<typeof ToolUseDefinitionSchema>;

export const CompletionRequestSchema = z.object({
	model: z.string().optional(),
	agent: z.string().optional(),
	threadName: z.string().optional(),
	newThread: z.boolean().optional(),
	input: z.array(MessageSchema).optional(),
	modelPreferences: z.any().optional(), // mcp.ModelPreferences
	systemPrompt: z.string().optional(),
	maxTokens: z.number().optional(),
	toolChoice: z.string().optional(),
	outputSchema: OutputSchemaSchema.optional(),
	temperature: z.number().optional(),
	truncation: z.string().optional(),
	topP: z.number().optional(),
	metadata: z.record(z.any()).optional(),
	tools: z.array(ToolUseDefinitionSchema).optional(),
	reasoning: AgentReasoningSchema.optional(),
});

export type CompletionRequest = z.infer<typeof CompletionRequestSchema>;

export const CompletionResponseSchema = z.object({
	output: MessageSchema.optional(),
	internalMessages: z.array(MessageSchema).optional(),
	chatResponse: z.boolean().optional(),
	agent: z.string().optional(),
	model: z.string().optional(),
	hasMore: z.boolean().optional(),
	error: z.string().optional(),
	progressToken: z.any().optional(),
});

export type CompletionResponse = z.infer<typeof CompletionResponseSchema>;

export const CompletionProgressSchema = z.object({
	model: z.string().optional(),
	agent: z.string().optional(),
	messageID: z.string().optional(),
	role: z.string().optional(),
	item: CompletionItemSchema.optional(),
});

export type CompletionProgress = z.infer<typeof CompletionProgressSchema>;

// ============================================================================
// Hook Types
// ============================================================================

export const AgentConfigHookMCPServerSchema = z.object({
	url: z.string(),
	headers: z.record(z.string()).optional(),
});

export type AgentConfigHookMCPServer = z.infer<
	typeof AgentConfigHookMCPServerSchema
>;

/**
 * AgentConfigHook is a hook that can be used to configure the agent.
 * Hook Name = "config"
 */
export const AgentConfigHookSchema = z.object({
	agent: AgentSchema.optional(),
	mcpServers: z.record(AgentConfigHookMCPServerSchema).optional(),
	sessionId: z.string().optional(),
	_meta: z
		.object({
			workspace: z
				.object({
					parentId: z.string().optional(),
					baseUri: z.string().optional(),
					id: z.string().optional(),
					supported: z.boolean().optional(),
				})
				.optional(),
		})
		.optional(),
});

export type AgentConfigHook = z.infer<typeof AgentConfigHookSchema>;

/**
 * AgentRequestHook is a hook that can be used to modify the request before it is sent to the MCP server.
 * Hook Name = "request"
 */
export const AgentRequestHookSchema = z.object({
	request: CompletionRequestSchema.optional(),
	response: CompletionResponseSchema.optional(),
});

export type AgentRequestHook = z.infer<typeof AgentRequestHookSchema>;

/**
 * AgentResponseHook is a hook that can be used to modify the response before it is returned to the agent.
 * Hook Name = "response"
 */
export const AgentResponseHookSchema = AgentRequestHookSchema;

export type AgentResponseHook = z.infer<typeof AgentResponseHookSchema>;
