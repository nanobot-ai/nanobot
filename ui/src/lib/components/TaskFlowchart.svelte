<script lang="ts">
	import type { TaskFlowchart, TaskNode } from '$lib/types';

	interface Props {
		flowchart: TaskFlowchart;
		selectedNodeId?: string;
		onNodeClick: (node: TaskNode) => void;
	}

	let { flowchart, selectedNodeId, onNodeClick }: Props = $props();

	// Calculate SVG viewBox based on node positions
	const svgBounds = $derived(() => {
		const nodes = flowchart.nodes;
		if (nodes.length === 0) return { minX: 0, minY: 0, width: 800, height: 600 };

		const minX = Math.min(...nodes.map((n) => n.position.x)) - 50;
		const minY = Math.min(...nodes.map((n) => n.position.y)) - 50;
		const maxX = Math.max(...nodes.map((n) => n.position.x)) + 200;
		const maxY = Math.max(...nodes.map((n) => n.position.y)) + 100;

		return {
			minX,
			minY,
			width: maxX - minX,
			height: maxY - minY
		};
	});

	function getNodeShape(node: TaskNode) {
		switch (node.type) {
			case 'start':
			case 'end':
				return 'rounded';
			case 'decision':
				return 'diamond';
			case 'process':
			default:
				return 'rectangle';
		}
	}

	function getNodeColor(node: TaskNode) {
		if (node.completed) return '#10b981'; // success green
		switch (node.type) {
			case 'start':
				return '#3b82f6'; // blue
			case 'end':
				return '#8b5cf6'; // purple
			case 'decision':
				return '#f59e0b'; // orange
			case 'process':
			default:
				return '#6b7280'; // gray
		}
	}

	// Calculate edge paths with anchor points
	function getEdgePath(edge: { source: string; target: string }) {
		const sourceNode = flowchart.nodes.find((n) => n.id === edge.source);
		const targetNode = flowchart.nodes.find((n) => n.id === edge.target);

		if (!sourceNode || !targetNode) return '';

		// Get anchor points for source (exit point) and target (entry point)
		const sourcePoint = getNodeAnchor(sourceNode, targetNode);
		const targetPoint = getNodeAnchor(targetNode, sourceNode);

		// Simple straight line connecting anchor points
		return `M ${sourcePoint.x} ${sourcePoint.y} L ${targetPoint.x} ${targetPoint.y}`;
	}

	// Get anchor point on a node based on relative position to another node
	function getNodeAnchor(node: TaskNode, otherNode: TaskNode): { x: number; y: number } {
		const nodeWidth = 150;
		const nodeHeight = node.type === 'decision' ? 80 : 60;
		const centerX = node.position.x + nodeWidth / 2;
		const centerY = node.position.y + nodeHeight / 2;

		// Calculate relative position of other node
		const dx = otherNode.position.x + 75 - centerX;
		const dy = otherNode.position.y + 30 - centerY;

		// Determine which anchor point to use based on angle
		const angle = Math.atan2(dy, dx) * (180 / Math.PI);

		// For exit points (source), prefer bottom/right. For entry points (target), prefer top/left
		if (angle >= -45 && angle < 45) {
			// Right
			return { x: node.position.x + nodeWidth, y: centerY };
		} else if (angle >= 45 && angle < 135) {
			// Bottom
			return { x: centerX, y: node.position.y + nodeHeight };
		} else if (angle >= 135 || angle < -135) {
			// Left
			return { x: node.position.x, y: centerY };
		} else {
			// Top
			return { x: centerX, y: node.position.y };
		}
	}

	function getEdgeLabelPosition(edge: { source: string; target: string; label?: string }) {
		if (!edge.label) return null;

		const sourceNode = flowchart.nodes.find((n) => n.id === edge.source);
		const targetNode = flowchart.nodes.find((n) => n.id === edge.target);

		if (!sourceNode || !targetNode) return null;

		// Use anchor points for label positioning
		const sourcePoint = getNodeAnchor(sourceNode, targetNode);
		const targetPoint = getNodeAnchor(targetNode, sourceNode);

		return {
			x: (sourcePoint.x + targetPoint.x) / 2,
			y: (sourcePoint.y + targetPoint.y) / 2
		};
	}
</script>

<svg
	viewBox="{svgBounds().minX} {svgBounds().minY} {svgBounds().width} {svgBounds().height}"
	width={svgBounds().width}
	height={svgBounds().height}
