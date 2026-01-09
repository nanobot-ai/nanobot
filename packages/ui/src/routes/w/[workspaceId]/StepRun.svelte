<script lang="ts">
	import Messages from "$lib/components/Messages.svelte";
	import type { ChatMessage } from "$lib/types";

    interface Props {
        messages: ChatMessage[];
        chatLoading: boolean;
        pending: boolean;
    }

    let { messages, pending, chatLoading }: Props = $props();

    let container: HTMLDivElement;
    let autoScroll = $state(true);

    // Check if user is at the bottom (with small threshold for floating point)
    function isAtBottom(el: HTMLDivElement): boolean {
        const threshold = 10;
        return el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
    }

    function handleScroll() {
        if (!container) return;
        autoScroll = isAtBottom(container);
    }

    // Auto-scroll when messages change
    $effect(() => {
        // Track messages length to trigger effect on change
        void messages.length;
        if (autoScroll && container) {
            container.scrollTop = container.scrollHeight;
        }
    });
</script>

<div
    bind:this={container}
    onscroll={handleScroll}
    class="mt-4 w-full step-agent max-h-52 shadow-inner overflow-y-auto bg-base-200 dark:bg-base-100 rounded-field p-4"
>
    {#if pending}
        <span class="skeleton skeleton-text w-full h-4 text-sm">
            Waiting for prior step to complete...
        </span>
    {:else if chatLoading && messages.length === 0}
        <span class="skeleton skeleton-text w-full h-4 text-sm">
            The step is processing...
        </span>
    {/if}
    <Messages inline messages={messages} />
</div>