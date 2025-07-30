import { AgentEditor } from "~/components/agent-editor";
import { useRouteLoaderData } from "react-router";
import type { Route as LayoutRoute } from "./+types/layout";
import Chat from "~/components/chat";

export default function Page() {
  const { chat, user } =
    useRouteLoaderData<LayoutRoute.ComponentProps["loaderData"]>(
      "routes/layout",
    );
  chat.agentEditor = true;
  return (
    <div className="grid grid-cols-2">
      <div className="col-span-1 p-4 grid">
        <AgentEditor chat={chat} customAgent={chat.customAgent || {}} />
      </div>
      <div className="col-span-1">
        <Chat chat={chat} disableHeader user={user} />
      </div>
    </div>
  );
}
