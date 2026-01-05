<script lang="ts">
	import { EllipsisVertical, EyeClosed, HandHelping, ReceiptText, Sparkles } from "@lucide/svelte";
	import type { Input, Task } from "./types";

    interface Props {
        id: string;
        task: Task;
        input: Input;
        inputDescription: Map<string | number, boolean>;
        inputDefault: Map<string | number, boolean>;
        onHideInput: (id: string) => void;
        onDeleteInput: (id: string) => void;
        onToggleInputDescription: (id: string, value: boolean) => void;
        onToggleInputDefault: (id: string, value: boolean) => void;
        onSuggestImprovement: (content: string) => void;
    }

    let { 
        id, 
        task,
        input = $bindable(),
        inputDescription,
        inputDefault,
        onHideInput,
        onDeleteInput,
        onToggleInputDescription,
        onToggleInputDefault,
        onSuggestImprovement,
    }: Props = $props();
</script>

<div class="flex flex-col gap-2 bg-base-100 dark:bg-base-200 shadow-xs rounded-box p-4 pb-8 task-step relative">
    <div class="absolute top-3 right-3 z-2">
        {@render inputMenu(input.id, input.name)}
    </div>
    
    <div class="flex flex-col gap-2 pr-12">
        <label class="input w-full">
            <span class="label h-full font-semibold text-primary bg-primary/15 mr-0">$</span>
            <input type="text" class="font-semibold placeholder:font-normal" bind:value={input.name} placeholder="Argument name (ex. CompanyName)"/>
        </label>

        <input name="input-description" class="input w-full placeholder:text-base-content/30" type="text" placeholder="What is this argument for?" bind:value={input.description} />
        <input name="input-default" class="input w-full placeholder:text-base-content/30" type="text" placeholder="Default value (ex. Obot)" bind:value={input.default} />
    </div>
</div>

{#snippet inputMenu(id: string, name: string)}
    <button class="btn btn-ghost btn-square btn-sm" popoverTarget={`input-${id}-action`} style={`anchor-name: --input-${id}-action-anchor;`}>
        <EllipsisVertical class="text-base-content/50" />
    </button>

    <ul class="dropdown flex flex-col gap-1 dropdown-end dropdown-bottom menu w-64 rounded-box bg-base-100 dark:bg-base-300 shadow-sm border border-base-300"
        popover="auto" id={`input-${id}-action`} style={`position-anchor: --input-${id}-action-anchor;`}>
        <li>
            <label for={`step-${id}-description`} class="flex gap-2 justify-between items-center">
                <span class="flex items-center gap-2">
                    <ReceiptText class="size-4" />
                    Description
                </span>
                <input type="checkbox" class="toggle toggle-sm" id={`step-${id}-description`} 
                    checked={inputDescription.get(id) ?? false}
                    onchange={(e) => {
                        onToggleInputDescription(id, (e.target as HTMLInputElement)?.checked ?? false);
                    }}
                />
            </label>
        </li>
        <li>
            <label for={`step-${id}-description`} class="flex gap-2 justify-between items-center">
                <span class="flex items-center gap-2">
                    <HandHelping class="size-4" />
                    Default value
                </span>
                <input type="checkbox" class="toggle toggle-sm" id={`step-${id}-description`} 
                    checked={inputDefault.get(id) ?? false}
                    onchange={(e) => {
                        onToggleInputDefault(id, (e.target as HTMLInputElement)?.checked ?? false);
                    }}
                />
            </label>
        </li>
        
        <li>
            <button class="flex items-center gap-2"
                onclick={() => onSuggestImprovement(`
The user is asking for an improvement to the following input:
Argument name: ${name}
Argument description: ${input.description}
Argument default value: ${input.default}

Please provide a detailed improvement to the input. If the user has not provided a description, suggest one based on the argument name and default value.
Do not suggest a default value if the user has not provided one.
                `)}
            >
                <Sparkles class="size-4" /> Improve with AI
            </button>
        </li>
        {#if name.length > 0 && task!.inputs.some((input) => input.name === name)}
            <li>
                <button class="flex items-center gap-2"
                    onclick={() => { onHideInput(id) }}
                >
                    <EyeClosed class="size-4" /> Hide argument
                </button>
            </li>
        {:else}
            <li>
                <button class="flex items-center gap-2"
                    onclick={() => onDeleteInput(id)}
                >
                    <EyeClosed class="size-4" /> Delete argument
                </button>
            </li>
        {/if}
    </ul>
{/snippet}