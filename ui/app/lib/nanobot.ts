import { uuidv4 } from "./utils";

export type CustomAgentsData = {
  customAgents: CustomAgentMeta[];
};

export type ChatsData = {
  chats: ChatMeta[];
};

export type ChatData = {
  id: string;
  messages?: Message[];
  currentAgent?: string;
  tools?: {
    [key: string]: ToolMeta;
  };
  agents?: {
    [key: string]: CustomAgentMeta;
  };
  agentEditor?: boolean;
  customAgent?: CustomAgent;
  votes?: Vote[];
  readonly?: boolean;
  visibility?: "public" | "private";
  capabilities?: {
    customAgents?: boolean;
  };
};

export interface CustomAgent extends CustomAgentMeta {
  remoteUrl?: string;
  icons?: {
    light?: string;
    dark?: string;
  };
  baseAgent?: string;
  instructions?: string;
  introductionMessage?: string;
  starterMessages?: string[];
  knowledgeResources?: string[];
  mcpServers?: CustomAgentMCPServer[];
}

export interface CustomAgentMCPServer {
  url: string;
  enabledTools?: string[];
  enabledPrompts?: string[];
}

export interface ChatMeta {
  id: string;
  title?: string;
  created: string;
  visibility: "public" | "private";
  readonly?: boolean;
}

export type ToolMeta = {
  mcpServer?: string;
  targetName?: string;
  target?: {
    name: string;
    description?: string;
    inputSchema?: unknown;
  };
};

export type CompletionProgress = {
  id?: string;
  agent?: string;
  model?: string;
  messageID?: string;
  item?: CompletionItem;
};

export type CustomAgentMeta = {
  id: string;
  name?: string;
  description?: string;
  isPublic?: boolean;
};

export function appendProgress(
  messages: Message[],
  progress: CompletionProgress,
) {
  if (!progress.messageID || !progress.item) {
    return;
  }

  let message = messages.find((m) => m.id === progress.messageID);
  if (!message) {
    message = {
      id: progress.messageID,
      role: "assistant",
      created: new Date().toISOString(),
      items: progress.item ? [progress.item] : [],
    };
    messages.push(message);
    return;
  }

  if (!message.items) {
    message.items = [];
  }

  const itemIndex = message.items?.findIndex((x) => x.id === progress.item?.id);
  if (itemIndex === undefined || itemIndex === -1) {
    message.items.push(progress.item);
    return;
  }

  const item = message.items[itemIndex];

  if (!item.partial) {
    // Already completed, no need to update
    item.partials = undefined;
    return;
  }

  if (!progress.item.partial) {
    message.items[itemIndex] = progress.item;
    return;
  }

  if (progress.id) {
    if (item.partials?.has(progress.id)) {
      // Already processed this partial update
      return;
    }
    if (!item.partials) {
      item.partials = new Set<string>();
    }
    item.partials?.add(progress.id);
  }

  message.revision = (message.revision || 0) + 1;

  item.hasMore = progress.item.hasMore;
  if (!item.hasMore) {
    item.partial = false;
  }

  if (item.type === "tool" && progress.item.type === "tool") {
    if (progress.item.arguments) {
      item.arguments = (item.arguments || "") + progress.item.arguments;
    }
    if (progress.item.output) {
      item.output = progress.item.output;
    }
  } else if (item.type === "text" && progress.item.type === "text") {
    item.text = (item.text || "") + (progress.item.text || "");
  }
}

export type CompletionItem = {
  id?: string;
  hasMore?: boolean;
  partial?: boolean;
  partials?: Set<string>;
} & (Content | ToolCall | Reasoning);

export interface Reasoning {
  type: "reasoning";
  summary?: [
    {
      text: string;
    },
  ];
}

export type ToolCall = {
  type: "tool";
  name: string;
  arguments?: string;
  callID: string;
  target?: string;
  targetType?: string;
  output?: CallResult;
};

export type CallResult = {
  content?: Content[];
  isError?: boolean;
  agent: string;
  model: string;
};

export type Vote = {
  messageId: string;
  isUpvoted?: boolean;
};

export type Message = {
  id?: string;
  role?: string;
  created?: string;
  items?: CompletionItem[];
  revision?: number;
};

export type Content = {
  type: "text" | "image" | "audio" | "resource" | "";
  text?: string;
  structuredContent?: unknown;
  data?: string;
  mimeType?: string;
  resource?: {
    uri: string;
    mimeType?: string;
    blob?: number;
    text?: number;
  };
};

export function events(
  id: string,
  cb: (event: CompletionProgress) => void,
): () => void {
  const es = new EventSource(`/mcp/session/${id}/events`);
  es.onmessage = (e) => {
    const event = JSON.parse(e.data);
    const progress =
      event?.params?.["_meta"]?.["ai.nanobot.progress/completion"];
    if (progress) {
      progress.id = event.params.progressToken + "-" + event.params.progress;
      cb(progress as CompletionProgress);
    }
  };
  return () => {
    if (es.readyState === EventSource.CLOSED) {
      return;
    }
    es.close();
  };
}

