<script lang="ts">
    import type { ChatService } from "$lib/chat.svelte";
	import type { ResourceContents } from "$lib/types";
	import { fade } from "svelte/transition";

    interface Props {
        chat: ChatService;
    }

    let { chat }: Props = $props();

    const progressUri = 'chat://progress';
    const todoUri = 'todo:///list';

    let progress = $state<ResourceContents | null>(null);
    let todo = $state<ResourceContents | null>(null);

    let hasProgressResource = $derived(chat.resources.some((r) => r.uri === progressUri));
    let hasTodoResource = $derived(chat.resources.some((r) => r.uri === todoUri));

    $effect(() => {
        if (!chat.chatId) return;
        if (hasProgressResource) {
            chat.readResource(progressUri).then((result) => {
                progress = result.contents?.[0] ?? null;
            });
        }
        if (hasTodoResource) {
            chat.readResource(todoUri).then((result) => {
                todo = result.contents?.[0] ?? null;
            });
        }

		// Subscribe to live updates
		const progressCleanup = hasProgressResource ? chat.watchResource(progressUri, (updatedResource) => {
			console.debug('[WorkflowSidebar] Resource updated:', {progressUri, updatedResource});
			progress = updatedResource;
		}) : null;

		const todoCleanup = hasTodoResource ? chat.watchResource(todoUri, (updatedResource) => {
			console.debug('[WorkflowSidebar] Resource updated:', {todoUri, updatedResource});
			todo = updatedResource;
		}) : null;

		// Cleanup subscription when component unmounts or filename changes
		return () => {
            progressCleanup?.();
            todoCleanup?.();
        };
    })

    $effect(() => {
        console.log({ chat });
    })

</script>

{#if chat.chatId}
    <div class="max-w-[300px] h-dvh overflow-hidden" in:fade={{ duration: 150 }}>
        <div class="w-full h-full bg-base-100 flex flex-col">
            <div class="flex-1 overflow-auto p-4 pt-0">
                <div class="flex flex-col gap-2">
                    <h2 class="text-lg font-bold">Progress</h2>
                    <p class="text-sm text-base-content/60">{progress?.text ?? 'No progress found'}</p>
                </div>
            </div>
            <div class="flex-1 overflow-auto p-4 pt-0">
                <div class="flex flex-col gap-2">
                    <h2 class="text-lg font-bold">TODO</h2>
                    <p class="text-sm text-base-content/60">{todo?.text ?? 'No TODOs found'}</p>
                </div>
            </div>
        </div>
    </div>
{/if}