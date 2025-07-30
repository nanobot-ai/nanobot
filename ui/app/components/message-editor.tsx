import { Button } from "./ui/button";
import {
  type Dispatch,
  type SetStateAction,
  useEffect,
  useRef,
  useState,
} from "react";
import { Textarea } from "./ui/textarea";
import type { Message } from "~/lib/nanobot";
import type { UseChatType } from "~/hooks/use-chat";

export type MessageEditorProps = {
  message: Message;
  itemIndex: number;
  setMode: Dispatch<SetStateAction<"view" | "edit">>;
  setMessage: UseChatType["updateMessage"];
  reload: UseChatType["reload"];
};

export function MessageEditor({
  message,
  itemIndex,
  setMode,
  setMessage,
  reload,
}: MessageEditorProps) {
  const item = message.items?.[itemIndex];
  const initialContent =
    item && item.type === "text" && item.text ? item.text : "";

  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [draftContent, setDraftContent] = useState<string>(initialContent);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (textareaRef.current) {
      adjustHeight();
    }
  }, []);

  const adjustHeight = () => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight + 2}px`;
    }
  };

  const handleInput = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    setDraftContent(event.target.value);
    adjustHeight();
  };

  return (
    <div className="flex flex-col gap-2 w-full">
      <Textarea
        data-testid="message-editor"
        ref={textareaRef}
        className="bg-transparent outline-none overflow-hidden resize-none !text-base rounded-xl w-full"
        value={draftContent}
        onChange={handleInput}
      />

      <div className="flex flex-row gap-2 justify-end">
        <Button
          variant="outline"
          className="h-fit py-2 px-3"
          onClick={() => {
            setMode("view");
          }}
        >
          Cancel
        </Button>
        <Button
          data-testid="message-editor-send-button"
          variant="default"
          className="h-fit py-2 px-3"
          disabled={isSubmitting}
          onClick={async () => {
            setIsSubmitting(true);

            if (message.id) {
              setMessage(message.id, (prevMessage) => {
                const updatedMessage = {
                  ...prevMessage,
                };

                if (updatedMessage.items && updatedMessage.items[itemIndex]) {
                  updatedMessage.items[itemIndex] = {
                    ...item,
                    type: "text",
                    text: draftContent,
                  };

                  return updatedMessage;
                }

                return prevMessage;
              });
            }

            setMode("view");
            reload();
          }}
        >
          {isSubmitting ? "Sending..." : "Send"}
        </Button>
      </div>
    </div>
  );
}
