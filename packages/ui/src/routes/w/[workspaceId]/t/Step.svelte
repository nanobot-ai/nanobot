<script lang="ts">
	import { EllipsisVertical, ReceiptText, Sparkles, ToolCase, Trash2 } from "@lucide/svelte";
	import type { Input, Step, Task } from "./types";
	import MarkdownEditor from "$lib/components/MarkdownEditor.svelte";
	import { createVariablePillPlugin } from "$lib/plugins/variablePillPlugin";
	import { getNotificationContext } from "$lib/context/notifications.svelte";

    interface Props {
        id: string;
        task: Task;
        step: Step;
        stepDescription: Map<string | number, boolean>;
        stepBlockEditing: Map<string | number, boolean>;
        onAddInput: (input: Input) => void;
        onDeleteStep: (filename: string) => void;
        onToggleStepDescription: (id: string, value: boolean) => void;
        onToggleStepBlockEditing: (id: string, value: boolean) => void;
        visibleInputs: Input[];
    }

    let { 
        id,
        task = $bindable(), 
        step, 
        stepDescription, 
        stepBlockEditing, 
        onAddInput,
        onDeleteStep,
        onToggleStepDescription, 
        onToggleStepBlockEditing,
        visibleInputs,
    }: Props = $props();
    const notifications = getNotificationContext();
    
    const variablePillPlugin = createVariablePillPlugin({
		onVariableAddition: (variable: string) => {
            const exists = task?.inputs.find((input) => input.name === variable) || visibleInputs.find((input) => input.name === variable);
            if (!exists) {
                const newInput: Input = {
                    id: crypto.randomUUID(),
                    name: variable,
                    description: '',
                    default: ''
                };
                task!.inputs.push(newInput);
                notifications.action(
                    `${variable}`, 
                    'A new variable has been added. Would you like to add more details to it now?',
                    () => {
                        onAddInput(newInput);
                    },
                );
            }
        },
		onVariableDeletion: (variable: string) => {
            const stillExists = task?.steps.some((step) => step.content.includes(`$${variable}`));
            if (!stillExists) {
                const hasVisible = visibleInputs.some((input) => input.name === variable);
                if (hasVisible) {
                    notifications.action(
                        `${variable}`,
                        'Would you like to remove the variable details from this task?',
                        () => {
                            task!.inputs = task!.inputs.filter((input) => input.name !== variable);
                            visibleInputs = visibleInputs.filter((input) => input.name !== variable);
                        }
                    )
                } else {
                    task!.inputs = task!.inputs.filter((input) => input.name !== variable);
                }
                
            }
        },
	});
</script>

<div class="flex flex-col gap-2 bg-base-100 dark:bg-base-200 shadow-xs rounded-box p-4 pb-8 task-step relative">
    <div class="absolute top-3 right-3 z-2">
        {@render stepMenu(step.id)}
    </div>
    
    <div class="flex flex-col pr-12">
        <input name="step-name" class="input input-ghost input-lg w-full font-semibold placeholder:text-base-content/30" type="text" placeholder="Step name" bind:value={step.name} />
        {#if stepDescription.get(step.id) ?? false}
            <input name="step-description" class="input text-[16px] input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Step description" bind:value={step.description} />
        {/if}
    </div>

    <MarkdownEditor 
        value={step.content} 
        blockEditEnabled={stepBlockEditing.get(step.id) ?? false} 
        plugins={[variablePillPlugin]} 
        onChange={(value) => {
            step.content = value;
        }}
    />
</div>


{#snippet stepMenu(id: string)}
    <button class="btn btn-ghost btn-square btn-sm" popoverTarget={`step-${id}-action`} style={`anchor-name: --step-${id}-action-anchor;`}>
        <EllipsisVertical class="text-base-content/50" />
    </button>

    <ul class="dropdown flex flex-col gap-1 dropdown-end dropdown-bottom menu w-64 rounded-box bg-base-100 dark:bg-base-300 shadow-sm border border-base-300"
        popover="auto" id={`step-${id}-action`} style={`position-anchor: --step-${id}-action-anchor;`}>
        <li>
            <label for={`step-${id}-description`} class="flex gap-2 justify-between items-center">
                <span class="flex items-center gap-2">
                    <ReceiptText class="size-4" />
                    Description
                </span>
                <input type="checkbox" class="toggle toggle-sm" id={`step-${id}-description`} 
                    checked={stepDescription.get(id) ?? false}
                    onchange={(e) => {
                        onToggleStepDescription(id, (e.target as HTMLInputElement)?.checked ?? false);
                    }}
                />
            </label>
        </li>
        <li>
            <label for={`step-${id}-block-editing`} class="flex gap-2 justify-between items-center">
                <span class="flex items-center gap-2">
                    <ToolCase class="size-4" />
                    Enable block editing
                </span>
                <input type="checkbox" class="toggle toggle-sm" id={`step-${id}-block-editing`} 
                    checked={stepBlockEditing.get(id) ?? false}
                    onchange={(e) => {
                        onToggleStepBlockEditing(id, (e.target as HTMLInputElement)?.checked ?? false);
                    }}
                />
            </label>
        </li>
        <li>
            <button class="flex items-center gap-2">
                <Sparkles class="size-4" /> Improve with AI
            </button>
        </li>
        <li>
            <button class="flex items-center gap-2"
                onclick={() => {
                    task!.steps = task!.steps.filter((step) => step.id !== id);
                    const filename = `.nanobot/tasks/${id}/${id}.md`;
                    onDeleteStep(filename);
                }}
            >
                <Trash2 class="size-4" /> Delete step
            </button>
        </li>
    </ul>
{/snippet}

<style>
    :root[data-theme=nanobotlight] {
        .task-step :global(.milkdown) {
            background: var(--color-base-100);
        }
    }

    :root[data-theme=nanobotdark] {
        .task-step :global(.milkdown) {
            background: var(--color-base-200);
        }
    }
</style>