export async function deleteChat(ctx: Context, id: string): Promise<void> {
  let url = "/mcp/" + id + "/chats/delete";
  if (typeof process !== "undefined") {
    if (process.env.NANOBOT_URL) {
      url = process.env.NANOBOT_URL + url;
    } else {
      url = `http://localhost:9999${url}`;
    }
  }

  const resp = await fetch(url, {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
      "X-Nanobot-Session-Id": id,
    },
  });

  if (!resp.ok) {
    throw new Error(`Failed to delete chat ${id}: ${resp.statusText}`);
  }
}

async function call(
  ctx: Context,
  id: string,
  tool: string,
  body?: object,
  opts?: {
    agentId?: string;
  },
): Promise<any> {
  let url =
    "/mcp/" + (opts?.agentId ? `agents/${opts.agentId}/` : "") + id + "/tools";

  if (typeof process !== "undefined") {
    if (process.env.NANOBOT_URL) {
      url = process.env.NANOBOT_URL + url;
    } else {
      url = `http://localhost:9999${url}`;
    }
  }

  const resp = await fetch(`${url}/${tool}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Nanobot-Session-Id": id,
    },
    body: JSON.stringify({
      jsonrpc: "2.0",
      method: "tools/call",
      params: {
        name: tool,
        arguments: body,
        _meta: {
          progressToken: uuidv4(),
        },
      },
    }),
  });

  const rpcResp = await resp.json();
  if (rpcResp.error) {
    throw new Error(
      `Error calling tool ${tool}: ${JSON.stringify(rpcResp.error)}`,
    );
  }
  const result = rpcResp.result;
  if (result.isError) {
    throw new Error(
      `Error calling tool ${tool}: ${JSON.stringify(result.content)}`,
    );
  }
  for (const content of result.content || []) {
    if (content.structuredContent) {
      return content.structuredContent;
    }
  }
  return null;
}

type Context = object;

export function getContext(request: Request): Context {
  // This function can be extended to extract more context from the request if needed
  return {};
}

export async function setAgent(ctx: Context, id: string, agent: string) {
  await call(ctx, id, "set_current_agent", { agent });
}

export interface Attachment {
  name?: string;
  description?: string;
  uri?: string;
  mimeType?: string;
}

export async function setVisibility(
  ctx: Context,
  id: string,
  visibility: "public" | "private",
) {
  await call(ctx, id, "set_visibility", { visibility });
}

export async function clone(ctx: Context, id: string): Promise<string> {
  return (await call(ctx, id, "clone", {})) as string;
}

export async function newCustomAgent(ctx: Context): Promise<CustomAgent> {
  return (await call(ctx, "new", "create_custom_agent", {})) as CustomAgent;
}

export async function updateCustomAgent(
  ctx: Context,
  id: string,
  agent: CustomAgent,
) {
  return (await call(ctx, id, "update_custom_agent", agent, {
    agentId: agent.id,
  })) as CustomAgent;
}

export async function chat(
  ctx: Context,
  id: string,
  prompt: string,
  opts?: {
    attachments?: Attachment[];
  },
) {
  await call(ctx, id, "chat", { prompt, attachments: opts?.attachments });
}

export async function getChat(
  ctx: Context,
  id: string,
  opts: {
    agentId?: string;
  },
): Promise<ChatData> {
  const result = await call(ctx, id, "get_chat", undefined, opts);
  if ("ai.nanobot/ext" in result) {
    const copy = { ...result, ...result["ai.nanobot/ext"] };
    delete result["ai.nanobot/ext"];
    return copy as ChatData;
  }
  return result as ChatData;
}

export async function listChats(ctx: Context, id: string): Promise<ChatsData> {
  return (await call(ctx, id, "list_chats")) as ChatsData;
}

export async function listCustomAgents(
  ctx: Context,
  id: string,
): Promise<CustomAgentsData> {
  return (await call(ctx, id, "list_custom_agents")) as CustomAgentsData;
}

export interface Resource {
  uri: string;
  mimeType?: string;
  name?: string;
  description?: string;
}

export async function createResource(
  ctx: Context,
  id: string,
  blob: string,
  opts?: {
    name?: string;
    description?: string;
    mimeType?: string;
  },
) {
  return (await call(ctx, id, "create_resource", {
    blob,
    ...(opts ?? {}),
  })) as Resource;
}

export async function deleteCustomAgent(
  ctx: Context,
  id: string,
  agentId: string,
) {
  await call(
    ctx,
    id,
    "delete_custom_agent",
    { id: agentId },
    {
      agentId,
    },
  );
}
