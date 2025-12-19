<script lang="ts">
	import Messages from './Messages.svelte';
	import MessageInput from './MessageInput.svelte';
	import type {
		Attachment,
		ChatMessage,
		ChatResult,
		Elicitation as ElicitationType,
		ElicitationResult,
		Prompt as PromptType,
		Agent,
		UploadingFile,
		UploadedFile,
		Resource
	} from '$lib/types';
	import Elicitation from '$lib/components/Elicitation.svelte';
	import Prompt from '$lib/components/Prompt.svelte';
	import { ChevronDown } from '@lucide/svelte';

	interface Props {
		messages: ChatMessage[];
		prompts: PromptType[];
		resources: Resource[];
		elicitations?: ElicitationType[];
		onElicitationResult?: (elicitation: ElicitationType, result: ElicitationResult) => void;
		onSendMessage?: (message: string, attachments?: Attachment[]) => Promise<ChatResult | void>;
		onFileUpload?: (file: File, opts?: { controller?: AbortController }) => Promise<Attachment>;
		cancelUpload?: (fileId: string) => void;
		uploadingFiles?: UploadingFile[];
		uploadedFiles?: UploadedFile[];
		isLoading?: boolean;
		agent?: Agent;
	}

	let {
		// Do not use _chat variable anywhere except these assignments
		messages,
		prompts,
		resources,
		onSendMessage,
		onFileUpload,
		cancelUpload,
		uploadingFiles,
		uploadedFiles,
		elicitations,
		onElicitationResult,
		agent,
		isLoading,
	}: Props = $props();

	let messagesContainer: HTMLElement;
	let showScrollButton = $state(false);
	let previousLastMessageId = $state<string | null>(null);
	let hasMessages = $derived(messages && messages.length > 0);
	let selectedPrompt = $state<string | undefined>();

	// Watch for changes to the last message ID and scroll to bottom
	$effect(() => {
		if (!messagesContainer) return;

		// Make this reactive to changes in messages
		void messages.length;

		const lastDiv = messagesContainer.querySelector('#message-groups > :last-child');
		const currentLastMessageId = lastDiv?.getAttribute('data-message-id');

		if (currentLastMessageId && currentLastMessageId !== previousLastMessageId) {
			// Wait for DOM update, then scroll to bottom
			setTimeout(() => {
				scrollToBottom();
			}, 10);
			previousLastMessageId = currentLastMessageId;
		}
	});

	function handleScroll() {
		if (!messagesContainer) return;

		const { scrollTop, scrollHeight, clientHeight } = messagesContainer;
		const isNearBottom = scrollTop + clientHeight >= scrollHeight - 10; // 10px threshold
		showScrollButton = !isNearBottom;
	}

	function scrollToBottom() {
		if (messagesContainer) {
			messagesContainer.scrollTo({
				top: messagesContainer.scrollHeight,
				behavior: 'smooth'
			});
		}
	}
</script>

<div class="flex h-dvh w-full flex-col md:relative peer-[.workspace]:md:w-1/4">
	<!-- Messages area - full height scrollable with bottom padding for floating input -->
	<div class="w-full overflow-y-auto" bind:this={messagesContainer} onscroll={handleScroll}>
		<div class="mx-auto max-w-4xl">
			<!-- Prompts section - show when prompts available and no messages -->
			{#if prompts && prompts.length > 0}
				<div class="mb-6">
					<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
						{#each prompts as prompt (prompt.name)}
							{#if selectedPrompt === prompt.name}
								<Prompt
									{prompt}
									onSend={async (m) => {
										selectedPrompt = undefined;
										if (onSendMessage) {
											return await onSendMessage(m);
										}
									}}
									onCancel={() => (selectedPrompt = undefined)}
									open
								/>
							{/if}
						{/each}
					</div>
				</div>
			{/if}

			<Messages {messages} onSend={onSendMessage} {isLoading} {agent} />
		</div>
	</div>

	<!-- Message input - centered when no messages, bottom when messages exist -->
	<div
		class="absolute right-0 bottom-0 left-0 flex flex-col transition-all duration-500 ease-in-out {hasMessages
			? 'bg-base-100/80 backdrop-blur-sm'
			: 'md:-translate-y-1/2 [@media(min-height:900px)]:md:top-1/2 [@media(min-height:900px)]:md:bottom-auto'}"
	>
		<!-- Scroll to bottom button -->
		{#if showScrollButton && hasMessages}
			<button
				class="btn mx-auto btn-circle border-base-300 bg-base-100 shadow-lg btn-md active:translate-y-0.5"
				onclick={scrollToBottom}
				aria-label="Scroll to bottom"
			>
				<ChevronDown class="size-5" />
			</button>
		{/if}
		<div class="mx-auto w-full max-w-4xl">
			<MessageInput
				placeholder={`Type your message...${prompts && prompts.length > 0 ? ' or / for prompts' : ''}`}
				onSend={onSendMessage}
				{resources}
				{messages}
				onPrompt={(p) => (selectedPrompt = p)}
				{onFileUpload}
				disabled={isLoading}
				{prompts}
				{cancelUpload}
				{uploadingFiles}
				{uploadedFiles}
			/>
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
