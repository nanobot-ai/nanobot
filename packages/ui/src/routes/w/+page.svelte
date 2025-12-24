<script lang="ts">
	import '$lib/../app.css';
	import DragDropList from '$lib/components/DragDropList.svelte';
	import MarkdownEditor from '$lib/components/MarkdownEditor.svelte';
	import MessageInput from '$lib/components/MessageInput.svelte';
	import { getLayoutContext } from '$lib/context/layout.svelte';
	import { EllipsisVertical, GripVertical, MessageCircleMore, Play, Plus, ReceiptText, Sparkles, ToolCase, Trash2, X } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { fade, fly, slide } from 'svelte/transition';
	import { createVariablePillPlugin } from '$lib/plugins/variablePillPlugin';
	import { onMount } from 'svelte';

    let scrollContainer = $state<HTMLElement | null>(null);

    let taskBlockEditing = new SvelteMap<number | string, boolean>();
    let taskDescription = new SvelteMap<number | string, boolean>();

    let currentRun = $state<unknown | null>(null);
    let showCurrentRun = $state(false);

    let showMessageInput = $state(false);
    let showAlternateHeader = $state(false);
    
    const layout = getLayoutContext();
    const variablePillPlugin = createVariablePillPlugin({
		onVariableAddition: (variable: string) => {
            console.log('variable added', variable);
        },
		onVariableDeletion: (variable: string) => {
            console.log('variable deleted', variable);
        },
	});

    let workflow = $state({
		name: 'Onboarding Workflow',
		description: 'This workflow is used to onboard new users to the platform.',
		prompt:
			'You are an assistant responsible for onboarding members to the CNCF without breaking the rules.\n\nRules:\n* You can read from Salesforce.\n* Never add a new record to Salesforce.\n* Follow the Action you are told.',
		arguments: [
			{
				id: '1',
				name: 'CompanyName',
				displayLabel: 'Company Name',
				description: 'The name of the company to onboard. Example: Obot',
				visible: true
			}
		],
		tasks: [
			{
				id: '1',
				name: 'Add Member to Google Sheet',
				description: 'This task will add the member to the google sheet',
				content: `
1. Get the account record for $CompanyName in Salesforce. You also need to get all related Contacts including roles and emails. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). Notes or special instructions are typically found in the Account’s Description field. To review everything created for a company, look for these related records and fields using the Account’s Id.

2. Get the Demo Workflow LF Google Sheet. Read the first few rows to understand the sheet and formats used in each column.

3. Follow the formatting and style, add a new row to the google sheet for the member $CompanyName based on the information we got previously in Salesforce. The join date should be the closed won date.
				`
			},
			{
				id: '2',
				name: 'Add New Member Contacts to Google Groups',
				description:
					'This will add the contacts of our new members to the appropriate Google Groups',
				content:
					'1. Get the account record for \$CompanyName in Salesforce. You also need to get all related Contacts including roles and emails.'
			},

			{
				name: 'Add member contacts to Slack',
				description: '',
				content: `
1. Get the account record for \$CompanyName in Salesforce. You also need to get all related Contacts including roles and emails.

2. List all channels including private ones.

3. Search for the marketing contacts in the slack workspace to see if they are members. If they are in the workspace add them to the private marketing channel.

4. Search for the business owner contacts in the slack workspace to see if they are members. If they are in the workspace, add them to the private business-owners channel.

5. Search for the technical contacts in the slack workspace to see if they are members. If they are members of the workspace, add them to the private technical-leads channel.

6. Report back who is a member of slack already.`,
				id: '3'
			},
			{
				name: 'Add Logo to site',
				description: 'Create a github PR to add the logo to the site',
				content: `
1. Get the account record for \$CompanyName in Salesforce. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). To review everything created for a company, look for these related records and fields using the Account’s Id.

2. Create a branch in the repo cloudnautique/obot-mcpserver-examples called add-\$CompanyName-logo.

3. Create a file in the workspace called logo.txt and write a story about a robot in markdown.

4. Add the file to the assets/img directory called \$CompanyName-logo.txt, and create a PR back into the main branch.
				`,
				id: '4'
			},
			{
				name: 'Send Welcome Email',
				description: 'Used to send welcome email when the org has been onboarded.',
				content: `
1. Get the account record for \$CompanyName in Salesforce. You also need to get the contacts and their email addresses. Also get the most recent opportunity to determine the membership.

2. Using gmail tools, create a draft email using the business owner contact, account, and opportunity info.

\`\`\` Markdown
# CNCF Onboarding Completion 

**Subject:** Welcome to the Cloud Native Computing Foundation (CNCF)!

---

Dear {{FirstName}} {{LastName}},

Congratulations, and welcome to the **Cloud Native Computing Foundation (CNCF)** community!

We’re pleased to let you know that all onboarding steps for **{{CompanyName}}** have been successfully completed. Your organization is now fully set up as a {{membership level}} member and ready to take advantage of CNCF programs, resources, and community benefits.

Congrats!
Dir. CNCF Onboarding Agent
\`\`\`

Send the drafted email.
				`,
				id: '5'
			}
		]
	});

    onMount(() => {
        workflow.tasks.forEach((task) => {
            if (task.description) {
                taskDescription.set(task.id, true);
            }
        })
    })

    function toggleTaskBlockEditing(id: number, enabled: boolean) {
        taskBlockEditing.set(id, enabled);
    }

    function toggleTaskDescription(id: number, enabled: boolean) {
        taskDescription.set(id, enabled);
    }
