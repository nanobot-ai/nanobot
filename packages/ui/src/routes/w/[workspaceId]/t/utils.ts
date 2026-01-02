import type { WorkspaceClient, WorkspaceFile } from "$lib/types";
import type { Input, ParsedContent, ParsedFile, Step, Task } from "./types";
import YAML from "yaml";

async function parseFrontmatterMarkdown(fileContent: Blob): Promise<ParsedContent> {
    const text = await fileContent.text();
    
    // Split by frontmatter delimiter (---)
    // Format: ---\n<frontmatter>\n---\n<content>
    const frontmatterRegex = /^---\s*\n([\s\S]*?)\n---\s*\n?([\s\S]*)$/;
    const match = text.match(frontmatterRegex);
    
    if (!match) {
        return {
            taskName: '',
            taskDescription: '',
            inputs: [],
            next: '',
            name: '',
            description: '',
            content: text.trim(),
            tools: [],
        };
    }

    const [, frontmatter, content] = match;
    
    // Parse YAML frontmatter
    const metadata = YAML.parse(frontmatter) ?? {};

    // Normalize inputs array
    const inputs: Input[] = (metadata.inputs ?? []).map((input: Partial<Input>) => ({
        name: input.name ?? '',
        description: input.description ?? '',
        default: input.default ?? '',
        id: crypto.randomUUID()
    }));

    const tools: string[] = (metadata.tools ?? []).map((tool: string) => tool) ?? [];

    return {
        taskName: metadata.task_name ?? '',
        taskDescription: metadata.task_description ?? '',
        inputs,
        tools,
        next: metadata.next ?? '',
        name: metadata.name ?? '',
        description: metadata.description ?? '',
        content: content.trim()
    };
}

export async function compileFileContents(workspace: WorkspaceClient, files: WorkspaceFile[], taskId: string = '') {
    if (!taskId) { return []; }

    const validFiles = files.filter((file) => file.name.startsWith(`.nanobot/tasks/${taskId}/`));
    const parsedFiles: ParsedFile[] = [];
    for (const file of validFiles) {
        const content = await workspace?.readFile(file.name);
        if (content) {
            const parsedContent = await parseFrontmatterMarkdown(content);
            parsedFiles.push({
                ...parsedContent,
                id: file.name.replace(`.nanobot/tasks/${taskId}/`, ''),
            });
        }
    }
    
    return parsedFiles;
}

export async function convertToTask(workspace: WorkspaceClient, files: WorkspaceFile[], taskId: string) {
    let name = '';
    let description = '';
    let inputs: Input[] = [];
    
    let parsedFiles: ParsedFile[] = [];
    if (files) {
        parsedFiles = await compileFileContents(workspace, files, taskId);
    }
    
    const steps: Step[] = [];
    let pointer: ParsedFile | undefined = parsedFiles.length > 1 
        ? parsedFiles.find((file) => file.id === 'TASK.md') 
        : parsedFiles?.[0];

    if (pointer) {
        name = pointer.taskName;
        description = pointer.taskDescription;
        inputs = pointer.inputs;

        steps.push({
            id: pointer.id,
            name: pointer.name,
            description: pointer.description,
            content: pointer.content,
            tools: pointer.tools,
        })
    }

    while (pointer) {
        pointer = pointer.next ? parsedFiles.find((file) => file.id === pointer?.next) : undefined;
        if (pointer) {
            steps.push({
                id: pointer.id,
                name: pointer.name,
                description: pointer.description,
                content: pointer.content,
                tools: pointer.tools,
            })
        }
    }

    return {
        name,
        description,
        inputs,
        steps,
    }
}

export function compileOutputFiles(task: Task, visibleInputs: Input[], taskId: string) {
    const { name: taskName, description: taskDescription, inputs } = task;
    const files = task.steps.map((step, index) => {
        let id = `.nanobot/tasks/${taskId}/STEP_${index}.md`;
        const metadata: Record<string, unknown> = {
            name: step.name,
            description: step.description,
            next: index !== task.steps.length - 1 ? `STEP_${index+1}.md` : '',
            tools: step.tools,
        };

        if (index === 0) {
            metadata['task_name'] = taskName;
            metadata['task_description'] = taskDescription;
            if (inputs.length > 0) {
                const metadataInputs = [
                    ...visibleInputs.filter((input) => input.name),
                    ...inputs.filter((input) => !visibleInputs.some((visibleInput) => visibleInput.name === input.name)),
                ];
                metadata['inputs'] = metadataInputs
                    .map((input) => ({
                        name: input.name,
                        description: input.description,
                        default: input.default,
                    }));
            }
            id = `.nanobot/tasks/${taskId}/TASK.md`;
        }

        // Remove empty string values for cleaner output
        const cleanedMetadata = Object.fromEntries(
            Object.entries(metadata).filter(([, value]) => value !== '' && value !== undefined)
        );

        // Serialize to YAML frontmatter
        const frontmatter = YAML.stringify(cleanedMetadata).trim();
        
        const content = `---\n${frontmatter}\n---\n${step.content}`;
        const data = new Blob([content], { type: 'text/markdown' });

        return {
            id,
            data,
        };
    });
    return files;
}

export function setupEmptyTask(): Task {
    return {
        name: '',
        description: '',
        inputs: [],
        steps: [
            {
                id: '',
                name: '',
                description: '',
                content: '',
                tools: [],
            }
        ],
    };
}