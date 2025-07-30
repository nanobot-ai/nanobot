import {
  type CustomAgentMeta,
  type ChatData,
  type Attachment,
  type Message,
  type Vote,
  events,
  appendProgress,
  type CompletionProgress,
} from "~/lib/nanobot";
import {
  type Dispatch,
  type SetStateAction,
  useDebugValue,
  useEffect,
  useState,
} from "react";
import { useFetcher, useNavigate } from "react-router";
import { uuidv4 } from "~/lib/utils";
import { useDebounce } from "use-debounce";
import { useDebounceCallback } from "usehooks-ts";

export type ChatVisibilityType = "public" | "private";

export interface UseChatType {
  messages: Message[];
  setMessages: Dispatch<SetStateAction<Message[]>>;
  updateMessage: (id: string, update: (prev: Message) => Message) => void;
  handleSubmit: (opt: {
    prompt?: string;
    attachments?: Attachment[];
  }) => Promise<void>;
  input: string;
  setInput: Dispatch<SetStateAction<string>>;
  status: "submitted" | "streaming" | "ready" | "error";
  stop: () => void;
  reload: () => Promise<string | null | undefined>;
  currentAgent: string;
  setCurrentAgent: Dispatch<SetStateAction<string>>;
  agents: Record<string, CustomAgentMeta>;
  visibilityType: ChatVisibilityType;
  setVisibilityType: Dispatch<SetStateAction<ChatVisibilityType>>;
  supportsCustomAgents: boolean;
  votes?: Vote[];
}

export function useChatDelete(): (chatId: string) => Promise<void> {
  const submitter = useFetcher({ key: "delete-chat" });
  return (chatId: string): Promise<void> => {
    return submitter.submit(
      {
        id: chatId,
      },
      {
        method: "delete",
        action: `/chat/${chatId}`,
      },
    );
  };
}

export function useChatVisibility({
  chatId,
}: {
  chatId: string;
}): [ChatVisibilityType, Dispatch<SetStateAction<ChatVisibilityType>>] {
  return [
    "private",
    () => {}, // Placeholder for setVisibilityType
  ];
}

export function useChat({
  chat,
  onError,
}: {
  chat: ChatData;
  onError?: (error: unknown) => void;
}): UseChatType {
  const [input, setInput] = useState<string>("");
  const [status, setStatus] = useState<
    "submitted" | "streaming" | "ready" | "error"
  >("ready");
  const [messages, setMessages] = useState<Message[]>([]);
  const submitter = useFetcher();
  const navigate = useNavigate();

  useEffect(() => {
    // Initialize messages from chat data, only when chat.id changes.
    if (chat.messages) {
      setMessages([...chat.messages]);
    } else {
      setMessages([]);
    }
  }, [chat.id]);

  useEffect(() => {
    if (chat.id) {
      return events(chat.id, (event) => {
        setMessages((prev) => {
          const messages = [...prev];
          appendProgress(messages, event);
          return messages;
        });
      });

      const eventQueue: CompletionProgress[] = [];

      const close1 = events(chat.id, (event) => {
        eventQueue.push(event);
      });

      const close2 = setInterval(function () {
        if (eventQueue.length === 0) {
          return;
        }
        setMessages((prev) => {
          const messages = [...prev];
          for (const event of eventQueue) {
            appendProgress(messages, event);
          }
          return messages;
        });
        eventQueue.length = 0; // Clear the queue after processing
      }, 100);

      return function () {
        close1();
        clearInterval(close2);
      };
    }
  }, [chat.id]);

  async function handleSubmit({
    prompt,
    attachments,
  }: {
    prompt?: string;
    attachments?: Attachment[];
  }) {
    setStatus("submitted");
    const id = uuidv4();
    const chatURL = chat.agentEditor
      ? `/chat/${chat.id}/agent/${chat.customAgent!.id}/edit`
      : `/chat/${chat.id}`;
    try {
      setInput("");
      setMessages((prev): Message[] => {
        return [
          ...prev,
          {
            id: id,
            role: "user",
            items: [
              {
                id: id + "_0",
                type: "text",
                text: prompt || input,
              },
            ],
          },
        ];
      });
      setStatus("streaming");
      await submitter.submit(
        {
          id: id,
          prompt: prompt || input,
          clone: !!chat.readonly,
        },
        {
          method: "post",
          action: chatURL,
        },
      );
      setStatus("ready");
      await navigate(chatURL);
    } catch (error) {
      console.error("Error submitting chat message:", error);
      setStatus("error");
      if (onError) {
        onError(error);
      }
    }
  }

  function getVisibilityType(): ChatVisibilityType {
    return chat.visibility || "private";
  }

  async function setVisibilityType(
    visibility: ChatVisibilityType | SetStateAction<ChatVisibilityType>,
  ) {
    if (typeof visibility === "function") {
      visibility = visibility(getVisibilityType());
    }

    await navigate(`/chat/${chat.id}`);
    await submitter.submit(
      {
        visibility: visibility,
      },
      {
        method: "post",
        action: `/chat/${chat.id}`,
      },
    );
  }

  async function setCurrentAgent(agent: string | SetStateAction<string>) {
    await navigate(`/chat/${chat.id}`);
    if (typeof agent === "function") {
      agent = agent(chat.currentAgent || "");
    }
    await submitter.submit(
      {
        agent: agent,
      },
      {
        method: "post",
        action: `/chat/${chat.id}`,
      },
    );
  }

  return {
    messages,
    setMessages,
    updateMessage: () => {}, // Placeholder for updateMessage
    handleSubmit,
    input,
    setInput,
    status,
    stop: () => {}, // Placeholder for stop
    reload: async () => null, // Placeholder for reload
    currentAgent: chat.currentAgent || "",
    setCurrentAgent,
    agents: chat.agents || {},
    visibilityType: getVisibilityType(),
    setVisibilityType,
    votes: chat.votes,
    supportsCustomAgents: chat.supportsCustomAgents || true,
  };
}
