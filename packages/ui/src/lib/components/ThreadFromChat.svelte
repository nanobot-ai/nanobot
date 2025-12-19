<script lang="ts">
    import type {ChatService} from "$lib/chat.svelte";
    import Thread from "$lib/components/Thread.svelte";

    interface Props {
        chat: ChatService
    }

    let {chat}: Props = $props();
</script>

{#key chat.chatId}
    <Thread
            messages={chat.messages}
            prompts={chat.prompts}
            resources={chat.resources}
            elicitations={chat.elicitations}
            onElicitationResult={chat.replyToElicitation}
            onSendMessage={chat.sendMessage}
            onFileUpload={chat.uploadFile}
            cancelUpload={chat.cancelUpload}
            uploadingFiles={chat.uploadingFiles}
            uploadedFiles={chat.uploadedFiles}
            isLoading={chat.isLoading}
            agent={chat.agent}
    />
{/key}
