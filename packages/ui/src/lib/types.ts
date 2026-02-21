import type { ChatService } from '$lib/chat.svelte';

// Re-export types from the official MCP SDK.
// Some SDK Zod schemas lag behind the spec interfaces (e.g. Resource missing `size`,
// PromptArgument missing `title`), so we extend the inferred types where needed.
export type { BaseMetadata, Icon } from '@modelcontextprotocol/sdk/types.js';

import type {
	Resource as SDKResource,
	Prompt as SDKPrompt,
	PromptArgument as SDKPromptArgument
} from '@modelcontextprotocol/sdk/types.js';

/** SDK Resource extended with `size` (present in spec but missing from Zod schema). */
export type Resource = SDKResource & { size?: number };

/** SDK PromptArgument extended with `title` (present in spec's BaseMetadata but missing from Zod schema). */
export type PromptArgument = SDKPromptArgument & { title?: string };

/** SDK Prompt with arguments typed using our extended PromptArgument. */
export type Prompt = Omit<SDKPrompt, 'arguments'> & { arguments?: PromptArgument[] };

// Re-export the ElicitResult type (named ElicitResult in SDK, was ElicitationResult here)
export type { ElicitResult as ElicitationResult } from '@modelcontextprotocol/sdk/types.js';

export interface Agent {
	id: string;
	name?: string;
	description?: string;
	icon?: string;
	iconDark?: string;
	starterMessages?: string[];
	current?: boolean;
}

export interface Agents {
	agents: Agent[];
}

export interface Chat {
	id: string;
	title: string;
	created: string;
	visibility?: 'public' | 'private';
	readonly?: boolean;
	workflowURIs?: string[];
}

export interface ChatMessage {
	id: string;
	created?: string;
	role: 'user' | 'assistant';
	items?: ChatMessageItem[];
	hasMore?: boolean;
}

export type ChatMessageItem = ToolOutputItem | ChatMessageItemToolCall | ChatMessageItemReasoning;

export type ToolOutputItem =
	| ChatMessageItemImage
	| ChatMessageItemAudio
	| ChatMessageItemText
	| ChatMessageItemResource
	| ChatMessageItemResourceLink;

export interface ChatMessageItemToolCall extends ChatMessageItemBase {
	type: 'tool';
	arguments?: string;
	callID?: string;
	name?: string;
	output?: {
		isError?: boolean;
		content?: ToolOutputItem[];
		structuredContent?: unknown;
	};
}

export interface ChatMessageItemImage extends ChatMessageItemBase {
	type: 'image';
	data: string;
	mimeType: string;
}

export interface ChatMessageItemAudio extends ChatMessageItemBase {
	type: 'audio';
	data: string;
	mimeType: string;
}

export interface ChatMessageItemText extends ChatMessageItemBase {
	type: 'text';
	text: string;
}

export interface ChatMessageItemResourceLink extends ChatMessageItemBase {
	type: 'resource_link';
	name?: string;
	description?: string;
	uri: string;
}

export interface ChatMessageItemReasoning extends ChatMessageItemBase {
	type: 'reasoning';
	summary?: {
		text: string;
	}[];
}

export interface ChatMessageItemResource extends ChatMessageItemBase {
	type: 'resource';
	resource: {
		uri: string;
		name?: string;
		description?: string;
		title?: string;
		mimeType: string;
		size?: number;
		text?: string;
		blob?: string;
		annotations?: {
			audience?: string[];
			priority?: number;
			lastModified?: string;
		};
		_meta?: { [key: string]: unknown };
	};
}

export interface ChatMessageItemBase {
	id: string;
	hasMore?: boolean;
	_meta?: { [key: string]: unknown };
	type:
		| 'text'
		| 'image'
		| 'audio'
		| 'toolCall'
		| 'resource'
		| 'resource_link'
		| 'tool'
		| 'reasoning';
}

export interface ChatRequest {
	id: string;
	threadId: string;
	message: string;
	agent?: string;
	attachments?: Attachment[];
}

export interface Attachment {
	name?: string;
	uri: string;
	mimeType?: string;
}

export interface ChatResult {
	message: ChatMessage;
}

export interface Event {
	id?: string | number;
	type:
		| 'message'
		| 'history-start'
		| 'history-end'
		| 'chat-in-progress'
		| 'chat-done'
		| 'error'
		| 'elicitation/create';
	message?: ChatMessage;
	data?: unknown;
	error?: string;
}

