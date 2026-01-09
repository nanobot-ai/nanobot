<script lang="ts">
	import '$lib/../app.css';
	import { ChatService } from '$lib/chat.svelte';
	import { onDestroy, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { setSharedChat } from '$lib/stores/chat.svelte';
	import ThreadFromChat from "$lib/components/ThreadFromChat.svelte";
	import { WorkspaceService } from '$lib/workspace.svelte';
	import { ChevronRight, Eye, File, Folder, ListTodo, Plus } from '@lucide/svelte';
	import { formatTimeAgo } from '$lib/utils/time';
	import * as mocks from '$lib/mocks';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	
	const chat = new ChatService();
	let doClose = true;
	let workspaceEnabled = $derived(chat.workspaceEnabled);
	const workspaceService = new WorkspaceService();
	
	let tasks = $state(mocks.tasks);
	let files = $state(mocks.files);
	let taskRuns = $state(mocks.taskRuns);
	let sharedWorkspaces = $state(mocks.sharedWorkspaces);

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

	const layout = getLayoutContext();
</script>

{#if workspaceEnabled}
<div class="h-dvh w-full overflow-y-auto">
	<div class="{layout.isSidebarCollapsed ? 'pt-18' : ''} mx-auto pt-6 p-8 flex flex-col gap-4 max-w-7xl w-full transition-[padding] duration-100 ease-out">
		<h1 class="text-3xl font-semibold">Dashboard</h1>
		<div class="grid grid-cols-12 gap-4">
			<div class="flex flex-col col-span-8 gap-4">
				<div class="card h-fit bg-base-100 dark:bg-base-200">
					<div class="card-body">
						<h2 class="card-title">Overview</h2>
						<div class="text-sm text-base-content/70">
							This is a summary of the most recent activity across your workspaces.
						</div>

						<div class="grid grid-cols-2 gap-3 bg-base-200 dark:bg-base-100 p-2 rounded-box">
							<div class="stats shadow bg-base-100 dark:bg-base-200">
								<div class="stat">
									<div class="stat-title">Total Workflow Runs</div>
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
										CNCF Onboarding
									</div>
									<div class="stat-desc self-end">CNCF Onboarding</div>
								</div>
							</div>

							<div class="stats shadow bg-base-100 dark:bg-base-200">
								<div class="stat">
									<div class="stat-title">Most Expensive Task</div>
									<div class="stat-value text-lg truncate self-end">
										CNCF Onboarding
									</div>
									<div class="stat-desc">8,500 average tokens used per run</div>
								</div>
							</div>
						</div>
					</div>
				</div>

				<div class="card h-fit bg-base-100 dark:bg-base-200">
					<div class="card-body">
						<h2 class="card-title text-base">Recent Workflow Activity</h2>

						<table class="table">
							<thead>
								<tr>
									<th>Workflow</th>
									<th>Created</th>
									<th>Time to Complete</th>
									<th>Tokens Used</th>
									<th>User</th>
									<th>Workspace</th>
								</tr>
							</thead>
							<tbody>
								{#each taskRuns as taskRun (taskRun.id)}
									<tr>
										<td>{taskRun.task}</td>
										<td>{formatTimeAgo(taskRun.created).relativeTime}</td>
										<td>{taskRun.averageCompletionTime}</td>
										<td>{taskRun.tokensUsed}</td>
										<td>{taskRun.user}</td>
										<td>{taskRun.workspace}</td>
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
						<h3 class="card-title justify-between">
							Workspaces
							<button class="btn btn-square btn-ghost btn-xs tooltip" data-tip="Create new workspace">
								<Plus />
							</button>
						</h3>
						<ul class="list bg-base-100 rounded-box">
							<li class="pb-2 text-xs opacity-60 tracking-wide">Most recently created workspaces</li>

							{#each [...workspaceService.workspaces, mocks.workspace] as workspace (workspace.id)}
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
							{#each sharedWorkspaces as workspace (workspace.id)}
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
						<h3 class="card-title">Workflows</h3>
						<ul class="list bg-base-100 rounded-box">
							<li class="pb-2 text-xs opacity-60 tracking-wide">Most recently created workflows</li>
							{#each tasks as task (task.id)}
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
							{#each files as file (file.id)}
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