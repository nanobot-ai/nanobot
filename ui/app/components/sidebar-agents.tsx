import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "~/components/ui/sidebar";
import { Form, Link, useNavigate } from "react-router";
import { Button } from "~/components/ui/button";
import {
  CheckCircleFillIcon,
  GlobeIcon,
  LockIcon,
  MoreHorizontalIcon,
  PlusIcon,
  ShareIcon,
  TrashIcon,
} from "~/components/icons";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import type { CustomAgent } from "~/lib/nanobot";
import { PencilIcon } from "lucide-react";
import type { User } from "~/services/auth.server";

function SidebarAgent({
  chatId,
  agent,
  user,
  isActive = false,
}: {
  chatId?: string;
  agent: CustomAgent;
  user?: User;
  isActive: boolean;
}) {
  const { setOpenMobile } = useSidebar();
  const navigate = useNavigate();

  function setVisibilityType(type: "private" | "public") {}

  return (
    <SidebarMenuItem>
      <SidebarMenuButton asChild isActive={isActive} key={agent.id}>
        <Link to={`/agent/${agent.id}`} onClick={() => setOpenMobile(false)}>
          <span>{agent.name}</span>
        </Link>
      </SidebarMenuButton>
      <DropdownMenu modal={true}>
        <DropdownMenuTrigger asChild>
          <SidebarMenuAction
            className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground mr-0.5"
            showOnHover={!isActive}
          >
            <MoreHorizontalIcon />
            <span className="sr-only">More</span>
          </SidebarMenuAction>
        </DropdownMenuTrigger>

        <DropdownMenuContent side="bottom" align="end">
          {user && (
            <DropdownMenuSub>
              <DropdownMenuSubTrigger className="cursor-pointer">
                <div className="flex flex-row gap-2 items-center">
                  <ShareIcon />
                  <span>Share</span>
                </div>
              </DropdownMenuSubTrigger>
              <DropdownMenuPortal>
                <DropdownMenuSubContent>
                  <DropdownMenuItem
                    className="cursor-pointer flex-row justify-between"
                    onClick={() => {
                      setVisibilityType("private");
                    }}
                  >
                    <div className="flex flex-row gap-2 items-center">
                      <LockIcon size={12} />
                      <span>Private</span>
                    </div>
                    {!agent.isPublic ? <CheckCircleFillIcon /> : null}
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className="cursor-pointer flex-row justify-between"
                    onClick={() => {
                      setVisibilityType("public");
                    }}
                  >
                    <div className="flex flex-row gap-2 items-center">
                      <GlobeIcon />
                      <span>Public</span>
                    </div>
                    {agent.isPublic ? <CheckCircleFillIcon /> : null}
                  </DropdownMenuItem>
                </DropdownMenuSubContent>
              </DropdownMenuPortal>
            </DropdownMenuSub>
          )}

          <DropdownMenuItem
            className="cursor-pointer"
            onSelect={() => navigate(`/agent/${agent.id}/edit`)}
          >
            <PencilIcon />
            <span>Edit</span>
          </DropdownMenuItem>
          {chatId && (
            <DropdownMenuItem
              className="cursor-pointer text-destructive focus:bg-destructive/15 focus:text-destructive dark:text-red-500"
              asChild
            >
              <Form method="DELETE" action={`/chat/${chatId}/delete-agent`}>
                <input type="hidden" name="agentId" value={agent.id} />
                <button
                  type="submit"
                  className="flex items-center gap-2 w-full text-left"
                >
                  <TrashIcon />
                  <span>Delete</span>
                </button>
              </Form>
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  );
}

export function SidebarAgents({
  chatId,
  customAgents,
  customAgent,
  user,
}: {
  chatId?: string;
  customAgents?: CustomAgent[];
  customAgent?: CustomAgent;
  user?: User;
}) {
  return (
    <SidebarGroup>
      <SidebarGroupLabel className="flex items-center justify-between">
        <span>Agents</span>
      </SidebarGroupLabel>
      <Form action={"/agent/new"} method="post" className="contents">
        <SidebarMenuAction
          type="submit"
          className="mr-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200"
        >
          <PlusIcon />
        </SidebarMenuAction>
      </Form>
      <SidebarGroupContent>
        <SidebarMenu>
          {customAgents?.map((agent) => (
            <SidebarAgent
              chatId={chatId}
              key={agent.id}
              agent={agent}
              user={user}
              isActive={agent.id === customAgent?.id}
            />
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
