import { renderMarkdown } from '$lib/markdown';
import type { ChatMessage, WorkspaceClient, WorkspaceFile } from '$lib/types';
import type {
	Input,
	ParsedContent,
	ParsedFile,
	SessionData,
	Step,
	StepSession,
	Task,
	Tool
} from './types';
import YAML from 'yaml';

export async function parseFrontmatterMarkdown(fileContent: Blob): Promise<ParsedContent> {
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
			tools: []
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

	const tools: Tool[] =
		(metadata.tools ?? []).map((tool: Tool) => ({
			name: tool.name,
			title: tool.title,
			url: tool.url
		})) ?? [];

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

export async function compileFileContents(
	workspace: WorkspaceClient,
	files: WorkspaceFile[],
	taskId: string = ''
) {
	if (!taskId) {
		return [];
	}

	const validFiles = files.filter((file) => file.name.startsWith(`.nanobot/tasks/${taskId}/`));
	const parsedFiles: ParsedFile[] = [];
	for (const file of validFiles) {
		const content = await workspace?.readFile(file.name);
		if (content) {
			const parsedContent = await parseFrontmatterMarkdown(content);
			parsedFiles.push({
				...parsedContent,
				id: file.name.replace(`.nanobot/tasks/${taskId}/`, '')
			});
		}
	}

	return parsedFiles;
}

export async function convertToTask(
	workspace: WorkspaceClient,
	files: WorkspaceFile[],
	taskId: string
) {
	let name = '';
	let description = '';
	let inputs: Input[] = [];

	let parsedFiles: ParsedFile[] = [];
	if (files) {
		parsedFiles = await compileFileContents(workspace, files, taskId);
	}

	const steps: Step[] = [];
	let pointer: ParsedFile | undefined =
		parsedFiles.length > 1 ? parsedFiles.find((file) => file.id === 'TASK.md') : parsedFiles?.[0];

	if (pointer) {
		name = pointer.taskName;
		description = pointer.taskDescription;
		inputs = pointer.inputs;

		steps.push({
			id: pointer.id,
			name: pointer.name,
			description: pointer.description,
			content: pointer.content,
			tools: pointer.tools
		});
	}

	while (pointer) {
		pointer = pointer.next ? parsedFiles.find((file) => file.id === pointer?.next) : undefined;
		if (pointer) {
			steps.push({
				id: pointer.id,
				name: pointer.name,
				description: pointer.description,
				content: pointer.content,
				tools: pointer.tools
			});
		}
	}

	return {
		name,
		description,
		inputs,
		steps
	};
}

export function compileOutputFiles(task: Task, visibleInputs: Input[], taskId: string) {
	const { name: taskName, description: taskDescription, inputs } = task;
	const files = task.steps.map((step, index) => {
		let id = `.nanobot/tasks/${taskId}/STEP_${index}.md`;
		const metadata: Record<string, unknown> = {
			name: step.name,
			description: step.description,
			next: index !== task.steps.length - 1 ? `STEP_${index + 1}.md` : '',
			tools: step.tools.filter((tool) => tool.name)
		};

		if (index === 0) {
			metadata['task_name'] = taskName;
			metadata['task_description'] = taskDescription;
			if (inputs.length > 0) {
				const metadataInputs = [
					...visibleInputs.filter((input) => input.name.trim().length > 0),
					...inputs.filter(
						(input) =>
							input.name.trim().length > 0 &&
							!visibleInputs.some((visibleInput) => visibleInput.name === input.name)
					)
				];
				metadata['inputs'] = metadataInputs
					.filter((input) => {
						if (input.default || input.description) return true;
						const variableRegex = new RegExp(`\\$${input.name}(?![a-zA-Z0-9_])`);
						return task.steps.some((step) => variableRegex.test(step.content));
					})
					.map((input) => ({
						name: input.name,
						description: input.description,
						default: input.default
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
			data
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
				tools: []
			}
		]
	};
}

function parseToolArgs(item: { arguments?: string }): Record<string, unknown> | null {
	try {
		return JSON.parse(item.arguments || '{}');
	} catch {
		return null;
	}
}

function getStepIdFromExecuteTask(args: Record<string, unknown>, steps: Step[]): string {
	if (args.filename) {
		const step = steps.find((s) => s.id === args.filename);
		return step?.id ?? '';
	}
	return steps[0]?.id ?? '';
}

function createStepSession(stepId: string): StepSession {
	return { stepId, messages: [], pending: true, completed: false };
}

function processExecuteTaskStep(
	item: { name?: string; arguments?: string },
	steps: Step[],
	sessionData: SessionData,
	currentStepId: string
): string {
	const args = parseToolArgs(item);
	if (!args) return currentStepId;

	const stepId = getStepIdFromExecuteTask(args, steps);
	if (stepId && !sessionData[stepId]) {
		sessionData[stepId] = createStepSession(stepId);
	}
	return stepId || currentStepId;
}

function processTaskStepStatus(
	item: { name?: string; arguments?: string },
	sessionData: SessionData,
	activeStepId: string
): void {
	if (!activeStepId) return;

	const args = parseToolArgs(item);
	if (!args) return;

	const session = sessionData[activeStepId];
	if (session && (args.status === 'succeeded' || args.status === 'failed')) {
		session.pending = false;
		session.completed = true;
	}
}

export function processMessage(
	message: ChatMessage,
	steps: Step[],
	sessionData: SessionData,
	activeStepId: string,
	skipMessage: boolean
): string {
	const items = message.items ?? [];

	// Process tool calls in this message
	for (const item of items) {
		if (item.type === 'tool' && 'name' in item) {
			if (item.name === 'ExecuteTaskStep') {
				activeStepId = processExecuteTaskStep(item, steps, sessionData, activeStepId);
			} else if (item.name === 'TaskStepStatus') {
				processTaskStepStatus(item, sessionData, activeStepId);
			}
		}
	}

	if (!skipMessage && activeStepId && sessionData[activeStepId]) {
		const session = sessionData[activeStepId];
		if (!session.messages.some((m) => m.id === message.id)) {
			session.messages.push(message);
		}
	}

	return activeStepId;
}

function isSummaryMessage(message: ChatMessage): boolean {
	const items = message?.items ?? [];
	return items.length > 0 && items.every((item) => item.type === 'text');
}

export function buildSessionData(messages: ChatMessage[], steps: Step[]): SessionData {
	const sessionData: SessionData = {};

	const firstStepId = steps[0]?.id ?? '';
	if (firstStepId) {
		sessionData[firstStepId] = createStepSession(firstStepId);
	}
	let activeStepId = firstStepId;

	const lastMessage = messages[messages.length - 1];
	const hasTrailingSummary = lastMessage && isSummaryMessage(lastMessage);

	for (let i = 0; i < messages.length; i++) {
		const isLastMessage = i === messages.length - 1;
		const skipMessage = isLastMessage && hasTrailingSummary;

		activeStepId = processMessage(messages[i], steps, sessionData, activeStepId, skipMessage);
	}

	return sessionData;
}

export function areAllStepsCompleted(steps: Step[], sessionData: SessionData): boolean {
	return steps.every((step) => sessionData[step.id]?.completed);
}

export function buildRunSummary(messages: ChatMessage[]): string {
	if (messages.length === 0) return '';
	const lastMessage = messages[messages.length - 1];
	if (!lastMessage.items) return '';
	// Find the text item in the last message
	const textItem = lastMessage.items.find((item) => item.type === 'text' && 'text' in item);
	const text = textItem && 'text' in textItem ? textItem.text : '';
	return text ? renderMarkdown(text) : '';
}
