import type { User } from "~/services/auth.server";

import { PlusIcon } from "~/components/icons";
import { SidebarHistory } from "~/components/sidebar-history";
import { SidebarUserNav } from "~/components/sidebar-user-nav";
import { Button } from "~/components/ui/button";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuAction,
  SidebarGroup,
  useSidebar,
} from "~/components/ui/sidebar";
import { Form, Link } from "react-router";
import { Tooltip, TooltipContent, TooltipTrigger } from "./ui/tooltip";
import type { CustomAgentMeta, ChatMeta, CustomAgent } from "~/lib/nanobot";
import { useNavigate } from "react-router";
import { useChatDelete } from "~/hooks/use-chat";
import { SidebarAgents } from "~/components/sidebar-agents";

export function AppSidebar({
  user,
  chatId,
  chats,
  customAgents,
  customAgent,
}: {
  user?: User;
  chatId?: string;
  chats: ChatMeta[];
  customAgents?: CustomAgentMeta[];
  customAgent?: CustomAgent;
}) {
  const deleteChat = useChatDelete();
  const navigate = useNavigate();
  const { setOpenMobile } = useSidebar();

  return (
    <Sidebar className="group-data-[side=left]:border-r-0">
      <SidebarHeader>
        <SidebarMenu>
          <div className="flex flex-row justify-between items-center">
            <Link
              to="/"
              onClick={() => {
                setOpenMobile(false);
              }}
              className="flex flex-row gap-3 items-center"
            >
              <span className="text-lg font-semibold px-2 hover:bg-muted rounded-md cursor-pointer">
                Nanobot
              </span>
            </Link>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  type="button"
                  className="p-2 h-fit"
                  onClick={() => {
                    setOpenMobile(false);
                    navigate("/");
                  }}
                >
                  <PlusIcon />
                </Button>
              </TooltipTrigger>
              <TooltipContent align="end">New Chat</TooltipContent>
            </Tooltip>
          </div>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        {customAgents && customAgents.length > 0 && (
          <SidebarAgents
            chatId={chatId}
            customAgents={customAgents}
            customAgent={customAgent}
            user={user}
          />
        )}
        <SidebarHistory
          chats={chats}
          chatId={chatId}
          deleteChat={deleteChat}
          user={user}
        />
        {!customAgents?.length && (
          <SidebarAgents
            chatId={chatId}
            customAgents={customAgents}
            customAgent={customAgent}
            user={user}
          />
        )}
      </SidebarContent>
      <SidebarFooter>{user && <SidebarUserNav user={user} />}</SidebarFooter>
    </Sidebar>
  );
}
