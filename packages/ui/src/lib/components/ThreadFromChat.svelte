<script lang="ts">
import type { ChatService } from "$lib/chat.svelte";
import Thread from "$lib/components/Thread.svelte";
	import { getSidebarContext } from "$lib/context/sidebar.svelte";
	import FileEditor from "./FileEditor.svelte";
	import WorkflowSidebar from "./WorkflowSidebar.svelte";

interface Props {
	chat: ChatService;
}

let { chat }: Props = $props();

let selectedFile = $state('');
let drawerInput = $state<HTMLInputElement | null>(null);

const sidebar = getSidebarContext();
$effect(() => {
	console.log({ agent: chat.agent });
})
</script>

{#key chat.chatId}
	<Thread
		messages={chat.messages}
		prompts={chat.prompts}
		resources={chat.resources}
		elicitations={chat.elicitations}
		agents={chat.agents}
		selectedAgentId={chat.selectedAgentId}
		onAgentChange={chat.selectAgent}
		onElicitationResult={chat.replyToElicitation}
		onSendMessage={chat.sendMessage}
		onFileUpload={chat.uploadFile}
		onFileOpen={(filename) => {
			sidebar?.collapse();
			drawerInput?.click();
			selectedFile = filename;
		}}
		cancelUpload={chat.cancelUpload}
		uploadingFiles={chat.uploadingFiles}
		uploadedFiles={chat.uploadedFiles}
		isLoading={chat.isLoading}
		agent={chat.agent}
	/>
{/key}

{#if selectedFile}
	<FileEditor filename={selectedFile} {chat} onClose={() => {
		selectedFile = '';
		sidebar?.expand();
	}}/>
{:else if chat.agent.id === 'planner' || chat.agent.id === 'executor'}
	<WorkflowSidebar {chat} />
{/if}
