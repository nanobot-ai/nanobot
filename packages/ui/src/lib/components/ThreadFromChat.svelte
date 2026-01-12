<script lang="ts">
    import type {ChatService} from "$lib/chat.svelte";
    import Thread from "$lib/components/Thread.svelte";
	import type { Agent, Attachment } from "$lib/types";

    interface Props {
        chat: ChatService
        inline?: boolean;
        files?: Attachment[];
        agent?: Agent;
    }

    let { chat, inline, files = [], agent }: Props = $props();

    function handleSendMessage(message: string, attachments?: Attachment[]) {
        return chat.sendMessage(message, [...files, ...(attachments || [])]);
    }
</script>

{#key chat.chatId}
    <Thread
        messages={chat.messages}
        prompts={chat.prompts}
        resources={chat.resources}
        elicitations={chat.elicitations}
        onElicitationResult={chat.replyToElicitation}
        onSendMessage={handleSendMessage}
        onFileUpload={chat.uploadFile}
        cancelUpload={chat.cancelUpload}
        uploadingFiles={chat.uploadingFiles}
        uploadedFiles={chat.uploadedFiles}
        isLoading={chat.isLoading}
        agent={agent ?? chat.agent}
        {inline}
    />
{/key}
