import * as path from "node:path";
import type { WorkspaceClient } from "@nanobot-ai/workspace-client";
import { load } from "js-yaml";

export type Task = {
	id: string;
	name: string;
	description?: string;
	instructions: string;
	inputs?: TaskInput[];
	baseDir: string;
};

export type TaskInput = {
	name: string;
	description: string;
	default?: string;
};

const emptyTask: Task = {
	id: "",
	name: "",
	description: "",
	instructions: "",
	baseDir: "",
};

function parseYAMLFrontMatter(text: string): {
	frontMatter: Record<string, string>;
	instructions: string;
} {
	const match = text.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
	if (!match) return { frontMatter: {}, instructions: text.trim() };
	try {
		const frontMatter = load(match[1]) as Record<string, string>;
		return { frontMatter, instructions: match[2].trim() };
	} catch {
		return { frontMatter: {}, instructions: text.trim() };
	}
}

const tasksRoot = ".nanobot/tasks";

export async function getTask(client: WorkspaceClient, taskName: string) {
	const task = await getTaskByDirectoryName(client, taskName);
	if (task.name) {
		return task;
	}

	// Check for task where name doesn't match directory name
	const tasks = await getTasks(client);
	return tasks.find((s) => s.name === taskName) || emptyTask;
}

async function getTaskByDirectoryName(
	client: WorkspaceClient,
	taskName: string,
) {
	const taskDir = await client.resolvePath(path.join(tasksRoot, taskName));
	const content = await client.readTextFile(
		path.join(tasksRoot, taskName, "TASK.md"),
		{ ignoreNotFound: true },
	);
	if (!content) {
		// Check for task where name doesn't match directory name
		return emptyTask;
	}

	const { frontMatter, instructions: parsedContent } =
		parseYAMLFrontMatter(content);
	return {
		id: taskName,
		name: frontMatter.task_name || frontMatter.name || taskName,
		description: frontMatter.task_description || frontMatter.description || "",
		instructions: parsedContent,
		baseDir: taskDir,
	};
}

export async function getTasksDescription(client: WorkspaceClient) {
	const tasks = await getTasks(client);
	const available = tasks
		.map((s) => {
			`Name: ${s.name}\nDescription: ${s.description}`;
			if (s.inputs) {
				return `${s.name}: ${s.description}\nInputs:\n${s.inputs.map((i) => ` - ${i.name}: ${i.description}`).join("\n")}`;
			} else {
				return `${s.name}: ${s.description}`;
			}
		})
		.join("---\n");
	if (!available) {
		return "No tasks available";
	}
	return;
}

export async function getTasks(client: WorkspaceClient) {
	const dirEntries = await client.listDir(tasksRoot, {
		ignoreNotFound: true,
	});

	const tasks = Promise.all(
		dirEntries.entries.map(async (entry): Promise<Task> => {
			if (!entry.isDirectory) {
				return emptyTask;
			}

			return getTaskByDirectoryName(client, entry.name);
		}),
	);

	return (await tasks).filter((x) => x.name);
}
