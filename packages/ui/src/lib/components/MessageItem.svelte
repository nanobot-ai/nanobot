<script lang="ts">
import type {
	Attachment,
	ChatMessageItem,
	ChatResult,
	ResourceContents,
} from "$lib/types";

interface Props {
	item: ChatMessageItem;
	role: "user" | "assistant";
	onSend?: (
		message: string,
		attachments?: Attachment[],
	) => Promise<ChatResult | undefined>;
	onReadResource?: (uri: string) => Promise<{ contents: ResourceContents[] }>;
}

const { item, role, onSend, onReadResource }: Props = $props();
</script>

{#if item.type === 'text'}
	<MessageItemText {item} {role} />
{:else if item.type === 'image'}
	<MessageItemImage {item} />
{:else if item.type === 'audio'}
	<MessageItemAudio {item} />
{:else if item.type === 'resource_link'}
	<MessageItemResourceLink {item} {onReadResource} />
{:else if item.type === 'resource'}
	<MessageItemResource {item} />
{:else if item.type === 'reasoning'}
	<MessageItemReasoning {item} />
{:else if item.type === 'tool'}
	<MessageItemTool {item} {onSend} />
{/if}
