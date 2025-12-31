<script lang="ts">
	import { ChevronRight, Plus } from "@lucide/svelte";
	import type { Input, Task, Step } from "./types";

    interface Props {
        task: Task;
        item: Step;
        availableInputs: Input[];
        onAddInput: (input: Input) => void;
        onOpenSelectTool: () => void;
    }

    let { task = $bindable(), item, availableInputs, onAddInput, onOpenSelectTool }: Props = $props();
    function blur() {
        document.getElementById(`add-to-step-${item.id}`)?.hidePopover();
    }
</script>

{#if item}
    <button class="btn btn-ghost btn-square btn-sm" popoverTarget={`add-to-step-${item.id}`} style="anchor-name: --add-to-step-${item.id}-anchor;">
        <Plus class="text-base-content/50" />
    </button>
    <ul class="dropdown menu w-72 rounded-box bg-base-100 dark:bg-base-300 shadow-sm overflow-visible"
        popover="auto" id={`add-to-step-${item.id}`} style="position-anchor: --add-to-step-${item.id}-anchor;">
        <li>
            <button class="justify-between"
                onclick={(e) => {
                    const currentIndex = task!.steps.findIndex((step) => step.id === item?.id);
                    const newStep = {
                        id: task!.steps.length.toString(),
                        name: '',
                        description: '',
                        content: '',
                        tools: [],
                    };
                    if (e.metaKey) {
                        task!.steps.splice(currentIndex, 0, newStep);
                    } else {
                        task!.steps.splice(currentIndex + 1, 0, newStep);
                    }
                    blur();
                }}
            >
                <span>Add new step</span>
                <span class="text-[11px] text-base-content/50">
                    click / <kbd class="kbd ">⌘</kbd> + click
                </span>
            </button>
        </li>
        <li>
            <button 
                onclick={() => {
                    const newInput = {
                        id: crypto.randomUUID(),
                        name: '',
                        description: '',
                        default: ''
                    };
                    task.inputs.push(newInput);
                    onAddInput(newInput);
                    blur();
                }}
            >
                Add new argument
            </button>
        </li>
        {#if availableInputs.length > 0}
            <li class="group/submenu relative">
                <span class="flex justify-between items-center">
                    Add existing argument
                    <ChevronRight class="size-3" />
                </span>
                <ul class="ml-0 menu -translate-y-2 bg-base-100 dark:bg-base-300 rounded-box shadow-md absolute left-full top-0 w-52 invisible opacity-0 group-hover/submenu:visible group-hover/submenu:opacity-100 transition-opacity z-50 before:hidden">
                    {#each availableInputs as input}
                        <li>
                            <button
                                onclick={() => {
                                    onAddInput(input);
                                    blur();
                                }}
                            >
                                {input.name}
                            </button>
                        </li>
                    {/each}
                </ul>
            </li>
        {/if}
        <li><button onclick={onOpenSelectTool}>Add a tool</button></li>
    </ul>
{/if}