</script>

<svelte:head>
    <title>Nanobot | Workflows</title>
</svelte:head>

<div class="flex w-full h-dvh">
    <div class="
        flex flex-col p-4 pt-0 overflow-y-auto max-h-dvh transition-all duration-200 ease-in-out 
        {layout.isSidebarCollapsed ? 'mt-10' : ''}
    " 
        bind:this={scrollContainer}
        onscroll={() => {
            showAlternateHeader = (scrollContainer?.scrollTop ?? 0) > 100;
        }}
    >
        <div class="sticky top-0 left-0 w-full bg-base-200 dark:bg-base-100 z-10 py-4">
            <div in:fade class="flex flex-col grow">
                <div class="flex w-full items-center gap-4">
                    {#if showAlternateHeader}
                        <p in:fade class="flex grow text-xl font-semibold">{workflow.name}</p>
                    {:else}
                        <input name="title" class="input input-ghost input-xl w-full placeholder:text-base-content/30 font-semibold" type="text" placeholder="Workflow title" 
                            bind:value={workflow.name}
                        />
                    {/if}
                    <button class="btn btn-primary w-48" onclick={() => {
                        showCurrentRun = true;
                    }}>
                        Run <Play class="size-4" /> 
                    </button>
                </div>
                {#if !showAlternateHeader}
                    <input out:slide={{ axis: 'y' }} name="description" class="input input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Workflow description"
                        bind:value={workflow.description}
                    />
                {/if}
            </div>
        </div>
        <DragDropList bind:items={workflow.tasks} scrollContainerEl={scrollContainer}
            class={showCurrentRun ? '' : 'md:pr-22'}
            classes={{
                dropIndicator: 'mx-22 my-2 h-2',
                item: 'pl-22'
            }}
        >
            {#snippet blockHandle({ startDrag, currentItem })}
                <div class="flex items-center gap-2">
                    <button class="btn btn-ghost btn-square btn-sm" popoverTarget="add-to-workflow" style="anchor-name: --add-to-workflow-anchor;">
                        <Plus class="text-base-content/50" />
                    </button>
                    
                    <ul class="dropdown menu w-72 rounded-box bg-base-100 dark:bg-base-300 shadow-sm"
                        popover="auto" id="add-to-workflow" style="position-anchor: --add-to-workflow-anchor;">
                        <li>
                            <button class="justify-between"
                                onclick={(e) => {
                                    const currentIndex = workflow.tasks.findIndex((task) => task.id === currentItem?.id);
                                    const newTask = {
                                        id: '',
                                        name: '',
                                        description: '',
                                        content: ''
                                    };
                                    if (e.metaKey) {
                                        workflow.tasks.splice(currentIndex, 0, newTask);
                                    } else {
                                        workflow.tasks.splice(currentIndex + 1, 0, newTask);
                                    }

                                    (document.activeElement as HTMLElement)?.blur();
                                }}
                            >
                                <span>Add new task</span>
                                <span class="text-[11px] text-base-content/50">
                                    click / <kbd class="kbd ">⌘</kbd> + click
                                </span>
                            </button>
                        </li>
                        <li><button>Add a tool</button></li>
                    </ul>

                    <button class="btn btn-ghost btn-square cursor-grab btn-sm" onmousedown={startDrag}><GripVertical class="text-base-content/50" /></button>
                </div>
            {/snippet}
            {#snippet children({ item: task })}
                <div class="flex flex-col gap-2 bg-base-100 dark:bg-base-200 shadow-xs rounded-box p-4 pb-8 workflow-task relative">
                    <div class="absolute top-3 right-3 z-2">
                        {@render menu(task.id)}
                    </div>
                    
                    <div class="flex flex-col pr-12">
                        <input name="task-name" class="input input-ghost input-lg w-full font-semibold placeholder:text-base-content/30" type="text" placeholder="Task name" bind:value={task.name} />
                        {#if taskDescription.get(task.id) ?? false}
                            <input name="task-description" class="input text-[16px] input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Task description" bind:value={task.description} />
                        {/if}
                    </div>
                   
                    <MarkdownEditor value={task.content} blockEditEnabled={taskBlockEditing.get(task.id) ?? false} plugins={[variablePillPlugin]} />
                </div>
            {/snippet}
        </DragDropList>
        
        {#if !showCurrentRun}
            <div in:fade={{ duration: 200 }} class="sticky bottom-0 right-0 self-end flex flex-col gap-4 z-10">
                {#if showMessageInput}
                    <div class="bg-base-100 dark:bg-base-200 border border-base-300 rounded-selector w-sm md:w-2xl"
                        transition:fly={{ x: 100, duration: 200 }}
                    >
                        <MessageInput />
                    </div>  
                {/if}

                <button class="float-right btn btn-lg btn-circle btn-primary self-end" onclick={() => showMessageInput = !showMessageInput}>
                    <MessageCircleMore class="size-6" />
                </button>
            </div>
        {/if}
    </div>

    {#if showCurrentRun}
        <div transition:fly={{ x: 100, duration: 200 }} class="md:min-w-[520px] bg-base-100 h-dvh">
            <div class="w-full h-full flex flex-col max-h-dvh overflow-y-auto">
                <div class="w-full flex justify-between items-center pr-4 bg-base-100">
                    <h4 class="text-lg font-semibold border-l-4 border-primary p-4 pr-0">{workflow.name} | Run {'{id}'}</h4>
                    <button class="btn btn-ghost btn-square btn-sm" onclick={() => showCurrentRun = false}>
                        <X class="size-4" />
                    </button>
                </div>
                <div class="flex grow p-4 pt-0">
                    Thread content here
                </div>
                <div class="sticky bottom-0 left-0 w-full">
                    <MessageInput />
                </div>
            </div>
        </div>
    {/if}
</div>

{#snippet menu(id: number)}
    <button class="btn btn-ghost btn-square btn-sm" popoverTarget={`task-${id}-action`} style={`anchor-name: --task-${id}-action-anchor;`}>
        <EllipsisVertical class="text-base-content/50" />
    </button>

    <ul class="dropdown flex flex-col gap-1 dropdown-end dropdown-bottom menu w-64 rounded-box bg-base-100 dark:bg-base-300 shadow-sm border border-base-300"
        popover="auto" id={`task-${id}-action`} style={`position-anchor: --task-${id}-action-anchor;`}>
        <li>
            <label for={`task-${id}-description`} class="flex gap-2 justify-between items-center">
                <span class="flex items-center gap-2">
                    <ReceiptText class="size-4" />
                    Description
                </span>
                <input type="checkbox" class="toggle toggle-sm" id={`task-${id}-description`} 
                    checked={taskDescription.get(id) ?? false}
                    onchange={(e) => toggleTaskDescription(id, (e.target as HTMLInputElement)?.checked ?? false)}
                />
            </label>
        </li>
        <li>
            <label for={`task-${id}-block-editing`} class="flex gap-2 justify-between items-center">
                <span class="flex items-center gap-2">
                    <ToolCase class="size-4" />
                    Enable block editing
                </span>
                <input type="checkbox" class="toggle toggle-sm" id={`task-${id}-block-editing`} 
                    checked={taskBlockEditing.get(id) ?? false}
                    onchange={(e) => toggleTaskBlockEditing(id, (e.target as HTMLInputElement)?.checked ?? false)}
                />
            </label>
        </li>
        <li>
            <button class="flex items-center gap-2">
                <Sparkles class="size-4" /> Improve with AI
            </button>
        </li>
        <li>
            <button class="flex items-center gap-2">
                <Trash2 class="size-4" /> Delete task
            </button>
        </li>
    </ul>
{/snippet}

<style>
    :root[data-theme=nanobotlight] {
        .workflow-task :global(.milkdown) {
            background: var(--color-base-100);
        }
    }

    :root[data-theme=nanobotdark] {
        .workflow-task :global(.milkdown) {
            background: var(--color-base-200);
        }
    }
</style>