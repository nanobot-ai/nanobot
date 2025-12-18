<script lang="ts">
	import '$lib/../app.css';
	import MarkdownEditor from '$lib/components/MarkdownEditor.svelte';
	import { GripVertical, Plus } from '@lucide/svelte';

    let currentFocusedElement = $state<HTMLElement | null>(null);
    let taskBlockHandle = $state<HTMLElement | null>(null);
    let taskBlockHandleYPosition = $state(0);
    let scrollContainer = $state<HTMLElement | null>(null);
    let tasksContainer = $state<HTMLElement | null>(null);
    let taskListContainer = $state<HTMLElement | null>(null);

    // Drag & drop state
    let isDragging = $state(false);
    let draggedTaskIndex = $state<number | null>(null);
    let dropTargetIndex = $state<number | null>(null);
    let dragGhost = $state<HTMLElement | null>(null);
    let dragOffsetY = $state(0);

    function updateHandlePosition() {
        if (!currentFocusedElement || !scrollContainer || !tasksContainer) return;
        
        const elementTop = currentFocusedElement.offsetTop;
        const elementBottom = elementTop + currentFocusedElement.offsetHeight;
        const scrollTop = scrollContainer.scrollTop;
        const tasksContainerOffset = tasksContainer.offsetTop;
        
        // Calculate where the viewport top is relative to the tasks container
        const viewportTopRelative = scrollTop - tasksContainerOffset;
        
        // Clamp: stay at element top, or stick to viewport top (but not past element bottom)
        const clampedPosition = Math.max(elementTop, Math.min(viewportTopRelative, elementBottom - 40));
        
        taskBlockHandleYPosition = clampedPosition;
    }

    function getTaskIndexFromElement(element: HTMLElement): number | null {
        if (!taskListContainer) return null;
        const taskElements = Array.from(taskListContainer.children).filter(
            (el) => !el.classList.contains('drop-indicator')
        );
        const index = taskElements.indexOf(element);
        return index >= 0 ? index : null;
    }

    function startDrag(e: MouseEvent) {
        if (!currentFocusedElement || !taskListContainer) return;
        
        e.preventDefault();
        
        const index = getTaskIndexFromElement(currentFocusedElement);
        if (index === null) return;
        
        isDragging = true;
        draggedTaskIndex = index;
        
        // Create ghost clone
        const rect = currentFocusedElement.getBoundingClientRect();
        dragOffsetY = e.clientY - rect.top;
        
        const ghost = currentFocusedElement.cloneNode(true) as HTMLElement;
        ghost.classList.add('drag-ghost');
        ghost.style.position = 'fixed';
        ghost.style.width = `${rect.width}px`;
        ghost.style.left = `${rect.left}px`;
        ghost.style.top = `${e.clientY - dragOffsetY}px`;
        ghost.style.pointerEvents = 'none';
        ghost.style.zIndex = '1000';
        ghost.style.opacity = '0.7';
        ghost.style.transform = 'scale(1.02)';
        ghost.style.boxShadow = '0 8px 32px rgba(0,0,0,0.2)';
        document.body.appendChild(ghost);
        dragGhost = ghost;
        
        // Add event listeners
        document.addEventListener('mousemove', handleDragMove);
        document.addEventListener('mouseup', endDrag);
    }

    function handleDragMove(e: MouseEvent) {
        if (!isDragging || !dragGhost || !taskListContainer || draggedTaskIndex === null) return;
        
        // Move ghost
        dragGhost.style.top = `${e.clientY - dragOffsetY}px`;
        
        // Calculate drop target - filter out drop indicators
        const taskElements = Array.from(taskListContainer.children).filter(
            (el) => !el.classList.contains('drop-indicator')
        ) as HTMLElement[];
        let newDropIndex: number | null = null;
        
        for (let i = 0; i < taskElements.length; i++) {
            const taskEl = taskElements[i];
            const rect = taskEl.getBoundingClientRect();
            const midY = rect.top + rect.height / 2;
            
            if (e.clientY < midY) {
                newDropIndex = i;
                break;
            }
        }
        
        // If we didn't find a position, drop at the end
        if (newDropIndex === null) {
            newDropIndex = taskElements.length;
        }
        
        // If hovering over the dragged item (index or index+1), keep indicator at original position
        // Both positions result in "no change", so show consistent indicator
        if (newDropIndex === draggedTaskIndex || newDropIndex === draggedTaskIndex + 1) {
            newDropIndex = draggedTaskIndex;
        }
        
        dropTargetIndex = newDropIndex;
    }

    function endDrag() {
        if (!isDragging) return;
        
        // Remove ghost
        if (dragGhost) {
            dragGhost.remove();
            dragGhost = null;
        }
        
        // Reorder tasks if we have a valid drop target and position changed
        if (dropTargetIndex !== null && draggedTaskIndex !== null) {
            // Skip if dropping at original position (no change)
            const isUnchanged = dropTargetIndex === draggedTaskIndex || dropTargetIndex === draggedTaskIndex + 1;
            
            if (!isUnchanged) {
                const tasks = [...workflow.tasks];
                const [removed] = tasks.splice(draggedTaskIndex, 1);
                
                // Adjust index if dropping after the original position
                const insertIndex = dropTargetIndex > draggedTaskIndex ? dropTargetIndex - 1 : dropTargetIndex;
                tasks.splice(insertIndex, 0, removed);
                
                workflow.tasks = tasks;
            }
        }
        
        // Reset state
        isDragging = false;
        draggedTaskIndex = null;
        dropTargetIndex = null;
        currentFocusedElement = null;
        
        // Remove event listeners
        document.removeEventListener('mousemove', handleDragMove);
        document.removeEventListener('mouseup', endDrag);
    }

	// The existing chat might have been set by / so don't recreate it because that will
	// loose the event stream.
	
    // const chat = page.data.chat || new ChatService();
	// const notification = getNotificationContext();

	// $effect(() => {
	// 	if (!page.params.id) return;
	// 	chat.setChatId(page.params.id).catch((e) => {
	// 		console.error('Error setting chat ID:', e);
	// 		notification.error(e.message);
	// 	});
	// });

	// onDestroy(() => {
	// 	chat.close();
	// });

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
</script>

