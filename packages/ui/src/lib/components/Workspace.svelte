<script lang="ts">
	import type { Attachment, ChatMessage, ChatResult, ChatMessageItemToolCall } from '$lib/types';
	import { getMcpAppsContext } from '$lib/context/mcpApps.svelte';
	import MessageItemApp from '$lib/components/MessageItemApp.svelte';

	interface Props {
		messages: ChatMessage[];
		onSendMessage?: (message: string, attachments?: Attachment[]) => Promise<ChatResult | void>;
	}

	let { messages }: Props = $props();

	const { tools } = getMcpAppsContext();

	// Find the most recent tool call that has workspace-flagged output and a matching app resourceUri
	let sidecar = $derived.by(() => {
		for (const message of messages.toReversed()) {
			for (const item of (message.items ?? []).toReversed()) {
				if (item.type === 'tool' && item.output && item.output?.content) {
					const hasWorkspaceOutput = item.output.content.some(
						(output) =>
							output.type === 'resource' &&
							'resource' in output &&
							output.resource?._meta?.['ai.nanobot.meta/workspace']
					);
					if (hasWorkspaceOutput) {
						const resourceUri = tools?.find(
							(t) => t.name === item.name
						)?._meta?.ui?.resourceUri;
						if (resourceUri) {
							return {
								item: item satisfies ChatMessageItemToolCall,
								resourceUri
							};
						}
					}
				}
			}
		}
		return null;
	});

	let key = $derived(sidecar ? `${sidecar.item.callID}-${sidecar.resourceUri}` : '');
</script>

{#if key && sidecar}
	{#key key}
		<div
			class="workspace peer m-3 h-[60vh] border-2 border-base-100/30 md:m-0 md:h-dvh md:max-h-dvh md:w-3/4"
		>
			<MessageItemApp item={sidecar.item} resourceUri={sidecar.resourceUri} />
		</div>
	{/key}
{/if}
