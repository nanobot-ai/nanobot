<script lang="ts">
	import type { Attachment, ChatResult, ChatMessageItem } from '$lib/types';
	import MessageItemText from './MessageItemText.svelte';
	import MessageItemImage from './MessageItemImage.svelte';
	import MessageItemAudio from './MessageItemAudio.svelte';
	import MessageItemResourceLink from './MessageItemResourceLink.svelte';
	import MessageItemResource from './MessageItemResource.svelte';
	import MessageItemReasoning from './MessageItemReasoning.svelte';
	import MessageItemTool from './MessageItemTool.svelte';
	import MessageItemWorkflowFile from './MessageItemWorkflowFile.svelte';

	interface Props {
		item: ChatMessageItem;
		role: 'user' | 'assistant';
		onSend?: (message: string, attachments?: Attachment[]) => Promise<ChatResult | void>;
		onFileOpen?: (filename: string) => void;
	}

	let { item, role, onSend, onFileOpen }: Props = $props();
</script>

{#if item.type === 'text'}
	<MessageItemText {item} {role} />
{:else if item.type === 'image'}
	<MessageItemImage {item} />
{:else if item.type === 'audio'}
	<MessageItemAudio {item} />
{:else if item.type === 'resource_link'}
	<MessageItemResourceLink {item} />
{:else if item.type === 'resource'}
	<MessageItemResource {item} />
{:else if item.type === 'reasoning'}
	<MessageItemReasoning {item} />
{:else if item.type === 'tool'}
	{#if item.name === 'write'}
		<MessageItemWorkflowFile {item} {onFileOpen} />
	{:else}
		<MessageItemTool {item} {onSend} />
	{/if}
{/if}
