<script lang="ts">
	import { onDestroy } from 'svelte';
	import { Monitor, X, Maximize2, Minimize2 } from '@lucide/svelte';
	import RFB from '@novnc/novnc/lib/rfb.js';

	interface Props {
		vncUrl?: string;
		visible?: boolean;
	}

	function defaultVNCUrl() {
		if (typeof window === 'undefined') {
			return 'ws://localhost/browser';
		}

		const url = new URL('/browser', window.location.origin);
		url.protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		return url.toString();
	}

	let { vncUrl = defaultVNCUrl(), visible = $bindable(true) }: Props = $props();

	let container = $state<HTMLDivElement | undefined>(undefined);
	let rfb = $state<RFB | null>(null);
	let connected = $state(false);
	let connecting = $state(false);
	let error = $state<string | null>(null);
	let isFullscreen = $state(false);
	let activeVNCUrl = $state<string | null>(null);
	let resizeTimer: ReturnType<typeof setTimeout> | null = null;
	let lastRequestedSize = $state<string | null>(null);
	let viewerActive = $state(false);

	async function connect() {
		if (!container || rfb || connecting) return;

		connecting = true;
		error = null;

		try {
			const nextRfb = new RFB(container, vncUrl);
			rfb = nextRfb;
			activeVNCUrl = vncUrl;

			nextRfb.addEventListener('connect', () => {
				if (rfb !== nextRfb) return;
				connected = true;
				connecting = false;
				error = null;
			});

			nextRfb.addEventListener('disconnect', () => {
				if (rfb === nextRfb) {
					rfb = null;
					activeVNCUrl = null;
				}
				connected = false;
				connecting = false;
			});

			nextRfb.addEventListener('credentialsrequired', () => {
				if (rfb !== nextRfb) return;
				connecting = false;
				error = 'Password required (but none configured)';
			});

			nextRfb.addEventListener('securityfailure', (e) => {
				if (rfb !== nextRfb) return;
				connecting = false;
				error = `Security failure: ${e.detail.status}`;
			});

			nextRfb.scaleViewport = true;
			nextRfb.resizeSession = true;
			nextRfb.dragViewport = false;
			nextRfb.clipViewport = false;
		} catch (err) {
			rfb = null;
			activeVNCUrl = null;
			connecting = false;
			error = err instanceof Error ? err.message : 'Connection failed';
		}
	}

	function disconnect() {
		if (rfb) {
			rfb.disconnect();
		}
		rfb = null;
		activeVNCUrl = null;
		connected = false;
		connecting = false;
	}

	async function syncRemoteClipboard(text: string) {
		if (!text || !rfb) return;
		rfb.focus();
		rfb.clipboardPasteFrom(text);
	}

	function sendRemotePasteShortcut() {
		if (!rfb) return;
		rfb.focus();
		rfb.sendKey(0xffe3, 'ControlLeft', true);
		rfb.sendKey(0x0076, 'KeyV', true);
		rfb.sendKey(0x0076, 'KeyV', false);
		rfb.sendKey(0xffe3, 'ControlLeft', false);
	}

	async function handleLocalPaste() {
		if (!viewerActive || !rfb || typeof navigator === 'undefined' || !navigator.clipboard) {
			return;
		}

		try {
			const text = await navigator.clipboard.readText();
			await syncRemoteClipboard(text);
			sendRemotePasteShortcut();
		} catch {
			// Clipboard reads are permission-gated in the browser; ignore failures.
		}
	}

	function resizeUrl(currentVNCUrl: string) {
		const url = new URL(currentVNCUrl);
		url.protocol = url.protocol === 'wss:' ? 'https:' : 'http:';
		url.pathname = `${url.pathname.replace(/\/$/, '')}/resize`;
		url.search = '';
		url.hash = '';
		return url.toString();
	}

	function queueResize(width: number, height: number) {
		const targetWidth = Math.max(640, Math.ceil(width * window.devicePixelRatio));
		const targetHeight = Math.max(480, Math.ceil(height * window.devicePixelRatio));
		const sizeKey = `${targetWidth}x${targetHeight}`;
		if (sizeKey === lastRequestedSize) return;

		if (resizeTimer) {
			clearTimeout(resizeTimer);
		}

		resizeTimer = setTimeout(async () => {
			try {
				const response = await fetch(resizeUrl(vncUrl), {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify({
						width: targetWidth,
						height: targetHeight,
					}),
				});

				if (!response.ok) {
					throw new Error(`resize failed: ${response.status}`);
				}

				lastRequestedSize = sizeKey;
			} catch {
				// Resize is best-effort and should not disrupt the viewer.
			}
		}, 150);
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

	onDestroy(() => {
		if (resizeTimer) {
			clearTimeout(resizeTimer);
		}
		disconnect();
	});

	$effect(() => {
		if (!visible || !container || typeof ResizeObserver === 'undefined') {
			return;
		}

		const observer = new ResizeObserver((entries) => {
			const entry = entries[0];
			if (!entry) return;
			queueResize(entry.contentRect.width, entry.contentRect.height);
		});

		observer.observe(container);
		const rect = container.getBoundingClientRect();
		queueResize(rect.width, rect.height);

		return () => {
			observer.disconnect();
		};
	});

	$effect(() => {
		if (!visible || !container) {
			viewerActive = false;
			return;
		}

		const handlePointerDown = (event: PointerEvent) => {
			if (!container) return;
			viewerActive = container.contains(event.target as Node);
		};

		const handlePaste = async (event: ClipboardEvent) => {
			if (!viewerActive || !rfb) return;

			const text = event.clipboardData?.getData('text/plain') ?? '';
			if (!text) return;

			event.preventDefault();
			await syncRemoteClipboard(text);
			sendRemotePasteShortcut();
		};

		const handleKeyDown = async (event: KeyboardEvent) => {
			const modifierPressed = event.metaKey || event.ctrlKey;
			if (!viewerActive || !modifierPressed || event.key.toLowerCase() !== 'v') {
				return;
			}

			event.preventDefault();
			await handleLocalPaste();
		};

		document.addEventListener('pointerdown', handlePointerDown, true);
		window.addEventListener('paste', handlePaste);
		window.addEventListener('keydown', handleKeyDown, true);

		return () => {
			document.removeEventListener('pointerdown', handlePointerDown, true);
			window.removeEventListener('paste', handlePaste);
			window.removeEventListener('keydown', handleKeyDown, true);
		};
	});

	$effect(() => {
		const desiredVisible = visible;
		const desiredUrl = vncUrl;
		const hasContainer = !!container;

		if (!desiredVisible || !hasContainer) {
			disconnect();
			return;
		}

		if (rfb && activeVNCUrl !== desiredUrl) {
			disconnect();
			return;
		}

		if (!rfb && !connecting) {
			void connect();
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
				{:else if connecting}
					<span class="status-badge connecting">Connecting...</span>
				{:else}
					<span class="status-badge connecting">Idle</span>
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
		overflow: hidden;
		background: #1a1a1a;
		border-radius: 0.5rem;
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
