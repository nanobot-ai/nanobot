import { useChat } from "~/hooks/use-chat";
import { useState } from "react";
import { ChatHeader } from "~/components/chat-header";
import { MultimodalInput } from "./multimodal-input";
import { Messages } from "./messages";
import { toast } from "./toast";
import { ChatSDKError } from "~/lib/errors";
import { type Attachment, type ChatData, getWidgets } from "~/lib/nanobot";
import CloneChat from "~/components/clone-chat";
import { Widgets } from "~/components/messages-widgets";

export default function Chat({
  chat,
  disableHeader,
}: {
  chat: ChatData;
  disableHeader?: boolean;
}) {
  const {
    messages,
    setMessages,
    updateMessage,
    handleSubmit,
    input,
    setInput,
    status,
    stop,
    reload,
    currentAgent,
    setCurrentAgent,
    agents,
    votes,
  } = useChat({
    chat,
    onError: (error) => {
      if (error instanceof ChatSDKError) {
        toast({
          type: "error",
          description: error.message,
        });
      }
    },
  });

  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const widgets = getWidgets(messages);

  return (
    <>
      <div className="flex flex-col min-w-0 h-dvh bg-background">
        {!disableHeader && (
          <ChatHeader
            currentAgent={currentAgent}
            setCurrentAgent={setCurrentAgent}
            agents={agents}
            customAgent={chat.customAgent}
            isReadonly={!!chat.readonly}
          />
        )}

        <div className="flex overflow-hidden h-full">
          <div className="flex flex-col flex-1 h-full">
            <Messages
              chatId={chat.id}
              status={status}
              votes={votes}
              messages={messages}
              updateMessage={updateMessage}
              reload={reload}
              isReadonly={!!chat.readonly}
            />

            <form className="flex mx-auto px-4 bg-background pb-4 md:pb-6 gap-2 w-full md:max-w-3xl">
              {!chat.readonly && (
                <MultimodalInput
                  chatId={chat.id}
                  input={input}
                  setInput={setInput}
                  handleSubmit={handleSubmit}
                  status={status}
                  stop={stop}
                  attachments={attachments}
                  setAttachments={setAttachments}
                  messages={messages}
                  setMessages={setMessages}
                />
              )}
              {!!chat.readonly && <CloneChat chatId={chat.id} />}
            </form>
          </div>
          {widgets.length ? (
            <div className="h-full w-1/2 p-2">
              <Widgets widgets={widgets} />
            </div>
          ) : null}
        </div>
      </div>
    </>
  );
}
