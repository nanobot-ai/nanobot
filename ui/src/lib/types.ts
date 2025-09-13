export interface Chat {
	id: string;
	title: string;
	created: string;
	visibility?: 'public' | 'private';
	readonly?: boolean;
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
	structuredContent?: unknown;
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
	attachments?: Array<{
		file: File;
		url: string;
	}>;
}

export interface ChatResult {
	message: ChatMessage;
}

export interface Event {
	id?: string;
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
	prompts: Prompt[];
}

export interface BaseMetadata {
	name: string;
	title?: string;
}

export interface Prompt extends BaseMetadata {
	description?: string;
	arguments?: PromptArgument[];
	_meta?: { [key: string]: unknown };
}

export interface PromptArgument extends BaseMetadata {
	description?: string;
	required?: boolean;
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

export interface ElicitationResult {
	action: 'accept' | 'decline' | 'cancel';
	content?: { [key: string]: string | number | boolean };
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
