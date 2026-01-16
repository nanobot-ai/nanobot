<script lang="ts">
	import { EllipsisVertical, ReceiptText, Sparkles, ToolCase, Trash2, Wrench, X } from "@lucide/svelte";
	import type { Input, Step, Task } from "./types";
	import MarkdownEditor from "$lib/components/MarkdownEditor.svelte";
	import { createVariablePillPlugin } from "$lib/plugins/variablePillPlugin";
	import { getNotificationContext } from "$lib/context/notifications.svelte";
	import { getRegistryContext } from "$lib/context/registry.svelte";
	import type { Snippet } from "svelte";

    interface Props {
        class?: string;
        taskId: string;
        task: Task;
        step: Step;
        showDescription?: boolean;
        showBlockEditing?: boolean;
        onAddInput?: (input: Input) => void;
        onAddTaskInput?: (input: Input) => void;
        onRemoveTaskInput?: (inputName: string) => void;
        onDeleteStep?: (stepId: string, filename: string) => void;
        onToggleStepDescription?: (id: string, value: boolean) => void;
        onToggleStepBlockEditing?: (id: string, value: boolean) => void;
        onUpdateStep?: (id: string, updates: Partial<Step>) => void;
        visibleInputs?: Input[];
        onUpdateVisibleInputs?: (inputs: Input[]) => void;
        onSuggestImprovement?: (opts: { uri: string, name: string, mimeType: string }) => void;
        readonly?: boolean;
        children?: Snippet<[{ item: Step }]>;
    }

    let { 
        class: klass,
        taskId,
        task, 
        step, 
        showDescription, 
        showBlockEditing, 
        onAddInput,
        onAddTaskInput,
        onRemoveTaskInput,
        onDeleteStep,
        onToggleStepDescription, 
        onToggleStepBlockEditing,
        onUpdateStep,
        visibleInputs = [],
        onUpdateVisibleInputs,
        onSuggestImprovement,
        readonly,
        children,
    }: Props = $props();
    const notifications = getNotificationContext();
    const registry = getRegistryContext();
    
    let tools = $derived(
        step.tools && step.tools.length > 0 && !registry.loading && registry.servers.length > 0 
            ? step.tools.map((tool) => registry.getServerByName(tool.name)).filter((tool) => tool !== undefined) 
            : []
    );
    
    function handleRemoveTool(toolName: string) {
        step.tools = step.tools.filter((tool) => tool.name !== toolName);
        onUpdateStep?.(step.id, { tools: step.tools });
    }

    function blur() {
        document.getElementById(`step-${step.id}-action`)?.hidePopover();
    }

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
                onAddTaskInput?.(newInput);
                notifications.action(
                    `${variable}`, 
                    'A new variable has been added. Would you like to add more details to it now?',
                    () => {
                        onAddInput?.(newInput);
                    },
                );
            }
        },
		onVariableDeletion: (variable: string) => {
            const variableRegex = new RegExp(`\\$${variable}(?![a-zA-Z0-9_])`);
            const stillExists = task?.steps.some((stepToCheck) => {
                if (stepToCheck.id === step.id) {
                    return false; // already know this step removed the variable so return false
                }
                return variableRegex.test(stepToCheck.content);
            });
            if (!stillExists) {
                const hasVisible = visibleInputs.some((input) => input.name === variable);
                if (hasVisible) {
                    notifications.action(
                        `${variable}`,
                        'Would you like to remove the variable details from this task?',
                        () => {
                            onRemoveTaskInput?.(variable);
                            onUpdateVisibleInputs?.(visibleInputs.filter((input) => input.name !== variable));
                        }
                    )
                } else {
                    onRemoveTaskInput?.(variable);
                }
            }
        },
        getAvailableVariables: () => {
            // Combine task inputs and visible inputs, returning unique variable names
            const taskInputNames = task?.inputs.map((input) => input.name) || [];
            const visibleInputNames = visibleInputs.map((input) => input.name);
            const allNames = [...new Set([...taskInputNames, ...visibleInputNames])];
            return allNames;
        },
	});
</script>

<div class="flex flex-col gap-2 bg-base-100 dark:bg-base-200 shadow-xs rounded-box p-4 pb-8 task-step relative {klass}">
    {#if !readonly}
        <div class="absolute top-3 right-3 z-2">
            {@render stepMenu(step.id)}
        </div>
    {/if}
    
    <div class="flex flex-col pr-12">
        <input name="step-name" class="input input-ghost input-lg w-full font-semibold placeholder:text-base-content/30" type="text" placeholder="Step name" 
            value={step.name} 
            oninput={(e) => onUpdateStep?.(step.id, { name: (e.target as HTMLInputElement).value })} 
            disabled={readonly}
        />
        {#if showDescription}
            <input name="step-description" class="input text-[16px] input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Step description" 
                value={step.description} 
                oninput={(e) => onUpdateStep?.(step.id, { description: (e.target as HTMLInputElement).value })} 
                disabled={readonly}
            />
        {/if}
    </div>

    <MarkdownEditor 
        value={step.content} 
        blockEditEnabled={showBlockEditing} 
        plugins={[variablePillPlugin]} 
        onChange={(value) => onUpdateStep?.(step.id, { content: value })}
        {readonly}
    />

    {#if tools.length > 0}
        <div class="flex flex-wrap gap-2">
            {#each tools as tool (tool.name)}
            <div class="indicator group">
                {#if !readonly}
                    <div class="indicator-item group-hover:opacity-100 opacity-0 transition-opacity duration-150">
                        <button class="btn btn-primary size-4 btn-circle tooltip" onclick={() => handleRemoveTool(tool.name)} data-tip="Remove tool">
                            <X class="size-2" />
                        </button>
                    </div>
                {/if}
                <div class="badge dark:bg-base-200 size-fit py-1">
                    {#if tool.icons?.[0]?.src}
                        <img alt={tool.title} src={tool.icons[0].src} class="size-4"/>
                    {:else}
                        <Wrench class="size-4" />
                    {/if}
                    {tool.title}
                </div>
            </div>
            {/each}
        </div>
    {/if}

    {#if children}
        {@render children({ item: step })}
    {/if}
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
                    checked={showDescription}
                    onchange={(e) => {
                        onToggleStepDescription?.(id, (e.target as HTMLInputElement)?.checked ?? false);
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
                    checked={showBlockEditing}
                    onchange={(e) => {
                        onToggleStepBlockEditing?.(id, (e.target as HTMLInputElement)?.checked ?? false);
                    }}
                />
            </label>
        </li>
        <li>
            <button class="flex items-center gap-2"
                onclick={() => {
                    onSuggestImprovement?.({
                        uri: `workspace://.nanobot/tasks/${taskId}/${id}`,
                        name: step.name,
                        mimeType: 'application/octet-stream',
                    });
                    blur();
                }}
            >
                <Sparkles class="size-4" /> Improve with AI
            </button>
        </li>
        <!-- <li>
            <button class="flex items-center gap-2 disabled:opacity-50 disabled:hover:bg-transparent disabled:cursor-default"
                onclick={() => {
                    // TODO: share step
                    blur();
                }}
                disabled
            >
                <Share class="size-4" /> Share step with...
            </button>
        </li> -->
        <li>
            <button class="flex items-center gap-2 menu-alert"
                onclick={() => {
                    const filename = `.nanobot/tasks/${taskId}/${id}`;
                    onDeleteStep?.(id, filename);
                    blur();
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