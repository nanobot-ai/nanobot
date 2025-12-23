export type Step = {
    id: string;
    name: string;
    description: string;
    content: string;
}

export type Task = {
    name: string;
    description: string;
    steps: Step[];
}

export type ParsedContent = {
    taskName: string;
    taskDescription: string;
    next: string;
    name: string;
    description: string;
    content: string;
};

export type ParsedFile = {
    id: string;
} & ParsedContent;
