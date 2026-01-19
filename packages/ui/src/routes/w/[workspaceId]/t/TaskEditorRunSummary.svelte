<script lang="ts">
	import { fade, slide } from 'svelte/transition';
	import type { Task } from './types';

	type Props = {
		task: Task;
		completed: boolean;
		error: boolean;
		runSummary: string;
		totalRunTime: number;
		totalRunTokens: number;
		showSidebarThread: boolean;
	};

	let {
		task,
		completed,
		error,
		runSummary,
		totalRunTime,
		totalRunTokens,
		showSidebarThread
	}: Props = $props();

	let containerRef = $state<HTMLElement | null>(null);

	export function getElement(): HTMLElement | null {
		return containerRef;
	}
</script>

<div
	bind:this={containerRef}
	in:fade
	out:slide={{ axis: 'y', duration: 150 }}
	class="w-full flex flex-col justify-center items-center pl-22 pb-4 {showSidebarThread
		? ''
		: 'md:pr-22'}"
>
	<div
		class="w-full flex flex-col justify-center items-center border border-transparent dark:border-base-300 bg-base-100 dark:bg-base-200 shadow-xs rounded-field p-6 pb-8"
	>
		{#if completed && !error}
			<h4 class="text-xl font-semibold">Workflow Completed</h4>
			<p class="text text-base-content/50 text-center mt-1">
				The workflow has completed successfully. Here are your summarized results:
			</p>

			{#if runSummary}
				<div class="prose mt-4 text-left w-full max-w-none">
					{@html runSummary}
				</div>
				<div class="divider"></div>
			{/if}

			<p class="text-sm text-center">
				The workflow completed <b>{task.steps.length}</b> out of <b>{task.steps.length}</b> steps.
			</p>
			<p class="text-sm text-center mt-1">
				It took a total time of <b>{(totalRunTime / 1000).toFixed(1)}s</b> to complete.
			</p>
			<p class="text-sm text-center mt-1">
				A total of <b>{totalRunTokens}</b> tokens were used.
			</p>
		{:else if error}
			<h4 class="text-xl font-semibold">Workflow Failed</h4>
			<p class="text-sm text-base-content/50 text-center mt-1">
				The workflow did not complete successfully. Please try again.
			</p>
		{:else}
			<div class="skeleton skeleton-text">The workflow is running...</div>
		{/if}
	</div>
</div>
