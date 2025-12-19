import { p as push, h as head, a as pop, e as escape_html } from "../../chunks/index.js";
import { C as ChatService } from "../../chunks/chat.svelte.js";
import { o as onDestroy, T as Thread } from "../../chunks/Thread.js";
import "@sveltejs/kit/internal";
import "../../chunks/exports.js";
import "../../chunks/utils.js";
import "clsx";
import "../../chunks/state.svelte.js";
import "../../chunks/client.js";
function _page($$payload, $$props) {
  push();
  const chat = new ChatService();
  onDestroy(() => {
    {
      chat.close();
    }
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
  Thread($$payload, {
    messages: chat.messages,
    isLoading: chat.isLoading,
    onSendMessage: chat.sendMessage,
    onFileUpload: chat.uploadFile,
    cancelUpload: chat.cancelUpload,
    prompts: chat.prompts,
    resources: chat.resources,
    elicitations: chat.elicitations,
    agent: chat.agent,
    uploadingFiles: chat.uploadingFiles,
    uploadedFiles: chat.uploadedFiles,
    onElicitationResult: chat.replyToElicitation
  });
  pop();
}
export {
  _page as default
};
