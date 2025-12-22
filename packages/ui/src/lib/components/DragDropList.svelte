<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props<T> {
		items: T[];
		getKey?: (item: T) => string | number;
		blockHandle?: Snippet<[{ startDrag: (e: MouseEvent) => void; currentItem: T | null }]>;
		children: Snippet<[{ item: T; index: number; isDragging: boolean }]>;
		onreorder?: (items: T[]) => void;
		scrollContainerEl?: HTMLElement | null;
		class?: string;
		classes?: {
			dropIndicator?: string;
			item?: string;
		};
		as?: 'div' | 'ul'
		childrenAs?: 'li' | 'div'
	}

	let {
		items = $bindable(),
		getKey = (item: any) => item.id,
		blockHandle,
		children,
		onreorder,
		scrollContainerEl = null,
		class: className = '',
		classes = {},
		as = 'div',
		childrenAs = 'div',
	}: Props<any> = $props();

	let currentFocusedElement = $state<HTMLElement | null>(null);
	let currentItem = $state<(typeof items)[number] | null>(null);
	let handleElement = $state<HTMLElement | null>(null);
	let handleYPosition = $state(0);
	let internalScrollContainer = $state<HTMLElement | null>(null);
	let itemsContainer = $state<HTMLElement | null>(null);
	let listContainer = $state<HTMLElement | null>(null);

	// Use external scroll container if provided, otherwise use internal
	let scrollContainer = $derived(scrollContainerEl ?? internalScrollContainer);

	// Drag & drop state
	let isDragging = $state(false);
	let draggedIndex = $state<number | null>(null);
	let dropTargetIndex = $state<number | null>(null);
	let dragGhost = $state<HTMLElement | null>(null);
	let dragOffsetY = $state(0);

	// Auto-scroll state
	const SCROLL_THRESHOLD = 80; // pixels from edge to start scrolling
	const SCROLL_SPEED = 12; // pixels per frame
	let autoScrollAnimationId: number | null = null;
	let lastMouseY = 0;

	function updateHandlePosition() {
		if (!currentFocusedElement || !scrollContainer || !itemsContainer) return;

		const elementTop = currentFocusedElement.offsetTop;
		const elementBottom = elementTop + currentFocusedElement.offsetHeight;
		const scrollTop = scrollContainer.scrollTop;
		const containerOffset = itemsContainer.offsetTop;

		// Calculate where the viewport top is relative to the items container
		const viewportTopRelative = scrollTop - containerOffset;

		// Clamp: stay at element top, or stick to viewport top (but not past element bottom)
		const clampedPosition = Math.max(elementTop, Math.min(viewportTopRelative, elementBottom - 40));

		handleYPosition = clampedPosition;
	}

	function getItemIndexFromElement(element: HTMLElement): number | null {
		if (!listContainer) return null;
		const itemElements = Array.from(listContainer.children).filter(
			(el) => !el.classList.contains('drop-indicator')
		);
		const index = itemElements.indexOf(element);
		return index >= 0 ? index : null;
	}

	function startDrag(e: MouseEvent) {
		if (!currentFocusedElement || !listContainer) return;

		e.preventDefault();

		const index = getItemIndexFromElement(currentFocusedElement);
		if (index === null) return;

		isDragging = true;
		draggedIndex = index;

		// Create ghost clone
		const rect = currentFocusedElement.getBoundingClientRect();
		dragOffsetY = e.clientY - rect.top;

		const ghost = currentFocusedElement.cloneNode(true) as HTMLElement;
		ghost.classList.add('drag-ghost');
		ghost.style.position = 'fixed';
		ghost.style.width = `${rect.width}px`;
		ghost.style.left = `${rect.left}px`;
		ghost.style.top = `${e.clientY - dragOffsetY}px`;
		ghost.style.pointerEvents = 'none';
		ghost.style.zIndex = '1000';
		ghost.style.opacity = '0.7';
		ghost.style.transform = 'scale(0.98)';
		document.body.appendChild(ghost);
		dragGhost = ghost;

		// Add event listeners
		document.addEventListener('mousemove', handleDragMove);
		document.addEventListener('mouseup', endDrag);
	}

	function startAutoScroll() {
		if (autoScrollAnimationId !== null) return;

		function scrollStep() {
			if (!isDragging || !scrollContainer) {
				stopAutoScroll();
				return;
			}

			const containerRect = scrollContainer.getBoundingClientRect();
			const distanceFromTop = lastMouseY - containerRect.top;
			const distanceFromBottom = containerRect.bottom - lastMouseY;

			let scrollAmount = 0;

			if (distanceFromTop < SCROLL_THRESHOLD && scrollContainer.scrollTop > 0) {
				// Scroll up - faster when closer to edge
				const intensity = 1 - distanceFromTop / SCROLL_THRESHOLD;
				scrollAmount = -SCROLL_SPEED * intensity;
			} else if (
				distanceFromBottom < SCROLL_THRESHOLD &&
				scrollContainer.scrollTop < scrollContainer.scrollHeight - scrollContainer.clientHeight
			) {
				// Scroll down - faster when closer to edge
				const intensity = 1 - distanceFromBottom / SCROLL_THRESHOLD;
				scrollAmount = SCROLL_SPEED * intensity;
			}

			if (scrollAmount !== 0) {
				scrollContainer.scrollTop += scrollAmount;

				// Update ghost position to stay with cursor
				if (dragGhost) {
					dragGhost.style.top = `${lastMouseY - dragOffsetY}px`;
				}

				// Re-calculate drop target after scrolling
				updateDropTarget(lastMouseY);
			}

			autoScrollAnimationId = requestAnimationFrame(scrollStep);
		}

		autoScrollAnimationId = requestAnimationFrame(scrollStep);
	}

	function stopAutoScroll() {
		if (autoScrollAnimationId !== null) {
			cancelAnimationFrame(autoScrollAnimationId);
			autoScrollAnimationId = null;
		}
	}

	function updateDropTarget(mouseY: number) {
		if (!listContainer || draggedIndex === null) return;

		// Calculate drop target - filter out drop indicators
		const itemElements = Array.from(listContainer.children).filter(
			(el) => !el.classList.contains('drop-indicator')
		) as HTMLElement[];
		let newDropIndex: number | null = null;

		for (let i = 0; i < itemElements.length; i++) {
			const itemEl = itemElements[i];
			const rect = itemEl.getBoundingClientRect();
			const midY = rect.top + rect.height / 2;

			if (mouseY < midY) {
				newDropIndex = i;
				break;
			}
		}

		// If we didn't find a position, drop at the end
		if (newDropIndex === null) {
			newDropIndex = itemElements.length;
		}

		// If hovering over the dragged item (index or index+1), keep indicator at original position
		// Both positions result in "no change", so show consistent indicator
		if (newDropIndex === draggedIndex || newDropIndex === draggedIndex + 1) {
			newDropIndex = draggedIndex;
		}

		dropTargetIndex = newDropIndex;
	}

	function handleDragMove(e: MouseEvent) {
		if (!isDragging || !dragGhost || !listContainer || draggedIndex === null) return;

		lastMouseY = e.clientY;

		// Move ghost
		dragGhost.style.top = `${e.clientY - dragOffsetY}px`;

		// Start auto-scroll if near edges
		if (scrollContainer) {
			const containerRect = scrollContainer.getBoundingClientRect();
			const distanceFromTop = e.clientY - containerRect.top;
			const distanceFromBottom = containerRect.bottom - e.clientY;

			if (distanceFromTop < SCROLL_THRESHOLD || distanceFromBottom < SCROLL_THRESHOLD) {
				startAutoScroll();
			} else {
				stopAutoScroll();
			}
		}

		updateDropTarget(e.clientY);
	}

	function endDrag() {
		if (!isDragging) return;

		// Stop auto-scroll
		stopAutoScroll();

		// Remove ghost
		if (dragGhost) {
			dragGhost.remove();
			dragGhost = null;
		}

		// Reorder items if we have a valid drop target and position changed
		if (dropTargetIndex !== null && draggedIndex !== null) {
			// Skip if dropping at original position (no change)
			const isUnchanged = dropTargetIndex === draggedIndex || dropTargetIndex === draggedIndex + 1;

			if (!isUnchanged) {
				const newItems = [...items];
				const [removed] = newItems.splice(draggedIndex, 1);

				// Adjust index if dropping after the original position
				const insertIndex = dropTargetIndex > draggedIndex ? dropTargetIndex - 1 : dropTargetIndex;
				newItems.splice(insertIndex, 0, removed);

				items = newItems;
				onreorder?.(newItems);
			}
		}

		// Reset state
		isDragging = false;
		draggedIndex = null;
		dropTargetIndex = null;

		// Remove event listeners
		document.removeEventListener('mousemove', handleDragMove);
		document.removeEventListener('mouseup', endDrag);
	}

	export function getIsDragging() {
		return isDragging;
	}
