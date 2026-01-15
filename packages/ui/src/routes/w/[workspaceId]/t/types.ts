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
