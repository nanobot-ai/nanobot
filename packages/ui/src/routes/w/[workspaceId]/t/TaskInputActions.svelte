<script lang="ts">
	import { ChevronRight, Plus } from "@lucide/svelte";
	import type { Input, Task, Step } from "./types";

    interface Props {
        task: Task;
        availableInputs: Input[];
        onAddInput: (input: Input) => void;
    }

    let { task = $bindable(), availableInputs, onAddInput }: Props = $props();

    function blur() {
        document.getElementById('add-to-input')?.hidePopover();
    }
</script>

<button class="btn btn-ghost btn-square btn-sm tooltip tooltip-right" data-tip="Add..." popoverTarget="add-to-input" style="anchor-name: --add-to-input-anchor;">
    <Plus class="text-base-content/50" />
</button>
<ul class="dropdown menu w-72 rounded-box bg-base-100 dark:bg-base-300 shadow-sm overflow-visible"
    popover="auto" id="add-to-input" style="position-anchor: --add-to-input-anchor;">
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
</ul>