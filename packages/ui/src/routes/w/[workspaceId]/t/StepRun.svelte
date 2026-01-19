<script lang="ts">
	import Messages from "$lib/components/Messages.svelte";
	import type { ChatMessage } from "$lib/types";
	import { TriangleAlert } from "@lucide/svelte";

    interface Props {
        messages: ChatMessage[];
        pending: boolean;
        error?: boolean;
    }

    let { messages, pending, error }: Props = $props();

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

<div class="flex flex-col gap-0">
    <p class="text-sm font-semibold py-2">Output:</p>
    <div
        bind:this={container}
        onscroll={handleScroll}
        class="w-full step-agent shadow-inner bg-base-200 dark:bg-base-100 rounded-field p-4"
    >
        {#if pending && messages.length === 0}
            <span class="skeleton skeleton-text w-full h-4 text-sm">
                Waiting for step to complete...
            </span>
        {:else}
            {#if error}
                <div class="mb-3 rounded-lg border border-error/20 bg-error/10 p-3">
                    <div class="mb-2 flex items-center gap-2 text-sm">
                        <TriangleAlert class="size-4 text-error" />
                        <span class="font-medium text-error">Error</span>
                    </div>
                    <pre class="mt-2 rounded bg-base-100 p-2 text-xs break-all whitespace-pre-wrap text-error">This step failed to complete.</pre>
                </div>
            {/if}
            <Messages inline messages={messages} />
        {/if}
    </div>
</div>