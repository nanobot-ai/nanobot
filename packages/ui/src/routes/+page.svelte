<script lang="ts">
	import '$lib/../app.css';
	import { ChatService } from '$lib/chat.svelte';
	import { onDestroy, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { setSharedChat } from '$lib/stores/chat.svelte';
	import ThreadFromChat from "$lib/components/ThreadFromChat.svelte";
	import { WorkspaceService } from '$lib/workspace.svelte';
	import { ChevronRight, Eye, File, Folder, ListTodo } from '@lucide/svelte';
	import { formatTimeAgo } from '$lib/utils/time';

	const chat = new ChatService();
	let doClose = true;
	let workspaceEnabled = $derived(chat.workspaceEnabled);
	const workspaceService = new WorkspaceService();
	const mockSharedWorkspaces = $state<{ id: string, name: string, color: string, created: string }[]>([
		{ id: '1', name: 'Jolly Roger', color: '#000', created: '2026-01-01' },
		{ id: '2', name: 'Matcha Latte', color: '#2ddcec', created: '2026-01-02' },
		{ id: '3', name: 'Pumpkin Spice', color: '#fdcc11', created: '2026-01-03' },
	]);
	const mockTasks = $state<{ id: string, name: string, created: string, workspace: string }[]>([
		{ id: '1', name: 'Onboarding Workflow', created: '2026-01-01', workspace: 'Adorable Akita' },
		{ id: '2', name: 'Customer Support', created: '2026-01-02', workspace: 'Adorable Akita' },
		{ id: '3', name: 'Marketing Campaign', created: '2026-01-03', workspace: 'Caramel Cookie' },
	]);
	const mockFiles = $state<{ id: string, name: string, created: string, workspace: string }[]>([
		{ id: '1', name: 'Example.pdf', created: '2026-01-01', workspace: 'Adorable Akita' },
		{ id: '2', name: 'Example.docx', created: '2026-01-02', workspace: 'Caramel Cookie' },
		{ id: '3', name: 'Example.xlsx', created: '2026-01-03', workspace: 'Adorable Akita' },
	]);
	const mockTaskRuns = $state<{ id: string, task: string, averageCompletionTime: string, user: string, workspace: string, tokensUsed: number; }[]>([
		{ id: '1', task: 'Onboarding Workflow', averageCompletionTime: '10m', user: 'John Doe', workspace: 'Adorable Akita', tokensUsed: 7000 },
		{ id: '2', task: 'Customer Support', averageCompletionTime: '10.1m', user: 'John Doe', workspace: 'Adorable Akita', tokensUsed: 8500 },
		{ id: '3', task: 'Marketing Campaign', averageCompletionTime: '10m', user: 'Jane Doe', workspace: 'Caramel Cookie', tokensUsed: 8000 },
		{ id: '4', task: 'Product Launch', averageCompletionTime: '11m', user: 'Jane Doe', workspace: 'Caramel Cookie', tokensUsed: 9000 },
		{ id: '5', task: 'Sales Pipeline', averageCompletionTime: '10m', user: 'John Doe', workspace: 'Adorable Akita', tokensUsed: 10000 },
		{ id: '6', task: 'Customer Support', averageCompletionTime: '6.5m', user: 'John Doe', workspace: 'Adorable Akita', tokensUsed: 11500 },
		{ id: '7', task: 'Marketing Campaign', averageCompletionTime: '10m', user: 'Jane Doe', workspace: 'Caramel Cookie', tokensUsed: 12000 },
		{ id: '8', task: 'Product Launch', averageCompletionTime: '10m', user: 'Jane Doe', workspace: 'Caramel Cookie', tokensUsed: 13000 },
		{ id: '9', task: 'Sales Pipeline', averageCompletionTime: '10m', user: 'John Doe', workspace: 'Adorable Akita', tokensUsed: 14000 },
		{ id: '10', task: 'Customer Support', averageCompletionTime: '10m', user: 'John Doe', workspace: 'Adorable Akita', tokensUsed: 15500 },
	]);

	// Share the chat instance immediately so layout can access it
	setSharedChat(chat);

	$effect(() => {
		if (chat.chatId) {
			doClose = false;
			goto(resolve(`/c/${chat.chatId}`));
		}
	});

	onMount(() => {
		if (window.location.search.includes('new')) {
			chat.newChat();
		}
	});

	$effect(() => {
		if (workspaceEnabled) {
			workspaceService.load();
		}
	})

	onDestroy(() => {
		if (doClose) {
			chat.close();
		}
	});
</script>

{#if workspaceEnabled}
<div class="h-dvh w-full overflow-y-auto">
	<div class="w-6xl mx-auto py-8 flex flex-col gap-4">
		<h1 class="text-3xl font-semibold">Dashboard</h1>
		<div class="grid grid-cols-12 gap-4">
			<div class="flex flex-col col-span-8 gap-4">
				<div class="card h-fit bg-base-100 dark:bg-base-200">
					<div class="card-body">
						<h2 class="card-title">Overview</h2>
						<div class="text-sm text-base-content/70">
							Welcome to your dashboard. This is a summary of your activity.
						</div>

						<div class="grid grid-cols-2 gap-3 bg-base-200 dark:bg-base-100 p-2 rounded-box">
							<div class="stats shadow bg-base-100 dark:bg-base-200">
								<div class="stat">
									<div class="stat-title">Total Task Runs</div>
									<div class="stat-value">121</div>
									<div class="stat-desc">10% more than last month</div>
								</div>
							</div>

							<div class="stats shadow bg-base-100 dark:bg-base-200">
								<div class="stat">
									<div class="stat-title">Total Token Usage</div>
									<div class="stat-value">89,400</div>
									<div class="stat-desc">21% more than last month</div>
								</div>
							</div>

							<div class="stats shadow bg-base-100 dark:bg-base-200">
								<div class="stat">
									<div class="stat-title">Most Used Task</div>
									<div class="stat-value text-lg truncate self-end">
										Onboarding Workflow
									</div>
									<div class="stat-desc self-end">Adorable Akita</div>
								</div>
							</div>

							<div class="stats shadow bg-base-100 dark:bg-base-200">
								<div class="stat">
									<div class="stat-title">Most Expensive Task</div>
									<div class="stat-value text-lg truncate self-end">
										Onboarding Workflow
									</div>
									<div class="stat-desc">8,500 average tokens used per run</div>
								</div>
							</div>
						</div>
					</div>
				</div>

				<div class="card h-fit bg-base-100 dark:bg-base-200">
					<div class="card-body">
						<h2 class="card-title text-base">Recent Task Run Activity</h2>

						<table class="table">
							<thead>
								<tr>
									<th>Task</th>
									<th>Time To Complete</th>
									<th>Tokens Used</th>
									<th>User</th>
									<th>Workspace</th>
									<th></th>
								</tr>
							</thead>
							<tbody>
								{#each mockTaskRuns as taskRun (taskRun.id)}
									<tr>
										<td>{taskRun.task}</td>
										<td>{taskRun.averageCompletionTime}</td>
										<td>{taskRun.tokensUsed}</td>
										<td>{taskRun.user}</td>
										<td>{taskRun.workspace}</td>
										<td>
											<button class="btn btn-ghost btn-xs">details</button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			</div>
			<div class="flex flex-col col-span-4 gap-4">
				<div class="card h-fit bg-base-100 dark:bg-base-200">
					<div class="card-body">
						<h3 class="card-title">Workspaces</h3>
						<ul class="list bg-base-100 rounded-box">
							<li class="pb-2 text-xs opacity-60 tracking-wide">Most recently created workspaces</li>
							{#each workspaceService.workspaces as workspace (workspace.id)}
								<li class="list-row">
									<div>
										<Folder class="size-6" style="color: {workspace.color || '#000'}; fill: color-mix(in oklab, {workspace.color || '#000'} 50%, var(--color-base-100))" />
									</div>
									<div>
									<div>{workspace.name}</div>
									<div class="text-xs uppercase font-semibold opacity-60">
										{formatTimeAgo(workspace.created).relativeTime}
									</div>
									</div>
								</li>
							{/each}
						</ul>

						<ul class="list bg-base-100 rounded-box">
							<li class="pb-2 text-xs opacity-60 tracking-wide">Most recently shared workspaces</li>
							{#each mockSharedWorkspaces as workspace (workspace.id)}
								<li class="list-row">
									<div>
										<Folder class="size-6" style="color: {workspace.color || '#000'}; fill: color-mix(in oklab, {workspace.color || '#000'} 50%, var(--color-base-100))" />
									</div>
									<div>
									<div>{workspace.name}</div>
									<div class="text-xs uppercase font-semibold opacity-60">
										{formatTimeAgo(workspace.created).relativeTime}
									</div>
									</div>
								</li>
							{/each}
						</ul>
					</div>
				</div>

				<div class="card h-fit bg-base-100 dark:bg-base-200">
					<div class="card-body">
						<h3 class="card-title">Tasks</h3>
						<ul class="list bg-base-100 rounded-box">
							<li class="pb-2 text-xs opacity-60 tracking-wide">Most recently created tasks</li>
							{#each mockTasks as task (task.id)}
								<button class="list-row hover:bg-base-200 text-left cursor-pointer transition-colors duration-250">
									<div class="self-center">
										<ListTodo class="size-6" />
									</div>
									<div>
									<div class="text-xs opacity-60">{task.workspace}</div>
									<div>{task.name}</div>
									<div class="text-xs uppercase font-semibold opacity-60">
										{formatTimeAgo(task.created).relativeTime}
									</div>
									</div>
									<div class="p-2">
										<ChevronRight class="size-4" />
									</div>
								</button>
							{/each}
						</ul>
					</div>
				</div>

				<div class="card h-fit bg-base-100 dark:bg-base-200">
					<div class="card-body">
						<h3 class="card-title">Files</h3>
						<ul class="list bg-base-100 rounded-box">
							<li class="pb-2 text-xs opacity-60 tracking-wide">Most recently uploaded files</li>
							{#each mockFiles as file (file.id)}
								<button class="list-row hover:bg-base-200 text-left cursor-pointer transition-colors duration-250">
									<div class="self-center">
										<File class="size-6" />
									</div>
									<div>
									<div class="text-xs opacity-60">{file.workspace}</div>
									<div>{file.name}</div>
									<div class="text-xs uppercase font-semibold opacity-60">
										{formatTimeAgo(file.created).relativeTime}
									</div>
									</div>
									<div class="p-2 self-center">
										<Eye class="size-4" />
									</div>
								</button>
							{/each}
						</ul>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>
{:else}
	<ThreadFromChat {chat} />
{/if}


<svelte:head>
	{#if workspaceEnabled}
		<title>Nanobot | Dashboard</title>
	{:else if chat.agent?.name}
		<title>{chat.agent.name}</title>
	{:else}
		<title>Nanobot</title>
	{/if}
</svelte:head>