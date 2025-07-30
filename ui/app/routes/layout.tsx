import type { Route } from "./+types/layout";
import {
  getChat,
  listChats,
  getContext,
  listCustomAgents,
} from "~/lib/nanobot";
import { SidebarInset, SidebarProvider } from "~/components/ui/sidebar";
import { sidebarCookie } from "~/cookies.server";
import { AppSidebar } from "~/components/app-sidebar";
import { Outlet } from "react-router";
import type { User } from "~/services/auth.server";

export async function loader({ request, params }: Route.LoaderArgs) {
  const ctx = getContext(request);
  const chat = await getChat(ctx, params.id || "new", {
    agentId: params.agentId,
  });
  const threads = await listChats(ctx, chat.id);
  const customAgents = await listCustomAgents(ctx, chat.id);
  const cookie: string | null = await sidebarCookie.parse(
    request.headers.get("Cookie"),
  );

  const currentChatMeta = threads.chats.find((c) => c.id === chat.id);
  if (currentChatMeta) {
    chat.visibility = currentChatMeta.visibility;
    chat.readonly = currentChatMeta.readonly;
  }

  return {
    threads,
    chat,
    customAgents,
    sidebar: cookie || "true",
    user: {
      name: "User",
      email: "user@example.com",
    } as User,
  };
}

export default function Layout({
  loaderData: { chat, threads, sidebar, customAgents, user },
}: Route.ComponentProps) {
  const isCollapsed = sidebar !== "true";
  return (
    <SidebarProvider defaultOpen={!isCollapsed}>
      <AppSidebar
        chats={threads.chats}
        chatId={chat.id}
        customAgents={customAgents.customAgents?.filter((c) => !!c.name)}
        customAgent={chat.customAgent}
        user={user}
      />
      <SidebarInset>
        <Outlet />
      </SidebarInset>
    </SidebarProvider>
  );
}