</script>

<div
	class="drag-drop-list {className}"
	bind:this={internalScrollContainer}
	onscroll={updateHandlePosition}
>
	<div
		class="drag-drop-items-container"
		bind:this={itemsContainer}
		onmouseleave={() => {
			currentFocusedElement = null;
			currentItem = null;
		}}
		role="presentation"
	>
		{#if blockHandle}
			<div
				class="block-handle"
				data-show={currentFocusedElement !== null && !isDragging}
				bind:this={handleElement}
				style="top: {handleYPosition}px;"
			>
				{#if blockHandle}
					{@render blockHandle({ startDrag, currentItem })}
				{/if}
			</div>
		{/if}
		<svelte:element this={as} class="drag-drop-list-container" bind:this={listContainer}>
			{#each items as item, index (getKey(item))}
				<div
					class="drop-indicator {classes.dropIndicator}"
					data-active={dropTargetIndex !== null && dropTargetIndex === index}
				></div>
				<svelte:element this={childrenAs}
					class="drag-drop-item w-full {classes.item}"
					class:dragging={isDragging && draggedIndex === index}
					onmouseenter={(e: MouseEvent) => {
						if (isDragging) return;
						currentFocusedElement = e.currentTarget as HTMLElement;
						currentItem = item;
						updateHandlePosition();
					}}
					role="presentation"
					onmousedown={!blockHandle ? startDrag : undefined}
				>
					{@render children({ item, index, isDragging: isDragging && draggedIndex === index })}
				</svelte:element>
			{/each}
			<div
				class="drop-indicator"
				data-active={dropTargetIndex !== null && dropTargetIndex === items.length}
			></div>
		</svelte:element>
	</div>
</div>

<style>
	.drag-drop-list {
		position: relative;
	}

	.drag-drop-items-container {
		position: relative;
	}

	.drag-drop-list-container {
		display: flex;
		flex-direction: column;
	}

	.block-handle {
		position: absolute;
		display: flex;
		top: 0;
		left: 0;
		z-index: 2;
		opacity: 0;
		transition: all 0.2s ease-in-out;
	}

	.block-handle[data-show='true'] {
		opacity: 1;
	}

	.dragging {
		opacity: 0.3;
	}

	.drop-indicator {
		background: transparent;
		border-radius: 0.25rem;
		position: relative;
	}

	.drop-indicator[data-active='true'] {
		background: var(--color-base-300);
	}
</style>
