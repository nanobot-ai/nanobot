export type Input = {
    name: string;
    description: string;
    default?: string;
    id: string;
}

export type Step = {
    id: string;
    name: string;
    description: string;
    content: string;
}

export type Task = {
    name: string;
    description: string;
    inputs: Input[];
    steps: Step[];
}

export type ParsedContent = {
    taskName: string;
    taskDescription: string;
    inputs: Input[];
    next: string;
    name: string;
    description: string;
    content: string;
};

export type ParsedFile = {
    id: string;
} & ParsedContent;
