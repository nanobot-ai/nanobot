import type { ChatMessage } from '$lib/types';

export type Input = {
	name: string;
	description: string;
	default?: string;
	id: string;
};

export type Step = {
	id: string;
	name: string;
	description: string;
	content: string;
	tools: Tool[];
};

export type Tool = {
	name: string;
	title: string;
	url: string;
};

export type Task = {
	name: string;
	description: string;
	inputs: Input[];
	steps: Step[];
};

export type ParsedContent = {
	taskName: string;
	taskDescription: string;
	inputs: Input[];
	next: string;
	name: string;
	description: string;
	content: string;
	tools: Tool[];
};

export type ParsedFile = {
	id: string;
} & ParsedContent;

export type StepSession = {
	stepId: string;
	messages: ChatMessage[];
	pending: boolean;
	completed: boolean;
};
export type SessionData = Record<string, StepSession>;

export type OngoingStep = {
	loading: boolean;
	completed: boolean;
	oauth: string;
	totalTime?: number;
	tokens?: number;
	error?: boolean;
};
