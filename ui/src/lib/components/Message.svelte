<script lang="ts">
	import { Bot } from '@lucide/svelte';
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
	<div class="flex w-full justify-end">
		<div class="max-w-md">
			<div class="flex flex-col items-end">
				<div class="mb-1 text-sm font-medium opacity-70">
					You
					<time class="ml-2 text-xs opacity-50">{displayTime.toLocaleTimeString()}</time>
				</div>
				<div class="rounded-lg bg-base-200 p-3">
					{#if message.items && message.items.length > 0}
						{#each message.items as item (item.id)}
							<MessageItem {item} role={message.role} />
						{/each}
					{:else}
						<!-- Fallback for messages without items -->
						<p>No content</p>
					{/if}
				</div>
			</div>
		</div>
	</div>
{:else}
	<div class="flex w-full items-start gap-3">
		<!-- Assistant avatar on left -->
		<div class="avatar flex-shrink-0">
			<div class="w-10 rounded-full">
				<div
					class="flex h-10 w-10 items-center justify-center rounded-full bg-secondary text-secondary-content"
				>
					<Bot class="h-6 w-6" />
				</div>
			</div>
		</div>

		<!-- Assistant message content -->
		<div class="flex min-w-0 flex-1 flex-col items-start">
			<div class="mb-1 text-sm font-medium opacity-70">
				Assistant
				<time class="ml-2 text-xs opacity-50">{displayTime.toLocaleTimeString()}</time>
			</div>

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