>
	<!-- Define arrow marker -->
	<defs>
		<marker id="arrowhead" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
			<polygon points="0 0, 10 3, 0 6" fill="#6b7280" />
		</marker>
	</defs>

	<!-- Draw edges first (so they appear behind nodes) -->
	{#each flowchart.edges as edge (edge.id)}
		<g>
			<path
				d={getEdgePath(edge)}
				stroke="#6b7280"
				stroke-width="2"
				fill="none"
				marker-end="url(#arrowhead)"
			/>
			{#if edge.label}
				{@const labelPos = getEdgeLabelPosition(edge)}
				{#if labelPos}
					<text
						x={labelPos.x}
						y={labelPos.y}
						text-anchor="middle"
						class="fill-base-content text-sm font-medium"
						dominant-baseline="middle"
					>
						{edge.label}
					</text>
				{/if}
			{/if}
		</g>
	{/each}

	<!-- Draw nodes -->
	{#each flowchart.nodes as node (node.id)}
		{@const shape = getNodeShape(node)}
		{@const color = getNodeColor(node)}
		{@const isSelected = selectedNodeId === node.id}
		{@const hasAssignments =
			(node.tools && node.tools.length > 0) ||
			(node.agents && node.agents.length > 0) ||
			(node.tasks && node.tasks.length > 0)}

		<g
			role="button"
			class="cursor-pointer transition-opacity hover:opacity-80"
			onclick={() => onNodeClick(node)}
			tabindex="0"
			onkeydown={(e) => {
				if (e.key === 'Enter' || e.key === ' ') {
					e.preventDefault();
					onNodeClick(node);
				}
			}}
		>
			{#if shape === 'diamond'}
				<!-- Decision node (diamond) -->
				<polygon
					points="{node.position.x + 75},
{node.position.y}
{node.position.x + 150},{node.position.y + 40}
{node.position.x + 75},{node.position.y + 80}
{node.position.x},{node.position.y + 40}"
					fill={color}
					stroke={isSelected ? '#000' : color}
					stroke-width={isSelected ? '3' : '1'}
					opacity={node.completed ? '0.7' : '1'}
				/>
				<text
					x={node.position.x + 75}
					y={node.position.y + 40}
					text-anchor="middle"
					dominant-baseline="middle"
					class="pointer-events-none fill-white text-sm font-medium"
				>
					{node.label}
				</text>
			{:else if shape === 'rounded'}
				<!-- Start/End node (rounded rectangle) -->
				<rect
					x={node.position.x}
					y={node.position.y}
					width="150"
					height="60"
					rx="30"
					ry="30"
					fill={color}
					stroke={isSelected ? '#000' : color}
					stroke-width={isSelected ? '3' : '1'}
					opacity={node.completed ? '0.7' : '1'}
				/>
				<text
					x={node.position.x + 75}
					y={node.position.y + 30}
					text-anchor="middle"
					dominant-baseline="middle"
					class="pointer-events-none fill-white text-sm font-medium"
				>
					{node.label}
				</text>
			{:else}
				<!-- Process node (rectangle) -->
				<rect
					x={node.position.x}
					y={node.position.y}
					width="150"
					height="60"
					rx="4"
					ry="4"
					fill={color}
					stroke={isSelected ? '#000' : color}
					stroke-width={isSelected ? '3' : '1'}
					opacity={node.completed ? '0.7' : '1'}
				/>
				<text
					x={node.position.x + 75}
					y={node.position.y + 30}
					text-anchor="middle"
					dominant-baseline="middle"
					class="pointer-events-none fill-white text-sm font-medium"
				>
					{node.label}
				</text>
			{/if}

			<!-- Completion checkmark -->
			{#if node.completed}
				<circle cx={node.position.x + 140} cy={node.position.y + 10} r="12" fill="#10b981" />
				<text
					x={node.position.x + 140}
					y={node.position.y + 10}
					text-anchor="middle"
					dominant-baseline="middle"
					class="pointer-events-none fill-white text-xs font-bold"
				>
					✓
				</text>
			{/if}

			<!-- Assignment indicators -->
			{#if hasAssignments}
				<g transform="translate({node.position.x + 5}, {node.position.y + 5})">
					{#if node.tools && node.tools.length > 0}
						<circle cx="0" cy="0" r="8" fill="#f59e0b" />
						<text
							x="0"
							y="0"
							text-anchor="middle"
							dominant-baseline="middle"
							class="pointer-events-none fill-white text-[10px] font-bold"
						>
							T
						</text>
					{/if}
					{#if node.agents && node.agents.length > 0}
						<circle cx="18" cy="0" r="8" fill="#3b82f6" />
						<text
							x="18"
							y="0"
							text-anchor="middle"
							dominant-baseline="middle"
							class="pointer-events-none fill-white text-[10px] font-bold"
						>
							A
						</text>
					{/if}
					{#if node.tasks && node.tasks.length > 0}
						<circle cx="36" cy="0" r="8" fill="#8b5cf6" />
						<text
							x="36"
							y="0"
							text-anchor="middle"
							dominant-baseline="middle"
							class="pointer-events-none fill-white text-[10px] font-bold"
						>
							K
						</text>
					{/if}
				</g>
			{/if}
		</g>
	{/each}
</svg>
