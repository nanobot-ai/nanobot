import { type Dispatch, type SetStateAction } from "react";

export type ChatVisibilityType = "public" | "private";

export function useAgentVisibility({
  customAgentId,
}: {
  customAgentId: string;
}): [ChatVisibilityType, Dispatch<SetStateAction<ChatVisibilityType>>] {
  return [
    "private",
    () => {}, // Placeholder for setVisibilityType
  ];
}
