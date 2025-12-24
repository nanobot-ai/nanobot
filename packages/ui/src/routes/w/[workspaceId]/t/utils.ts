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
            content: text.trim()
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
    }));

    return {
        taskName: metadata.task_name ?? '',
        taskDescription: metadata.task_description ?? '',
        inputs,
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
    console.log({ parsedFiles });
    
    const steps: Step[] = [];
    let pointer: ParsedFile | undefined = parsedFiles.length > 1 ? parsedFiles.find((file) => {
        // find the first file
        // it can have a next but is not the next of any other file
        return !parsedFiles.some((compareFile) => compareFile.next === file.id);
    }) : parsedFiles?.[0];

    if (pointer) {
        name = pointer.taskName;
        description = pointer.taskDescription;
        inputs = pointer.inputs;

        steps.push({
            id: pointer.id,
            name: pointer.name,
            description: pointer.description,
            content: pointer.content,
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
            })
        }
    }

    return {
        name,
        description,
        inputs,
        steps
    }
}

export function compileOutputFiles(task: Task, taskId: string) {
    const { name: taskName, description: taskDescription, inputs } = task;
    const files = task.steps.map((step, index) => {
        let id = `.nanobot/tasks/${taskId}/STEP_${index}.md`;
        const metadata: Record<string, unknown> = {
            name: step.name,
            description: step.description,
            next: index !== task.steps.length - 1 ? `STEP_${index+1}.md` : '',
        };

        if (index === 0) {
            metadata['task_name'] = taskName;
            metadata['task_description'] = taskDescription;
            if (inputs.length > 0) {
                metadata['inputs'] = inputs;
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
            }
        ],
    };
}

export function compileArguments(steps: Step[]) {
    return steps.reduce<Record<string, string>>((acc, step) => {
        const contentArguments = step.content.match(/\$\w+/g);
        if (contentArguments) {
            for (const arg of contentArguments) {
                acc[arg] = '';
            }
        }
        return acc;
    }, {});
}