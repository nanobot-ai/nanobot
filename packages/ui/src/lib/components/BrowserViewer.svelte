<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Monitor, X, Maximize2, Minimize2 } from '@lucide/svelte';

	interface Props {
		vncUrl?: string;
		visible?: boolean;
	}

	let { vncUrl = 'ws://localhost:6080', visible = $bindable(true) }: Props = $props();

	// State
	let container: HTMLDivElement;
	let rfb: any = null;
	let connected = $state(false);
	let error = $state<string | null>(null);
	let isFullscreen = $state(false);

	// Import noVNC library dynamically
	async function loadNoVNC() {
		try {
			// Load noVNC from CDN
			const RFB = (await import('https://cdn.jsdelivr.net/npm/@novnc/novnc@1.4.0/core/rfb.js')).default;
			return RFB;
		} catch (err) {
			console.error('Failed to load noVNC:', err);
			error = 'Failed to load VNC library';
			return null;
		}
	}

	async function connect() {
		if (!container) return;

		const RFB = await loadNoVNC();
		if (!RFB) return;

		try {
			// Create RFB connection
			rfb = new RFB(container, vncUrl, {
				credentials: { password: '' },
			});

			// Event handlers
			rfb.addEventListener('connect', () => {
				connected = true;
				error = null;
				console.log('VNC connected');
			});

			rfb.addEventListener('disconnect', () => {
				connected = false;
				console.log('VNC disconnected');
			});

			rfb.addEventListener('credentialsrequired', () => {
				error = 'Password required (but none configured)';
			});

			rfb.addEventListener('securityfailure', (e: any) => {
				error = `Security failure: ${e.detail.status}`;
			});

			// Set scale mode
			rfb.scaleViewport = true;
			rfb.resizeSession = false;
		} catch (err) {
			console.error('VNC connection error:', err);
			error = err instanceof Error ? err.message : 'Connection failed';
		}
	}

	function disconnect() {
		if (rfb) {
			rfb.disconnect();
			rfb = null;
		}
		connected = false;
	}

	function toggleFullscreen() {
		if (!container) return;

		if (!document.fullscreenElement) {
			container.requestFullscreen();
			isFullscreen = true;
		} else {
			document.exitFullscreen();
			isFullscreen = false;
		}
	}

	onMount(() => {
		if (visible) {
			connect();
		}
	});

	onDestroy(() => {
		disconnect();
	});

	// Reconnect when visibility changes
	$effect(() => {
		if (visible && !connected) {
			connect();
		} else if (!visible && connected) {
			disconnect();
		}
	});
</script>

{#if visible}
	<div class="browser-viewer">
		<div class="viewer-header">
			<div class="header-title">
				<Monitor size={16} />
				<span>Browser View</span>
				{#if connected}
					<span class="status-badge connected">Connected</span>
				{:else if error}
					<span class="status-badge error">Error</span>
				{:else}
					<span class="status-badge connecting">Connecting...</span>
				{/if}
			</div>

			<div class="header-actions">
				<button
					class="btn btn-sm btn-ghost"
					onclick={toggleFullscreen}
					title="Toggle fullscreen"
				>
					{#if isFullscreen}
						<Minimize2 size={16} />
					{:else}
						<Maximize2 size={16} />
					{/if}
				</button>

				<button
					class="btn btn-sm btn-ghost"
					onclick={() => (visible = false)}
					title="Close browser view"
				>
					<X size={16} />
				</button>
			</div>
		</div>

		{#if error}
			<div class="error-message">
				<p>{error}</p>
				<button class="btn btn-sm btn-primary" onclick={connect}>Retry</button>
			</div>
		{/if}

		<div class="viewer-container" bind:this={container}></div>
	</div>
{/if}

<style>
	.browser-viewer {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: #1a1a1a;
		border-radius: 0.5rem;
		overflow: hidden;
	}

	.viewer-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 1rem;
		background: #2a2a2a;
		border-bottom: 1px solid #3a3a3a;
	}

	.header-title {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 600;
		color: #e0e0e0;
	}

	.header-actions {
		display: flex;
		gap: 0.25rem;
	}

	.status-badge {
		padding: 0.125rem 0.5rem;
		font-size: 0.75rem;
		border-radius: 0.25rem;
		font-weight: 500;
	}

	.status-badge.connected {
		background: #10b981;
		color: white;
	}

	.status-badge.connecting {
		background: #f59e0b;
		color: white;
	}

	.status-badge.error {
		background: #ef4444;
		color: white;
	}

	.viewer-container {
		flex: 1;
		position: relative;
		overflow: hidden;
		background: #000;
	}

	.viewer-container :global(canvas) {
		width: 100% !important;
		height: 100% !important;
		object-fit: contain;
	}

	.error-message {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		padding: 2rem;
		color: #ef4444;
	}

	.error-message p {
		margin: 0;
	}
</style>
