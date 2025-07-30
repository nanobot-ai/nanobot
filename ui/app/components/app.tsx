import { type ChatData } from "~/lib/nanobot";
import Chat from "~/components/chat";

export default function App({ chat }: { chat: ChatData }) {
  return (
    <>
      <Chat key={chat.id} chat={chat} />
    </>
  );
}
