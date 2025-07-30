import type { UseChatType } from "~/hooks/use-chat";
import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { LucideWrench, Plus, Trash2 } from "lucide-react";

import { Switch } from "~/components/ui/switch";
import { useState } from "react";
import { NewCustomAgent } from "~/components/multimodal-input-customagent";

export function OptionsButton({
  status,
  chatId,
  supportsCustomAgents,
}: {
  status: UseChatType["status"];
  chatId: string;
  supportsCustomAgents?: boolean;
}) {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            className="rounded-md rounded-bl-lg p-[7px] h-fit dark:border-zinc-700 hover:dark:bg-zinc-900 hover:bg-zinc-200"
            disabled={status !== "ready"}
            variant="ghost"
          >
            <LucideWrench size={14} />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56" align="start">
          <DropdownMenuGroup>
            <DropdownMenuItem
              onSelect={(e) => e.preventDefault()}
              className="justify-between"
            >
              Gmail
              <Switch id="airplane-mode" />
            </DropdownMenuItem>
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>Advanced</DropdownMenuSubTrigger>
              <DropdownMenuPortal>
                <DropdownMenuSubContent>
                  <DropdownMenuItem>Instructions</DropdownMenuItem>
                  <DropdownMenuItem>Prompts</DropdownMenuItem>
                  <DropdownMenuSub>
                    <DropdownMenuSubTrigger>
                      Custom Tools
                    </DropdownMenuSubTrigger>
                    <DropdownMenuPortal>
                      <DropdownMenuSubContent>
                        <DropdownMenuSub>
                          <DropdownMenuSubTrigger>
                            GitHub
                          </DropdownMenuSubTrigger>
                          <DropdownMenuPortal>
                            <DropdownMenuSubContent>
                              <DropdownMenuCheckboxItem checked>
                                Clone
                              </DropdownMenuCheckboxItem>
                              <DropdownMenuItem>
                                <Trash2 />
                                Remove
                              </DropdownMenuItem>
                            </DropdownMenuSubContent>
                          </DropdownMenuPortal>
                        </DropdownMenuSub>
                        <DropdownMenuItem>
                          <Plus />
                          Add Remote Server
                        </DropdownMenuItem>
                      </DropdownMenuSubContent>
                    </DropdownMenuPortal>
                  </DropdownMenuSub>
                  {supportsCustomAgents && (
                    <>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => setIsOpen(true)}>
                        Save as New Agent
                      </DropdownMenuItem>
                    </>
                  )}
                </DropdownMenuSubContent>
              </DropdownMenuPortal>
            </DropdownMenuSub>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
      {supportsCustomAgents && (
        <NewCustomAgent
          chatId={chatId}
          open={isOpen}
          onOpenChange={setIsOpen}
        />
      )}
    </>
  );
}
