<script lang="ts">
	import { AppFrame, AppBridge } from '@mcp-ui/client';
	import type { CallToolResult } from '@modelcontextprotocol/sdk/types.js';
	import React from 'react';
	import ReactDOM from 'react-dom/client';
	import { onMount, onDestroy } from 'svelte';
	import { getMcpAppsContext } from '$lib/context/mcpApps.svelte';
	import type { ChatMessageItemToolCall } from '$lib/types';

	interface Props {
		item: ChatMessageItemToolCall;
		resourceUri: string;
	}

	let { item, resourceUri }: Props = $props();
	let container: HTMLDivElement;
	let reactRoot: ReactDOM.Root | undefined;
	let displayMode = $state<'inline' | 'fullscreen' | 'pip'>('inline');
	let prefersBorder = $state<boolean | undefined>(undefined);
	let appErrors = $state<string[]>([]);
	let bridgeRef: AppBridge | undefined;
	let cspHandler: ((e: SecurityPolicyViolationEvent) => void) | undefined;

	const mcpApps = getMcpAppsContext();
	const sandboxUrl = new URL('/sandbox_proxy.html', window.location.origin);

	// Debounced error sending to LLM
	let errorDebounceTimer: ReturnType<typeof setTimeout> | undefined;
	let pendingErrors: string[] = [];

	function flushErrorsToLLM() {
		if (pendingErrors.length === 0 || !mcpApps.sendMessage) return;
		const errors = pendingErrors.slice();
		pendingErrors = [];
		const msg = `[App Error] The app reported errors:\n${errors.map((e) => `- ${e}`).join('\n')}`;
		mcpApps.sendMessage(msg);
	}

	function queueErrorForLLM(error: string) {
		pendingErrors.push(error);
		appErrors = [...appErrors, error];
		clearTimeout(errorDebounceTimer);
		errorDebounceTimer = setTimeout(flushErrorsToLLM, 500);
	}

	function dismissErrors() {
		appErrors = [];
	}

	function buildHostContext(el: HTMLElement) {
		const root = document.documentElement;
		const cs = getComputedStyle(root);
		const dataTheme = root.getAttribute('data-theme') ?? '';
		const isDark = dataTheme.includes('dark') || cs.colorScheme === 'dark';

		const v = (prop: string) => cs.getPropertyValue(prop).trim() || undefined;

		const theme = isDark ? ('dark' as const) : ('light' as const);

		return {
			theme,
			displayMode,
			availableDisplayModes: ['inline', 'fullscreen', 'pip'],
			platform: 'web' as const,
			locale: navigator.language,
			timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
			containerDimensions: { maxWidth: el.clientWidth },
			styles: {
				variables: {
					'--color-background-primary': v('--color-base-100'),
					'--color-background-secondary': v('--color-base-200'),
					'--color-background-tertiary': v('--color-base-300'),
					'--color-text-primary': v('--color-base-content'),
					'--color-background-info': v('--color-info'),
					'--color-background-danger': v('--color-error'),
					'--color-background-success': v('--color-success'),
					'--color-background-warning': v('--color-warning'),
					'--color-text-info': v('--color-info-content'),
					'--color-text-danger': v('--color-error-content'),
					'--color-text-success': v('--color-success-content'),
					'--color-text-warning': v('--color-warning-content'),
					'--color-border-primary': v('--color-base-300'),
					'--font-sans': cs.fontFamily || undefined,
					'--border-radius-md': v('--radius-box'),
					'--border-radius-sm': v('--radius-field')
				} as Record<string, string | undefined>
			}
		};
	}

	function applyDisplayMode(mode: string) {
		if (mode !== 'inline' && container) {
			container.style.height = ''; // clear inline style so CSS classes apply
		}
		requestAnimationFrame(() => {
			if (bridgeRef && container) {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				bridgeRef.setHostContext({
					...buildHostContext(container),
					displayMode: mode as 'inline' | 'fullscreen' | 'pip'
				} as any);
			}
		});
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && displayMode !== 'inline') {
			displayMode = 'inline';
			applyDisplayMode('inline');
		}
	}

	// Watch for theme changes and push updated context to the app
	let themeObserver: MutationObserver | undefined;

	onMount(async () => {
		if (!container) return;

		document.addEventListener('keydown', handleKeydown);

		// Observe theme changes on <html> element
		themeObserver = new MutationObserver(() => {
			if (bridgeRef && container) {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				bridgeRef.setHostContext(buildHostContext(container) as any);
			}
		});
		themeObserver.observe(document.documentElement, {
			attributes: true,
			attributeFilter: ['data-theme', 'class', 'style']
		});

		const client = await mcpApps.ensureClient();

		// Pre-fetch resource to extract _meta.ui (CSP, prefersBorder)
		let html: string | undefined;
		let resourceCsp: { connectDomains?: string[]; resourceDomains?: string[] } | undefined;

		if (client && resourceUri) {
			try {
				const result = await client.readResource({ uri: resourceUri });
				const content = result.contents?.[0];
				if (content && 'text' in content && content.mimeType === 'text/html;profile=mcp-app') {
					html = content.text as string;
					// Extract _meta.ui from resource response
					// eslint-disable-next-line @typescript-eslint/no-explicit-any
					const uiMeta = (content as any)._meta?.ui;
					if (uiMeta) {
						resourceCsp = uiMeta.csp;
						prefersBorder = uiMeta.prefersBorder;
					}
				}
			} catch (e) {
				console.error('[MCP App] Failed to pre-fetch resource:', e);
			}
		}

		// Create AppBridge with full control over callbacks
		const capabilities = client?.getServerCapabilities();
		const bridge = new AppBridge(
			client ?? null,
			{ name: 'Nanobot', version: '1.0.0' },
			{
				openLinks: {},
				serverTools: capabilities?.tools,
				serverResources: capabilities?.resources,
				logging: {}
			},
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			{ hostContext: buildHostContext(container) as any }
		);
		bridgeRef = bridge;

		bridge.onopenlink = async ({ url }) => {
			window.open(url, '_blank');
			return {};
		};

		bridge.onmessage = async ({ role, content }) => {
			const text = content
				.filter(
					(c): c is { type: 'text'; text: string } => c.type === 'text' && typeof c.text === 'string'
				)
				.map((c) => c.text)
				.join('\n');
			if (text && mcpApps.sendMessage) {
				await mcpApps.sendMessage(text);
			}
			return {};
		};

		bridge.onloggingmessage = ({ level, logger, data }) => {
			const prefix = logger ? `[${logger}]` : '[MCP App]';
			if (['error', 'critical', 'alert', 'emergency'].includes(level)) {
				console.error(prefix, data);
				const errorText = typeof data === 'string' ? data : JSON.stringify(data);
				queueErrorForLLM(`${prefix} ${errorText}`);
			} else {
				console.log(prefix, level, data);
			}
		};

		bridge.onrequestdisplaymode = async ({ mode }) => {
			const supported = ['inline', 'fullscreen', 'pip'];
			const granted = supported.includes(mode) ? mode : 'inline';
			displayMode = granted as 'inline' | 'fullscreen' | 'pip';
			applyDisplayMode(granted);
			return { mode: granted as 'inline' | 'fullscreen' | 'pip' };
		};

		reactRoot = ReactDOM.createRoot(container);

		let toolResult: CallToolResult | undefined;
		if (item.output) {
			toolResult = {
				content: (item.output.content ?? []).map((c) => {
					if (c.type === 'text' && 'text' in c) return { type: 'text' as const, text: c.text };
					if (c.type === 'image' && 'data' in c)
						return { type: 'image' as const, data: c.data, mimeType: c.mimeType };
					return { type: 'text' as const, text: JSON.stringify(c) };
				}),
				// Deep-clone to strip Svelte 5 reactivity Proxies — postMessage can't clone Proxy objects.
				structuredContent: item.output.structuredContent
					? JSON.parse(JSON.stringify(item.output.structuredContent))
					: undefined,
				isError: item.output.isError
			};
		}

		let toolInput: Record<string, unknown> | undefined;
		if (item.arguments) {
			try {
				toolInput = JSON.parse(item.arguments);
			} catch {
				/* ignore */
			}
		}

		reactRoot.render(
			React.createElement(AppFrame, {
				html: html ?? '',
				sandbox: {
					url: sandboxUrl,
					...(resourceCsp ? { csp: resourceCsp } : {})
				},
				appBridge: bridge,
				toolInput,
				toolResult,
				onSizeChanged: ({ height }: { width?: number; height?: number }) => {
					if (container && height != null && displayMode === 'inline') {
						container.style.height = `${height}px`;
					}
				},
				onError: (error: Error) => {
					console.error('[MCP App Error]', error);
					queueErrorForLLM(error.message || String(error));
				}
			})
		);

		// Listen for CSP violations and report them
		cspHandler = (e: SecurityPolicyViolationEvent) => {
			const blocked = e.blockedURI || 'unknown';
			console.warn('[MCP App CSP Violation]', `Blocked: ${blocked}, Directive: ${e.violatedDirective}`);
			queueErrorForLLM(
				`[CSP Violation] App tried to load blocked resource from: ${blocked} (directive: ${e.violatedDirective}). The server may need to declare this domain in _meta.ui.csp.resourceDomains.`
			);
		};
		document.addEventListener('securitypolicyviolation', cspHandler);
	});

	onDestroy(() => {
		document.removeEventListener('keydown', handleKeydown);
		if (cspHandler) {
			document.removeEventListener('securitypolicyviolation', cspHandler);
		}
		themeObserver?.disconnect();
		clearTimeout(errorDebounceTimer);
		bridgeRef?.close();
		reactRoot?.unmount();
	});
</script>

{#if appErrors.length > 0}
	<div class="alert alert-error mb-2 text-sm">
		<div class="flex-1">
			<p class="font-semibold">App errors:</p>
			{#each appErrors as error}
				<p class="text-xs opacity-80">{error}</p>
			{/each}
		</div>
		<button class="btn btn-ghost btn-xs" onclick={dismissErrors}>Dismiss</button>
	</div>
{/if}

<div
	bind:this={container}
	class="w-full overflow-hidden transition-all"
	class:rounded={prefersBorder === true}
	class:border={prefersBorder === true}
	class:border-base-300={prefersBorder === true}
	class:fixed={displayMode !== 'inline'}
	class:inset-0={displayMode === 'fullscreen'}
	class:z-50={displayMode !== 'inline'}
	class:bg-base-100={displayMode === 'fullscreen'}
	class:bottom-4={displayMode === 'pip'}
	class:right-4={displayMode === 'pip'}
	class:w-96={displayMode === 'pip'}
	class:h-72={displayMode === 'pip'}
	class:shadow-2xl={displayMode === 'pip'}
	class:rounded-lg={displayMode === 'pip'}
	class:resize={displayMode === 'pip'}
></div>
