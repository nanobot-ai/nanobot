import type { WorkspaceClient, WorkspaceFile } from "$lib/types";
import type { ParsedContent, ParsedFile, Step, Task } from "./types";

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
            next: '',
            name: '',
            description: '',
            content: text.trim()
        };
    }

    const [, frontmatter, content] = match;
    
    // Parse simple frontmatter key-value pairs
    const metadata: Record<string, string> = {};
    for (const line of frontmatter.split('\n')) {
        const colonIndex = line.indexOf(':');
        if (colonIndex === -1) continue;
        
        const key = line.slice(0, colonIndex).trim();
        const value = line.slice(colonIndex + 1).trim();
        metadata[key] = value;
    }

    return {
        taskName: metadata['task_name'] ?? '',
        taskDescription: metadata['task_description'] ?? '',
        next: metadata['next'] ?? '',
        name: metadata['name'] ?? '',
        description: metadata['description'] ?? '',
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
                id: file.name.replace(`.nanobot/tasks/${taskId}/`, '').replace('.md', ''),
            });
        }
    }
    
    return parsedFiles;
}

export async function convertToTask(workspace: WorkspaceClient, files: WorkspaceFile[], taskId: string) {
    let name = '';
    let description = '';
    
    let parsedFiles: ParsedFile[] = [];
    if (files) {
        parsedFiles = await compileFileContents(workspace, files, taskId);
    }
    
    const steps: Step[] = [];
    let pointer: ParsedFile | undefined = parsedFiles.length > 1 ? parsedFiles.find((file) => {
        // find the first file
        // it can have a next but is not the next of any other file
        return !parsedFiles.some((compareFile) => compareFile.next === file.id);
    }) : parsedFiles?.[0];

    if (pointer) {
        name = pointer.taskName;
        description = pointer.taskDescription;

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
        steps
    }
}

export function compileOutputFiles(task: Task, taskId: string) {
    const { name: taskName, description: taskDescription } = task;
    const files = task.steps.map((step, index) => {
        const metadata: Record<string, string> = {
            name: step.name,
            description: step.description,
            next: index !== task.steps.length - 1 ? task.steps[index + 1].id : '',
        };

        if (index === 0) {
            metadata['task_name'] = taskName;
            metadata['task_description'] = taskDescription;
        }

        // Build Markdown frontmatter
        const frontmatterLines = Object.entries(metadata)
            .filter(([, value]) => value) // Skip empty values
            .map(([key, value]) => `${key}: ${value}`);
        
        const content = `---\n${frontmatterLines.join('\n')}\n---\n${step.content}`;
        const data = new Blob([content], { type: 'text/markdown' });
        const stepId = step.id || `step${index}`

        return {
            id: `.nanobot/tasks/${taskId}/${stepId}.md`,
            data,
        };
    });
    return files;
}

export function setupEmptyTask() {
    return {
        name: '',
        description: '',
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