export interface Notification {
	id: string;
	type: 'success' | 'error' | 'warning' | 'info';
	title: string;
	message?: string;
	timestamp: Date;
	autoClose?: boolean;
	duration?: number; // milliseconds
}

export interface Prompts {
	prompts: import('@modelcontextprotocol/sdk/types.js').Prompt[];
}

export interface Resources {
	resources: import('@modelcontextprotocol/sdk/types.js').Resource[];
}

/**
 * Convenience type for resource contents that allows checking both text and blob.
 * The SDK splits these into TextResourceContents and BlobResourceContents,
 * but our code commonly checks both fields on a single object.
 */
export interface ResourceContents {
	uri: string;
	mimeType?: string;
	text?: string;
	blob?: string;
	_meta?: { [key: string]: unknown };
}

export interface Elicitation {
	id: string | number;
	message: string;
	requestedSchema: {
		type: 'object';
		properties: {
			[key: string]: PrimitiveSchemaDefinition;
		};
		required?: string[];
	};
	_meta?: { [key: string]: unknown };
}

export type PrimitiveSchemaDefinition = StringSchema | NumberSchema | BooleanSchema | EnumSchema;

export interface StringSchema {
	type: 'string';
	title?: string;
	description?: string;
	minLength?: number;
	maxLength?: number;
	format?: 'email' | 'uri' | 'date' | 'date-time';
}

export interface NumberSchema {
	type: 'number' | 'integer';
	title?: string;
	description?: string;
	minimum?: number;
	maximum?: number;
}

export interface BooleanSchema {
	type: 'boolean';
	title?: string;
	description?: string;
	default?: boolean;
}

export interface EnumSchema {
	type: 'string';
	title?: string;
	description?: string;
	enum: string[];
	enumNames?: string[]; // Display names for enum values
}

export const MessageMimeType = 'application/vnd.nanobot.chat.message+json';
export const HistoryMimeType = 'application/vnd.nanobot.chat.history+json';
export const ToolResultMimeType = 'application/vnd.nanobot.tool.result+json';

export interface UploadedFile {
	id: string;
	file: File;
	uri: string;
	mimeType?: string;
}

export interface UploadingFile {
	id: string;
	file: File;
	controller?: AbortController;
}

// Workspace types
export const SessionMimeType = 'application/vnd.nanobot.session+json';

// Forward declaration for WorkspaceClient - export both names
export type { NanobotClient, NanobotClient as SimpleClient } from './mcpclient';

export interface Workspace {
	id: string;
	name: string;
	created: string;
	order?: number;
	color?: string;
	icons?: import('@modelcontextprotocol/sdk/types.js').Icon[];
}

export interface Session {
	id: string;
	title: string;
}

export interface SessionDetails {
	id: string;
	title?: string;
	createdAt: string;
	updatedAt?: string;
	workspaceId?: string;
	sessionWorkspaceId?: string;
}

export interface WorkspaceFile {
	name: string;
}

export interface WorkspaceClient {
	// Properties
	readonly id: string;
	readonly workspace: Workspace;
	readonly files: WorkspaceFile[];
	readonly sessions: Session[];
	readonly loading: boolean;

	readFile(path: string): Promise<Blob>;
	writeFile(path: string, data: Blob | string): Promise<void>;
	createFile(path: string, data: Blob | string): Promise<void>;
	deleteFile(path: string): Promise<void>;
	deleteSession(sessionId: string): Promise<void>;
	getSessionDetails(sessionId: string): Promise<SessionDetails>;
	getSession(sessionId: string): Promise<ChatService>;
	newSession(opts?: { editor?: boolean }): Promise<ChatService>;
}

export interface InitializationResult {
	capabilities?: {
		experimental?: {
			'ai.nanobot'?: {
				session?: {
					ui?: boolean;
					workspace?: {
						id?: string;
						baseUri?: string;
						base?: string;
						supported?: boolean;
					};
				};
			};
		};
		[key: string]: unknown;
	};
	[key: string]: unknown;
}

export interface ToolDef {
	name: string;
	description?: string;
	_meta?: { ui?: { resourceUri?: string }; [key: string]: unknown };
}

export const UIPath = '/mcp?ui';
