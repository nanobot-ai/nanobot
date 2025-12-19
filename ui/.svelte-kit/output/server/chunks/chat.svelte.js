import { p as push, b as ensure_array_like, j as spread_attributes, k as clsx, l as element, a as pop, m as spread_props } from "./index.js";
import "clsx";
import { a as SimpleClient, g as getNotificationContext, b as SvelteDate } from "./mcpclient.js";
/**
 * @license @lucide/svelte v0.540.0 - ISC
 *
 * ISC License
 * 
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 * 
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 * 
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 * 
 * ---
 * 
 * The MIT License (MIT) (for portions derived from Feather)
 * 
 * Copyright (c) 2013-2023 Cole Bemis
 * 
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 * 
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 * 
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 * 
 */
const defaultAttributes = {
  xmlns: "http://www.w3.org/2000/svg",
  width: 24,
  height: 24,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  "stroke-width": 2,
  "stroke-linecap": "round",
  "stroke-linejoin": "round"
};
function Icon($$payload, $$props) {
  push();
  const {
    name,
    color = "currentColor",
    size = 24,
    strokeWidth = 2,
    absoluteStrokeWidth = false,
    iconNode = [],
    children,
    $$slots,
    $$events,
    ...props
  } = $$props;
  const each_array = ensure_array_like(iconNode);
  $$payload.out.push(`<svg${spread_attributes(
    {
      ...defaultAttributes,
      ...props,
      width: size,
      height: size,
      stroke: color,
      "stroke-width": absoluteStrokeWidth ? Number(strokeWidth) * 24 / Number(size) : strokeWidth,
      class: clsx(["lucide-icon lucide", name && `lucide-${name}`, props.class])
    },
    null,
    void 0,
    void 0,
    3
  )}><!--[-->`);
  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
    let [tag, attrs] = each_array[$$index];
    element($$payload, tag, () => {
      $$payload.out.push(`${spread_attributes({ ...attrs }, null, void 0, void 0, 3)}`);
    });
  }
  $$payload.out.push(`<!--]-->`);
  children?.($$payload);
  $$payload.out.push(`<!----></svg>`);
  pop();
}
function Copy($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    [
      "rect",
      {
        "width": "14",
        "height": "14",
        "x": "8",
        "y": "8",
        "rx": "2",
        "ry": "2"
      }
    ],
    [
      "path",
      {
        "d": "M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"
      }
    ]
  ];
  Icon($$payload, spread_props([
    { name: "copy" },
    /**
     * @component @name Copy
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cmVjdCB3aWR0aD0iMTQiIGhlaWdodD0iMTQiIHg9IjgiIHk9IjgiIHJ4PSIyIiByeT0iMiIgLz4KICA8cGF0aCBkPSJNNCAxNmMtMS4xIDAtMi0uOS0yLTJWNGMwLTEuMS45LTIgMi0yaDEwYzEuMSAwIDIgLjkgMiAyIiAvPgo8L3N2Zz4K) - https://lucide.dev/icons/copy
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
function Triangle_alert($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    [
      "path",
      {
        "d": "m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"
      }
    ],
    ["path", { "d": "M12 9v4" }],
    ["path", { "d": "M12 17h.01" }]
  ];
  Icon($$payload, spread_props([
    { name: "triangle-alert" },
    /**
     * @component @name TriangleAlert
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cGF0aCBkPSJtMjEuNzMgMTgtOC0xNGEyIDIgMCAwIDAtMy40OCAwbC04IDE0QTIgMiAwIDAgMCA0IDIxaDE2YTIgMiAwIDAgMCAxLjczLTMiIC8+CiAgPHBhdGggZD0iTTEyIDl2NCIgLz4KICA8cGF0aCBkPSJNMTIgMTdoLjAxIiAvPgo8L3N2Zz4K) - https://lucide.dev/icons/triangle-alert
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
function X($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    ["path", { "d": "M18 6 6 18" }],
    ["path", { "d": "m6 6 12 12" }]
  ];
  Icon($$payload, spread_props([
    { name: "x" },
    /**
     * @component @name X
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cGF0aCBkPSJNMTggNiA2IDE4IiAvPgogIDxwYXRoIGQ9Im02IDYgMTIgMTIiIC8+Cjwvc3ZnPgo=) - https://lucide.dev/icons/x
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
class ChatAPI {
  baseUrl;
  mcpClient;
  constructor(baseUrl = "", opts) {
    this.baseUrl = baseUrl;
    this.mcpClient = new SimpleClient({ baseUrl, fetcher: opts?.fetcher, sessionId: opts?.sessionId });
  }
  #getClient(sessionId) {
    if (sessionId) {
      return new SimpleClient({ baseUrl: this.baseUrl, sessionId });
    }
    return this.mcpClient;
  }
  async reply(id, result, opts) {
    const client = this.#getClient(opts?.sessionId);
    await client.reply(id, result);
  }
  async exchange(method, params, opts) {
    const client = this.#getClient(opts?.sessionId);
    return await client.exchange(method, params);
  }
  async callMCPTool(name, opts) {
    const client = this.#getClient(opts?.sessionId);
    try {
      const result = await client.exchange(
        "tools/call",
        {
          name,
          arguments: opts?.payload || {},
          ...opts?.async && {
            _meta: { "ai.nanobot.async": true, progressToken: opts?.progressToken }
          }
        },
        { abort: opts?.abort }
      );
      if (opts?.parseResponse) {
        return opts.parseResponse(result);
      }
      if (result && typeof result === "object" && "structuredContent" in result) {
        return result.structuredContent;
      }
      return result;
    } catch (error) {
      try {
        const notifications = getNotificationContext();
        const message = error instanceof Error ? error.message : String(error);
        notifications.error("API Error", message);
      } catch {
        console.error("MCP Tool Error:", error);
      }
      throw error;
    }
  }
  async capabilities() {
    const client = this.#getClient();
    const { initializeResult } = await client.getSessionDetails();
    return initializeResult?.capabilities?.experimental?.["ai.nanobot"]?.session ?? {};
  }
  async deleteThread(threadId) {
    const client = this.#getClient(threadId);
    return client.deleteSession();
  }
  async renameThread(threadId, title) {
    return await this.callMCPTool("update_chat", { payload: { chatId: threadId, title } });
  }
  async listAgents(opts) {
    return await this.callMCPTool("list_agents", opts);
  }
  async getThreads() {
    return (await this.callMCPTool("list_chats")).chats;
  }
  async createThread() {
    const client = this.#getClient("new");
    const { id } = await client.getSessionDetails();
    return {
      id,
      title: "New Chat",
      created: new SvelteDate().toISOString()
    };
  }
  async createResource(name, mimeType, blob, opts) {
    return await this.callMCPTool("create_resource", {
      payload: {
        blob,
        mimeType,
        name,
        ...opts?.description && { description: opts.description }
      },
      sessionId: opts?.sessionId,
      abort: opts?.abort,
      parseResponse: (resp) => {
        if (resp.content?.[0]?.type === "resource_link") {
          return { uri: resp.content[0].uri };
        }
        return { uri: "" };
      }
    });
  }
  async sendMessage(request) {
    await this.callMCPTool("chat", {
      payload: {
        prompt: request.message,
        attachments: request.attachments?.map((a) => {
          return { name: a.name, url: a.uri, mimeType: a.mimeType };
        })
      },
      sessionId: request.threadId,
      progressToken: request.id,
      async: true
    });
    const message = {
      id: request.id,
      role: "user",
      created: now(),
      items: [
        { id: request.id + "_0", type: "text", text: request.message }
      ]
    };
    return { message };
  }
  subscribe(threadId, onEvent, opts) {
    console.log("Subscribing to thread:", threadId);
    const eventSource = new EventSource(`${this.baseUrl}/api/events/${threadId}`);
    eventSource.onmessage = (e) => {
      const data = JSON.parse(e.data);
      onEvent({ type: "message", message: data });
    };
    for (const type of opts?.events ?? []) {
      eventSource.addEventListener(type, (e) => {
        const idInt = parseInt(e.lastEventId);
        onEvent({ id: idInt || e.lastEventId, type, data: JSON.parse(e.data) });
      });
    }
    eventSource.onerror = (e) => {
      onEvent({ type: "error", error: String(e) });
      console.error("EventSource failed:", e);
      eventSource.close();
    };
    eventSource.onopen = () => {
      console.log("EventSource connected for thread:", threadId);
    };
    return () => eventSource.close();
  }
}
function appendMessage(messages, newMessage) {
  let found = false;
  if (newMessage.id) {
    messages = messages.map((oldMessage) => {
      if (oldMessage.id === newMessage.id) {
        found = true;
        return newMessage;
      }
      return oldMessage;
    });
  }
  if (!found) {
    messages = [...messages, newMessage];
  }
  return messages;
}
const defaultChatApi = new ChatAPI();
class ChatService {
  messages;
  prompts;
  resources;
  agent;
  elicitations;
  isLoading;
  chatId;
  uploadedFiles;
  uploadingFiles;
  api;
  closer = () => {
  };
  history;
  onChatDone = [];
  constructor(opts) {
    this.api = opts?.api || defaultChatApi;
    this.messages = [];
    this.history = void 0;
    this.isLoading = false;
    this.elicitations = [];
    this.prompts = [];
    this.resources = [];
    this.chatId = "";
    this.agent = {};
    this.uploadedFiles = [];
    this.uploadingFiles = [];
    this.setChatId(opts?.chatId);
  }
  close = () => {
    this.closer();
    this.setChatId("");
  };
  setChatId = async (chatId) => {
    if (chatId === this.chatId) {
      return;
    }
    this.messages = [];
    this.prompts = [];
    this.resources = [];
    this.elicitations = [];
    this.history = void 0;
    this.isLoading = false;
    this.uploadedFiles = [];
    this.uploadingFiles = [];
    if (chatId) {
      this.chatId = chatId;
      this.subscribe(chatId);
    }
    this.listResources().then((r) => {
      if (r && r.resources) {
        this.resources = r.resources;
      }
    });
    this.listPrompts().then((prompts) => {
      if (prompts && prompts.prompts) {
        this.prompts = prompts.prompts;
      }
    });
    await this.reloadAgent();
  };
  reloadAgent = async () => {
    const agents = await this.api.listAgents({ sessionId: this.chatId });
    if (agents.agents?.length > 0) {
      this.agent = agents.agents[0];
    }
  };
  listPrompts = async () => {
    return await this.api.exchange("prompts/list", {}, { sessionId: this.chatId });
  };
  listResources = async () => {
    return await this.api.exchange("resources/list", {}, { sessionId: this.chatId });
  };
  subscribe(chatId) {
    this.closer();
    if (!chatId) {
      return;
    }
    this.closer = this.api.subscribe(
      chatId,
      (event) => {
        if (event.type == "message" && event.message?.id) {
          if (this.history) {
            this.history = appendMessage(this.history, event.message);
          } else {
            this.messages = appendMessage(this.messages, event.message);
          }
        } else if (event.type == "history-start") {
          this.history = [];
        } else if (event.type == "history-end") {
          this.messages = this.history || [];
          this.history = void 0;
        } else if (event.type == "chat-in-progress") {
          this.isLoading = true;
        } else if (event.type == "chat-done") {
          this.isLoading = false;
          for (const waiting of this.onChatDone) {
            waiting();
          }
          this.onChatDone = [];
        } else if (event.type == "elicitation/create") {
          this.elicitations = [...this.elicitations, { id: event.id, ...event.data }];
        }
        console.debug("Received event:", event);
      },
      {
        events: [
          "history-start",
          "history-end",
          "chat-in-progress",
          "chat-done",
          "elicitation/create"
        ]
      }
    );
  }
  replyToElicitation = async (elicitation, result) => {
    await this.api.reply(elicitation.id, result, { sessionId: this.chatId });
    this.elicitations = this.elicitations.filter((e) => e.id !== elicitation.id);
  };
  newChat = async () => {
    const thread = await this.api.createThread();
    await this.setChatId(thread.id);
  };
  sendMessage = async (message, attachments) => {
    if (!message.trim() || this.isLoading) return;
    this.isLoading = true;
    if (!this.chatId) {
      await this.newChat();
    }
    try {
      const response = await this.api.sendMessage({
        id: crypto.randomUUID(),
        threadId: this.chatId,
        message,
        attachments: [...this.uploadedFiles, ...attachments || []]
      });
      this.uploadedFiles = [];
      this.messages = appendMessage(this.messages, response.message);
      return new Promise((resolve) => {
        this.onChatDone.push(() => {
          this.isLoading = false;
          const i = this.messages.findIndex((m) => m.id === response.message.id);
          if (i !== -1 && i <= this.messages.length) {
            resolve({ message: this.messages[i + 1] });
          } else {
            resolve();
          }
        });
      });
    } catch (error) {
      this.isLoading = false;
      this.messages = appendMessage(this.messages, {
        id: crypto.randomUUID(),
        role: "assistant",
        created: now(),
        items: [
          {
            id: crypto.randomUUID(),
            type: "text",
            text: `Sorry, I couldn't send your message. Please try again. Error: ${error}`
          }
        ]
      });
    }
  };
  cancelUpload = (fileId) => {
    this.uploadingFiles = this.uploadingFiles.filter((f) => {
      if (f.id !== fileId) {
        return true;
      }
      if (f.controller) {
        f.controller.abort();
      }
      return false;
    });
    this.uploadedFiles = this.uploadedFiles.filter((f) => f.id !== fileId);
  };
  uploadFile = async (file, opts) => {
    if (!this.chatId) {
      const thread = await this.api.createThread();
      await this.setChatId(thread.id);
    }
    const fileId = crypto.randomUUID();
    const controller = opts?.controller || new AbortController();
    this.uploadingFiles.push({ file, id: fileId, controller });
    try {
      const result = await this.doUploadFile(file, controller);
      this.uploadedFiles.push({ file, uri: result.uri, id: fileId, mimeType: result.mimeType });
      return result;
    } finally {
      this.uploadingFiles = this.uploadingFiles.filter((f) => f.id !== fileId);
    }
  };
  doUploadFile = async (file, controller) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    await new Promise((resolve, reject) => {
      reader.onloadend = resolve;
      reader.onerror = reject;
    });
    const base64 = reader.result.split(",")[1];
    if (!this.chatId) {
      throw new Error("Chat ID not set");
    }
    return await this.api.createResource(file.name, file.type, base64, {
      description: file.name,
      sessionId: this.chatId,
      abort: controller
    });
  };
}
function now() {
  return /* @__PURE__ */ (/* @__PURE__ */ new Date()).toISOString();
}
export {
  ChatService as C,
  Icon as I,
  Triangle_alert as T,
  X,
  Copy as a
};
