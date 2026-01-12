import * as mocks from '$lib/mocks';
import { browser } from '$app/environment';

type TaskArgument = {
	name: string;
	value: string;
	description: string;
};

type Task = {
	id: string;
	runs: {
		id: string;
		created: string;
		arguments?: TaskArgument[];
		stepSessions?: {
			stepId: string;
			threadId: string;
		}[];
	}[];
};

type TaskState = {
	current: {
		tasks: Task[];
	};
	addRun: (
		taskId: string,
		run: {
			id: string;
			created: string;
			arguments?: TaskArgument[];
			stepSessions?: { stepId: string; threadId: string }[];
		}
	) => void;
	updateRun: (
		taskId: string,
		runId: string,
		stepSessions: { stepId: string; threadId: string }[]
	) => void;
	deleteRun: (taskId: string, runId: string) => void;
};

function getInitialTasks(): Task[] {
	if (browser) {
		const json = localStorage.getItem('mock-tasks');
		if (json) {
			return JSON.parse(json);
		}
	}
	return mocks.tasks.map((task) => ({
		id: task.id,
		runs: []
	}));
}

const taskState = $state<TaskState>({
	current: {
		tasks: getInitialTasks()
	},
	addRun,
	updateRun,
	deleteRun
});

export const mockTasks = taskState;

function deleteRun(taskId: string, runId: string): void {
	if (!browser) return;

	const task = taskState.current.tasks.find((t) => t.id === taskId);
	if (!task) return;

	task.runs = task.runs.filter((r) => r.id !== runId);
	localStorage.setItem('mock-tasks', JSON.stringify(taskState.current.tasks));
}

function addRun(
	taskId: string,
	run: {
		id: string;
		created: string;
		arguments?: TaskArgument[];
		stepSessions?: { stepId: string; threadId: string }[];
	}
): void {
	if (!browser) return;

	let task = taskState.current.tasks.find((t) => t.id === taskId);
	if (!task) {
		// Create task entry if it doesn't exist
		task = { id: taskId, runs: [] };
		taskState.current.tasks.push(task);
	}
	task.runs.push(run);
	localStorage.setItem('mock-tasks', JSON.stringify(taskState.current.tasks));
}

function updateRun(
	taskId: string,
	runId: string,
	stepSessions: { stepId: string; threadId: string }[]
): void {
	if (!browser) return;

	const task = taskState.current.tasks.find((t) => t.id === taskId);
	if (!task) return;

	const run = task.runs.find((r) => r.id === runId);
	if (!run) return;

	run.stepSessions = stepSessions;
	localStorage.setItem('mock-tasks', JSON.stringify(taskState.current.tasks));
}
