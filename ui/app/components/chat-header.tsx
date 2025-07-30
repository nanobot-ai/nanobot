"use client";

import { useWindowSize } from "usehooks-ts";

import { AgentSelector } from "~/components/agent-selector";
import { SidebarToggle } from "~/components/sidebar-toggle";
import { Button } from "~/components/ui/button";
import { PlusIcon } from "./icons";
import { useSidebar } from "./ui/sidebar";
import { type Dispatch, type SetStateAction, memo } from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from "./ui/tooltip";
import { useNavigate } from "react-router";
import type { CustomAgent, CustomAgentMeta } from "~/lib/nanobot";

function PureChatHeader({
  currentAgent,
  setCurrentAgent,
  agents,
  customAgent,
  isReadonly,
}: {
  currentAgent: string;
  setCurrentAgent: Dispatch<SetStateAction<string>>;
  agents: Record<string, CustomAgentMeta>;
  customAgent?: CustomAgent;
  isReadonly: boolean;
}) {
  const navigate = useNavigate();
  const { open } = useSidebar();

  const { width: windowWidth } = useWindowSize();

  return (
    <header className="flex sticky top-0 bg-background py-1.5 items-center px-2 md:px-2 gap-2">
      <SidebarToggle />

      {(!open || windowWidth < 768) && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              className="order-2 md:order-1 md:px-2 px-2 md:h-fit ml-auto md:ml-0"
              onClick={() => {
                if (customAgent?.id) {
                  navigate(`/agent/${customAgent?.id}`);
                } else {
                  navigate("/");
                }
              }}
            >
              <PlusIcon />
              <span className="md:sr-only">New Chat</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>New Chat</TooltipContent>
        </Tooltip>
      )}

      {!isReadonly && !customAgent && Object.keys(agents).length > 1 && (
        <AgentSelector
          currentAgent={currentAgent}
          setCurrentAgent={setCurrentAgent}
          agents={agents}
          className="order-1 md:order-2"
        />
      )}
    </header>
  );
}

export const ChatHeader = memo(PureChatHeader, (prevProps, nextProps) => {
  if (prevProps.customAgent?.id !== nextProps.customAgent?.id) return false;
  if (
    Object.keys(prevProps.agents || {}) !== Object.keys(nextProps.agents || {})
  )
    return false;
  return prevProps.currentAgent === nextProps.currentAgent;
});
