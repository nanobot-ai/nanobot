import { AgentEditor } from "~/components/agent-editor";
import { redirect, useRouteLoaderData } from "react-router";
import Chat from "~/components/chat";
import type { ActionFunctionArgs } from "react-router";
import { type CustomAgent, getContext, updateCustomAgent } from "~/lib/nanobot";
import type { Route as LayoutRoute } from "./+types/layout";

export async function action({ request, params }: ActionFunctionArgs) {
  const formData = await request.formData();

  const servers: string[] = [];
  for (const [key, value] of formData.entries()) {
    if (key.startsWith("serverUrls[") && key.endsWith("]") && value) {
      servers.push(value as string);
    }
  }

  const customAgent: CustomAgent = {
    id: params.agentId || "",
    remoteUrl: formData.get("remoteUrl") as string,
    name: formData.get("name") as string,
    description: formData.get("description") as string,
    icons: {
      light: formData.get("icon") as string,
      dark: formData.get("iconDark") as string,
    },
    isPublic: formData.get("isPublic") === "true",
    instructions: formData.get("instructions") as string,
    introductionMessage: formData.get("introductionMessage") as string,
    baseAgent: formData.get("model") as string,
    mcpServers: servers
      .map((x) => {
        return {
          url: x,
        };
      })
      .filter((x) => x.url),
  };

  if (request.method === "POST" && params.agentId) {
    return redirect(`/agent/${params.agentId}`);
  }

  return await updateCustomAgent(
    getContext(request),
    params.id || "new",
    customAgent,
  );
}

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
