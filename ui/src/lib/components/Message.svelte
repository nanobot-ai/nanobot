<script lang="ts">
	import MessageItem from './MessageItem.svelte';
	import type { ChatMessage } from '$lib/types';

	interface Props {
		message: ChatMessage;
		timestamp?: Date;
		onSend?: (message: string) => void;
	}

	let { message, timestamp, onSend }: Props = $props();

	const displayTime = $derived(
		timestamp || (message.created ? new Date(message.created) : new Date())
	);
</script>

{#if message.role === 'user'}
	<div class="group flex w-full justify-end">
		<div class="max-w-md">
			<div class="flex flex-col items-end">
				<div class="rounded-box bg-base-200 p-2">
					{#if message.items && message.items.length > 0}
						{#each message.items as item (item.id)}
							<MessageItem {item} role={message.role} />
						{/each}
					{:else}
						<!-- Fallback for messages without items -->
						<p>No content</p>
					{/if}
				</div>
				<div
					class="transition-duration-500 mb-1 text-sm font-medium opacity-0 transition-opacity group-hover:opacity-100"
				>
					<time class="ml-2 text-xs opacity-50">{displayTime.toLocaleTimeString()}</time>
				</div>
			</div>
		</div>
	</div>
{:else}
	<div class="flex w-full items-start gap-3">
		<!-- Assistant message content -->
		<div class="flex min-w-0 flex-1 flex-col items-start">
			<!-- Render all message items -->
			{#if message.items && message.items.length > 0}
				{#each message.items as item (item.id)}
					<MessageItem {item} role={message.role} {onSend} />
				{/each}
			{:else}
				<!-- Fallback for messages without items -->
				<div class="prose w-full max-w-full rounded-lg bg-base-200 p-3 prose-invert">
					<p>No content</p>
				</div>
			{/if}
		</div>
	</div>
{/if}