<svelte:head>
    <title>Nanobot | Workflows</title>
</svelte:head>

<div class="flex flex-col gap-4 p-4 overflow-y-auto max-h-dvh" bind:this={scrollContainer} onscroll={updateHandlePosition}>
    <div class="flex flex-col w-full">
        <input name="title" class="input input-ghost input-xl w-full placeholder:text-base-content/30" type="text" placeholder="Workflow title" />
        <input name="description" class="input input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Workflow description" />
    </div>
    <div class="w-full relative" bind:this={tasksContainer} onmouseleave={() => currentFocusedElement = null} role="presentation">
        <div class="task-block-handle" data-show={currentFocusedElement !== null && !isDragging} bind:this={taskBlockHandle} style="top: {taskBlockHandleYPosition}px;">
            <button class="btn btn-ghost btn-square"><Plus class="size-6 text-base-content/50" /></button>
            <button class="btn btn-ghost btn-square cursor-grab" onmousedown={startDrag}><GripVertical class="size-6 text-base-content/50" /></button>
        </div>
        <div class="flex flex-col gap-4" bind:this={taskListContainer}>
            {#each workflow.tasks as task, index (task.id)}
                {#if dropTargetIndex === index}
                    <div class="drop-indicator"></div>
                {/if}
                <div class="w-full px-22" 
                    class:dragging={isDragging && draggedTaskIndex === index}
                    onmouseenter={(e) => {
                        if (isDragging) return;
                        currentFocusedElement = e.currentTarget as HTMLElement;
                        updateHandlePosition();
                    }}
                    role="presentation"
                >
                    <div class="flex flex-col gap-4 bg-base-200 rounded-box p-4 pb-8 workflow-task">
                        <div class="flex flex-col gap-2">
                            <input name="task-name" class="input input-ghost input-lg w-full font-semibold placeholder:text-base-content/30" type="text" placeholder="Task name" />
                            <input name="task-description" class="input input-ghost w-full placeholder:text-base-content/30" type="text" placeholder="Task description" />
                        </div>
                        <MarkdownEditor value={task.content} />
                    </div>
                </div>
            {/each}
            {#if dropTargetIndex === workflow.tasks.length}
                <div class="drop-indicator"></div>
            {/if}
        </div>
    </div>
</div>

<style>
    .workflow-task :global(.milkdown) {
        background: var(--color-base-200);
    }

    .task-block-handle {
        position: absolute;
        display: flex;
        top: 0;
        left: 0;
        z-index: 2;
        opacity: 0;
        transition: all 0.2s ease-in-out;
    }

    .task-block-handle[data-show="true"] {
        opacity: 1;
    }

    .dragging {
        opacity: 0.3;
    }

    .drop-indicator {
        height: 12px;
        background: var(--color-primary);
        border-radius: 2px;
        margin: -2px 5.5rem;
        position: relative;
    }

    .drop-indicator::before,
    .drop-indicator::after {
        content: '';
        position: absolute;
        width: 12px;
        height: 12px;
        background: var(--color-primary);
        border-radius: 50%;
        top: 50%;
        transform: translateY(-50%);
    }

    .drop-indicator::before {
        left: -6px;
    }

    .drop-indicator::after {
        right: -6px;
    }
</style>