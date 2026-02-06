<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import Threads from '$lib/components/Threads.svelte';
	import Notifications from '$lib/components/Notifications.svelte';
	import { defaultChatApi } from '$lib/chat.svelte';
	import { NotificationStore } from '$lib/stores/notifications.svelte';
	import { setNotificationContext } from '$lib/context/notifications.svelte';
	import { SidebarStore } from '$lib/stores/sidebar.svelte';
	import { setSidebarContext } from '$lib/context/sidebar.svelte';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { Menu, X, SidebarOpen, SidebarClose, Sun, Moon, SquarePen } from '@lucide/svelte';
	import type { Chat } from '$lib/types';
	import { resolve } from '$app/paths';

	let { children } = $props();

	let threads = $state<Chat[]>([]);
	let isLoading = $state(true);
	let currentTheme = $state('nanobotlight');
	let currentLogoUrl = $state('/assets/nanobot.svg');
	const root = resolve('/');
	const newThread = resolve('/');
	const notifications = new NotificationStore();
	const sidebar = new SidebarStore();

	// Set contexts for global access
	setNotificationContext(notifications);
	setSidebarContext(sidebar);

	onMount(async () => {
		// Load theme from localStorage or detect system preference
		if (browser) {
			const savedTheme = localStorage.getItem('theme');
			if (savedTheme) {
				currentTheme = savedTheme;
			} else {
				// Default to system preference
				const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
				currentTheme = prefersDark ? 'nanobotdark' : 'nanobotlight';
			}
			// Set theme on document
			document.documentElement.setAttribute('data-theme', currentTheme);
		}

		threads = await defaultChatApi.getThreads()
		isLoading = false;
	});

	$effect(() => {
		if (currentTheme) {
			requestAnimationFrame(() => {
				const logoUrlAttribute = getComputedStyle(document.documentElement).getPropertyValue('--logo-url');
				currentLogoUrl = logoUrlAttribute || '/assets/nanobot.svg';
			});
		}
	})

	async function handleRenameThread(threadId: string, newTitle: string) {
		try {
			await defaultChatApi.renameThread(threadId, newTitle);
			const threadIndex = threads.findIndex((t) => t.id === threadId);
			if (threadIndex !== -1) {
				threads[threadIndex].title = newTitle;
			}
			notifications.success('Thread Renamed', `Successfully renamed to "${newTitle}"`);
		} catch (error) {
			notifications.error('Rename Failed', 'Unable to rename the thread. Please try again.');
			console.error('Failed to rename thread:', error);
		}
	}

	async function handleDeleteThread(threadId: string) {
		try {
			await defaultChatApi.deleteThread(threadId);
			const threadToDelete = threads.find((t) => t.id === threadId);
			threads = threads.filter((t) => t.id !== threadId);
			notifications.success('Thread Deleted', `Deleted "${threadToDelete?.title || 'thread'}"`);
		} catch (error) {
			notifications.error('Delete Failed', 'Unable to delete the thread. Please try again.');
			console.error('Failed to delete thread:', error);
		}
	}

	function toggleTheme() {
		if (browser) {
			currentTheme = currentTheme === 'nanobotlight' ? 'nanobotdark' : 'nanobotlight';
			document.documentElement.setAttribute('data-theme', currentTheme);
			localStorage.setItem('theme', currentTheme);
		}
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link rel="stylesheet" href="/assets/theme.css" />
</svelte:head>

<!-- Unified responsive layout -->
<div class="relative flex h-dvh">
	<!-- Sidebar - responsive behavior -->
	<div
		class="
		bg-base-200 transition-all duration-300 ease-in-out
		{sidebar.isCollapsed ? 'hidden lg:block lg:w-0' : 'hidden lg:block lg:w-80'}
		{sidebar.isMobileOpen ? 'fixed inset-y-0 left-0 z-40 block! w-80' : 'lg:relative'}
	"
	>
		<div class="flex h-full flex-col {sidebar.isCollapsed ? 'lg:overflow-hidden' : ''}">
			<!-- Sidebar header -->
			<div
				class="flex h-15 items-center justify-between p-2 {!sidebar.isCollapsed ? 'min-w-80' : ''}"
			>
				<a href={root} class="flex items-center gap-2 text-xl font-bold hover:opacity-80">
					<img src={currentLogoUrl} alt="Nanobot" class="h-12" />
				</a>
				<div class="flex items-center gap-1">
					<a href={newThread} class="btn p-1 btn-ghost btn-sm" aria-label="New thread">
						<SquarePen class="h-5 w-5" />
					</a>
					<button
						onclick={() => {
							if (window.innerWidth >= 1024) {
								sidebar.toggle();
							} else {
								sidebar.closeMobile();
							}
						}}
						class="btn p-1 btn-ghost btn-sm"
						aria-label={sidebar.isCollapsed ? 'Open sidebar' : 'Close sidebar'}
					>
						<!-- Desktop collapsed state -->
						<span class="hidden lg:inline">
							{#if sidebar.isCollapsed}
								<SidebarOpen class="h-5 w-5" />
							{:else}
								<SidebarClose class="h-5 w-5" />
							{/if}
						</span>
						<!-- Mobile state -->
						<span class="lg:hidden">
							<X class="h-5 w-5" />
						</span>
					</button>
				</div>
			</div>

			<!-- Threads and Workspaces list -->
			<div class="flex-1 overflow-hidden {!sidebar.isCollapsed ? 'min-w-80' : ''}">
				<div class="flex h-full flex-col">
					<!-- Threads section (takes up ~40% of available space) -->
					<div class='flex-shrink-0 overflow-y-auto'>
						<Threads
							{threads}
							onRename={handleRenameThread}
							onDelete={handleDeleteThread}
							{isLoading}
							onThreadClick={() => sidebar.closeMobile()}
						/>
					</div>
				</div>
			</div>

			<!-- Theme switcher - bottom left corner -->
			<div class="absolute bottom-4 left-4 z-50">
				<button
					onclick={toggleTheme}
					class="btn btn-circle border-base-300 bg-base-100 shadow-lg btn-sm"
					aria-label="Toggle theme"
				>
					{#if currentTheme === 'nanobotlight'}
						<Moon class="h-4 w-4" />
					{:else}
						<Sun class="h-4 w-4" />
					{/if}
				</button>
			</div>
		</div>
	</div>

	<!-- Mobile sidebar backdrop -->
	{#if sidebar.isMobileOpen}
		<div
			class="fixed inset-0 z-30 bg-black/50 lg:hidden"
			role="button"
			tabindex="0"
			onclick={() => sidebar.closeMobile()}
			onkeydown={(e) => (e.key === 'Enter' || e.key === ' ' ? sidebar.closeMobile() : null)}
		></div>
	{/if}

	<!-- Collapsed sidebar toggle (desktop only) -->
	{#if sidebar.isCollapsed}
		<div class="absolute top-0 left-0 z-10 hidden h-15 items-center bg-transparent p-2 lg:flex">
			<div class="flex items-center gap-2">
				<a href={root} class="flex items-center gap-2 text-xl font-bold hover:opacity-80">
					<img src={currentLogoUrl} alt="Nanobot" class="h-12" />
				</a>
				<a href={newThread} class="btn p-1 btn-ghost btn-sm" aria-label="New thread">
					<SquarePen class="h-4 w-4" />
				</a>
				<button
					onclick={() => sidebar.toggle()}
					class="btn p-1 btn-ghost btn-sm"
					aria-label="Open sidebar"
				>
					<SidebarOpen class="h-4 w-4" />
				</button>
			</div>
		</div>
	{/if}

	<!-- Mobile menu button -->
	{#if !sidebar.isMobileOpen}
		<div class="absolute top-4 left-4 z-50 flex gap-2 lg:hidden">
			<a
				href={newThread}
				class="btn border border-base-300 bg-base-100/80 btn-ghost backdrop-blur-sm btn-sm"
				aria-label="New thread"
			>
				<SquarePen class="h-5 w-5" />
			</a>
			<button
				onclick={() => sidebar.toggleMobile()}
				class="btn border border-base-300 bg-base-100/80 btn-ghost backdrop-blur-sm btn-sm"
				aria-label="Open sidebar"
			>
				<Menu class="h-5 w-5" />
			</button>
		</div>
	{/if}

	<!-- Main content area -->
	<div class="h-dvh flex-1">
		{@render children?.()}
	</div>
</div>

<!-- Notifications -->
<Notifications />
