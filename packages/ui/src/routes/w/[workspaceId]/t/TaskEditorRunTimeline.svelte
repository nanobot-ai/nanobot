<script lang="ts">
	import { CircleCheck, LoaderCircle, Circle, CircleAlert, X, CircleMinus } from '@lucide/svelte';
	import { fade } from 'svelte/transition';
	import type { Task } from './types';

	type Props = {
		task: Task;
		currentRunStepId: string | null;
		completed: boolean;
		error: boolean;
		runSummary: string;
		onStepClick: (stepId: string) => void;
		onSummaryClick: () => void;
		onClose: () => void;
		parentElement: HTMLElement | null;
	};

	let {
		task,
		currentRunStepId,
		completed,
		error,
		runSummary,
		onStepClick,
		onSummaryClick,
		onClose,
		parentElement
	}: Props = $props();

	// Draggable state
	let position = $state({ x: 16, y: 16 });
	let isDragging = $state(false);
	let containerRef = $state<HTMLElement | null>(null);

	function startDrag(e: MouseEvent) {
		if (e.button !== 0) return;
		const target = e.target as HTMLElement;
		if (target.closest('button') || target.closest('a') || target.closest('input')) return;

		e.preventDefault();
		isDragging = true;

		const startX = e.clientX;
		const startY = e.clientY;
		const startPosX = position.x;
		const startPosY = position.y;

		function onMouseMove(e: MouseEvent) {
			if (!containerRef || !parentElement) return;

			const parentRect = parentElement.getBoundingClientRect();
			const timelineRect = containerRef.getBoundingClientRect();

			const deltaX = e.clientX - startX;
			const deltaY = startY - e.clientY;

			const newX = Math.max(0, Math.min(startPosX + deltaX, parentRect.width - timelineRect.width));
			const newY = Math.max(0, Math.min(startPosY + deltaY, parentRect.height - timelineRect.height));

			position = { x: newX, y: newY };
		}

		function onMouseUp() {
			isDragging = false;
			document.removeEventListener('mousemove', onMouseMove);
			document.removeEventListener('mouseup', onMouseUp);
		}

		document.addEventListener('mousemove', onMouseMove);
		document.addEventListener('mouseup', onMouseUp);
	}
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
	bind:this={containerRef}
	class="absolute select-none {isDragging ? 'cursor-grabbing' : 'cursor-grab'}"
	style="left: {position.x}px; bottom: {position.y}px;"
	in:fade
	role="region"
	aria-label="Workflow progress timeline"
	onmousedown={startDrag}
>
	<button
		class="cursor-pointer p-0.5 tooltip tooltip-left text-base-content/50 hover:text-base-content"
		data-tip="Close"
		onclick={onClose}
	>
		<X class="size-4" />
	</button>
	<ul in:fade class="timeline timeline-snap-icon timeline-vertical timeline-compact grow">
		{#each task.steps as step, index (step.id)}
			{@const isBeforeCurrentStep = index < task.steps.findIndex((s) => s.id === currentRunStepId)}
			{@const isAfterCurrentStep = index > task.steps.findIndex((s) => s.id === currentRunStepId)}
			<li>
				{#if index > 0}
					<hr class="timeline-connector w-0.5 {isBeforeCurrentStep ? 'completed' : ''}" />
				{/if}
				<div class="timeline-middle">
					{#if isBeforeCurrentStep || (currentRunStepId === step.id && completed && !error)}
						<CircleCheck class="size-5 text-primary" />
					{:else if currentRunStepId === step.id && !error && !completed}
						<LoaderCircle class="size-5 animate-spin shrink-0 text-base-content/50" />
					{:else if currentRunStepId === step.id && error}
						<CircleAlert class="size-5 text-error/50" />
					{:else if isAfterCurrentStep && error}
						<CircleMinus class="size-5 text-error/50" />
					{:else}
						<Circle class="size-5 text-base-content/50" />
					{/if}
				</div>
				<button
					class="timeline-end timeline-box py-2 cursor-pointer hover:bg-base-200
						{currentRunStepId === step.id ? error ? 'border-error' : !completed ? 'border-primary' : '' : ''}"
					onclick={() => onStepClick(step.id)}
				>
					<div>
						{step.name}
					</div>
				</button>
				<hr class="timeline-connector w-0.5 {isBeforeCurrentStep ? 'completed' : ''}" />
			</li>
		{/each}
		<li>
			<hr class="timeline-connector w-0.5 {runSummary && completed ? 'completed' : ''}" />
			<div class="timeline-middle">
				{#if completed && runSummary}
					<CircleCheck class="size-5 text-primary" />
				{:else if error}
					<CircleMinus class="size-5 text-error/50" />
				{:else}
					<Circle class="size-5 text-base-content/50" />
				{/if}
			</div>
			<button
				class="timeline-end timeline-box py-2 {runSummary
					? 'border-primary cursor-pointer hover:bg-base-200'
					: 'cursor-default opacity-50'}
				"
				onclick={onSummaryClick}
			>
				<div>Run Summary</div>
			</button>
		</li>
	</ul>
</div>
