import type { WorkspaceClient, WorkspaceFile } from "$lib/types";
import type { ParsedContent, ParsedFile, Task } from "./types";

async function parseYaml(fileContent: Blob): Promise<ParsedContent> {
    const text = await fileContent.text();
    
    // Split by frontmatter delimiter (---)
    // Format: ---\n<yaml>\n---\n<content>
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
    
    // Parse simple YAML key-value pairs
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

export async function compileFileContents(workspace: WorkspaceClient, files: WorkspaceFile[]) {
    const validFiles = files.filter((file) => file.name.startsWith('tasks/'));
    
    const parsedFiles: ParsedFile[] = [];
    for (const file of validFiles) {
        const content = await workspace?.readFile(file.name);
        if (content) {
            const parsedContent = await parseYaml(content);
            parsedFiles.push({
                ...parsedContent,
                fileName: file.name,
            });
        }
    }
    
    return parsedFiles;
}

export function compileOutputFiles(task: Task) {
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

        // Build YAML frontmatter
        const frontmatterLines = Object.entries(metadata)
            .filter(([, value]) => value) // Skip empty values
            .map(([key, value]) => `${key}: ${value}`);
        
        const yamlContent = `---\n${frontmatterLines.join('\n')}\n---\n${step.content}`;
        const data = new Blob([yamlContent], { type: 'text/yaml' });

        return {
            id: step.id,
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
                id: `tasks/${crypto.randomUUID()}.yaml`,
                name: '',
                description: '',
                content: '',
            }
        ],
    };
}