<script lang="ts">
	import { EllipsisVertical, PencilLine, Play, ReceiptText, Square } from '@lucide/svelte';
	import { fade, slide } from 'svelte/transition';
	import type { Task } from './types';

	type Props = {
		task: Task;
		running: boolean;
		completed: boolean;
		showAlternateHeader: boolean;
		showTaskTitle: boolean;
		showTaskDescription: boolean;
		isSidebarCollapsed: boolean;
		onRun: () => void;
		onCancel: () => void;
		onToggleTitle: (value: boolean) => void;
		onToggleDescription: (value: boolean) => void;
	};

	let {
		task,
		running,
		completed,
		showAlternateHeader,
		showTaskTitle,
		showTaskDescription,
		isSidebarCollapsed,
		onRun,
		onCancel,
		onToggleTitle,
		onToggleDescription
	}: Props = $props();
</script>

<div class="sticky top-0 left-0 w-full bg-base-200 dark:bg-base-100 z-10 py-4">
	<div in:fade class="flex flex-col grow">
		<div class="flex w-full items-center gap-4 {isSidebarCollapsed ? 'pl-68' : ''}">
			{#if showAlternateHeader}
				<p in:fade class="flex grow text-xl font-semibold">{task.name}</p>
			{:else if showTaskTitle}
				<input
					name="title"
					class="input input-ghost input-lg w-full placeholder:text-base-content/30 font-semibold"
					type="text"
					placeholder="Workflow title"
					bind:value={task.name}
				/>
			{:else}
				<div class="w-full"></div>
			{/if}
			<div class="flex shrink-0 items-center gap-2">
				<div class="flex">
					<button
						class="btn btn-primary w-48 {running ? 'tooltip tooltip-bottom' : ''}"
						data-tip="Cancel current run"
						onclick={() => {
							if (running && !completed) {
								onCancel();
							} else {
								onRun();
							}
						}}
					>
						{#if running && !completed}
							<Square class="size-4" />
						{:else}
							Run <Play class="size-4" />
						{/if}
					</button>
				</div>
				<button
					class="btn btn-ghost btn-square"
					popoverTarget="task-actions"
					style="anchor-name: --task-actions-anchor;"
				>
					<EllipsisVertical class="text-base-content/50" />
				</button>

				<ul
					class="dropdown flex flex-col gap-1 dropdown-end dropdown-bottom menu w-64 rounded-box bg-base-100 dark:bg-base-300 shadow-sm border border-base-300"
					popover="auto"
					id="task-actions"
					style="position-anchor: --task-actions-anchor;"
				>
					<li>
						<label for="task-title" class="flex gap-2 justify-between items-center">
							<span class="flex items-center gap-2">
								<PencilLine class="size-4" />
								Workflow title
							</span>
							<input
								type="checkbox"
								class="toggle toggle-sm"
								id="task-title"
								checked={showTaskTitle}
								onchange={(e) => onToggleTitle((e.target as HTMLInputElement)?.checked ?? false)}
							/>
						</label>
					</li>
					<li>
						<label for="task-description" class="flex gap-2 justify-between items-center">
							<span class="flex items-center gap-2">
								<ReceiptText class="size-4" />
								Workflow description
							</span>
							<input
								type="checkbox"
								class="toggle toggle-sm"
								id="task-description"
								checked={showTaskDescription}
								onchange={(e) =>
									onToggleDescription((e.target as HTMLInputElement)?.checked ?? false)}
							/>
						</label>
					</li>
				</ul>
			</div>
		</div>
		{#if !showAlternateHeader && showTaskDescription}
			<input
				out:slide={{ axis: 'y' }}
				name="description"
				class="input input-ghost w-full placeholder:text-base-content/30"
				type="text"
				placeholder="Workflow description"
				bind:value={task.description}
			/>
		{/if}
	</div>
</div>
