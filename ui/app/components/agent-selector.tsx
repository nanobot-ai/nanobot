import {
  type Dispatch,
  type SetStateAction,
  startTransition,
  useState,
  type ComponentProps,
} from "react";

import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { cn } from "~/lib/utils";

import { CheckCircleFillIcon, ChevronDownIcon } from "./icons";
import type { CustomAgentMeta } from "~/lib/nanobot";

export function AgentSelector({
  currentAgent,
  setCurrentAgent,
  agents,
  className,
}: {
  currentAgent: string;
  agents: Record<string, CustomAgentMeta>;
  setCurrentAgent: Dispatch<SetStateAction<string>>;
} & ComponentProps<typeof Button>) {
  const [open, setOpen] = useState(false);

  const currentAgentMeta = agents?.[currentAgent];
  if (!currentAgentMeta) {
    return null;
  }

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger
        asChild
        className={cn(
          "w-fit data-[state=open]:bg-accent data-[state=open]:text-accent-foreground",
          className,
        )}
      >
        <Button
          data-testid="model-selector"
          variant="outline"
          className="md:px-2 md:h-[34px]"
        >
          {currentAgentMeta?.name}
          <ChevronDownIcon />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-[300px]">
        {Object.entries(agents).map(([id, agent]) => {
          return (
            <DropdownMenuItem
              data-testid={`model-selector-item-${id}`}
              key={id}
              onSelect={() => {
                setOpen(false);
                startTransition(() => {
                  setCurrentAgent(id);
                });
              }}
              data-active={id === currentAgent}
              asChild
            >
              <button
                type="button"
                className="gap-4 group/item flex flex-row justify-between items-center w-full"
              >
                <div className="flex flex-col gap-1 items-start">
                  <div>{agent.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {agent.description}
                  </div>
                </div>

                <div className="text-foreground dark:text-foreground opacity-0 group-data-[active=true]/item:opacity-100">
                  <CheckCircleFillIcon />
                </div>
              </button>
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
