<script lang="ts">
import type { ChatService } from "$lib/chat.svelte";
import { setMcpAppsContext } from "$lib/context/mcpApps.svelte";
import Thread from "$lib/components/Thread.svelte";

interface Props {
	chat: ChatService;
}

const { chat }: Props = $props();

setMcpAppsContext({
	get client() { return chat.client; },
	get tools() { return chat.tools; },
	ensureClient: () => chat.ensureClient(),
	sendMessage: (msg: string) => chat.sendMessage(msg),
});
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
		cancelUpload={chat.cancelUpload}
		uploadingFiles={chat.uploadingFiles}
		uploadedFiles={chat.uploadedFiles}
		isLoading={chat.isLoading}
		agent={chat.agent}
		onCancel={chat.cancelChat}
		client={chat.client}
	/>
{/key}
