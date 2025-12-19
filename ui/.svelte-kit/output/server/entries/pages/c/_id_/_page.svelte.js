import { p as push, a as pop, h as head, e as escape_html } from "../../../../chunks/index.js";
import { C as ChatService } from "../../../../chunks/chat.svelte.js";
import { p as page } from "../../../../chunks/index2.js";
import { M as MessageItemUI, o as onDestroy, T as Thread } from "../../../../chunks/Thread.js";
import { g as getNotificationContext } from "../../../../chunks/mcpclient.js";
import "clsx";
import { isUIResource } from "@mcp-ui/client";
function Workspace($$payload, $$props) {
  push();
  let { messages, onSendMessage } = $$props;
  let sidecar = (() => {
    for (const message of messages.toReversed()) {
      for (const item of (message.items ?? []).toReversed()) {
        if (item.type === "tool" && item.output && item.output?.content) {
          for (const output of item.output.content.toReversed()) {
            if (isUIResource(output) && output.resource._meta?.["ai.nanobot.meta/workspace"]) {
              return output;
            }
          }
        }
      }
    }
    return null;
  })();
  let key = sidecar?.resource?.text ?? sidecar?.resource?.blob ?? "";
  if (key && sidecar) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<!---->`);
    {
      $$payload.out.push(`<div class="workspace peer m-3 h-[60vh] border-2 border-base-100/30 md:m-0 md:h-dvh md:max-h-dvh md:w-3/4">`);
      MessageItemUI($$payload, { item: sidecar, onSend: onSendMessage });
      $$payload.out.push(`<!----></div>`);
    }
    $$payload.out.push(`<!---->`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function _page($$payload, $$props) {
  push();
  const chat = page.data.chat || new ChatService();
  getNotificationContext();
  onDestroy(() => {
    chat.close();
  });
  head($$payload, ($$payload2) => {
    if (chat.agent?.name) {
      $$payload2.out.push("<!--[-->");
      $$payload2.title = `<title>${escape_html(chat.agent.name)}</title>`;
    } else {
      $$payload2.out.push("<!--[!-->");
      $$payload2.title = `<title>Nanobot</title>`;
    }
    $$payload2.out.push(`<!--]-->`);
  });
  $$payload.out.push(`<div class="grid grid-cols-1 md:flex md:flex-row">`);
  Workspace($$payload, { messages: chat.messages, onSendMessage: chat.sendMessage });
  $$payload.out.push(`<!----> `);
  Thread($$payload, {
    messages: chat.messages,
    isLoading: chat.isLoading,
    onFileUpload: chat.uploadFile,
    onSendMessage: chat.sendMessage,
    cancelUpload: chat.cancelUpload,
    prompts: chat.prompts,
    resources: chat.resources,
    elicitations: chat.elicitations,
    agent: chat.agent,
    uploadingFiles: chat.uploadingFiles,
    uploadedFiles: chat.uploadedFiles,
    onElicitationResult: chat.replyToElicitation
  });
  $$payload.out.push(`<!----></div>`);
  pop();
}
export {
  _page as default
};
