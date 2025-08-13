<script lang="ts">
	import Messages from './Messages.svelte';
	import MessageInput from './MessageInput.svelte';
	import type {
		ChatMessage,
		Elicitation as ElicitationType,
		ElicitationResult,
		Prompt as PromptType
	} from '$lib/types';
	import Elicitation from '$lib/components/Elicitation.svelte';
	import Prompt from '$lib/components/Prompt.svelte';

	interface Props {
		messages: ChatMessage[];
		prompts: PromptType[];
		elicitations?: ElicitationType[];
		onElicitationResult?: (elicitation: ElicitationType, result: ElicitationResult) => void;
		onSendMessage?: (message: string) => void;
		onFileUpload?: (file: File, url: string) => void;
		isLoading?: boolean;
	}

	let {
		messages,
		prompts,
		onSendMessage,
		onFileUpload,
		elicitations,
		onElicitationResult,
		isLoading = false
	}: Props = $props();
</script>

<div class="flex size-full flex-col">
	<!-- Messages area - scrollable -->
	<div class="flex-1 overflow-y-auto">
		<div class="mx-auto max-w-4xl p-4">
			<!-- Prompts section - show when prompts available and no messages -->
			{#if prompts && prompts.length > 0}
				<div class="mb-6">
					<h2 class="mb-4 text-xl font-semibold">Available Prompts</h2>
					<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
						{#each prompts as prompt (prompt.name)}
							<Prompt {prompt} onSend={onSendMessage} />
						{/each}
					</div>
				</div>
			{/if}

			<Messages {messages} onSend={onSendMessage} />
		</div>
	</div>

	<!-- Message input - fixed at bottom -->
	<div class="flex-shrink-0 border-t bg-base-100">
		<div class="mx-auto max-w-4xl">
			<MessageInput onSend={onSendMessage} {onFileUpload} disabled={isLoading} />
		</div>
	</div>

	{#if elicitations && elicitations.length > 0}
		{#key elicitations[0].id}
			<Elicitation
				elicitation={elicitations[0]}
				open
				onresult={(result) => {
					onElicitationResult?.(elicitations[0], result);
				}}
			/>
		{/key}
	{/if}
</div>
