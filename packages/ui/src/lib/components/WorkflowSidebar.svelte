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

    $effect(() => {
        Promise.all([
            chat.readResource(progressUri),
            chat.readResource(todoUri)
        ]).then(([progressResult, todoResult]) => {
            if (progressResult.contents?.length) {
                progress = progressResult.contents[0];
            }
            if (todoResult.contents?.length) {
                todo = todoResult.contents[0];
            }
        });

		// Subscribe to live updates
		const progressCleanup = chat.watchResource(progressUri, (updatedResource) => {
			console.debug('[WorkflowSidebar] Resource updated:', {progressUri, updatedResource});
			progress = updatedResource;
		});

		const todoCleanup = chat.watchResource(todoUri, (updatedResource) => {
			console.debug('[WorkflowSidebar] Resource updated:', {todoUri, updatedResource});
			todo = updatedResource;
		});

		// Cleanup subscription when component unmounts or filename changes
		return () => {
            progressCleanup();
            todoCleanup();
        };
    })

</script>

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