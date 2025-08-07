var __defProp = Object.defineProperty;
var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
var __publicField = (obj, key, value) => __defNormalProp(obj, typeof key !== "symbol" ? key + "" : key, value);
import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import { PassThrough } from "node:stream";
import { createReadableStreamFromReadable } from "@react-router/node";
import { ServerRouter, createCookieSessionStorage, UNSAFE_withComponentProps, useLoaderData, UNSAFE_withErrorBoundaryProps, isRouteErrorResponse, Meta, Links, Outlet, ScrollRestoration, Scripts, redirect, Form, createCookie, useFetcher, useNavigate, Link, useRouteLoaderData, useSubmit } from "react-router";
import { isbot } from "isbot";
import { renderToPipeableStream } from "react-dom/server";
import { Toaster as Toaster$1, toast as toast$1 } from "sonner";
import { useTheme, createThemeSessionResolver, ThemeProvider, PreventFlashOnWrongTheme, Theme } from "remix-themes";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { FormStrategy } from "remix-auth-form";
import { Authenticator } from "remix-auth";
import * as React from "react";
import { useState, useEffect, memo, startTransition, useRef, useCallback, useMemo } from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva } from "class-variance-authority";
import * as SheetPrimitive from "@radix-ui/react-dialog";
import { XIcon, ChevronRightIcon, ChevronUp, PencilIcon, ArrowDown, Hammer, ChevronDownIcon as ChevronDownIcon$1, Minimize, Maximize, CheckIcon, ChevronUpIcon, MinusIcon } from "lucide-react";
import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import { subWeeks, subMonths, isToday, isYesterday } from "date-fns";
import * as AlertDialogPrimitive from "@radix-ui/react-alert-dialog";
import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu";
import { useWindowSize, useLocalStorage, useCopyToClipboard } from "usehooks-ts";
import { motion, AnimatePresence } from "framer-motion";
import equal from "fast-deep-equal";
import useSWR, { useSWRConfig } from "swr";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { UIResourceRenderer } from "@mcp-ui/client";
import * as LabelPrimitive from "@radix-ui/react-label";
import * as SwitchPrimitive from "@radix-ui/react-switch";
import * as SelectPrimitive from "@radix-ui/react-select";
import * as TabsPrimitive from "@radix-ui/react-tabs";
const streamTimeout = 5e3;
function handleRequest(request, responseStatusCode, responseHeaders, routerContext, loadContext) {
  return new Promise((resolve, reject) => {
    let shellRendered = false;
    let userAgent = request.headers.get("user-agent");
    let readyOption = userAgent && isbot(userAgent) || routerContext.isSpaMode ? "onAllReady" : "onShellReady";
    const { pipe, abort } = renderToPipeableStream(
      /* @__PURE__ */ jsx(ServerRouter, { context: routerContext, url: request.url }),
      {
        [readyOption]() {
          shellRendered = true;
          const body = new PassThrough();
          const stream = createReadableStreamFromReadable(body);
          responseHeaders.set("Content-Type", "text/html");
          resolve(
            new Response(stream, {
              headers: responseHeaders,
              status: responseStatusCode
            })
          );
          pipe(body);
        },
        onShellError(error) {
          reject(error);
        },
        onError(error) {
          responseStatusCode = 500;
          if (shellRendered) {
            console.error(error);
          }
        }
      }
    );
    setTimeout(abort, streamTimeout + 1e3);
  });
}
const entryServer = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  default: handleRequest,
  streamTimeout
}, Symbol.toStringTag, { value: "Module" }));
const Toaster = ({ ...props }) => {
  const [theme = "system"] = useTheme();
  return /* @__PURE__ */ jsx(
    Toaster$1,
    {
      theme,
      className: "toaster group",
      style: {
        "--normal-bg": "var(--popover)",
        "--normal-text": "var(--popover-foreground)",
        "--normal-border": "var(--border)"
      },
      ...props
    }
  );
};
const themeSessionResolver = createThemeSessionResolver(
  createCookieSessionStorage({
    cookie: {
      name: "__remix-themes",
      path: "/",
      httpOnly: true,
      sameSite: "lax",
      secrets: [process.env.COOKIE_SECRET || ""],
      secure: process.env.NODE_ENV === "production"
    }
  })
);
const links = () => [{
  rel: "preconnect",
  href: "https://fonts.googleapis.com"
}, {
  rel: "preconnect",
  href: "https://fonts.gstatic.com",
  crossOrigin: "anonymous"
}, {
  rel: "stylesheet",
  href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap"
}, {
  rel: "stylesheet",
  href: "https://fonts.googleapis.com/css2?family=Manrope:wght@200..800&family=Poppins:ital,wght@0,100;0,200;0,300;0,400;0,500;0,600;0,700;0,800;0,900;1,100;1,200;1,300;1,400;1,500;1,600;1,700;1,800;1,900&display=swap"
}];
async function loader$3({
  request
}) {
  const {
    getTheme
  } = await themeSessionResolver(request);
  return {
    theme: getTheme()
  };
}
function App() {
  const data = useLoaderData();
  const [theme] = useTheme();
  return /* @__PURE__ */ jsxs("html", {
    lang: "en",
    className: theme ?? "",
    children: [/* @__PURE__ */ jsxs("head", {
      children: [/* @__PURE__ */ jsx("meta", {
        charSet: "utf-8"
      }), /* @__PURE__ */ jsx("meta", {
        name: "viewport",
        content: "width=device-width, initial-scale=1"
      }), /* @__PURE__ */ jsx(Meta, {}), /* @__PURE__ */ jsx(PreventFlashOnWrongTheme, {
        ssrTheme: Boolean(data.theme)
      }), /* @__PURE__ */ jsx(Links, {})]
    }), /* @__PURE__ */ jsxs("body", {
      className: "antialiased",
      style: {
        overflow: "hidden",
        height: "100vh"
      },
      children: [/* @__PURE__ */ jsx(Toaster, {
        position: "top-center"
      }), /* @__PURE__ */ jsx(Outlet, {}), /* @__PURE__ */ jsx(ScrollRestoration, {}), /* @__PURE__ */ jsx(Scripts, {})]
    })]
  });
}
const root = UNSAFE_withComponentProps(function AppWithProviders() {
  const data = useLoaderData();
  return /* @__PURE__ */ jsx(ThemeProvider, {
    specifiedTheme: data.theme,
    themeAction: "/action/set-theme",
    disableTransitionOnThemeChange: true,
    children: /* @__PURE__ */ jsx(App, {})
  });
});
const ErrorBoundary = UNSAFE_withErrorBoundaryProps(function ErrorBoundary2({
  error
}) {
  let message = "Oops!";
  let details = "An unexpected error occurred.";
  let stack;
  if (isRouteErrorResponse(error)) {
    message = error.status === 404 ? "404" : "Error";
    details = error.status === 404 ? "The requested page could not be found." : error.statusText || details;
  }
  return /* @__PURE__ */ jsxs("main", {
    className: "pt-16 p-4 container mx-auto",
    children: [/* @__PURE__ */ jsx("h1", {
      children: message
    }), /* @__PURE__ */ jsx("p", {
      children: details
    }), stack]
  });
});
const route0 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  App,
  ErrorBoundary,
  default: root,
  links,
  loader: loader$3
}, Symbol.toStringTag, { value: "Module" }));
const visibilityBySurface = {
  database: "log",
  chat: "response",
  auth: "response",
  stream: "response",
  api: "response",
  history: "response",
  vote: "response",
  document: "response",
  suggestions: "response"
};
class ChatSDKError extends Error {
  constructor(errorCode, cause) {
    super();
    __publicField(this, "type");
    __publicField(this, "surface");
    __publicField(this, "statusCode");
    const [type, surface] = errorCode.split(":");
    this.type = type;
    this.cause = cause;
    this.surface = surface;
    this.message = getMessageByErrorCode(errorCode);
    this.statusCode = getStatusCodeByType(this.type);
  }
  toResponse() {
    const code = `${this.type}:${this.surface}`;
    const visibility = visibilityBySurface[this.surface];
    const { message, cause, statusCode } = this;
    if (visibility === "log") {
      console.error({
        code,
        message,
        cause
      });
      return Response.json(
        { code: "", message: "Something went wrong. Please try again later." },
        { status: statusCode }
      );
    }
    return Response.json({ code, message, cause }, { status: statusCode });
  }
}
function getMessageByErrorCode(errorCode) {
  if (errorCode.includes("database")) {
    return "An error occurred while executing a database query.";
  }
  switch (errorCode) {
    case "bad_request:api":
      return "The request couldn't be processed. Please check your input and try again.";
    case "unauthorized:auth":
      return "You need to sign in before continuing.";
    case "forbidden:auth":
      return "Your account does not have access to this feature.";
    case "rate_limit:chat":
      return "You have exceeded your maximum number of messages for the day. Please try again later.";
    case "not_found:chat":
      return "The requested chat was not found. Please check the chat ID and try again.";
    case "forbidden:chat":
      return "This chat belongs to another user. Please check the chat ID and try again.";
    case "unauthorized:chat":
      return "You need to sign in to view this chat. Please sign in and try again.";
    case "offline:chat":
      return "We're having trouble sending your message. Please check your internet connection and try again.";
    case "not_found:document":
      return "The requested document was not found. Please check the document ID and try again.";
    case "forbidden:document":
      return "This document belongs to another user. Please check the document ID and try again.";
    case "unauthorized:document":
      return "You need to sign in to view this document. Please sign in and try again.";
    case "bad_request:document":
      return "The request to create or update the document was invalid. Please check your input and try again.";
    default:
      return "Something went wrong. Please try again later.";
  }
}
function getStatusCodeByType(type) {
  switch (type) {
    case "bad_request":
      return 400;
    case "unauthorized":
      return 401;
    case "forbidden":
      return 403;
    case "not_found":
      return 404;
    case "rate_limit":
      return 429;
    case "offline":
      return 503;
    default:
      return 500;
  }
}
function cn(...inputs) {
  return twMerge(clsx(inputs));
}
function sanitizeText(text) {
  return text.replace("<has_function_call>", "");
}
function uuidv4() {
  let d = (/* @__PURE__ */ new Date()).getTime();
  if (typeof performance !== "undefined" && typeof performance.now === "function") {
    d += performance.now();
  }
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function(c) {
    const r = (d + Math.random() * 16) % 16 | 0;
    d = Math.floor(d / 16);
    return (c === "x" ? r : r & 3 | 8).toString(16);
  });
}
function getWidgets(messages) {
  var _a, _b, _c;
  const widgets = [];
  const uriToIndex = {};
  for (const message of messages) {
    for (const item of message.items || []) {
      if (item.type === "tool" && ((_a = item.output) == null ? void 0 : _a.content)) {
        for (const output of item.output.content) {
          if ((_c = (_b = output.resource) == null ? void 0 : _b.uri) == null ? void 0 : _c.startsWith("ui://widget/")) {
            const currentIdx = uriToIndex[output.resource.uri];
            if (currentIdx !== void 0) {
              widgets[currentIdx] = output;
            } else {
              widgets.push(output);
              uriToIndex[output.resource.uri] = widgets.length - 1;
            }
          }
        }
      }
    }
  }
  return widgets;
}
function appendProgress(messages, progress) {
  var _a;
  if (!progress.messageID || !progress.item) {
    return;
  }
  console.log("appendProgress", progress);
  const messageIndex = messages.findIndex((m) => m.id === progress.messageID);
  if (messageIndex < 0) {
    const message2 = {
      id: progress.messageID,
      role: "assistant",
      created: (/* @__PURE__ */ new Date()).toISOString(),
      items: progress.item ? [progress.item] : []
    };
    messages.push(message2);
    return;
  }
  const prevMessage = messages[messageIndex];
  const message = {
    ...prevMessage,
    items: [...prevMessage.items ?? []],
    revision: (prevMessage.revision || 0) + 1
  };
  messages[messageIndex] = message;
  const itemIndex = (_a = message.items) == null ? void 0 : _a.findIndex((x) => {
    var _a2;
    return x.id === ((_a2 = progress.item) == null ? void 0 : _a2.id);
  });
  if (itemIndex === void 0 || itemIndex === -1) {
    message.items.push(progress.item);
    return;
  }
  const item = { ...message.items[itemIndex] };
  message.items[itemIndex] = item;
  if (!progress.item.partial) {
    message.items[itemIndex] = progress.item;
    return;
  }
  if (!item.partial) {
    return;
  }
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
function events(id, cb) {
  const es = new EventSource(`/mcp/session/${id}/events`);
  es.onmessage = (e) => {
    var _a, _b;
    const event = JSON.parse(e.data);
    const progress = (_b = (_a = event == null ? void 0 : event.params) == null ? void 0 : _a["_meta"]) == null ? void 0 : _b["ai.nanobot.progress/completion"];
    if (progress) {
      progress.id = event.params.progressToken + "-" + event.params.progress;
      cb(progress);
    }
  };
  return () => {
    if (es.readyState === EventSource.CLOSED) {
      return;
    }
    es.close();
  };
}
async function deleteChat(ctx, id) {
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
      "X-Nanobot-Session-Id": id
    }
  });
  if (!resp.ok) {
    throw new Error(`Failed to delete chat ${id}: ${resp.statusText}`);
  }
}
async function call(ctx, id, tool, body, opts) {
  let url = "/mcp/" + ((opts == null ? void 0 : opts.agentId) ? `agents/${opts.agentId}/` : "") + id + "/tools";
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
      "X-Nanobot-Session-Id": id
    },
    body: JSON.stringify({
      jsonrpc: "2.0",
      method: "tools/call",
      params: {
        name: tool,
        arguments: body,
        _meta: {
          progressToken: uuidv4()
        }
      }
    })
  });
  const rpcResp = await resp.json();
  if (rpcResp.error) {
    throw new Error(
      `Error calling tool ${tool}: ${JSON.stringify(rpcResp.error)}`
    );
  }
  const result = rpcResp.result;
  if (result.isError) {
    throw new Error(
      `Error calling tool ${tool}: ${JSON.stringify(result.content)}`
    );
  }
  for (const content of result.content || []) {
    if (content.structuredContent) {
      return content.structuredContent;
    }
  }
  return null;
}
function getContext(request) {
  return {};
}
async function setAgent(ctx, id, agent2) {
  await call(ctx, id, "set_current_agent", { agent: agent2 });
}
async function setVisibility(ctx, id, visibility) {
  await call(ctx, id, "set_visibility", { visibility });
}
async function clone(ctx, id) {
  return await call(ctx, id, "clone", {});
}
async function newCustomAgent(ctx) {
  return await call(ctx, "new", "create_custom_agent", {});
}
async function updateCustomAgent(ctx, id, agent2) {
  return await call(ctx, id, "update_custom_agent", agent2, {
    agentId: agent2.id
  });
}
async function chat(ctx, id, prompt, opts) {
  await call(ctx, id, "chat", { prompt, attachments: opts == null ? void 0 : opts.attachments });
}
async function getChat(ctx, id, opts) {
  const result = await call(ctx, id, "get_chat", void 0, opts);
  if ("ai.nanobot/ext" in result) {
    const copy = { ...result, ...result["ai.nanobot/ext"] };
    delete result["ai.nanobot/ext"];
    return copy;
  }
  return result;
}
async function listChats(ctx, id) {
  return await call(ctx, id, "list_chats");
}
async function listCustomAgents(ctx, id) {
  return await call(ctx, id, "list_custom_agents");
}
async function createResource(ctx, id, blob, opts) {
  return await call(ctx, id, "create_resource", {
    blob,
    ...opts ?? {}
  });
}
async function deleteCustomAgent(ctx, id, agentId) {
  await call(
    ctx,
    id,
    "delete_custom_agent",
    { id: agentId },
    {
      agentId
    }
  );
}
async function action$6({
  request,
  params
}) {
  const formData = await request.formData();
  const file = formData.get("file");
  const mimeType = formData.get("mimeType");
  if (!(file instanceof File)) {
    return {
      error: "Invalid file upload."
    };
  }
  if (!mimeType || typeof mimeType !== "string") {
    return {
      error: "Invalid or missing mimeType."
    };
  }
  try {
    const buffer = await file.arrayBuffer();
    const base64String = Buffer.from(buffer).toString("base64");
    return createResource(getContext(request), params.id, base64String, {
      mimeType
    });
  } catch (error) {
    return {
      error: `Failed to read file. ${error}`
    };
  }
}
const route1 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  action: action$6
}, Symbol.toStringTag, { value: "Module" }));
async function action$5({
  request,
  params
}) {
  const id = await clone(getContext(), params.id || "");
  return redirect(`/chat/${id}`);
}
const route2 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  action: action$5
}, Symbol.toStringTag, { value: "Module" }));
async function action$4({
  request,
  params
}) {
  const formData = await request.formData();
  const agentId = formData.get("agentId");
  if (request.method !== "DELETE") {
    throw new Error("Invalid request method. Use DELETE to remove an agent.");
  }
  if (!agentId) {
    throw new Error("Agent ID is required");
  }
  await deleteCustomAgent(getContext(), params.id || "new", agentId);
  return redirect("/");
}
const route3 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  action: action$4
}, Symbol.toStringTag, { value: "Module" }));
async function action$3({
  request
}) {
  const agent2 = await newCustomAgent(getContext());
  return redirect(`/agent/${agent2.id}/edit`);
}
const route4 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  action: action$3
}, Symbol.toStringTag, { value: "Module" }));
const sessionStorage = createCookieSessionStorage({
  cookie: {
    name: "__session",
    httpOnly: true,
    path: "/",
    sameSite: "lax",
    secrets: ["s3cr3t"],
    // replace this with an actual secret
    secure: process.env.NODE_ENV === "production"
  }
});
const authenticator = new Authenticator();
async function login$1(email, password) {
  return {
    id: "12345",
    // Replace with actual user ID
    email,
    name: "John Doe"
    // Replace with actual user name
  };
}
authenticator.use(
  new FormStrategy(async ({ form }) => {
    const email = form.get("email");
    const password = form.get("password");
    if (!email || !password) {
      throw new Error("Email and password are required");
    }
    return await login$1(email);
  }),
  // each strategy has a name and can be changed to use the same strategy
  // multiple times, especially useful for the OAuth2 strategy.
  "user-pass"
);
async function action$2({
  request
}) {
  try {
    const user = await authenticator.authenticate("user-pass", request);
    const session = await sessionStorage.getSession(request.headers.get("cookie"));
    session.set("user", user);
    redirect("/", {
      headers: {
        "Set-Cookie": await sessionStorage.commitSession(session)
      }
    });
    return {};
  } catch (error) {
    if (error instanceof Error) {
      return {
        error: error.message
      };
    }
    throw error;
  }
}
async function loader$2({
  request
}) {
  const session = await sessionStorage.getSession(request.headers.get("cookie"));
  const user = session.get("user");
  if (user) return redirect("/");
  return {};
}
const login = UNSAFE_withComponentProps(function Login({
  actionData
}) {
  return /* @__PURE__ */ jsxs("div", {
    children: [/* @__PURE__ */ jsx("h1", {
      children: "Login"
    }), (actionData == null ? void 0 : actionData.error) ? /* @__PURE__ */ jsx("div", {
      className: "error",
      children: actionData.error
    }) : null, /* @__PURE__ */ jsxs(Form, {
      method: "post",
      target: "/login",
      children: [/* @__PURE__ */ jsxs("div", {
        children: [/* @__PURE__ */ jsx("label", {
          htmlFor: "email",
          children: "Email"
        }), /* @__PURE__ */ jsx("input", {
          type: "email",
          name: "email",
          id: "email",
          required: true
        })]
      }), /* @__PURE__ */ jsxs("div", {
        children: [/* @__PURE__ */ jsx("label", {
          htmlFor: "password",
          children: "Password"
        }), /* @__PURE__ */ jsx("input", {
          type: "password",
          name: "password",
          id: "password",
          autoComplete: "current-password",
          required: true
        })]
      }), /* @__PURE__ */ jsx("button", {
        type: "submit",
        children: "Sign In"
      })]
    })]
  });
});
const route5 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  action: action$2,
  default: login,
  loader: loader$2
}, Symbol.toStringTag, { value: "Module" }));
async function loader$1({
  request
}) {
  const session = await sessionStorage.getSession(request.headers.get("cookie"));
  return redirect("/login", {
    headers: {
      "Set-Cookie": await sessionStorage.destroySession(session)
    }
  });
}
const logout = UNSAFE_withComponentProps(function Logout() {
  return /* @__PURE__ */ jsx("div", {
    children: "Logging out..."
  });
});
const route6 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  default: logout,
  loader: loader$1
}, Symbol.toStringTag, { value: "Module" }));
const MOBILE_BREAKPOINT = 768;
function useIsMobile() {
  const [isMobile, setIsMobile] = React.useState(
    void 0
  );
  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`);
    const onChange = () => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
    };
    mql.addEventListener("change", onChange);
    setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
    return () => mql.removeEventListener("change", onChange);
  }, []);
  return !!isMobile;
}
const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground shadow-xs hover:bg-primary/90",
        destructive: "bg-destructive text-white shadow-xs hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60",
        outline: "border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground dark:bg-input/30 dark:border-input dark:hover:bg-input/50",
        secondary: "bg-secondary text-secondary-foreground shadow-xs hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50",
        link: "text-primary underline-offset-4 hover:underline"
      },
      size: {
        default: "h-9 px-4 py-2 has-[>svg]:px-3",
        sm: "h-8 rounded-md gap-1.5 px-3 has-[>svg]:px-2.5",
        lg: "h-10 rounded-md px-6 has-[>svg]:px-4",
        icon: "size-9"
      }
    },
    defaultVariants: {
      variant: "default",
      size: "default"
    }
  }
);
function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}) {
  const Comp = asChild ? Slot : "button";
  return /* @__PURE__ */ jsx(
    Comp,
    {
      "data-slot": "button",
      className: cn(buttonVariants({ variant, size, className })),
      ...props
    }
  );
}
function Input({ className, type, ...props }) {
  return /* @__PURE__ */ jsx(
    "input",
    {
      type,
      "data-slot": "input",
      className: cn(
        "file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 border-input flex h-9 w-full min-w-0 rounded-md border bg-transparent px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
        "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
        "aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
        className
      ),
      ...props
    }
  );
}
function Sheet({ ...props }) {
  return /* @__PURE__ */ jsx(SheetPrimitive.Root, { "data-slot": "sheet", ...props });
}
function SheetPortal({
  ...props
}) {
  return /* @__PURE__ */ jsx(SheetPrimitive.Portal, { "data-slot": "sheet-portal", ...props });
}
function SheetOverlay({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    SheetPrimitive.Overlay,
    {
      "data-slot": "sheet-overlay",
      className: cn(
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/50",
        className
      ),
      ...props
    }
  );
}
function SheetContent({
  className,
  children,
  side = "right",
  ...props
}) {
  return /* @__PURE__ */ jsxs(SheetPortal, { children: [
    /* @__PURE__ */ jsx(SheetOverlay, {}),
    /* @__PURE__ */ jsxs(
      SheetPrimitive.Content,
      {
        "data-slot": "sheet-content",
        className: cn(
          "bg-background data-[state=open]:animate-in data-[state=closed]:animate-out fixed z-50 flex flex-col gap-4 shadow-lg transition ease-in-out data-[state=closed]:duration-300 data-[state=open]:duration-500",
          side === "right" && "data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right inset-y-0 right-0 h-full w-3/4 border-l sm:max-w-sm",
          side === "left" && "data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left inset-y-0 left-0 h-full w-3/4 border-r sm:max-w-sm",
          side === "top" && "data-[state=closed]:slide-out-to-top data-[state=open]:slide-in-from-top inset-x-0 top-0 h-auto border-b",
          side === "bottom" && "data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom inset-x-0 bottom-0 h-auto border-t",
          className
        ),
        ...props,
        children: [
          children,
          /* @__PURE__ */ jsxs(SheetPrimitive.Close, { className: "ring-offset-background focus:ring-ring data-[state=open]:bg-secondary absolute top-4 right-4 rounded-xs opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 focus:outline-hidden disabled:pointer-events-none", children: [
            /* @__PURE__ */ jsx(XIcon, { className: "size-4" }),
            /* @__PURE__ */ jsx("span", { className: "sr-only", children: "Close" })
          ] })
        ]
      }
    )
  ] });
}
function SheetHeader({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sheet-header",
      className: cn("flex flex-col gap-1.5 p-4", className),
      ...props
    }
  );
}
function SheetTitle({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    SheetPrimitive.Title,
    {
      "data-slot": "sheet-title",
      className: cn("text-foreground font-semibold", className),
      ...props
    }
  );
}
function SheetDescription({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    SheetPrimitive.Description,
    {
      "data-slot": "sheet-description",
      className: cn("text-muted-foreground text-sm", className),
      ...props
    }
  );
}
function TooltipProvider({
  delayDuration = 0,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    TooltipPrimitive.Provider,
    {
      "data-slot": "tooltip-provider",
      delayDuration,
      ...props
    }
  );
}
function Tooltip({
  ...props
}) {
  return /* @__PURE__ */ jsx(TooltipProvider, { children: /* @__PURE__ */ jsx(TooltipPrimitive.Root, { "data-slot": "tooltip", ...props }) });
}
function TooltipTrigger({
  ...props
}) {
  return /* @__PURE__ */ jsx(TooltipPrimitive.Trigger, { "data-slot": "tooltip-trigger", ...props });
}
function TooltipContent({
  className,
  sideOffset = 0,
  children,
  ...props
}) {
  return /* @__PURE__ */ jsx(TooltipPrimitive.Portal, { children: /* @__PURE__ */ jsxs(
    TooltipPrimitive.Content,
    {
      "data-slot": "tooltip-content",
      sideOffset,
      className: cn(
        "bg-primary text-primary-foreground animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 w-fit origin-(--radix-tooltip-content-transform-origin) rounded-md px-3 py-1.5 text-xs text-balance",
        className
      ),
      ...props,
      children: [
        children,
        /* @__PURE__ */ jsx(TooltipPrimitive.Arrow, { className: "bg-primary fill-primary z-50 size-2.5 translate-y-[calc(-50%_-_2px)] rotate-45 rounded-[2px]" })
      ]
    }
  ) });
}
const SIDEBAR_COOKIE_NAME = "sidebar_state";
const SIDEBAR_COOKIE_MAX_AGE = 60 * 60 * 24 * 7;
const SIDEBAR_WIDTH = "16rem";
const SIDEBAR_WIDTH_MOBILE = "18rem";
const SIDEBAR_WIDTH_ICON = "3rem";
const SIDEBAR_KEYBOARD_SHORTCUT = "b";
const SidebarContext = React.createContext(null);
function useSidebar() {
  const context = React.useContext(SidebarContext);
  if (!context) {
    throw new Error("useSidebar must be used within a SidebarProvider.");
  }
  return context;
}
function SidebarProvider({
  defaultOpen = true,
  open: openProp,
  onOpenChange: setOpenProp,
  className,
  style,
  children,
  ...props
}) {
  const isMobile = useIsMobile();
  const [openMobile, setOpenMobile] = React.useState(false);
  const [_open, _setOpen] = React.useState(defaultOpen);
  const open = openProp ?? _open;
  const setOpen = React.useCallback(
    (value) => {
      const openState = typeof value === "function" ? value(open) : value;
      if (setOpenProp) {
        setOpenProp(openState);
      } else {
        _setOpen(openState);
      }
      document.cookie = `${SIDEBAR_COOKIE_NAME}=${openState}; path=/; max-age=${SIDEBAR_COOKIE_MAX_AGE}`;
    },
    [setOpenProp, open]
  );
  const toggleSidebar = React.useCallback(() => {
    return isMobile ? setOpenMobile((open2) => !open2) : setOpen((open2) => !open2);
  }, [isMobile, setOpen, setOpenMobile]);
  React.useEffect(() => {
    const handleKeyDown = (event) => {
      if (event.key === SIDEBAR_KEYBOARD_SHORTCUT && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        toggleSidebar();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [toggleSidebar]);
  const state = open ? "expanded" : "collapsed";
  const contextValue = React.useMemo(
    () => ({
      state,
      open,
      setOpen,
      isMobile,
      openMobile,
      setOpenMobile,
      toggleSidebar
    }),
    [state, open, setOpen, isMobile, openMobile, setOpenMobile, toggleSidebar]
  );
  return /* @__PURE__ */ jsx(SidebarContext.Provider, { value: contextValue, children: /* @__PURE__ */ jsx(TooltipProvider, { delayDuration: 0, children: /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-wrapper",
      style: {
        "--sidebar-width": SIDEBAR_WIDTH,
        "--sidebar-width-icon": SIDEBAR_WIDTH_ICON,
        ...style
      },
      className: cn(
        "group/sidebar-wrapper has-data-[variant=inset]:bg-sidebar flex min-h-svh w-full",
        className
      ),
      ...props,
      children
    }
  ) }) });
}
function Sidebar({
  side = "left",
  variant = "sidebar",
  collapsible = "offcanvas",
  className,
  children,
  ...props
}) {
  const { isMobile, state, openMobile, setOpenMobile } = useSidebar();
  if (collapsible === "none") {
    return /* @__PURE__ */ jsx(
      "div",
      {
        "data-slot": "sidebar",
        className: cn(
          "bg-sidebar text-sidebar-foreground flex h-full w-(--sidebar-width) flex-col",
          className
        ),
        ...props,
        children
      }
    );
  }
  if (isMobile) {
    return /* @__PURE__ */ jsx(Sheet, { open: openMobile, onOpenChange: setOpenMobile, ...props, children: /* @__PURE__ */ jsxs(
      SheetContent,
      {
        "data-sidebar": "sidebar",
        "data-slot": "sidebar",
        "data-mobile": "true",
        className: "bg-sidebar text-sidebar-foreground w-(--sidebar-width) p-0 [&>button]:hidden",
        style: {
          "--sidebar-width": SIDEBAR_WIDTH_MOBILE
        },
        side,
        children: [
          /* @__PURE__ */ jsxs(SheetHeader, { className: "sr-only", children: [
            /* @__PURE__ */ jsx(SheetTitle, { children: "Sidebar" }),
            /* @__PURE__ */ jsx(SheetDescription, { children: "Displays the mobile sidebar." })
          ] }),
          /* @__PURE__ */ jsx("div", { className: "flex h-full w-full flex-col", children })
        ]
      }
    ) });
  }
  return /* @__PURE__ */ jsxs(
    "div",
    {
      className: "group peer text-sidebar-foreground hidden md:block",
      "data-state": state,
      "data-collapsible": state === "collapsed" ? collapsible : "",
      "data-variant": variant,
      "data-side": side,
      "data-slot": "sidebar",
      children: [
        /* @__PURE__ */ jsx(
          "div",
          {
            "data-slot": "sidebar-gap",
            className: cn(
              "relative w-(--sidebar-width) bg-transparent transition-[width] duration-200 ease-linear",
              "group-data-[collapsible=offcanvas]:w-0",
              "group-data-[side=right]:rotate-180",
              variant === "floating" || variant === "inset" ? "group-data-[collapsible=icon]:w-[calc(var(--sidebar-width-icon)+(--spacing(4)))]" : "group-data-[collapsible=icon]:w-(--sidebar-width-icon)"
            )
          }
        ),
        /* @__PURE__ */ jsx(
          "div",
          {
            "data-slot": "sidebar-container",
            className: cn(
              "fixed inset-y-0 z-10 hidden h-svh w-(--sidebar-width) transition-[left,right,width] duration-200 ease-linear md:flex",
              side === "left" ? "left-0 group-data-[collapsible=offcanvas]:left-[calc(var(--sidebar-width)*-1)]" : "right-0 group-data-[collapsible=offcanvas]:right-[calc(var(--sidebar-width)*-1)]",
              // Adjust the padding for floating and inset variants.
              variant === "floating" || variant === "inset" ? "p-2 group-data-[collapsible=icon]:w-[calc(var(--sidebar-width-icon)+(--spacing(4))+2px)]" : "group-data-[collapsible=icon]:w-(--sidebar-width-icon) group-data-[side=left]:border-r group-data-[side=right]:border-l",
              className
            ),
            ...props,
            children: /* @__PURE__ */ jsx(
              "div",
              {
                "data-sidebar": "sidebar",
                "data-slot": "sidebar-inner",
                className: "bg-sidebar group-data-[variant=floating]:border-sidebar-border flex h-full w-full flex-col group-data-[variant=floating]:rounded-lg group-data-[variant=floating]:border group-data-[variant=floating]:shadow-sm",
                children
              }
            )
          }
        )
      ]
    }
  );
}
function SidebarInset({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "main",
    {
      "data-slot": "sidebar-inset",
      className: cn(
        "bg-background relative flex w-full flex-1 flex-col",
        "md:peer-data-[variant=inset]:m-2 md:peer-data-[variant=inset]:ml-0 md:peer-data-[variant=inset]:rounded-xl md:peer-data-[variant=inset]:shadow-sm md:peer-data-[variant=inset]:peer-data-[state=collapsed]:ml-2",
        className
      ),
      ...props
    }
  );
}
function SidebarHeader({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-header",
      "data-sidebar": "header",
      className: cn("flex flex-col gap-2 p-2", className),
      ...props
    }
  );
}
function SidebarFooter({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-footer",
      "data-sidebar": "footer",
      className: cn("flex flex-col gap-2 p-2", className),
      ...props
    }
  );
}
function SidebarContent({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-content",
      "data-sidebar": "content",
      className: cn(
        "flex min-h-0 flex-1 flex-col gap-2 overflow-auto group-data-[collapsible=icon]:overflow-hidden",
        className
      ),
      ...props
    }
  );
}
function SidebarGroup({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-group",
      "data-sidebar": "group",
      className: cn("relative flex w-full min-w-0 flex-col p-2", className),
      ...props
    }
  );
}
function SidebarGroupLabel({
  className,
  asChild = false,
  ...props
}) {
  const Comp = asChild ? Slot : "div";
  return /* @__PURE__ */ jsx(
    Comp,
    {
      "data-slot": "sidebar-group-label",
      "data-sidebar": "group-label",
      className: cn(
        "text-sidebar-foreground/70 ring-sidebar-ring flex h-8 shrink-0 items-center rounded-md px-2 text-xs font-medium outline-hidden transition-[margin,opacity] duration-200 ease-linear focus-visible:ring-2 [&>svg]:size-4 [&>svg]:shrink-0",
        "group-data-[collapsible=icon]:-mt-8 group-data-[collapsible=icon]:opacity-0",
        className
      ),
      ...props
    }
  );
}
function SidebarGroupContent({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-group-content",
      "data-sidebar": "group-content",
      className: cn("w-full text-sm", className),
      ...props
    }
  );
}
function SidebarMenu({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "ul",
    {
      "data-slot": "sidebar-menu",
      "data-sidebar": "menu",
      className: cn("flex w-full min-w-0 flex-col gap-1", className),
      ...props
    }
  );
}
function SidebarMenuItem({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "li",
    {
      "data-slot": "sidebar-menu-item",
      "data-sidebar": "menu-item",
      className: cn("group/menu-item relative", className),
      ...props
    }
  );
}
const sidebarMenuButtonVariants = cva(
  "peer/menu-button flex w-full items-center gap-2 overflow-hidden rounded-md p-2 text-left text-sm outline-hidden ring-sidebar-ring transition-[width,height,padding] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-2 active:bg-sidebar-accent active:text-sidebar-accent-foreground disabled:pointer-events-none disabled:opacity-50 group-has-data-[sidebar=menu-action]/menu-item:pr-8 aria-disabled:pointer-events-none aria-disabled:opacity-50 data-[active=true]:bg-sidebar-accent data-[active=true]:font-medium data-[active=true]:text-sidebar-accent-foreground data-[state=open]:hover:bg-sidebar-accent data-[state=open]:hover:text-sidebar-accent-foreground group-data-[collapsible=icon]:size-8! group-data-[collapsible=icon]:p-2! [&>span:last-child]:truncate [&>svg]:size-4 [&>svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
        outline: "bg-background shadow-[0_0_0_1px_hsl(var(--sidebar-border))] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground hover:shadow-[0_0_0_1px_hsl(var(--sidebar-accent))]"
      },
      size: {
        default: "h-8 text-sm",
        sm: "h-7 text-xs",
        lg: "h-12 text-sm group-data-[collapsible=icon]:p-0!"
      }
    },
    defaultVariants: {
      variant: "default",
      size: "default"
    }
  }
);
function SidebarMenuButton({
  asChild = false,
  isActive = false,
  variant = "default",
  size = "default",
  tooltip,
  className,
  ...props
}) {
  const Comp = asChild ? Slot : "button";
  const { isMobile, state } = useSidebar();
  const button = /* @__PURE__ */ jsx(
    Comp,
    {
      "data-slot": "sidebar-menu-button",
      "data-sidebar": "menu-button",
      "data-size": size,
      "data-active": isActive,
      className: cn(sidebarMenuButtonVariants({ variant, size }), className),
      ...props
    }
  );
  if (!tooltip) {
    return button;
  }
  if (typeof tooltip === "string") {
    tooltip = {
      children: tooltip
    };
  }
  return /* @__PURE__ */ jsxs(Tooltip, { children: [
    /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: button }),
    /* @__PURE__ */ jsx(
      TooltipContent,
      {
        side: "right",
        align: "center",
        hidden: state !== "collapsed" || isMobile,
        ...tooltip
      }
    )
  ] });
}
function SidebarMenuAction({
  className,
  asChild = false,
  showOnHover = false,
  ...props
}) {
  const Comp = asChild ? Slot : "button";
  return /* @__PURE__ */ jsx(
    Comp,
    {
      "data-slot": "sidebar-menu-action",
      "data-sidebar": "menu-action",
      className: cn(
        "text-sidebar-foreground ring-sidebar-ring hover:bg-sidebar-accent hover:text-sidebar-accent-foreground peer-hover/menu-button:text-sidebar-accent-foreground absolute top-1.5 right-1 flex aspect-square w-5 items-center justify-center rounded-md p-0 outline-hidden transition-transform focus-visible:ring-2 [&>svg]:size-4 [&>svg]:shrink-0",
        // Increases the hit area of the button on mobile.
        "after:absolute after:-inset-2 md:after:hidden",
        "peer-data-[size=sm]/menu-button:top-1",
        "peer-data-[size=default]/menu-button:top-1.5",
        "peer-data-[size=lg]/menu-button:top-2.5",
        "group-data-[collapsible=icon]:hidden",
        showOnHover && "peer-data-[active=true]/menu-button:text-sidebar-accent-foreground group-focus-within/menu-item:opacity-100 group-hover/menu-item:opacity-100 data-[state=open]:opacity-100 md:opacity-0",
        className
      ),
      ...props
    }
  );
}
const sidebarCookie = createCookie("sidebar:state", {
  maxAge: 604800
  // one week
});
const LoaderIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsxs(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: [
        /* @__PURE__ */ jsxs("g", { clipPath: "url(#clip0_2393_1490)", children: [
          /* @__PURE__ */ jsx("path", { d: "M8 0V4", stroke: "currentColor", strokeWidth: "1.5" }),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.5",
              d: "M8 16V12",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          ),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.9",
              d: "M3.29773 1.52783L5.64887 4.7639",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          ),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.1",
              d: "M12.7023 1.52783L10.3511 4.7639",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          ),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.4",
              d: "M12.7023 14.472L10.3511 11.236",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          ),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.6",
              d: "M3.29773 14.472L5.64887 11.236",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          ),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.2",
              d: "M15.6085 5.52783L11.8043 6.7639",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          ),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.7",
              d: "M0.391602 10.472L4.19583 9.23598",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          ),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.3",
              d: "M15.6085 10.4722L11.8043 9.2361",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          ),
          /* @__PURE__ */ jsx(
            "path",
            {
              opacity: "0.8",
              d: "M0.391602 5.52783L4.19583 6.7639",
              stroke: "currentColor",
              strokeWidth: "1.5"
            }
          )
        ] }),
        /* @__PURE__ */ jsx("defs", { children: /* @__PURE__ */ jsx("clipPath", { id: "clip0_2393_1490", children: /* @__PURE__ */ jsx("rect", { width: "16", height: "16", fill: "white" }) }) })
      ]
    }
  );
};
const PencilEditIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M11.75 0.189331L12.2803 0.719661L15.2803 3.71966L15.8107 4.24999L15.2803 4.78032L5.15901 14.9016C4.45575 15.6049 3.50192 16 2.50736 16H0.75H0V15.25V13.4926C0 12.4981 0.395088 11.5442 1.09835 10.841L11.2197 0.719661L11.75 0.189331ZM11.75 2.31065L9.81066 4.24999L11.75 6.18933L13.6893 4.24999L11.75 2.31065ZM2.15901 11.9016L8.75 5.31065L10.6893 7.24999L4.09835 13.841C3.67639 14.2629 3.1041 14.5 2.50736 14.5H1.5V13.4926C1.5 12.8959 1.73705 12.3236 2.15901 11.9016ZM9 16H16V14.5H9V16Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const TrashIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M6.75 2.75C6.75 2.05964 7.30964 1.5 8 1.5C8.69036 1.5 9.25 2.05964 9.25 2.75V3H6.75V2.75ZM5.25 3V2.75C5.25 1.23122 6.48122 0 8 0C9.51878 0 10.75 1.23122 10.75 2.75V3H12.9201H14.25H15V4.5H14.25H13.8846L13.1776 13.6917C13.0774 14.9942 11.9913 16 10.6849 16H5.31508C4.00874 16 2.92263 14.9942 2.82244 13.6917L2.11538 4.5H1.75H1V3H1.75H3.07988H5.25ZM4.31802 13.5767L3.61982 4.5H12.3802L11.682 13.5767C11.6419 14.0977 11.2075 14.5 10.6849 14.5H5.31508C4.79254 14.5 4.3581 14.0977 4.31802 13.5767Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const ArrowUpIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M8.70711 1.39644C8.31659 1.00592 7.68342 1.00592 7.2929 1.39644L2.21968 6.46966L1.68935 6.99999L2.75001 8.06065L3.28034 7.53032L7.25001 3.56065V14.25V15H8.75001V14.25V3.56065L12.7197 7.53032L13.25 8.06065L14.3107 6.99999L13.7803 6.46966L8.70711 1.39644Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const StopIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M3 3H13V13H3V3Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const PaperclipIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      className: "-rotate-45",
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M10.8591 1.70735C10.3257 1.70735 9.81417 1.91925 9.437 2.29643L3.19455 8.53886C2.56246 9.17095 2.20735 10.0282 2.20735 10.9222C2.20735 11.8161 2.56246 12.6734 3.19455 13.3055C3.82665 13.9376 4.68395 14.2927 5.57786 14.2927C6.47178 14.2927 7.32908 13.9376 7.96117 13.3055L14.2036 7.06304L14.7038 6.56287L15.7041 7.56321L15.204 8.06337L8.96151 14.3058C8.06411 15.2032 6.84698 15.7074 5.57786 15.7074C4.30875 15.7074 3.09162 15.2032 2.19422 14.3058C1.29682 13.4084 0.792664 12.1913 0.792664 10.9222C0.792664 9.65305 1.29682 8.43592 2.19422 7.53852L8.43666 1.29609C9.07914 0.653606 9.95054 0.292664 10.8591 0.292664C11.7678 0.292664 12.6392 0.653606 13.2816 1.29609C13.9241 1.93857 14.2851 2.80997 14.2851 3.71857C14.2851 4.62718 13.9241 5.49858 13.2816 6.14106L13.2814 6.14133L7.0324 12.3835C7.03231 12.3836 7.03222 12.3837 7.03213 12.3838C6.64459 12.7712 6.11905 12.9888 5.57107 12.9888C5.02297 12.9888 4.49731 12.7711 4.10974 12.3835C3.72217 11.9959 3.50444 11.4703 3.50444 10.9222C3.50444 10.3741 3.72217 9.8484 4.10974 9.46084L4.11004 9.46054L9.877 3.70039L10.3775 3.20051L11.3772 4.20144L10.8767 4.70131L5.11008 10.4612C5.11005 10.4612 5.11003 10.4612 5.11 10.4613C4.98779 10.5835 4.91913 10.7493 4.91913 10.9222C4.91913 11.0951 4.98782 11.2609 5.11008 11.3832C5.23234 11.5054 5.39817 11.5741 5.57107 11.5741C5.74398 11.5741 5.9098 11.5054 6.03206 11.3832L6.03233 11.3829L12.2813 5.14072C12.2814 5.14063 12.2815 5.14054 12.2816 5.14045C12.6586 4.7633 12.8704 4.25185 12.8704 3.71857C12.8704 3.18516 12.6585 2.6736 12.2813 2.29643C11.9041 1.91925 11.3926 1.70735 10.8591 1.70735Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const MoreHorizontalIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M4 8C4 8.82843 3.32843 9.5 2.5 9.5C1.67157 9.5 1 8.82843 1 8C1 7.17157 1.67157 6.5 2.5 6.5C3.32843 6.5 4 7.17157 4 8ZM9.5 8C9.5 8.82843 8.82843 9.5 8 9.5C7.17157 9.5 6.5 8.82843 6.5 8C6.5 7.17157 7.17157 6.5 8 6.5C8.82843 6.5 9.5 7.17157 9.5 8ZM13.5 9.5C14.3284 9.5 15 8.82843 15 8C15 7.17157 14.3284 6.5 13.5 6.5C12.6716 6.5 12 7.17157 12 8C12 8.82843 12.6716 9.5 13.5 9.5Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const SidebarLeftIcon = ({ size = 16 }) => /* @__PURE__ */ jsx(
  "svg",
  {
    height: size,
    strokeLinejoin: "round",
    viewBox: "0 0 16 16",
    width: size,
    style: { color: "currentcolor" },
    children: /* @__PURE__ */ jsx(
      "path",
      {
        fillRule: "evenodd",
        clipRule: "evenodd",
        d: "M6.245 2.5H14.5V12.5C14.5 13.0523 14.0523 13.5 13.5 13.5H6.245V2.5ZM4.995 2.5H1.5V12.5C1.5 13.0523 1.94772 13.5 2.5 13.5H4.995V2.5ZM0 1H1.5H14.5H16V2.5V12.5C16 13.8807 14.8807 15 13.5 15H2.5C1.11929 15 0 13.8807 0 12.5V2.5V1Z",
        fill: "currentColor"
      }
    )
  }
);
const PlusIcon = ({ size = 16 }) => /* @__PURE__ */ jsx(
  "svg",
  {
    height: size,
    strokeLinejoin: "round",
    viewBox: "0 0 16 16",
    width: size,
    style: { color: "currentcolor" },
    children: /* @__PURE__ */ jsx(
      "path",
      {
        fillRule: "evenodd",
        clipRule: "evenodd",
        d: "M 8.75,1 H7.25 V7.25 H1.5 V8.75 H7.25 V15 H8.75 V8.75 H14.5 V7.25 H8.75 V1.75 Z",
        fill: "currentColor"
      }
    )
  }
);
const CopyIcon = ({ size = 16 }) => /* @__PURE__ */ jsx(
  "svg",
  {
    height: size,
    strokeLinejoin: "round",
    viewBox: "0 0 16 16",
    width: size,
    style: { color: "currentcolor" },
    children: /* @__PURE__ */ jsx(
      "path",
      {
        fillRule: "evenodd",
        clipRule: "evenodd",
        d: "M2.75 0.5C1.7835 0.5 1 1.2835 1 2.25V9.75C1 10.7165 1.7835 11.5 2.75 11.5H3.75H4.5V10H3.75H2.75C2.61193 10 2.5 9.88807 2.5 9.75V2.25C2.5 2.11193 2.61193 2 2.75 2H8.25C8.38807 2 8.5 2.11193 8.5 2.25V3H10V2.25C10 1.2835 9.2165 0.5 8.25 0.5H2.75ZM7.75 4.5C6.7835 4.5 6 5.2835 6 6.25V13.75C6 14.7165 6.7835 15.5 7.75 15.5H13.25C14.2165 15.5 15 14.7165 15 13.75V6.25C15 5.2835 14.2165 4.5 13.25 4.5H7.75ZM7.5 6.25C7.5 6.11193 7.61193 6 7.75 6H13.25C13.3881 6 13.5 6.11193 13.5 6.25V13.75C13.5 13.8881 13.3881 14 13.25 14H7.75C7.61193 14 7.5 13.8881 7.5 13.75V6.25Z",
        fill: "currentColor"
      }
    )
  }
);
const ThumbUpIcon = ({ size = 16 }) => /* @__PURE__ */ jsx(
  "svg",
  {
    height: size,
    strokeLinejoin: "round",
    viewBox: "0 0 16 16",
    width: size,
    style: { color: "currentcolor" },
    children: /* @__PURE__ */ jsx(
      "path",
      {
        fillRule: "evenodd",
        clipRule: "evenodd",
        d: "M6.89531 2.23972C6.72984 2.12153 6.5 2.23981 6.5 2.44315V5.25001C6.5 6.21651 5.7165 7.00001 4.75 7.00001H2.5V13.5H12.1884C12.762 13.5 13.262 13.1096 13.4011 12.5532L14.4011 8.55318C14.5984 7.76425 14.0017 7.00001 13.1884 7.00001H9.25H8.5V6.25001V3.51458C8.5 3.43384 8.46101 3.35807 8.39531 3.31114L6.89531 2.23972ZM5 2.44315C5 1.01975 6.6089 0.191779 7.76717 1.01912L9.26717 2.09054C9.72706 2.41904 10 2.94941 10 3.51458V5.50001H13.1884C14.9775 5.50001 16.2903 7.18133 15.8563 8.91698L14.8563 12.917C14.5503 14.1412 13.4503 15 12.1884 15H1.75H1V14.25V6.25001V5.50001H1.75H4.75C4.88807 5.50001 5 5.38808 5 5.25001V2.44315Z",
        fill: "currentColor"
      }
    )
  }
);
const ThumbDownIcon = ({ size = 16 }) => /* @__PURE__ */ jsx(
  "svg",
  {
    height: size,
    strokeLinejoin: "round",
    viewBox: "0 0 16 16",
    width: size,
    style: { color: "currentcolor" },
    children: /* @__PURE__ */ jsx(
      "path",
      {
        fillRule: "evenodd",
        clipRule: "evenodd",
        d: "M6.89531 13.7603C6.72984 13.8785 6.5 13.7602 6.5 13.5569V10.75C6.5 9.7835 5.7165 9 4.75 9H2.5V2.5H12.1884C12.762 2.5 13.262 2.89037 13.4011 3.44683L14.4011 7.44683C14.5984 8.23576 14.0017 9 13.1884 9H9.25H8.5V9.75V12.4854C8.5 12.5662 8.46101 12.6419 8.39531 12.6889L6.89531 13.7603ZM5 13.5569C5 14.9803 6.6089 15.8082 7.76717 14.9809L9.26717 13.9095C9.72706 13.581 10 13.0506 10 12.4854V10.5H13.1884C14.9775 10.5 16.2903 8.81868 15.8563 7.08303L14.8563 3.08303C14.5503 1.85882 13.4503 1 12.1884 1H1.75H1V1.75V9.75V10.5H1.75H4.75C4.88807 10.5 5 10.6119 5 10.75V13.5569Z",
        fill: "currentColor"
      }
    )
  }
);
const ChevronDownIcon = ({ size = 16 }) => /* @__PURE__ */ jsx(
  "svg",
  {
    height: size,
    strokeLinejoin: "round",
    viewBox: "0 0 16 16",
    width: size,
    style: { color: "currentcolor" },
    children: /* @__PURE__ */ jsx(
      "path",
      {
        fillRule: "evenodd",
        clipRule: "evenodd",
        d: "M12.0607 6.74999L11.5303 7.28032L8.7071 10.1035C8.31657 10.4941 7.68341 10.4941 7.29288 10.1035L4.46966 7.28032L3.93933 6.74999L4.99999 5.68933L5.53032 6.21966L7.99999 8.68933L10.4697 6.21966L11 5.68933L12.0607 6.74999Z",
        fill: "currentColor"
      }
    )
  }
);
const SparklesIcon = ({ size = 16 }) => /* @__PURE__ */ jsxs(
  "svg",
  {
    height: size,
    strokeLinejoin: "round",
    viewBox: "0 0 16 16",
    width: size,
    style: { color: "currentcolor" },
    children: [
      /* @__PURE__ */ jsx(
        "path",
        {
          d: "M2.5 0.5V0H3.5V0.5C3.5 1.60457 4.39543 2.5 5.5 2.5H6V3V3.5H5.5C4.39543 3.5 3.5 4.39543 3.5 5.5V6H3H2.5V5.5C2.5 4.39543 1.60457 3.5 0.5 3.5H0V3V2.5H0.5C1.60457 2.5 2.5 1.60457 2.5 0.5Z",
          fill: "currentColor"
        }
      ),
      /* @__PURE__ */ jsx(
        "path",
        {
          d: "M14.5 4.5V5H13.5V4.5C13.5 3.94772 13.0523 3.5 12.5 3.5H12V3V2.5H12.5C13.0523 2.5 13.5 2.05228 13.5 1.5V1H14H14.5V1.5C14.5 2.05228 14.9477 2.5 15.5 2.5H16V3V3.5H15.5C14.9477 3.5 14.5 3.94772 14.5 4.5Z",
          fill: "currentColor"
        }
      ),
      /* @__PURE__ */ jsx(
        "path",
        {
          d: "M8.40706 4.92939L8.5 4H9.5L9.59294 4.92939C9.82973 7.29734 11.7027 9.17027 14.0706 9.40706L15 9.5V10.5L14.0706 10.5929C11.7027 10.8297 9.82973 12.7027 9.59294 15.0706L9.5 16H8.5L8.40706 15.0706C8.17027 12.7027 6.29734 10.8297 3.92939 10.5929L3 10.5V9.5L3.92939 9.40706C6.29734 9.17027 8.17027 7.29734 8.40706 4.92939Z",
          fill: "currentColor"
        }
      )
    ]
  }
);
const CheckCircleFillIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M16 8C16 12.4183 12.4183 16 8 16C3.58172 16 0 12.4183 0 8C0 3.58172 3.58172 0 8 0C12.4183 0 16 3.58172 16 8ZM11.5303 6.53033L12.0607 6L11 4.93934L10.4697 5.46967L6.5 9.43934L5.53033 8.46967L5 7.93934L3.93934 9L4.46967 9.53033L5.96967 11.0303C6.26256 11.3232 6.73744 11.3232 7.03033 11.0303L11.5303 6.53033Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const GlobeIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M10.268 14.0934C11.9051 13.4838 13.2303 12.2333 13.9384 10.6469C13.1192 10.7941 12.2138 10.9111 11.2469 10.9925C11.0336 12.2005 10.695 13.2621 10.268 14.0934ZM8 16C12.4183 16 16 12.4183 16 8C16 3.58172 12.4183 0 8 0C3.58172 0 0 3.58172 0 8C0 12.4183 3.58172 16 8 16ZM8.48347 14.4823C8.32384 14.494 8.16262 14.5 8 14.5C7.83738 14.5 7.67616 14.494 7.51654 14.4823C7.5132 14.4791 7.50984 14.4759 7.50647 14.4726C7.2415 14.2165 6.94578 13.7854 6.67032 13.1558C6.41594 12.5744 6.19979 11.8714 6.04101 11.0778C6.67605 11.1088 7.33104 11.125 8 11.125C8.66896 11.125 9.32395 11.1088 9.95899 11.0778C9.80021 11.8714 9.58406 12.5744 9.32968 13.1558C9.05422 13.7854 8.7585 14.2165 8.49353 14.4726C8.49016 14.4759 8.4868 14.4791 8.48347 14.4823ZM11.4187 9.72246C12.5137 9.62096 13.5116 9.47245 14.3724 9.28806C14.4561 8.87172 14.5 8.44099 14.5 8C14.5 7.55901 14.4561 7.12828 14.3724 6.71194C13.5116 6.52755 12.5137 6.37904 11.4187 6.27753C11.4719 6.83232 11.5 7.40867 11.5 8C11.5 8.59133 11.4719 9.16768 11.4187 9.72246ZM10.1525 6.18401C10.2157 6.75982 10.25 7.36805 10.25 8C10.25 8.63195 10.2157 9.24018 10.1525 9.81598C9.46123 9.85455 8.7409 9.875 8 9.875C7.25909 9.875 6.53877 9.85455 5.84749 9.81598C5.7843 9.24018 5.75 8.63195 5.75 8C5.75 7.36805 5.7843 6.75982 5.84749 6.18401C6.53877 6.14545 7.25909 6.125 8 6.125C8.74091 6.125 9.46123 6.14545 10.1525 6.18401ZM11.2469 5.00748C12.2138 5.08891 13.1191 5.20593 13.9384 5.35306C13.2303 3.7667 11.9051 2.51622 10.268 1.90662C10.695 2.73788 11.0336 3.79953 11.2469 5.00748ZM8.48347 1.51771C8.4868 1.52089 8.49016 1.52411 8.49353 1.52737C8.7585 1.78353 9.05422 2.21456 9.32968 2.84417C9.58406 3.42562 9.80021 4.12856 9.95899 4.92219C9.32395 4.89118 8.66896 4.875 8 4.875C7.33104 4.875 6.67605 4.89118 6.04101 4.92219C6.19978 4.12856 6.41594 3.42562 6.67032 2.84417C6.94578 2.21456 7.2415 1.78353 7.50647 1.52737C7.50984 1.52411 7.51319 1.52089 7.51653 1.51771C7.67615 1.50597 7.83738 1.5 8 1.5C8.16262 1.5 8.32384 1.50597 8.48347 1.51771ZM5.73202 1.90663C4.0949 2.51622 2.76975 3.7667 2.06159 5.35306C2.88085 5.20593 3.78617 5.08891 4.75309 5.00748C4.96639 3.79953 5.30497 2.73788 5.73202 1.90663ZM4.58133 6.27753C3.48633 6.37904 2.48837 6.52755 1.62761 6.71194C1.54392 7.12828 1.5 7.55901 1.5 8C1.5 8.44099 1.54392 8.87172 1.62761 9.28806C2.48837 9.47245 3.48633 9.62096 4.58133 9.72246C4.52807 9.16768 4.5 8.59133 4.5 8C4.5 7.40867 4.52807 6.83232 4.58133 6.27753ZM4.75309 10.9925C3.78617 10.9111 2.88085 10.7941 2.06159 10.6469C2.76975 12.2333 4.0949 13.4838 5.73202 14.0934C5.30497 13.2621 4.96639 12.2005 4.75309 10.9925Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const LockIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M10 4.5V6H6V4.5C6 3.39543 6.89543 2.5 8 2.5C9.10457 2.5 10 3.39543 10 4.5ZM4.5 6V4.5C4.5 2.567 6.067 1 8 1C9.933 1 11.5 2.567 11.5 4.5V6H12.5H14V7.5V12.5C14 13.8807 12.8807 15 11.5 15H4.5C3.11929 15 2 13.8807 2 12.5V7.5V6H3.5H4.5ZM11.5 7.5H10H6H4.5H3.5V12.5C3.5 13.0523 3.94772 13.5 4.5 13.5H11.5C12.0523 13.5 12.5 13.0523 12.5 12.5V7.5H11.5Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const ShareIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M15 11.25V10.5H13.5V11.25V12.75C13.5 13.1642 13.1642 13.5 12.75 13.5H3.25C2.83579 13.5 2.5 13.1642 2.5 12.75L2.5 3.25C2.5 2.83579 2.83579 2.5 3.25 2.5H5.75H6.5V1H5.75H3.25C2.00736 1 1 2.00736 1 3.25V12.75C1 13.9926 2.00736 15 3.25 15H12.75C13.9926 15 15 13.9926 15 12.75V11.25ZM15 5.5L10.5 1V4C7.46243 4 5 6.46243 5 9.5V10L5.05855 9.91218C6.27146 8.09281 8.31339 7 10.5 7V10L15 5.5Z",
          fill: "currentColor"
        }
      )
    }
  );
};
const WarningIcon = ({ size = 16 }) => {
  return /* @__PURE__ */ jsx(
    "svg",
    {
      height: size,
      strokeLinejoin: "round",
      viewBox: "0 0 16 16",
      width: size,
      style: { color: "currentcolor" },
      children: /* @__PURE__ */ jsx(
        "path",
        {
          fillRule: "evenodd",
          clipRule: "evenodd",
          d: "M8.55846 0.5C9.13413 0.5 9.65902 0.829456 9.90929 1.34788L15.8073 13.5653C16.1279 14.2293 15.6441 15 14.9068 15H1.09316C0.355835 15 -0.127943 14.2293 0.192608 13.5653L6.09065 1.34787C6.34092 0.829454 6.86581 0.5 7.44148 0.5H8.55846ZM8.74997 4.75V5.5V8V8.75H7.24997V8V5.5V4.75H8.74997ZM7.99997 12C8.55226 12 8.99997 11.5523 8.99997 11C8.99997 10.4477 8.55226 10 7.99997 10C7.44769 10 6.99997 10.4477 6.99997 11C6.99997 11.5523 7.44769 12 7.99997 12Z",
          fill: "currentColor"
        }
      )
    }
  );
};
function AlertDialog({
  ...props
}) {
  return /* @__PURE__ */ jsx(AlertDialogPrimitive.Root, { "data-slot": "alert-dialog", ...props });
}
function AlertDialogPortal({
  ...props
}) {
  return /* @__PURE__ */ jsx(AlertDialogPrimitive.Portal, { "data-slot": "alert-dialog-portal", ...props });
}
function AlertDialogOverlay({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    AlertDialogPrimitive.Overlay,
    {
      "data-slot": "alert-dialog-overlay",
      className: cn(
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/50",
        className
      ),
      ...props
    }
  );
}
function AlertDialogContent({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsxs(AlertDialogPortal, { children: [
    /* @__PURE__ */ jsx(AlertDialogOverlay, {}),
    /* @__PURE__ */ jsx(
      AlertDialogPrimitive.Content,
      {
        "data-slot": "alert-dialog-content",
        className: cn(
          "bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-4 rounded-lg border p-6 shadow-lg duration-200 sm:max-w-lg",
          className
        ),
        ...props
      }
    )
  ] });
}
function AlertDialogHeader({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "alert-dialog-header",
      className: cn("flex flex-col gap-2 text-center sm:text-left", className),
      ...props
    }
  );
}
function AlertDialogFooter({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "alert-dialog-footer",
      className: cn(
        "flex flex-col-reverse gap-2 sm:flex-row sm:justify-end",
        className
      ),
      ...props
    }
  );
}
function AlertDialogTitle({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    AlertDialogPrimitive.Title,
    {
      "data-slot": "alert-dialog-title",
      className: cn("text-lg font-semibold", className),
      ...props
    }
  );
}
function AlertDialogDescription({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    AlertDialogPrimitive.Description,
    {
      "data-slot": "alert-dialog-description",
      className: cn("text-muted-foreground text-sm", className),
      ...props
    }
  );
}
function AlertDialogAction({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    AlertDialogPrimitive.Action,
    {
      className: cn(buttonVariants(), className),
      ...props
    }
  );
}
function AlertDialogCancel({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    AlertDialogPrimitive.Cancel,
    {
      className: cn(buttonVariants({ variant: "outline" }), className),
      ...props
    }
  );
}
function DropdownMenu({
  ...props
}) {
  return /* @__PURE__ */ jsx(DropdownMenuPrimitive.Root, { "data-slot": "dropdown-menu", ...props });
}
function DropdownMenuPortal({
  ...props
}) {
  return /* @__PURE__ */ jsx(DropdownMenuPrimitive.Portal, { "data-slot": "dropdown-menu-portal", ...props });
}
function DropdownMenuTrigger({
  ...props
}) {
  return /* @__PURE__ */ jsx(
    DropdownMenuPrimitive.Trigger,
    {
      "data-slot": "dropdown-menu-trigger",
      ...props
    }
  );
}
function DropdownMenuContent({
  className,
  sideOffset = 4,
  ...props
}) {
  return /* @__PURE__ */ jsx(DropdownMenuPrimitive.Portal, { children: /* @__PURE__ */ jsx(
    DropdownMenuPrimitive.Content,
    {
      "data-slot": "dropdown-menu-content",
      sideOffset,
      className: cn(
        "bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 max-h-(--radix-dropdown-menu-content-available-height) min-w-[8rem] origin-(--radix-dropdown-menu-content-transform-origin) overflow-x-hidden overflow-y-auto rounded-md border p-1 shadow-md",
        className
      ),
      ...props
    }
  ) });
}
function DropdownMenuItem({
  className,
  inset,
  variant = "default",
  ...props
}) {
  return /* @__PURE__ */ jsx(
    DropdownMenuPrimitive.Item,
    {
      "data-slot": "dropdown-menu-item",
      "data-inset": inset,
      "data-variant": variant,
      className: cn(
        "focus:bg-accent focus:text-accent-foreground data-[variant=destructive]:text-destructive data-[variant=destructive]:focus:bg-destructive/10 dark:data-[variant=destructive]:focus:bg-destructive/20 data-[variant=destructive]:focus:text-destructive data-[variant=destructive]:*:[svg]:!text-destructive [&_svg:not([class*='text-'])]:text-muted-foreground relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50 data-[inset]:pl-8 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        className
      ),
      ...props
    }
  );
}
function DropdownMenuSeparator({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    DropdownMenuPrimitive.Separator,
    {
      "data-slot": "dropdown-menu-separator",
      className: cn("bg-border -mx-1 my-1 h-px", className),
      ...props
    }
  );
}
function DropdownMenuSub({
  ...props
}) {
  return /* @__PURE__ */ jsx(DropdownMenuPrimitive.Sub, { "data-slot": "dropdown-menu-sub", ...props });
}
function DropdownMenuSubTrigger({
  className,
  inset,
  children,
  ...props
}) {
  return /* @__PURE__ */ jsxs(
    DropdownMenuPrimitive.SubTrigger,
    {
      "data-slot": "dropdown-menu-sub-trigger",
      "data-inset": inset,
      className: cn(
        "focus:bg-accent focus:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground flex cursor-default items-center rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-[inset]:pl-8",
        className
      ),
      ...props,
      children: [
        children,
        /* @__PURE__ */ jsx(ChevronRightIcon, { className: "ml-auto size-4" })
      ]
    }
  );
}
function DropdownMenuSubContent({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    DropdownMenuPrimitive.SubContent,
    {
      "data-slot": "dropdown-menu-sub-content",
      className: cn(
        "bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 min-w-[8rem] origin-(--radix-dropdown-menu-content-transform-origin) overflow-hidden rounded-md border p-1 shadow-lg",
        className
      ),
      ...props
    }
  );
}
function useChatDelete() {
  const submitter = useFetcher({ key: "delete-chat" });
  return (chatId) => {
    return submitter.submit(
      {
        id: chatId
      },
      {
        method: "delete",
        action: `/chat/${chatId}`
      }
    );
  };
}
function useChatVisibility({
  chatId
}) {
  return [
    "private",
    () => {
    }
    // Placeholder for setVisibilityType
  ];
}
function useChat({
  chat: chat2,
  onError
}) {
  const [input, setInput] = useState("");
  const [status, setStatus] = useState("ready");
  const [messages, setMessages] = useState([]);
  const submitter = useFetcher();
  const navigate = useNavigate();
  useEffect(() => {
    if (chat2.messages) {
      setMessages([...chat2.messages]);
    } else {
      setMessages([]);
    }
  }, [chat2.id]);
  useEffect(() => {
    if (chat2.id) {
      return events(chat2.id, (event) => {
        setMessages((prev) => {
          const messages2 = [...prev];
          appendProgress(messages2, event);
          return messages2;
        });
      });
    }
  }, [chat2.id]);
  async function handleSubmit({
    prompt,
    attachments
  }) {
    setStatus("submitted");
    const id = uuidv4();
    const chatURL = chat2.agentEditor ? `/chat/${chat2.id}/agent/${chat2.customAgent.id}/edit` : `/chat/${chat2.id}`;
    try {
      setInput("");
      setMessages((prev) => {
        return [
          ...prev,
          {
            id,
            role: "user",
            items: [
              {
                id: id + "_0",
                type: "text",
                text: prompt || input
              }
            ]
          }
        ];
      });
      setStatus("streaming");
      await submitter.submit(
        {
          id,
          prompt: prompt || input,
          clone: !!chat2.readonly
        },
        {
          method: "post",
          action: chatURL
        }
      );
      setStatus("ready");
      await navigate(chatURL);
    } catch (error) {
      console.error("Error submitting chat message:", error);
      setStatus("error");
      if (onError) {
        onError(error);
      }
    }
  }
  function getVisibilityType() {
    return chat2.visibility || "private";
  }
  async function setVisibilityType(visibility) {
    if (typeof visibility === "function") {
      visibility = visibility(getVisibilityType());
    }
    await navigate(`/chat/${chat2.id}`);
    await submitter.submit(
      {
        visibility
      },
      {
        method: "post",
        action: `/chat/${chat2.id}`
      }
    );
  }
  async function setCurrentAgent(agent2) {
    await navigate(`/chat/${chat2.id}`);
    if (typeof agent2 === "function") {
      agent2 = agent2(chat2.currentAgent || "");
    }
    await submitter.submit(
      {
        agent: agent2
      },
      {
        method: "post",
        action: `/chat/${chat2.id}`
      }
    );
  }
  return {
    messages,
    setMessages,
    updateMessage: () => {
    },
    // Placeholder for updateMessage
    handleSubmit,
    input,
    setInput,
    status,
    stop: () => {
    },
    // Placeholder for stop
    reload: async () => null,
    // Placeholder for reload
    currentAgent: chat2.currentAgent || "",
    setCurrentAgent,
    agents: chat2.agents || {},
    visibilityType: getVisibilityType(),
    setVisibilityType,
    votes: chat2.votes,
    supportsCustomAgents: true
  };
}
const PureChatItem = ({
  chat: chat2,
  isActive,
  onDelete,
  setOpenMobile,
  user
}) => {
  const [visibilityType, setVisibilityType] = useChatVisibility({
    chatId: chat2.id
  });
  return /* @__PURE__ */ jsxs(SidebarMenuItem, { children: [
    /* @__PURE__ */ jsx(SidebarMenuButton, { asChild: true, isActive, children: /* @__PURE__ */ jsx(Link, { to: `/chat/${chat2.id}`, onClick: () => setOpenMobile(false), children: /* @__PURE__ */ jsx("span", { children: chat2.title }) }) }),
    /* @__PURE__ */ jsxs(DropdownMenu, { modal: true, children: [
      /* @__PURE__ */ jsx(DropdownMenuTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(
        SidebarMenuAction,
        {
          className: "data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground mr-0.5",
          showOnHover: !isActive,
          children: [
            /* @__PURE__ */ jsx(MoreHorizontalIcon, {}),
            /* @__PURE__ */ jsx("span", { className: "sr-only", children: "More" })
          ]
        }
      ) }),
      /* @__PURE__ */ jsxs(DropdownMenuContent, { side: "bottom", align: "end", children: [
        user && /* @__PURE__ */ jsxs(DropdownMenuSub, { children: [
          /* @__PURE__ */ jsx(DropdownMenuSubTrigger, { className: "cursor-pointer", children: /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-center", children: [
            /* @__PURE__ */ jsx(ShareIcon, {}),
            /* @__PURE__ */ jsx("span", { children: "Share" })
          ] }) }),
          /* @__PURE__ */ jsx(DropdownMenuPortal, { children: /* @__PURE__ */ jsxs(DropdownMenuSubContent, { children: [
            /* @__PURE__ */ jsxs(
              DropdownMenuItem,
              {
                className: "cursor-pointer flex-row justify-between",
                onClick: () => {
                  setVisibilityType("private");
                },
                children: [
                  /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-center", children: [
                    /* @__PURE__ */ jsx(LockIcon, { size: 12 }),
                    /* @__PURE__ */ jsx("span", { children: "Private" })
                  ] }),
                  visibilityType === "private" ? /* @__PURE__ */ jsx(CheckCircleFillIcon, {}) : null
                ]
              }
            ),
            /* @__PURE__ */ jsxs(
              DropdownMenuItem,
              {
                className: "cursor-pointer flex-row justify-between",
                onClick: () => {
                  setVisibilityType("public");
                },
                children: [
                  /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-center", children: [
                    /* @__PURE__ */ jsx(GlobeIcon, {}),
                    /* @__PURE__ */ jsx("span", { children: "Public" })
                  ] }),
                  visibilityType === "public" ? /* @__PURE__ */ jsx(CheckCircleFillIcon, {}) : null
                ]
              }
            )
          ] }) })
        ] }),
        /* @__PURE__ */ jsxs(
          DropdownMenuItem,
          {
            className: "cursor-pointer text-destructive focus:bg-destructive/15 focus:text-destructive dark:text-red-500",
            onSelect: () => onDelete(chat2.id),
            children: [
              /* @__PURE__ */ jsx(TrashIcon, {}),
              /* @__PURE__ */ jsx("span", { children: "Delete" })
            ]
          }
        )
      ] })
    ] })
  ] });
};
const ChatItem = memo(PureChatItem, (prevProps, nextProps) => {
  if (prevProps.isActive !== nextProps.isActive) return false;
  if (prevProps.chat.title !== nextProps.chat.title) return false;
  if (prevProps.user !== nextProps.user) return false;
  return true;
});
function groupChatsByDate(chats) {
  const now = /* @__PURE__ */ new Date();
  const oneWeekAgo = subWeeks(now, 1);
  const oneMonthAgo = subMonths(now, 1);
  return chats.reduce(
    (groups, chat2) => {
      const chatDate = new Date(chat2.created);
      if (!chat2.title) {
        return groups;
      }
      if (isToday(chatDate)) {
        groups.today.push(chat2);
      } else if (isYesterday(chatDate)) {
        groups.yesterday.push(chat2);
      } else if (chatDate > oneWeekAgo) {
        groups.lastWeek.push(chat2);
      } else if (chatDate > oneMonthAgo) {
        groups.lastMonth.push(chat2);
      } else {
        groups.older.push(chat2);
      }
      return groups;
    },
    {
      today: [],
      yesterday: [],
      lastWeek: [],
      lastMonth: [],
      older: []
    }
  );
}
function SidebarHistory({
  chatId,
  chats,
  deleteChat: deleteChat2,
  user
}) {
  const { setOpenMobile } = useSidebar();
  const nav = useNavigate();
  const [deleteId, setDeleteId] = useState(null);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const handleDelete = async () => {
    const deletePromise = deleteChat2(deleteId);
    toast$1.promise(deletePromise, {
      loading: "Deleting chat...",
      success: () => {
        return "Chat deleted successfully";
      },
      error: (e) => `Failed to delete chat ${e}`
    });
    setShowDeleteDialog(false);
    if (deleteId === chatId) {
      nav("/");
    }
  };
  if (chats.length === 0) {
    return /* @__PURE__ */ jsx(SidebarGroup, { children: /* @__PURE__ */ jsx(SidebarGroupContent, { children: /* @__PURE__ */ jsx("div", { className: "px-2 text-zinc-500 w-full flex flex-row justify-center items-center text-sm gap-2", children: "Your conversations will appear here once you start chatting!" }) }) });
  }
  return /* @__PURE__ */ jsxs(Fragment, { children: [
    /* @__PURE__ */ jsx(SidebarGroup, { children: /* @__PURE__ */ jsx(SidebarGroupContent, { children: /* @__PURE__ */ jsx(SidebarMenu, { children: (() => {
      const groupedChats = groupChatsByDate(chats);
      return /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-6", children: [
        groupedChats.today.length > 0 && /* @__PURE__ */ jsxs("div", { children: [
          /* @__PURE__ */ jsx("div", { className: "px-2 py-1 text-xs text-sidebar-foreground/50", children: "Today" }),
          groupedChats.today.map((chat2) => /* @__PURE__ */ jsx(
            ChatItem,
            {
              chat: chat2,
              user,
              isActive: chat2.id === chatId,
              onDelete: (chatId2) => {
                setDeleteId(chatId2);
                setShowDeleteDialog(true);
              },
              setOpenMobile
            },
            chat2.id
          ))
        ] }),
        groupedChats.yesterday.length > 0 && /* @__PURE__ */ jsxs("div", { children: [
          /* @__PURE__ */ jsx("div", { className: "px-2 py-1 text-xs text-sidebar-foreground/50", children: "Yesterday" }),
          groupedChats.yesterday.map((chat2) => /* @__PURE__ */ jsx(
            ChatItem,
            {
              chat: chat2,
              user,
              isActive: chat2.id === chatId,
              onDelete: (chatId2) => {
                setDeleteId(chatId2);
                setShowDeleteDialog(true);
              },
              setOpenMobile
            },
            chat2.id
          ))
        ] }),
        groupedChats.lastWeek.length > 0 && /* @__PURE__ */ jsxs("div", { children: [
          /* @__PURE__ */ jsx("div", { className: "px-2 py-1 text-xs text-sidebar-foreground/50", children: "Last 7 days" }),
          groupedChats.lastWeek.map((chat2) => /* @__PURE__ */ jsx(
            ChatItem,
            {
              chat: chat2,
              user,
              isActive: chat2.id === chatId,
              onDelete: (chatId2) => {
                setDeleteId(chatId2);
                setShowDeleteDialog(true);
              },
              setOpenMobile
            },
            chat2.id
          ))
        ] }),
        groupedChats.lastMonth.length > 0 && /* @__PURE__ */ jsxs("div", { children: [
          /* @__PURE__ */ jsx("div", { className: "px-2 py-1 text-xs text-sidebar-foreground/50", children: "Last 30 days" }),
          groupedChats.lastMonth.map((chat2) => /* @__PURE__ */ jsx(
            ChatItem,
            {
              chat: chat2,
              user,
              isActive: chat2.id === chatId,
              onDelete: (chatId2) => {
                setDeleteId(chatId2);
                setShowDeleteDialog(true);
              },
              setOpenMobile
            },
            chat2.id
          ))
        ] }),
        groupedChats.older.length > 0 && /* @__PURE__ */ jsxs("div", { children: [
          /* @__PURE__ */ jsx("div", { className: "px-2 py-1 text-xs text-sidebar-foreground/50", children: "Older than last month" }),
          groupedChats.older.map((chat2) => /* @__PURE__ */ jsx(
            ChatItem,
            {
              chat: chat2,
              user,
              isActive: chat2.id === chatId,
              onDelete: (chatId2) => {
                setDeleteId(chatId2);
                setShowDeleteDialog(true);
              },
              setOpenMobile
            },
            chat2.id
          ))
        ] })
      ] });
    })() }) }) }),
    /* @__PURE__ */ jsx(AlertDialog, { open: showDeleteDialog, onOpenChange: setShowDeleteDialog, children: /* @__PURE__ */ jsxs(AlertDialogContent, { children: [
      /* @__PURE__ */ jsxs(AlertDialogHeader, { children: [
        /* @__PURE__ */ jsx(AlertDialogTitle, { children: "Are you absolutely sure?" }),
        /* @__PURE__ */ jsx(AlertDialogDescription, { children: "This action cannot be undone. This will permanently delete your chat and remove it from our servers." })
      ] }),
      /* @__PURE__ */ jsxs(AlertDialogFooter, { children: [
        /* @__PURE__ */ jsx(AlertDialogCancel, { children: "Cancel" }),
        /* @__PURE__ */ jsx(AlertDialogAction, { onClick: handleDelete, children: "Continue" })
      ] })
    ] }) })
  ] });
}
function SidebarUserNav({ user }) {
  const navigate = useNavigate();
  const [resolvedTheme, setTheme] = useTheme();
  const isGuest = !!user.id;
  return /* @__PURE__ */ jsx(SidebarMenu, { children: /* @__PURE__ */ jsx(SidebarMenuItem, { children: /* @__PURE__ */ jsxs(DropdownMenu, { children: [
    /* @__PURE__ */ jsx(DropdownMenuTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(
      SidebarMenuButton,
      {
        "data-testid": "user-nav-button",
        className: "data-[state=open]:bg-sidebar-accent bg-background data-[state=open]:text-sidebar-accent-foreground h-10",
        children: [
          /* @__PURE__ */ jsx(
            "img",
            {
              src: `https://avatar.vercel.sh/${user.email}`,
              alt: user.email ?? "User Avatar",
              width: 24,
              height: 24,
              className: "rounded-full"
            }
          ),
          /* @__PURE__ */ jsx("span", { "data-testid": "user-email", className: "truncate", children: isGuest ? "Guest" : user == null ? void 0 : user.email }),
          /* @__PURE__ */ jsx(ChevronUp, { className: "ml-auto" })
        ]
      }
    ) }),
    /* @__PURE__ */ jsxs(
      DropdownMenuContent,
      {
        "data-testid": "user-nav-menu",
        side: "top",
        className: "w-[--radix-popper-anchor-width]",
        children: [
          /* @__PURE__ */ jsx(
            DropdownMenuItem,
            {
              "data-testid": "user-nav-item-theme",
              className: "cursor-pointer",
              onSelect: () => setTheme(resolvedTheme === "dark" ? Theme.LIGHT : Theme.DARK),
              children: `Toggle ${resolvedTheme === "light" ? "dark" : "light"} mode`
            }
          ),
          /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
          /* @__PURE__ */ jsx(DropdownMenuItem, { asChild: true, "data-testid": "user-nav-item-auth", children: /* @__PURE__ */ jsx(
            "button",
            {
              type: "button",
              className: "w-full cursor-pointer",
              onClick: () => {
                if (isGuest) {
                  navigate("/login");
                } else {
                  navigate("/logout");
                }
              },
              children: isGuest ? "Login to your account" : "Sign out"
            }
          ) })
        ]
      }
    )
  ] }) }) });
}
function SidebarAgent({
  chatId,
  agent: agent2,
  user,
  isActive = false
}) {
  const { setOpenMobile } = useSidebar();
  const navigate = useNavigate();
  return /* @__PURE__ */ jsxs(SidebarMenuItem, { children: [
    /* @__PURE__ */ jsx(SidebarMenuButton, { asChild: true, isActive, children: /* @__PURE__ */ jsx(Link, { to: `/agent/${agent2.id}`, onClick: () => setOpenMobile(false), children: /* @__PURE__ */ jsx("span", { children: agent2.name }) }) }, agent2.id),
    /* @__PURE__ */ jsxs(DropdownMenu, { modal: true, children: [
      /* @__PURE__ */ jsx(DropdownMenuTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(
        SidebarMenuAction,
        {
          className: "data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground mr-0.5",
          showOnHover: !isActive,
          children: [
            /* @__PURE__ */ jsx(MoreHorizontalIcon, {}),
            /* @__PURE__ */ jsx("span", { className: "sr-only", children: "More" })
          ]
        }
      ) }),
      /* @__PURE__ */ jsxs(DropdownMenuContent, { side: "bottom", align: "end", children: [
        user && /* @__PURE__ */ jsxs(DropdownMenuSub, { children: [
          /* @__PURE__ */ jsx(DropdownMenuSubTrigger, { className: "cursor-pointer", children: /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-center", children: [
            /* @__PURE__ */ jsx(ShareIcon, {}),
            /* @__PURE__ */ jsx("span", { children: "Share" })
          ] }) }),
          /* @__PURE__ */ jsx(DropdownMenuPortal, { children: /* @__PURE__ */ jsxs(DropdownMenuSubContent, { children: [
            /* @__PURE__ */ jsxs(
              DropdownMenuItem,
              {
                className: "cursor-pointer flex-row justify-between",
                onClick: () => {
                },
                children: [
                  /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-center", children: [
                    /* @__PURE__ */ jsx(LockIcon, { size: 12 }),
                    /* @__PURE__ */ jsx("span", { children: "Private" })
                  ] }),
                  !agent2.isPublic ? /* @__PURE__ */ jsx(CheckCircleFillIcon, {}) : null
                ]
              }
            ),
            /* @__PURE__ */ jsxs(
              DropdownMenuItem,
              {
                className: "cursor-pointer flex-row justify-between",
                onClick: () => {
                },
                children: [
                  /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-center", children: [
                    /* @__PURE__ */ jsx(GlobeIcon, {}),
                    /* @__PURE__ */ jsx("span", { children: "Public" })
                  ] }),
                  agent2.isPublic ? /* @__PURE__ */ jsx(CheckCircleFillIcon, {}) : null
                ]
              }
            )
          ] }) })
        ] }),
        /* @__PURE__ */ jsxs(
          DropdownMenuItem,
          {
            className: "cursor-pointer",
            onSelect: () => navigate(`/agent/${agent2.id}/edit`),
            children: [
              /* @__PURE__ */ jsx(PencilIcon, {}),
              /* @__PURE__ */ jsx("span", { children: "Edit" })
            ]
          }
        ),
        chatId && /* @__PURE__ */ jsx(
          DropdownMenuItem,
          {
            className: "cursor-pointer text-destructive focus:bg-destructive/15 focus:text-destructive dark:text-red-500",
            asChild: true,
            children: /* @__PURE__ */ jsxs(Form, { method: "DELETE", action: `/chat/${chatId}/delete-agent`, children: [
              /* @__PURE__ */ jsx("input", { type: "hidden", name: "agentId", value: agent2.id }),
              /* @__PURE__ */ jsxs(
                "button",
                {
                  type: "submit",
                  className: "flex items-center gap-2 w-full text-left",
                  children: [
                    /* @__PURE__ */ jsx(TrashIcon, {}),
                    /* @__PURE__ */ jsx("span", { children: "Delete" })
                  ]
                }
              )
            ] })
          }
        )
      ] })
    ] })
  ] });
}
function SidebarAgents({
  chatId,
  customAgents,
  customAgent,
  user
}) {
  return /* @__PURE__ */ jsxs(SidebarGroup, { children: [
    /* @__PURE__ */ jsx(SidebarGroupLabel, { className: "flex items-center justify-between", children: /* @__PURE__ */ jsx("span", { children: "Agents" }) }),
    /* @__PURE__ */ jsx(Form, { action: "/agent/new", method: "post", className: "contents", children: /* @__PURE__ */ jsx(
      SidebarMenuAction,
      {
        type: "submit",
        className: "mr-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200",
        children: /* @__PURE__ */ jsx(PlusIcon, {})
      }
    ) }),
    /* @__PURE__ */ jsx(SidebarGroupContent, { children: /* @__PURE__ */ jsx(SidebarMenu, { children: customAgents == null ? void 0 : customAgents.map((agent2) => /* @__PURE__ */ jsx(
      SidebarAgent,
      {
        chatId,
        agent: agent2,
        user,
        isActive: agent2.id === (customAgent == null ? void 0 : customAgent.id)
      },
      agent2.id
    )) }) })
  ] });
}
function AppSidebar({
  user,
  chatId,
  chats,
  customAgents,
  customAgent
}) {
  const deleteChat2 = useChatDelete();
  const navigate = useNavigate();
  const { setOpenMobile } = useSidebar();
  return /* @__PURE__ */ jsxs(Sidebar, { className: "group-data-[side=left]:border-r-0", children: [
    /* @__PURE__ */ jsx(SidebarHeader, { children: /* @__PURE__ */ jsx(SidebarMenu, { children: /* @__PURE__ */ jsxs("div", { className: "flex flex-row justify-between items-center", children: [
      /* @__PURE__ */ jsx(
        Link,
        {
          to: "/",
          onClick: () => {
            setOpenMobile(false);
          },
          className: "flex flex-row gap-3 items-center",
          children: /* @__PURE__ */ jsx("span", { className: "text-lg font-semibold px-2 hover:bg-muted rounded-md cursor-pointer", children: "Nanobot" })
        }
      ),
      /* @__PURE__ */ jsxs(Tooltip, { children: [
        /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: /* @__PURE__ */ jsx(
          Button,
          {
            variant: "ghost",
            type: "button",
            className: "p-2 h-fit",
            onClick: () => {
              setOpenMobile(false);
              navigate("/");
            },
            children: /* @__PURE__ */ jsx(PlusIcon, {})
          }
        ) }),
        /* @__PURE__ */ jsx(TooltipContent, { align: "end", children: "New Chat" })
      ] })
    ] }) }) }),
    /* @__PURE__ */ jsxs(SidebarContent, { children: [
      customAgents && customAgents.length > 0 && /* @__PURE__ */ jsx(
        SidebarAgents,
        {
          chatId,
          customAgents,
          customAgent,
          user
        }
      ),
      /* @__PURE__ */ jsx(
        SidebarHistory,
        {
          chats,
          chatId,
          deleteChat: deleteChat2,
          user
        }
      ),
      !(customAgents == null ? void 0 : customAgents.length) && /* @__PURE__ */ jsx(
        SidebarAgents,
        {
          chatId,
          customAgents,
          customAgent,
          user
        }
      )
    ] }),
    /* @__PURE__ */ jsx(SidebarFooter, { children: user && /* @__PURE__ */ jsx(SidebarUserNav, { user }) })
  ] });
}
async function loader({
  request,
  params
}) {
  const ctx = getContext();
  const chat2 = await getChat(ctx, params.id || "new", {
    agentId: params.agentId
  });
  const threads = await listChats(ctx, chat2.id);
  const customAgents = await listCustomAgents(ctx, chat2.id);
  const cookie = await sidebarCookie.parse(request.headers.get("Cookie"));
  const currentChatMeta = threads.chats.find((c) => c.id === chat2.id);
  if (currentChatMeta) {
    chat2.visibility = currentChatMeta.visibility;
    chat2.readonly = currentChatMeta.readonly;
  }
  return {
    threads,
    chat: chat2,
    customAgents,
    sidebar: cookie || "true",
    user: {
      name: "User",
      email: "user@example.com"
    }
  };
}
const layout = UNSAFE_withComponentProps(function Layout({
  loaderData: {
    chat: chat2,
    threads,
    sidebar,
    customAgents,
    user
  }
}) {
  var _a;
  const isCollapsed = sidebar !== "true";
  return /* @__PURE__ */ jsxs(SidebarProvider, {
    defaultOpen: !isCollapsed,
    children: [/* @__PURE__ */ jsx(AppSidebar, {
      chats: threads.chats,
      chatId: chat2.id,
      customAgents: (_a = customAgents.customAgents) == null ? void 0 : _a.filter((c) => !!c.name),
      customAgent: chat2.customAgent,
      user
    }), /* @__PURE__ */ jsx(SidebarInset, {
      children: /* @__PURE__ */ jsx(Outlet, {})
    })]
  });
});
const route7 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  default: layout,
  loader
}, Symbol.toStringTag, { value: "Module" }));
function AgentSelector({
  currentAgent,
  setCurrentAgent,
  agents,
  className
}) {
  const [open, setOpen] = useState(false);
  const currentAgentMeta = agents == null ? void 0 : agents[currentAgent];
  if (!currentAgentMeta) {
    return null;
  }
  return /* @__PURE__ */ jsxs(DropdownMenu, { open, onOpenChange: setOpen, children: [
    /* @__PURE__ */ jsx(
      DropdownMenuTrigger,
      {
        asChild: true,
        className: cn(
          "w-fit data-[state=open]:bg-accent data-[state=open]:text-accent-foreground",
          className
        ),
        children: /* @__PURE__ */ jsxs(
          Button,
          {
            "data-testid": "model-selector",
            variant: "outline",
            className: "md:px-2 md:h-[34px]",
            children: [
              currentAgentMeta == null ? void 0 : currentAgentMeta.name,
              /* @__PURE__ */ jsx(ChevronDownIcon, {})
            ]
          }
        )
      }
    ),
    /* @__PURE__ */ jsx(DropdownMenuContent, { align: "start", className: "min-w-[300px]", children: Object.entries(agents).map(([id, agent2]) => {
      return /* @__PURE__ */ jsx(
        DropdownMenuItem,
        {
          "data-testid": `model-selector-item-${id}`,
          onSelect: () => {
            setOpen(false);
            startTransition(() => {
              setCurrentAgent(id);
            });
          },
          "data-active": id === currentAgent,
          asChild: true,
          children: /* @__PURE__ */ jsxs(
            "button",
            {
              type: "button",
              className: "gap-4 group/item flex flex-row justify-between items-center w-full",
              children: [
                /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-1 items-start", children: [
                  /* @__PURE__ */ jsx("div", { children: agent2.name }),
                  /* @__PURE__ */ jsx("div", { className: "text-xs text-muted-foreground", children: agent2.description })
                ] }),
                /* @__PURE__ */ jsx("div", { className: "text-foreground dark:text-foreground opacity-0 group-data-[active=true]/item:opacity-100", children: /* @__PURE__ */ jsx(CheckCircleFillIcon, {}) })
              ]
            }
          )
        },
        id
      );
    }) })
  ] });
}
function SidebarToggle({
  className
}) {
  const { toggleSidebar } = useSidebar();
  return /* @__PURE__ */ jsxs(Tooltip, { children: [
    /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: /* @__PURE__ */ jsx(
      Button,
      {
        "data-testid": "sidebar-toggle-button",
        onClick: toggleSidebar,
        variant: "outline",
        className: cn("md:px-2 md:h-fit", className),
        children: /* @__PURE__ */ jsx(SidebarLeftIcon, { size: 16 })
      }
    ) }),
    /* @__PURE__ */ jsx(TooltipContent, { align: "start", children: "Toggle Sidebar" })
  ] });
}
function PureChatHeader({
  currentAgent,
  setCurrentAgent,
  agents,
  customAgent,
  isReadonly
}) {
  const navigate = useNavigate();
  const { open } = useSidebar();
  const { width: windowWidth } = useWindowSize();
  return /* @__PURE__ */ jsxs("header", { className: "flex sticky top-0 bg-background py-1.5 items-center px-2 md:px-2 gap-2", children: [
    /* @__PURE__ */ jsx(SidebarToggle, {}),
    (!open || windowWidth < 768) && /* @__PURE__ */ jsxs(Tooltip, { children: [
      /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(
        Button,
        {
          variant: "outline",
          className: "order-2 md:order-1 md:px-2 px-2 md:h-fit ml-auto md:ml-0",
          onClick: () => {
            if (customAgent == null ? void 0 : customAgent.id) {
              navigate(`/agent/${customAgent == null ? void 0 : customAgent.id}`);
            } else {
              navigate("/");
            }
          },
          children: [
            /* @__PURE__ */ jsx(PlusIcon, {}),
            /* @__PURE__ */ jsx("span", { className: "md:sr-only", children: "New Chat" })
          ]
        }
      ) }),
      /* @__PURE__ */ jsx(TooltipContent, { children: "New Chat" })
    ] }),
    !isReadonly && !customAgent && Object.keys(agents).length > 1 && /* @__PURE__ */ jsx(
      AgentSelector,
      {
        currentAgent,
        setCurrentAgent,
        agents,
        className: "order-1 md:order-2"
      }
    )
  ] });
}
const ChatHeader = memo(PureChatHeader, (prevProps, nextProps) => {
  var _a, _b;
  if (((_a = prevProps.customAgent) == null ? void 0 : _a.id) !== ((_b = nextProps.customAgent) == null ? void 0 : _b.id)) return false;
  if (Object.keys(prevProps.agents || {}) !== Object.keys(nextProps.agents || {}))
    return false;
  return prevProps.currentAgent === nextProps.currentAgent;
});
const PreviewAttachment = ({
  attachment: { uri },
  isUploading = false
}) => {
  return /* @__PURE__ */ jsx("div", { "data-testid": "input-attachment-preview", className: "flex flex-col gap-2", children: /* @__PURE__ */ jsxs("div", { className: "w-20 h-16 aspect-video bg-muted rounded-md relative flex flex-col items-center justify-center", children: [
    /* @__PURE__ */ jsx(
      "img",
      {
        src: uri,
        alt: "An image attachment",
        className: "rounded-md size-full object-cover"
      },
      uri
    ),
    (isUploading || !uri) && /* @__PURE__ */ jsx(
      "div",
      {
        "data-testid": "input-attachment-loader",
        className: "animate-spin absolute text-zinc-500",
        children: /* @__PURE__ */ jsx(LoaderIcon, {})
      }
    )
  ] }) });
};
function Textarea({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "textarea",
    {
      "data-slot": "textarea",
      className: cn(
        "border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 flex field-sizing-content min-h-16 w-full rounded-md border bg-transparent px-3 py-2 text-base shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
        className
      ),
      ...props
    }
  );
}
function PureSuggestedActions({ chatId, handleSubmit }) {
  const suggestedActions = [
    {
      title: "What are the advantages",
      label: "of using Next.js?",
      action: "What are the advantages of using Next.js?"
    },
    {
      title: "Write code to",
      label: `demonstrate djikstra's algorithm`,
      action: `Write code to demonstrate djikstra's algorithm`
    },
    {
      title: "Help me write an essay",
      label: `about silicon valley`,
      action: `Help me write an essay about silicon valley`
    },
    {
      title: "What is the weather",
      label: "in San Francisco?",
      action: "What is the weather in San Francisco?"
    }
  ];
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-testid": "suggested-actions",
      className: "grid sm:grid-cols-2 gap-2 w-full",
      children: suggestedActions.map((suggestedAction, index2) => /* @__PURE__ */ jsx(
        motion.div,
        {
          initial: { opacity: 0, y: 20 },
          animate: { opacity: 1, y: 0 },
          exit: { opacity: 0, y: 20 },
          transition: { delay: 0.05 * index2 },
          className: index2 > 1 ? "hidden sm:block" : "block",
          children: /* @__PURE__ */ jsxs(
            Button,
            {
              variant: "ghost",
              onClick: async () => {
                window.history.replaceState({}, "", `/chat/${chatId}`);
                handleSubmit(suggestedAction.action);
              },
              className: "text-left border rounded-xl px-4 py-3.5 text-sm flex-1 gap-1 sm:flex-col w-full h-auto justify-start items-start",
              children: [
                /* @__PURE__ */ jsx("span", { className: "font-medium", children: suggestedAction.title }),
                /* @__PURE__ */ jsx("span", { className: "text-muted-foreground", children: suggestedAction.label })
              ]
            }
          )
        },
        `suggested-action-${suggestedAction.title}-${index2}`
      ))
    }
  );
}
const SuggestedActions = memo(
  PureSuggestedActions,
  (prevProps, nextProps) => {
    if (prevProps.chatId !== nextProps.chatId) return false;
    return true;
  }
);
function useScrollToBottom() {
  const containerRef = useRef(null);
  const endRef = useRef(null);
  const { data: isAtBottom = false, mutate: setIsAtBottom } = useSWR(
    "messages:is-at-bottom",
    null,
    { fallbackData: false }
  );
  const { data: scrollBehavior = false, mutate: setScrollBehavior } = useSWR("messages:should-scroll", null, { fallbackData: false });
  useEffect(() => {
    var _a;
    if (scrollBehavior) {
      (_a = endRef.current) == null ? void 0 : _a.scrollIntoView({ behavior: scrollBehavior });
      setScrollBehavior(false);
    }
  }, [setScrollBehavior, scrollBehavior]);
  const scrollToBottom = useCallback(
    (scrollBehavior2 = "smooth") => {
      setScrollBehavior(scrollBehavior2);
    },
    [setScrollBehavior]
  );
  function onViewportEnter() {
    setIsAtBottom(true);
  }
  function onViewportLeave() {
    setIsAtBottom(false);
  }
  return {
    containerRef,
    endRef,
    isAtBottom,
    scrollToBottom,
    onViewportEnter,
    onViewportLeave
  };
}
function PureMultimodalInput({
  chatId,
  input,
  setInput,
  status,
  stop,
  attachments,
  setAttachments,
  messages,
  setMessages,
  handleSubmit,
  className
}) {
  const textareaRef = useRef(null);
  const { width } = useWindowSize();
  const fetcher = useFetcher();
  useEffect(() => {
    if (textareaRef.current) {
      adjustHeight();
    }
  }, []);
  const adjustHeight = () => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight + 2}px`;
    }
  };
  const resetHeight = () => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = "98px";
    }
  };
  const [localStorageInput, setLocalStorageInput] = useLocalStorage(
    "input",
    ""
  );
  useEffect(() => {
    if (textareaRef.current) {
      const domValue = textareaRef.current.value;
      const finalValue = domValue || localStorageInput || "";
      setInput(finalValue);
      adjustHeight();
    }
  }, []);
  useEffect(() => {
    setLocalStorageInput(input);
  }, [input, setLocalStorageInput]);
  const handleInput = (event) => {
    setInput(event.target.value);
    adjustHeight();
  };
  const fileInputRef = useRef(null);
  const [uploadQueue, setUploadQueue] = useState([]);
  const submitForm = useCallback(() => {
    var _a;
    handleSubmit({
      attachments
    });
    setAttachments([]);
    setLocalStorageInput("");
    resetHeight();
    if (width && width > 768) {
      (_a = textareaRef.current) == null ? void 0 : _a.focus();
    }
  }, [
    attachments,
    handleSubmit,
    setAttachments,
    setLocalStorageInput,
    width,
    chatId
  ]);
  const uploadFile = async (file) => {
    const formData = new FormData();
    formData.append("file", file);
    try {
      await fetcher.submit(formData, {
        method: "POST",
        action: "/api/files/upload"
      });
      if (fetcher.data.error) {
        toast$1.error(fetcher.data.error);
      }
      return fetcher.data;
    } catch (error) {
      toast$1.error(`Failed to upload file, please try again! ${error}`);
    }
  };
  const handleFileChange = useCallback(
    async (event) => {
      const files = Array.from(event.target.files || []);
      setUploadQueue(files.map((file) => file.name));
      try {
        const uploadPromises = files.map((file) => uploadFile(file));
        const uploadedAttachments = await Promise.all(uploadPromises);
        const successfullyUploadedAttachments = uploadedAttachments.filter(
          (attachment) => attachment !== void 0
        );
        setAttachments((currentAttachments) => [
          ...currentAttachments,
          ...successfullyUploadedAttachments
        ]);
      } catch (error) {
        console.error("Error uploading files!", error);
      } finally {
        setUploadQueue([]);
      }
    },
    [setAttachments]
  );
  const { isAtBottom, scrollToBottom } = useScrollToBottom();
  useEffect(() => {
    if (status === "submitted") {
      scrollToBottom();
    }
  }, [status, scrollToBottom]);
  return /* @__PURE__ */ jsxs("div", { className: "relative w-full flex flex-col gap-4", children: [
    /* @__PURE__ */ jsx(AnimatePresence, { children: !isAtBottom && /* @__PURE__ */ jsx(
      motion.div,
      {
        initial: { opacity: 0, y: 10 },
        animate: { opacity: 1, y: 0 },
        exit: { opacity: 0, y: 10 },
        transition: { type: "spring", stiffness: 300, damping: 20 },
        className: "absolute left-1/2 bottom-28 -translate-x-1/2 z-50",
        children: /* @__PURE__ */ jsx(
          Button,
          {
            "data-testid": "scroll-to-bottom-button",
            className: "rounded-full",
            size: "icon",
            variant: "outline",
            onClick: (event) => {
              event.preventDefault();
              scrollToBottom();
            },
            children: /* @__PURE__ */ jsx(ArrowDown, {})
          }
        )
      }
    ) }),
    messages.length === 0 && attachments.length === 0 && uploadQueue.length === 0 && /* @__PURE__ */ jsx(SuggestedActions, { handleSubmit, chatId }),
    /* @__PURE__ */ jsx(
      "input",
      {
        type: "file",
        className: "fixed -top-4 -left-4 size-0.5 opacity-0 pointer-events-none",
        ref: fileInputRef,
        multiple: true,
        onChange: handleFileChange,
        tabIndex: -1
      }
    ),
    (attachments.length > 0 || uploadQueue.length > 0) && /* @__PURE__ */ jsxs(
      "div",
      {
        "data-testid": "attachments-preview",
        className: "flex flex-row gap-2 overflow-x-scroll items-end",
        children: [
          attachments.map((attachment) => /* @__PURE__ */ jsx(PreviewAttachment, { attachment }, attachment.uri)),
          uploadQueue.map((filename) => /* @__PURE__ */ jsx(
            PreviewAttachment,
            {
              attachment: {
                name: filename
              },
              isUploading: true
            },
            filename
          ))
        ]
      }
    ),
    /* @__PURE__ */ jsx(
      Textarea,
      {
        "data-testid": "multimodal-input",
        ref: textareaRef,
        placeholder: "Send a message...",
        value: input,
        onChange: handleInput,
        className: cn(
          "min-h-[24px] max-h-[calc(75dvh)] overflow-hidden resize-none rounded-2xl !text-base bg-muted pb-10 dark:border-zinc-700",
          className
        ),
        rows: 2,
        autoFocus: true,
        onKeyDown: (event) => {
          if (event.key === "Enter" && !event.shiftKey && !event.nativeEvent.isComposing) {
            event.preventDefault();
            if (status !== "ready") {
              toast$1.error("Please wait for the model to finish its response!");
            } else {
              submitForm();
            }
          }
        }
      }
    ),
    /* @__PURE__ */ jsx("div", { className: "absolute bottom-0 p-2 w-fit flex flex-row justify-start", children: /* @__PURE__ */ jsx(AttachmentsButton, { fileInputRef, status }) }),
    /* @__PURE__ */ jsx("div", { className: "absolute bottom-0 right-0 p-2 w-fit flex flex-row justify-end", children: status === "submitted" ? /* @__PURE__ */ jsx(StopButton, { stop, setMessages }) : /* @__PURE__ */ jsx(
      SendButton,
      {
        input,
        submitForm,
        uploadQueue
      }
    ) })
  ] });
}
const MultimodalInput = memo(
  PureMultimodalInput,
  (prevProps, nextProps) => {
    var _a, _b;
    if (prevProps.input !== nextProps.input) return false;
    if (prevProps.status !== nextProps.status) return false;
    if (((_a = prevProps.messages) == null ? void 0 : _a.length) !== ((_b = nextProps.messages) == null ? void 0 : _b.length)) return false;
    if (!equal(prevProps.attachments, nextProps.attachments)) return false;
    return true;
  }
);
function PureAttachmentsButton({
  fileInputRef,
  status
}) {
  return /* @__PURE__ */ jsx(
    Button,
    {
      "data-testid": "attachments-button",
      className: "rounded-md rounded-bl-lg p-[7px] h-fit dark:border-zinc-700 hover:dark:bg-zinc-900 hover:bg-zinc-200",
      onClick: (event) => {
        var _a;
        event.preventDefault();
        (_a = fileInputRef.current) == null ? void 0 : _a.click();
      },
      disabled: status !== "ready",
      variant: "ghost",
      children: /* @__PURE__ */ jsx(PaperclipIcon, { size: 14 })
    }
  );
}
const AttachmentsButton = memo(PureAttachmentsButton);
function PureStopButton({
  stop,
  setMessages
}) {
  return /* @__PURE__ */ jsx(
    Button,
    {
      "data-testid": "stop-button",
      className: "rounded-full p-1.5 h-fit border dark:border-zinc-600",
      onClick: (event) => {
        event.preventDefault();
        stop();
        setMessages((messages) => messages);
      },
      children: /* @__PURE__ */ jsx(StopIcon, { size: 14 })
    }
  );
}
const StopButton = memo(PureStopButton);
function PureSendButton({
  submitForm,
  input,
  uploadQueue
}) {
  return /* @__PURE__ */ jsx(
    Button,
    {
      "data-testid": "send-button",
      className: "rounded-full p-1.5 h-fit border dark:border-zinc-600",
      onClick: (event) => {
        event.preventDefault();
        submitForm();
      },
      disabled: input.length === 0 || uploadQueue.length > 0,
      children: /* @__PURE__ */ jsx(ArrowUpIcon, { size: 14 })
    }
  );
}
const SendButton = memo(PureSendButton, (prevProps, nextProps) => {
  if (prevProps.uploadQueue.length !== nextProps.uploadQueue.length)
    return false;
  if (prevProps.input !== nextProps.input) return false;
  return true;
});
function CodeBlock({
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  node,
  inline,
  className,
  children,
  ...props
}) {
  if (!inline) {
    return /* @__PURE__ */ jsx("div", { className: "not-prose flex flex-col", children: /* @__PURE__ */ jsx(
      "pre",
      {
        ...props,
        className: `text-sm w-full overflow-x-auto dark:bg-zinc-900 p-4 border border-zinc-200 dark:border-zinc-700 rounded-xl dark:text-zinc-50 text-zinc-900`,
        children: /* @__PURE__ */ jsx("code", { className: "whitespace-pre-wrap break-words", children })
      }
    ) });
  } else {
    return /* @__PURE__ */ jsx(
      "code",
      {
        className: `${className} text-sm bg-zinc-100 dark:bg-zinc-800 py-0.5 px-1 rounded-md`,
        ...props,
        children
      }
    );
  }
}
const components = {
  // @ts-expect-error I dunno why this is needed, but it was just here
  code: CodeBlock,
  pre: ({ children }) => /* @__PURE__ */ jsx(Fragment, { children }),
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  ol: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("ol", { className: "list-decimal list-outside ml-4", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  li: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("li", { className: "py-1", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  ul: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("ul", { className: "list-decimal list-outside ml-4", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  strong: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("span", { className: "font-semibold", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  a: ({ node, children, ...props }) => {
    return props.href && /* @__PURE__ */ jsx(
      Link,
      {
        className: "text-blue-500 hover:underline",
        target: "_blank",
        rel: "noreferrer",
        to: props.href,
        ...props,
        children
      }
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h1: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("h1", { className: "text-3xl font-semibold mt-6 mb-2", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h2: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("h2", { className: "text-2xl font-semibold mt-6 mb-2", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h3: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("h3", { className: "text-xl font-semibold mt-6 mb-2", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h4: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("h4", { className: "text-lg font-semibold mt-6 mb-2", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h5: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("h5", { className: "text-base font-semibold mt-6 mb-2", ...props, children });
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h6: ({ node, children, ...props }) => {
    return /* @__PURE__ */ jsx("h6", { className: "text-sm font-semibold mt-6 mb-2", ...props, children });
  }
};
const remarkPlugins = [remarkGfm];
const NonMemoizedMarkdown = ({ children }) => {
  return /* @__PURE__ */ jsx(ReactMarkdown, { remarkPlugins, components, children });
};
const Markdown = memo(
  NonMemoizedMarkdown,
  (prevProps, nextProps) => prevProps.children === nextProps.children
);
function PureMessageActions({
  chatId,
  message,
  vote,
  isLoading
}) {
  const { mutate } = useSWRConfig();
  const [_, copyToClipboard] = useCopyToClipboard();
  if (isLoading) return null;
  if (message.role === "user") return null;
  return /* @__PURE__ */ jsx(TooltipProvider, { delayDuration: 0, children: /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2", children: [
    /* @__PURE__ */ jsxs(Tooltip, { children: [
      /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: /* @__PURE__ */ jsx(
        Button,
        {
          className: "py-1 px-2 h-fit text-muted-foreground",
          variant: "outline",
          onClick: async () => {
            var _a;
            const textFromParts = (_a = message.parts) == null ? void 0 : _a.filter((part) => part.type === "text").map((part) => part.text).join("\n").trim();
            if (!textFromParts) {
              toast$1.error("There's no text to copy!");
              return;
            }
            await copyToClipboard(textFromParts);
            toast$1.success("Copied to clipboard!");
          },
          children: /* @__PURE__ */ jsx(CopyIcon, {})
        }
      ) }),
      /* @__PURE__ */ jsx(TooltipContent, { children: "Copy" })
    ] }),
    /* @__PURE__ */ jsxs(Tooltip, { children: [
      /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: /* @__PURE__ */ jsx(
        Button,
        {
          "data-testid": "message-upvote",
          className: "py-1 px-2 h-fit text-muted-foreground !pointer-events-auto",
          disabled: vote == null ? void 0 : vote.isUpvoted,
          variant: "outline",
          onClick: async () => {
            const upvote = fetch("/api/vote", {
              method: "PATCH",
              body: JSON.stringify({
                chatId,
                messageId: message.id,
                type: "up"
              })
            });
            toast$1.promise(upvote, {
              loading: "Upvoting Response...",
              success: () => {
                mutate(
                  `/api/vote?chatId=${chatId}`,
                  (currentVotes) => {
                    if (!currentVotes) return [];
                    const votesWithoutCurrent = currentVotes.filter(
                      (vote2) => vote2.messageId !== message.id
                    );
                    return [
                      ...votesWithoutCurrent,
                      {
                        chatId,
                        messageId: message.id,
                        isUpvoted: true
                      }
                    ];
                  },
                  { revalidate: false }
                );
                return "Upvoted Response!";
              },
              error: "Failed to upvote response."
            });
          },
          children: /* @__PURE__ */ jsx(ThumbUpIcon, {})
        }
      ) }),
      /* @__PURE__ */ jsx(TooltipContent, { children: "Upvote Response" })
    ] }),
    /* @__PURE__ */ jsxs(Tooltip, { children: [
      /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: /* @__PURE__ */ jsx(
        Button,
        {
          "data-testid": "message-downvote",
          className: "py-1 px-2 h-fit text-muted-foreground !pointer-events-auto",
          variant: "outline",
          disabled: vote && !vote.isUpvoted,
          onClick: async () => {
            const downvote = fetch("/api/vote", {
              method: "PATCH",
              body: JSON.stringify({
                chatId,
                messageId: message.id,
                type: "down"
              })
            });
            toast$1.promise(downvote, {
              loading: "Downvoting Response...",
              success: () => {
                mutate(
                  `/api/vote?chatId=${chatId}`,
                  (currentVotes) => {
                    if (!currentVotes) return [];
                    const votesWithoutCurrent = currentVotes.filter(
                      (vote2) => vote2.messageId !== message.id
                    );
                    return [
                      ...votesWithoutCurrent,
                      {
                        chatId,
                        messageId: message.id,
                        isUpvoted: false
                      }
                    ];
                  },
                  { revalidate: false }
                );
                return "Downvoted Response!";
              },
              error: "Failed to downvote response."
            });
          },
          children: /* @__PURE__ */ jsx(ThumbDownIcon, {})
        }
      ) }),
      /* @__PURE__ */ jsx(TooltipContent, { children: "Downvote Response" })
    ] })
  ] }) });
}
const MessageActions = memo(
  PureMessageActions,
  (prevProps, nextProps) => {
    if (!equal(prevProps.vote, nextProps.vote)) return false;
    if (prevProps.isLoading !== nextProps.isLoading) return false;
    return true;
  }
);
function MessageEditor({
  message,
  itemIndex,
  setMode,
  setMessage,
  reload
}) {
  var _a;
  const item = (_a = message.items) == null ? void 0 : _a[itemIndex];
  const initialContent = item && item.type === "text" && item.text ? item.text : "";
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [draftContent, setDraftContent] = useState(initialContent);
  const textareaRef = useRef(null);
  useEffect(() => {
    if (textareaRef.current) {
      adjustHeight();
    }
  }, []);
  const adjustHeight = () => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight + 2}px`;
    }
  };
  const handleInput = (event) => {
    setDraftContent(event.target.value);
    adjustHeight();
  };
  return /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-2 w-full", children: [
    /* @__PURE__ */ jsx(
      Textarea,
      {
        "data-testid": "message-editor",
        ref: textareaRef,
        className: "bg-transparent outline-none overflow-hidden resize-none !text-base rounded-xl w-full",
        value: draftContent,
        onChange: handleInput
      }
    ),
    /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 justify-end", children: [
      /* @__PURE__ */ jsx(
        Button,
        {
          variant: "outline",
          className: "h-fit py-2 px-3",
          onClick: () => {
            setMode("view");
          },
          children: "Cancel"
        }
      ),
      /* @__PURE__ */ jsx(
        Button,
        {
          "data-testid": "message-editor-send-button",
          variant: "default",
          className: "h-fit py-2 px-3",
          disabled: isSubmitting,
          onClick: async () => {
            setIsSubmitting(true);
            if (message.id) {
              setMessage(message.id, (prevMessage) => {
                const updatedMessage = {
                  ...prevMessage
                };
                if (updatedMessage.items && updatedMessage.items[itemIndex]) {
                  updatedMessage.items[itemIndex] = {
                    ...item,
                    type: "text",
                    text: draftContent
                  };
                  return updatedMessage;
                }
                return prevMessage;
              });
            }
            setMode("view");
            reload();
          },
          children: isSubmitting ? "Sending..." : "Send"
        }
      )
    ] })
  ] });
}
function MessageReasoning({
  isLoading,
  reasoning
}) {
  const [isExpanded, setIsExpanded] = useState(true);
  const variants = {
    collapsed: {
      height: 0,
      opacity: 0,
      marginTop: 0,
      marginBottom: 0
    },
    expanded: {
      height: "auto",
      opacity: 1,
      marginTop: "1rem",
      marginBottom: "0.5rem"
    }
  };
  return /* @__PURE__ */ jsxs("div", { className: "flex flex-col", children: [
    isLoading ? /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-center", children: [
      /* @__PURE__ */ jsx("div", { className: "font-medium", children: "Reasoning" }),
      /* @__PURE__ */ jsx("div", { className: "animate-spin", children: /* @__PURE__ */ jsx(LoaderIcon, {}) })
    ] }) : /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-center", children: [
      /* @__PURE__ */ jsx("div", { className: "font-medium", children: "Reasoned for a few seconds" }),
      /* @__PURE__ */ jsx(
        "button",
        {
          "data-testid": "message-reasoning-toggle",
          type: "button",
          className: "cursor-pointer",
          onClick: () => {
            setIsExpanded(!isExpanded);
          },
          children: /* @__PURE__ */ jsx(ChevronDownIcon, {})
        }
      )
    ] }),
    /* @__PURE__ */ jsx(AnimatePresence, { initial: false, children: isExpanded && /* @__PURE__ */ jsx(
      motion.div,
      {
        "data-testid": "message-reasoning",
        initial: "collapsed",
        animate: "expanded",
        exit: "collapsed",
        variants,
        transition: { duration: 0.2, ease: "easeInOut" },
        style: { overflow: "hidden" },
        className: "pl-4 text-zinc-600 dark:text-zinc-400 border-l flex flex-col gap-4",
        children: /* @__PURE__ */ jsx(Markdown, { children: reasoning })
      },
      "content"
    ) })
  ] });
}
function isValidJson(str) {
  try {
    JSON.parse(str);
    return true;
  } catch (e) {
    return false;
  }
}
function JsonTable({ jsonString }) {
  const data = JSON.parse(jsonString);
  return /* @__PURE__ */ jsx("table", { className: "w-full", children: /* @__PURE__ */ jsx("tbody", { children: Object.entries(data).map(([key, value]) => /* @__PURE__ */ jsxs("tr", { className: "border-b border-border/50 last:border-0", children: [
    /* @__PURE__ */ jsx("td", { className: "py-1 pr-4 text-muted-foreground", children: key }),
    /* @__PURE__ */ jsx("td", { className: "py-1", children: typeof value === "string" ? value : JSON.stringify(value) })
  ] }, key)) }) });
}
function ResourceTable({ resource }) {
  return /* @__PURE__ */ jsxs("table", { className: "w-full", children: [
    /* @__PURE__ */ jsx("thead", { children: /* @__PURE__ */ jsx("tr", { children: /* @__PURE__ */ jsx("th", { className: "text-left py-2 font-medium", colSpan: 2, children: "Resource" }) }) }),
    /* @__PURE__ */ jsx("tbody", { children: Object.entries(resource ?? {}).map(([key, value]) => /* @__PURE__ */ jsxs("tr", { className: "border-b border-border/50 last:border-0", children: [
      /* @__PURE__ */ jsx("td", { className: "py-1 pr-4 text-muted-foreground", children: key }),
      /* @__PURE__ */ jsx("td", { className: "py-1", children: typeof value === "string" ? value : JSON.stringify(value) })
    ] }, key)) })
  ] });
}
function ToolCall({ toolCall }) {
  var _a, _b;
  const { name, arguments: args, output, target, targetType } = toolCall;
  const [isExpanded, setIsExpanded] = useState(false);
  const isCompleted = !!output;
  return /* @__PURE__ */ jsxs(Fragment, { children: [
    /* @__PURE__ */ jsx(
      "div",
      {
        className: cn({
          "p-4 my-2": true,
          "border rounded-lg bg-muted/30": isExpanded
        }),
        children: /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-3", children: [
          /* @__PURE__ */ jsxs(
            "button",
            {
              onClick: () => setIsExpanded(!isExpanded),
              className: "flex items-center gap-2 text-sm font-medium w-full",
              children: [
                /* @__PURE__ */ jsx("div", { className: "size-5 flex items-center rounded-full justify-center ring-1 shrink-0 ring-border bg-background", children: /* @__PURE__ */ jsx(Hammer, { size: 12 }) }),
                /* @__PURE__ */ jsxs(
                  "span",
                  {
                    className: cn(
                      "font-semibold",
                      isCompleted ? "text-foreground" : "text-muted-foreground"
                    ),
                    children: [
                      isCompleted ? "" : "Calling:",
                      " ",
                      name
                    ]
                  }
                ),
                target && targetType && /* @__PURE__ */ jsxs("span", { className: "text-xs text-muted-foreground", children: [
                  "(",
                  targetType,
                  ": ",
                  target,
                  ")"
                ] }),
                /* @__PURE__ */ jsx(
                  ChevronDownIcon$1,
                  {
                    size: 16,
                    className: cn(
                      "ml-auto transition-transform",
                      isExpanded ? "rotate-0" : "-rotate-90"
                    )
                  }
                )
              ]
            }
          ),
          isExpanded && /* @__PURE__ */ jsxs("div", { className: "animate-in fade-in-0 slide-in-from-top-2", children: [
            args && /* @__PURE__ */ jsxs("div", { className: "pl-7", children: [
              /* @__PURE__ */ jsx("div", { className: "text-xs text-muted-foreground mb-1", children: "Arguments:" }),
              /* @__PURE__ */ jsx("div", { className: "bg-background rounded p-2 text-sm font-mono overflow-x-auto", children: isValidJson(args) ? /* @__PURE__ */ jsx(JsonTable, { jsonString: args }) : args })
            ] }),
            output && /* @__PURE__ */ jsxs("div", { className: "pl-7 mt-1", children: [
              /* @__PURE__ */ jsxs("div", { className: "text-xs text-muted-foreground mb-1", children: [
                "Result",
                output.isError ? " (Error)" : "",
                output.agent ? ` from ${output.agent}` : "",
                output.model ? ` using {output.model}` : "",
                ":"
              ] }),
              /* @__PURE__ */ jsx(
                "div",
                {
                  className: cn(
                    "rounded p-3 text-sm",
                    output.isError ? "bg-destructive/10 border border-destructive/20" : "bg-background"
                  ),
                  children: (_a = output.content) == null ? void 0 : _a.map((content, index2) => {
                    var _a2;
                    if (content.type === "text" && content.text) {
                      return /* @__PURE__ */ jsxs("div", { className: "mb-2", children: [
                        /* @__PURE__ */ jsx("div", { className: "text-xs text-muted-foreground mb-1", children: "Text Output:" }),
                        /* @__PURE__ */ jsx(Markdown, { children: sanitizeText(content.text) })
                      ] }, index2);
                    } else if (content.type === "resource" && content.resource) {
                      return ((_a2 = content.resource.uri) == null ? void 0 : _a2.startsWith("ui://")) ? /* @__PURE__ */ jsxs("div", { className: "mb-2", children: [
                        /* @__PURE__ */ jsx("div", { className: "text-xs text-muted-foreground mb-1", children: "MCP-UI:" }),
                        /* @__PURE__ */ jsx(UIResourceRenderer, { resource: content.resource })
                      ] }, index2) : /* @__PURE__ */ jsx("div", { className: "mb-2", children: /* @__PURE__ */ jsx(ResourceTable, { resource: content.resource }) }, index2);
                    } else if (content.structuredContent) {
                      return /* @__PURE__ */ jsxs("div", { className: "mb-2", children: [
                        /* @__PURE__ */ jsx("div", { className: "text-xs text-muted-foreground mb-1", children: "Structured Output:" }),
                        /* @__PURE__ */ jsx("pre", { className: "whitespace-pre-wrap overflow-x-auto", children: JSON.stringify(
                          content.structuredContent,
                          null,
                          2
                        ) })
                      ] }, index2);
                    }
                    return null;
                  })
                }
              )
            ] }),
            !isCompleted && /* @__PURE__ */ jsxs("div", { className: "pl-7 flex items-center gap-2 text-muted-foreground", children: [
              /* @__PURE__ */ jsx("div", { className: "animate-pulse h-2 w-2 rounded-full bg-muted-foreground" }),
              /* @__PURE__ */ jsx("div", { className: "animate-pulse h-2 w-2 rounded-full bg-muted-foreground animation-delay-200" }),
              /* @__PURE__ */ jsx("div", { className: "animate-pulse h-2 w-2 rounded-full bg-muted-foreground animation-delay-400" })
            ] })
          ] })
        ] })
      }
    ),
    !isExpanded && (((_b = output == null ? void 0 : output.content) == null ? void 0 : _b.flatMap((content) => {
      var _a2;
      if (content.type === "resource" && content.resource && ((_a2 = content.resource.uri) == null ? void 0 : _a2.startsWith("ui://")) && !content.resource.uri.startsWith("ui://widget/")) {
        return [content.resource];
      }
      return [];
    })) || []).map((resource, index2) => /* @__PURE__ */ jsx(UIResourceRenderer, { resource }, index2))
  ] });
}
const PurePreviewMessage = ({
  chatId,
  message,
  vote,
  isLoading,
  setMessage,
  reload,
  isReadonly,
  requiresScrollPadding
}) => {
  var _a, _b;
  const [mode, setMode] = useState("view");
  const images = ((_a = message.items) == null ? void 0 : _a.filter((i) => i.type === "image")) || [];
  return /* @__PURE__ */ jsx(AnimatePresence, { children: /* @__PURE__ */ jsx(
    motion.div,
    {
      "data-testid": `message-${message.role}`,
      className: "w-full mx-auto max-w-3xl px-4 group/message",
      initial: { y: 5, opacity: 0 },
      animate: { y: 0, opacity: 1 },
      "data-role": message.role,
      children: /* @__PURE__ */ jsxs(
        "div",
        {
          className: cn(
            "flex gap-4 w-full group-data-[role=user]/message:ml-auto group-data-[role=user]/message:max-w-2xl",
            {
              "w-full": mode === "edit",
              "group-data-[role=user]/message:w-fit": mode !== "edit"
            }
          ),
          children: [
            message.role === "assistant" && /* @__PURE__ */ jsx("div", { className: "size-8 flex items-center rounded-full justify-center ring-1 shrink-0 ring-border bg-background", children: /* @__PURE__ */ jsx("div", { className: "translate-y-px", children: /* @__PURE__ */ jsx(SparklesIcon, { size: 14 }) }) }),
            /* @__PURE__ */ jsxs(
              "div",
              {
                className: cn("flex flex-col gap-4 w-full", {
                  "min-h-96": message.role === "assistant" && requiresScrollPadding
                }),
                children: [
                  images && images.length > 0 && /* @__PURE__ */ jsx(
                    "div",
                    {
                      "data-testid": `message-attachments`,
                      className: "flex flex-row justify-end gap-2",
                      children: images.map(
                        (attachment) => attachment.type == "image" && /* @__PURE__ */ jsx(
                          PreviewAttachment,
                          {
                            attachment
                          },
                          attachment.data
                        )
                      )
                    }
                  ),
                  (_b = message.items) == null ? void 0 : _b.map((part, index2) => {
                    var _a2;
                    const { type } = part;
                    const key = `message-${message.id}-part-${index2}`;
                    if (type === "reasoning") {
                      return /* @__PURE__ */ jsx(
                        MessageReasoning,
                        {
                          isLoading,
                          reasoning: ((_a2 = part.summary) == null ? void 0 : _a2.map((x) => x.text).join("\n")) || ""
                        },
                        key
                      );
                    }
                    if (type === "text") {
                      if (mode === "view") {
                        return /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-start", children: [
                          message.role === "user" && !isReadonly && /* @__PURE__ */ jsxs(Tooltip, { children: [
                            /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: /* @__PURE__ */ jsx(
                              Button,
                              {
                                variant: "ghost",
                                className: "px-2 h-fit rounded-full text-muted-foreground opacity-0 group-hover/message:opacity-100",
                                onClick: () => {
                                  setMode("edit");
                                },
                                children: /* @__PURE__ */ jsx(PencilEditIcon, {})
                              }
                            ) }),
                            /* @__PURE__ */ jsx(TooltipContent, { children: "Edit message" })
                          ] }),
                          /* @__PURE__ */ jsx(
                            "div",
                            {
                              "data-testid": "message-content",
                              className: cn("flex flex-col gap-4", {
                                "bg-primary text-primary-foreground px-3 py-2 rounded-xl": message.role === "user"
                              }),
                              children: /* @__PURE__ */ jsx(Markdown, { children: sanitizeText(part.text || "") })
                            }
                          )
                        ] }, key);
                      }
                      if (mode === "edit") {
                        return /* @__PURE__ */ jsxs("div", { className: "flex flex-row gap-2 items-start", children: [
                          /* @__PURE__ */ jsx("div", { className: "size-8" }),
                          /* @__PURE__ */ jsx(
                            MessageEditor,
                            {
                              itemIndex: index2,
                              message,
                              setMode,
                              setMessage,
                              reload
                            },
                            message.id
                          )
                        ] }, key);
                      }
                    }
                    if (type === "tool") {
                      return /* @__PURE__ */ jsx(ToolCall, { toolCall: part }, key);
                    }
                  }),
                  !isReadonly && /* @__PURE__ */ jsx(
                    MessageActions,
                    {
                      chatId,
                      message,
                      vote,
                      isLoading
                    },
                    `action-${message.id}`
                  )
                ]
              }
            )
          ]
        }
      )
    }
  ) });
};
const PreviewMessage = memo(
  PurePreviewMessage,
  (prevProps, nextProps) => {
    if (prevProps.isLoading !== nextProps.isLoading) return false;
    if (prevProps.message.id !== nextProps.message.id) return false;
    if (prevProps.requiresScrollPadding !== nextProps.requiresScrollPadding)
      return false;
    if (prevProps.message.revision !== nextProps.message.revision) return false;
    if (!equal(prevProps.vote, nextProps.vote)) return false;
    return true;
  }
);
const ThinkingMessage = () => {
  const role = "assistant";
  return /* @__PURE__ */ jsx(
    motion.div,
    {
      "data-testid": "message-assistant-loading",
      className: "w-full mx-auto max-w-3xl px-4 group/message min-h-96",
      initial: { y: 5, opacity: 0 },
      animate: { y: 0, opacity: 1, transition: { delay: 1 } },
      "data-role": role,
      children: /* @__PURE__ */ jsxs(
        "div",
        {
          className: cn(
            "flex gap-4 group-data-[role=user]/message:px-3 w-full group-data-[role=user]/message:w-fit group-data-[role=user]/message:ml-auto group-data-[role=user]/message:max-w-2xl group-data-[role=user]/message:py-2 rounded-xl",
            {
              "group-data-[role=user]/message:bg-muted": true
            }
          ),
          children: [
            /* @__PURE__ */ jsx("div", { className: "size-8 flex items-center rounded-full justify-center ring-1 shrink-0 ring-border", children: /* @__PURE__ */ jsx(SparklesIcon, { size: 14 }) }),
            /* @__PURE__ */ jsx("div", { className: "flex flex-col gap-2 w-full", children: /* @__PURE__ */ jsx("div", { className: "flex flex-col gap-4 text-muted-foreground", children: "Hmm..." }) })
          ]
        }
      )
    }
  );
};
function Greeting() {
  return /* @__PURE__ */ jsxs(
    "div",
    {
      className: "max-w-3xl mx-auto md:mt-20 px-8 size-full flex flex-col justify-center",
      children: [
        /* @__PURE__ */ jsx(
          motion.div,
          {
            initial: { opacity: 0, y: 10 },
            animate: { opacity: 1, y: 0 },
            exit: { opacity: 0, y: 10 },
            transition: { delay: 0.5 },
            className: "text-2xl font-semibold",
            children: "Hello there!"
          }
        ),
        /* @__PURE__ */ jsx(
          motion.div,
          {
            initial: { opacity: 0, y: 10 },
            animate: { opacity: 1, y: 0 },
            exit: { opacity: 0, y: 10 },
            transition: { delay: 0.6 },
            className: "text-2xl text-zinc-500",
            children: "How can I help you today?"
          }
        )
      ]
    },
    "overview"
  );
}
function useMessages({
  chatId,
  status
}) {
  const {
    containerRef,
    endRef,
    isAtBottom,
    scrollToBottom,
    onViewportEnter,
    onViewportLeave
  } = useScrollToBottom();
  const [hasSentMessage, setHasSentMessage] = useState(false);
  useEffect(() => {
    if (chatId) {
      scrollToBottom("instant");
      setHasSentMessage(false);
    }
  }, [chatId, scrollToBottom]);
  useEffect(() => {
    if (status === "submitted") {
      setHasSentMessage(true);
    }
  }, [status]);
  return {
    containerRef,
    endRef,
    isAtBottom,
    scrollToBottom,
    onViewportEnter,
    onViewportLeave,
    hasSentMessage
  };
}
function PureMessages({
  chatId,
  status,
  votes,
  messages,
  updateMessage,
  reload,
  isReadonly
}) {
  const {
    containerRef: messagesContainerRef,
    endRef: messagesEndRef,
    onViewportEnter,
    onViewportLeave,
    hasSentMessage
  } = useMessages({
    chatId,
    status
  });
  return /* @__PURE__ */ jsxs(
    "div",
    {
      ref: messagesContainerRef,
      className: "flex flex-col min-w-0 gap-6 flex-1 overflow-y-scroll pt-4 relative",
      children: [
        messages.length === 0 && /* @__PURE__ */ jsx(Greeting, {}),
        messages.map((message, index2) => /* @__PURE__ */ jsx(
          PreviewMessage,
          {
            chatId,
            message,
            isLoading: status === "streaming" && messages.length - 1 === index2,
            vote: votes ? votes.find((vote) => vote.messageId === message.id) : void 0,
            setMessage: updateMessage,
            reload,
            isReadonly,
            requiresScrollPadding: hasSentMessage && index2 === messages.length - 1
          },
          message.id
        )),
        status === "submitted" && messages.length > 0 && messages[messages.length - 1].role === "user" && /* @__PURE__ */ jsx(ThinkingMessage, {}),
        /* @__PURE__ */ jsx(
          motion.div,
          {
            ref: messagesEndRef,
            className: "shrink-0 min-w-[24px] min-h-[24px]",
            onViewportLeave,
            onViewportEnter
          }
        )
      ]
    }
  );
}
const Messages = memo(PureMessages, (prevProps, nextProps) => {
  if (prevProps.status !== nextProps.status) return false;
  if (prevProps.status && nextProps.status) return false;
  if (prevProps.messages.length !== nextProps.messages.length) return false;
  if (!equal(prevProps.messages, nextProps.messages)) return false;
  if (!equal(prevProps.votes, nextProps.votes)) return false;
  return true;
});
const iconsByType = {
  success: /* @__PURE__ */ jsx(CheckCircleFillIcon, {}),
  error: /* @__PURE__ */ jsx(WarningIcon, {})
};
function toast(props) {
  return toast$1.custom((id) => /* @__PURE__ */ jsx(Toast, { id, type: props.type, description: props.description }));
}
function Toast(props) {
  const { id, type, description } = props;
  const descriptionRef = useRef(null);
  const [multiLine, setMultiLine] = useState(false);
  useEffect(() => {
    const el = descriptionRef.current;
    if (!el) return;
    const update = () => {
      const lineHeight = Number.parseFloat(getComputedStyle(el).lineHeight);
      const lines = Math.round(el.scrollHeight / lineHeight);
      setMultiLine(lines > 1);
    };
    update();
    const ro = new ResizeObserver(update);
    ro.observe(el);
    return () => ro.disconnect();
  }, [description]);
  return /* @__PURE__ */ jsx("div", { className: "flex w-full toast-mobile:w-[356px] justify-center", children: /* @__PURE__ */ jsxs(
    "div",
    {
      "data-testid": "toast",
      className: cn(
        "bg-zinc-100 p-3 rounded-lg w-full toast-mobile:w-fit flex flex-row gap-3",
        multiLine ? "items-start" : "items-center"
      ),
      children: [
        /* @__PURE__ */ jsx(
          "div",
          {
            "data-type": type,
            className: cn(
              "data-[type=error]:text-red-600 data-[type=success]:text-green-600",
              { "pt-1": multiLine }
            ),
            children: iconsByType[type]
          }
        ),
        /* @__PURE__ */ jsx("div", { ref: descriptionRef, className: "text-zinc-950 text-sm", children: description })
      ]
    },
    id
  ) });
}
function CloneChat({ chatId }) {
  return /* @__PURE__ */ jsx(Form, { method: "post", action: `/chat/${chatId}/clone`, className: "contents", children: /* @__PURE__ */ jsx(Button, { type: "submit", className: "mt-4 mx-auto", children: "Continue Chat in New Thread" }) });
}
function Widgets({ widgets }) {
  var _a;
  const [fullScreenWidget, setFullScreenWidget] = useState(null);
  const toggleFullScreen = (index2) => {
    if (fullScreenWidget === index2) {
      setFullScreenWidget(null);
    } else {
      setFullScreenWidget(index2);
    }
  };
  if (fullScreenWidget !== null || widgets.length === 1) {
    const widget = widgets[widgets.length === 1 ? 0 : fullScreenWidget ?? 0];
    return /* @__PURE__ */ jsxs("div", { className: "border border-[#e0e0e0] rounded-md p-6 flex flex-col h-full", children: [
      /* @__PURE__ */ jsxs("div", { className: "flex justify-between items-center mb-4 pb-2 border-b", children: [
        /* @__PURE__ */ jsx("h3", { className: "font-medium", children: (_a = widget.resource) == null ? void 0 : _a.uri }),
        widgets.length !== 1 && /* @__PURE__ */ jsx(
          "button",
          {
            className: "p-1.5 hover:bg-[#f0f0f0] rounded-md cursor-pointer",
            onClick: () => setFullScreenWidget(null),
            "aria-label": "Exit full screen",
            children: /* @__PURE__ */ jsx(Minimize, { size: 18 })
          }
        )
      ] }),
      /* @__PURE__ */ jsx("div", { className: "flex-1", children: widget.resource && /* @__PURE__ */ jsx(
        UIResourceRenderer,
        {
          htmlProps: {
            iframeProps: {
              className: "h-full"
            }
          },
          resource: widget.resource
        }
      ) })
    ] });
  }
  return /* @__PURE__ */ jsx("div", { className: "grid grid-cols-2 gap-4", children: widgets.map((widget, index2) => {
    var _a2, _b;
    return /* @__PURE__ */ jsxs(
      "div",
      {
        className: "border border-[#e0e0e0] rounded-md p-4 flex flex-col",
        children: [
          /* @__PURE__ */ jsxs("div", { className: "flex justify-between items-center mb-4 pb-2 border-b", children: [
            /* @__PURE__ */ jsx("h3", { className: "font-medium text-muted-foreground", children: ((_a2 = widget.resource) == null ? void 0 : _a2.name) || ((_b = widget.resource) == null ? void 0 : _b.uri) }),
            /* @__PURE__ */ jsx(
              "button",
              {
                className: "p-1.5 hover:bg-[#f0f0f0] rounded-md cursor-pointer",
                onClick: () => toggleFullScreen(index2),
                "aria-label": "Full screen",
                children: /* @__PURE__ */ jsx(Maximize, { size: 18 })
              }
            )
          ] }),
          /* @__PURE__ */ jsx("div", { className: "flex-1 h-full", children: widget.resource && /* @__PURE__ */ jsx(
            UIResourceRenderer,
            {
              htmlProps: {
                iframeProps: {
                  className: "h-full"
                }
              },
              resource: widget.resource
            }
          ) })
        ]
      },
      index2
    );
  }) });
}
function Chat({
  chat: chat2,
  disableHeader
}) {
  const {
    messages,
    setMessages,
    updateMessage,
    handleSubmit,
    input,
    setInput,
    status,
    stop,
    reload,
    currentAgent,
    setCurrentAgent,
    agents,
    votes
  } = useChat({
    chat: chat2,
    onError: (error) => {
      if (error instanceof ChatSDKError) {
        toast({
          type: "error",
          description: error.message
        });
      }
    }
  });
  const [attachments, setAttachments] = useState([]);
  const widgets = getWidgets(messages);
  return /* @__PURE__ */ jsx(Fragment, { children: /* @__PURE__ */ jsxs("div", { className: "flex flex-col min-w-0 h-dvh bg-background", children: [
    !disableHeader && /* @__PURE__ */ jsx(
      ChatHeader,
      {
        currentAgent,
        setCurrentAgent,
        agents,
        customAgent: chat2.customAgent,
        isReadonly: !!chat2.readonly
      }
    ),
    /* @__PURE__ */ jsxs("div", { className: "flex overflow-hidden h-full", children: [
      /* @__PURE__ */ jsxs("div", { className: "flex flex-col flex-1 h-full", children: [
        /* @__PURE__ */ jsx(
          Messages,
          {
            chatId: chat2.id,
            status,
            votes,
            messages,
            updateMessage,
            reload,
            isReadonly: !!chat2.readonly
          }
        ),
        /* @__PURE__ */ jsxs("form", { className: "flex mx-auto px-4 bg-background pb-4 md:pb-6 gap-2 w-full md:max-w-3xl", children: [
          !chat2.readonly && /* @__PURE__ */ jsx(
            MultimodalInput,
            {
              chatId: chat2.id,
              input,
              setInput,
              handleSubmit,
              status,
              stop,
              attachments,
              setAttachments,
              messages,
              setMessages
            }
          ),
          !!chat2.readonly && /* @__PURE__ */ jsx(CloneChat, { chatId: chat2.id })
        ] })
      ] }),
      widgets.length ? /* @__PURE__ */ jsx("div", { className: "h-full w-1/2 p-2", children: /* @__PURE__ */ jsx(Widgets, { widgets }) }) : null
    ] })
  ] }) });
}
async function action$1({
  request,
  params
}) {
  if (request.method === "DELETE") {
    if (!params.id) {
      throw new Error("Chat ID is required for deletion.");
    }
    return await deleteChat(getContext(), params.id);
  }
  const formData = await request.formData();
  const agent2 = formData.get("agent");
  if (agent2) {
    return await setAgent(getContext(), params.id || "", agent2);
  }
  const prompt = formData.get("prompt");
  if (prompt) {
    await chat(getContext(), params.id || "", prompt);
    return;
  }
  const visibility = formData.get("visibility");
  if (visibility) {
    if (visibility !== "public" && visibility !== "private") {
      throw new Error("Invalid visibility type. Must be 'public' or 'private'.");
    }
    return await setVisibility(getContext(), params.id || "", visibility);
  }
}
const ChatPage = UNSAFE_withComponentProps(function Page() {
  const data = useRouteLoaderData("routes/layout");
  return /* @__PURE__ */ jsx(Chat, {
    chat: data.chat,
    user: data.user
  });
});
const route9 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  action: action$1,
  default: ChatPage
}, Symbol.toStringTag, { value: "Module" }));
const index = UNSAFE_withComponentProps(ChatPage);
const route8 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  default: index
}, Symbol.toStringTag, { value: "Module" }));
const agent = UNSAFE_withComponentProps(ChatPage);
const route10 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  default: agent
}, Symbol.toStringTag, { value: "Module" }));
function Label({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    LabelPrimitive.Root,
    {
      "data-slot": "label",
      className: cn(
        "flex items-center gap-2 text-sm leading-none font-medium select-none group-data-[disabled=true]:pointer-events-none group-data-[disabled=true]:opacity-50 peer-disabled:cursor-not-allowed peer-disabled:opacity-50",
        className
      ),
      ...props
    }
  );
}
function Switch({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    SwitchPrimitive.Root,
    {
      "data-slot": "switch",
      className: cn(
        "peer data-[state=checked]:bg-primary data-[state=unchecked]:bg-input focus-visible:border-ring focus-visible:ring-ring/50 dark:data-[state=unchecked]:bg-input/80 inline-flex h-[1.15rem] w-8 shrink-0 items-center rounded-full border border-transparent shadow-xs transition-all outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50",
        className
      ),
      ...props,
      children: /* @__PURE__ */ jsx(
        SwitchPrimitive.Thumb,
        {
          "data-slot": "switch-thumb",
          className: cn(
            "bg-background dark:data-[state=unchecked]:bg-foreground dark:data-[state=checked]:bg-primary-foreground pointer-events-none block size-4 rounded-full ring-0 transition-transform data-[state=checked]:translate-x-[calc(100%-2px)] data-[state=unchecked]:translate-x-0"
          )
        }
      )
    }
  );
}
function Select({
  ...props
}) {
  return /* @__PURE__ */ jsx(SelectPrimitive.Root, { "data-slot": "select", ...props });
}
function SelectValue({
  ...props
}) {
  return /* @__PURE__ */ jsx(SelectPrimitive.Value, { "data-slot": "select-value", ...props });
}
function SelectTrigger({
  className,
  size = "default",
  children,
  ...props
}) {
  return /* @__PURE__ */ jsxs(
    SelectPrimitive.Trigger,
    {
      "data-slot": "select-trigger",
      "data-size": size,
      className: cn(
        "border-input data-[placeholder]:text-muted-foreground [&_svg:not([class*='text-'])]:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 dark:hover:bg-input/50 flex w-fit items-center justify-between gap-2 rounded-md border bg-transparent px-3 py-2 text-sm whitespace-nowrap shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 data-[size=default]:h-9 data-[size=sm]:h-8 *:data-[slot=select-value]:line-clamp-1 *:data-[slot=select-value]:flex *:data-[slot=select-value]:items-center *:data-[slot=select-value]:gap-2 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        className
      ),
      ...props,
      children: [
        children,
        /* @__PURE__ */ jsx(SelectPrimitive.Icon, { asChild: true, children: /* @__PURE__ */ jsx(ChevronDownIcon$1, { className: "size-4 opacity-50" }) })
      ]
    }
  );
}
function SelectContent({
  className,
  children,
  position = "popper",
  ...props
}) {
  return /* @__PURE__ */ jsx(SelectPrimitive.Portal, { children: /* @__PURE__ */ jsxs(
    SelectPrimitive.Content,
    {
      "data-slot": "select-content",
      className: cn(
        "bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 relative z-50 max-h-(--radix-select-content-available-height) min-w-[8rem] origin-(--radix-select-content-transform-origin) overflow-x-hidden overflow-y-auto rounded-md border shadow-md",
        position === "popper" && "data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1",
        className
      ),
      position,
      ...props,
      children: [
        /* @__PURE__ */ jsx(SelectScrollUpButton, {}),
        /* @__PURE__ */ jsx(
          SelectPrimitive.Viewport,
          {
            className: cn(
              "p-1",
              position === "popper" && "h-[var(--radix-select-trigger-height)] w-full min-w-[var(--radix-select-trigger-width)] scroll-my-1"
            ),
            children
          }
        ),
        /* @__PURE__ */ jsx(SelectScrollDownButton, {})
      ]
    }
  ) });
}
function SelectItem({
  className,
  children,
  ...props
}) {
  return /* @__PURE__ */ jsxs(
    SelectPrimitive.Item,
    {
      "data-slot": "select-item",
      className: cn(
        "focus:bg-accent focus:text-accent-foreground [&_svg:not([class*='text-'])]:text-muted-foreground relative flex w-full cursor-default items-center gap-2 rounded-sm py-1.5 pr-8 pl-2 text-sm outline-hidden select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 *:[span]:last:flex *:[span]:last:items-center *:[span]:last:gap-2",
        className
      ),
      ...props,
      children: [
        /* @__PURE__ */ jsx("span", { className: "absolute right-2 flex size-3.5 items-center justify-center", children: /* @__PURE__ */ jsx(SelectPrimitive.ItemIndicator, { children: /* @__PURE__ */ jsx(CheckIcon, { className: "size-4" }) }) }),
        /* @__PURE__ */ jsx(SelectPrimitive.ItemText, { children })
      ]
    }
  );
}
function SelectScrollUpButton({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    SelectPrimitive.ScrollUpButton,
    {
      "data-slot": "select-scroll-up-button",
      className: cn(
        "flex cursor-default items-center justify-center py-1",
        className
      ),
      ...props,
      children: /* @__PURE__ */ jsx(ChevronUpIcon, { className: "size-4" })
    }
  );
}
function SelectScrollDownButton({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    SelectPrimitive.ScrollDownButton,
    {
      "data-slot": "select-scroll-down-button",
      className: cn(
        "flex cursor-default items-center justify-center py-1",
        className
      ),
      ...props,
      children: /* @__PURE__ */ jsx(ChevronDownIcon$1, { className: "size-4" })
    }
  );
}
function Tabs({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    TabsPrimitive.Root,
    {
      "data-slot": "tabs",
      className: cn("flex flex-col gap-2", className),
      ...props
    }
  );
}
function TabsList({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    TabsPrimitive.List,
    {
      "data-slot": "tabs-list",
      className: cn(
        "bg-muted text-muted-foreground inline-flex h-9 w-fit items-center justify-center rounded-lg p-[3px]",
        className
      ),
      ...props
    }
  );
}
function TabsTrigger({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    TabsPrimitive.Trigger,
    {
      "data-slot": "tabs-trigger",
      className: cn(
        "data-[state=active]:bg-background dark:data-[state=active]:text-foreground focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:outline-ring dark:data-[state=active]:border-input dark:data-[state=active]:bg-input/30 text-foreground dark:text-muted-foreground inline-flex h-[calc(100%-1px)] flex-1 items-center justify-center gap-1.5 rounded-md border border-transparent px-2 py-1 text-sm font-medium whitespace-nowrap transition-[color,box-shadow] focus-visible:ring-[3px] focus-visible:outline-1 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:shadow-sm [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        className
      ),
      ...props
    }
  );
}
function TabsContent({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    TabsPrimitive.Content,
    {
      "data-slot": "tabs-content",
      className: cn("flex-1 outline-none", className),
      ...props
    }
  );
}
function Card({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "card",
      className: cn(
        "bg-card text-card-foreground flex flex-col gap-6 rounded-xl border py-6 shadow-sm",
        className
      ),
      ...props
    }
  );
}
function CardHeader({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "card-header",
      className: cn(
        "@container/card-header grid auto-rows-min grid-rows-[auto_auto] items-start gap-1.5 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6",
        className
      ),
      ...props
    }
  );
}
function CardTitle({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "card-title",
      className: cn("leading-none font-semibold", className),
      ...props
    }
  );
}
function CardContent({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "card-content",
      className: cn("px-6", className),
      ...props
    }
  );
}
function CardFooter({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "card-footer",
      className: cn("flex items-center px-6 [.border-t]:pt-6", className),
      ...props
    }
  );
}
function AgentEditor({
  customAgent,
  chat: chat2
}) {
  var _a, _b, _c;
  const [isPublic, setIsPublic] = useState(customAgent.isPublic);
  const [serverUrls, setServerUrls] = useState(
    ((_a = customAgent == null ? void 0 : customAgent.mcpServers) == null ? void 0 : _a.map((x) => x.url)) || [""]
  );
  const [selectedModel, setSelectedModel] = useState(
    (customAgent == null ? void 0 : customAgent.baseAgent) || chat2.currentAgent
  );
  const formRef = useRef(null);
  const submit = useSubmit();
  const autoSaveTimeoutRef = useRef(null);
  const handleFormChange = () => {
    if (formRef.current) {
      const newFormData = new FormData(formRef.current);
      if (autoSaveTimeoutRef.current) {
        clearTimeout(autoSaveTimeoutRef.current);
      }
      autoSaveTimeoutRef.current = setTimeout(() => {
        console.log("Auto-saving form...");
        submit(newFormData, {
          action: `/chat/${chat2.id}/agent/${customAgent.id}/edit`,
          method: "put"
        });
      }, 1e3);
    }
  };
  useEffect(() => {
    return () => {
      if (autoSaveTimeoutRef.current) {
        clearTimeout(autoSaveTimeoutRef.current);
      }
    };
  }, []);
  const handleUrlChange = (index2, value) => {
    const newUrls = [...serverUrls];
    newUrls[index2] = value;
    setServerUrls(newUrls);
    setTimeout(handleFormChange, 0);
  };
  const addNewUrlInput = useMemo(() => {
    const lastUrl = serverUrls[serverUrls.length - 1];
    return lastUrl !== "";
  }, [serverUrls]);
  return /* @__PURE__ */ jsx(Card, { children: /* @__PURE__ */ jsxs(
    Form,
    {
      action: `/chat/${chat2.id}/agent/${customAgent.id}/edit`,
      method: "post",
      ref: formRef,
      onChange: handleFormChange,
      children: [
        /* @__PURE__ */ jsx(CardHeader, { children: /* @__PURE__ */ jsx(CardTitle, { children: "Agent Configuration" }) }),
        /* @__PURE__ */ jsx(CardContent, { className: "space-y-6 mt-6", children: /* @__PURE__ */ jsxs(Tabs, { defaultValue: "local", className: "w-full", children: [
          /* @__PURE__ */ jsxs(TabsList, { className: "mb-4", children: [
            /* @__PURE__ */ jsx(TabsTrigger, { value: "local", children: "Local" }),
            /* @__PURE__ */ jsx(TabsTrigger, { value: "remote", children: "Remote" })
          ] }),
          /* @__PURE__ */ jsxs(TabsContent, { value: "local", className: "space-y-6", children: [
            /* @__PURE__ */ jsxs("div", { className: "space-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "name", children: "Name" }),
              /* @__PURE__ */ jsx(
                Input,
                {
                  id: "name",
                  name: "name",
                  placeholder: "Agent name",
                  defaultValue: customAgent.name
                }
              )
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "space-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "description", children: "Description" }),
              /* @__PURE__ */ jsx(
                Input,
                {
                  id: "description",
                  name: "description",
                  placeholder: "Brief description of the agent",
                  defaultValue: customAgent.description
                }
              )
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "space-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "icon", children: "Icon" }),
              /* @__PURE__ */ jsx(
                Input,
                {
                  id: "icon",
                  name: "icon",
                  placeholder: "Icon URL or path",
                  defaultValue: (_b = customAgent.icons) == null ? void 0 : _b.light
                }
              )
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "space-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "darkIcon", children: "Dark Icon" }),
              /* @__PURE__ */ jsx(
                Input,
                {
                  id: "darkIcon",
                  name: "darkIcon",
                  placeholder: "Dark mode icon URL or path",
                  defaultValue: (_c = customAgent.icons) == null ? void 0 : _c.dark
                }
              )
            ] }),
            Object.keys(chat2.agents || {}).length > 1 && /* @__PURE__ */ jsxs("div", { className: "space-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "model", children: "Model" }),
              /* @__PURE__ */ jsxs(
                Select,
                {
                  value: selectedModel,
                  onValueChange: (value) => {
                    setSelectedModel(value);
                    setTimeout(handleFormChange, 0);
                  },
                  children: [
                    /* @__PURE__ */ jsx(SelectTrigger, { children: /* @__PURE__ */ jsx(SelectValue, { placeholder: "Select a model" }) }),
                    /* @__PURE__ */ jsx(SelectContent, { children: Object.entries(chat2.agents ?? {}).map(([id, agent2]) => /* @__PURE__ */ jsx(SelectItem, { value: id, children: agent2.name || id }, id)) })
                  ]
                }
              ),
              /* @__PURE__ */ jsx("input", { type: "hidden", name: "model", value: selectedModel })
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "space-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "instructions", children: "Instructions" }),
              /* @__PURE__ */ jsx(
                Textarea,
                {
                  defaultValue: customAgent.instructions,
                  id: "instructions",
                  name: "instructions",
                  placeholder: "Enter the system prompt",
                  className: "min-h-[100px]"
                }
              )
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "flex items-center space-x-2", children: [
              /* @__PURE__ */ jsx(
                Switch,
                {
                  id: "visibility",
                  checked: isPublic,
                  onCheckedChange: (checked) => {
                    setIsPublic(checked);
                    setTimeout(handleFormChange, 0);
                  }
                }
              ),
              /* @__PURE__ */ jsx(Label, { htmlFor: "visibility", children: "Make this agent public" }),
              /* @__PURE__ */ jsx(
                "input",
                {
                  type: "hidden",
                  name: "isPublic",
                  value: isPublic ? "true" : "false"
                }
              )
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "space-y-2", children: [
              /* @__PURE__ */ jsx(Label, { children: "MCP Server URLs" }),
              serverUrls.map((url, index2) => /* @__PURE__ */ jsxs("div", { className: "flex gap-2 mb-2", children: [
                /* @__PURE__ */ jsx(
                  Input,
                  {
                    placeholder: "Enter MCP server URL",
                    value: url,
                    name: `serverUrls[${index2}]`,
                    onChange: (e) => handleUrlChange(index2, e.target.value)
                  }
                ),
                /* @__PURE__ */ jsx(
                  Button,
                  {
                    type: "button",
                    variant: "ghost",
                    className: "px-2 h-10",
                    onClick: () => {
                      const newUrls = serverUrls.filter(
                        (_, i) => i !== index2
                      );
                      setServerUrls(newUrls.length ? newUrls : [""]);
                    },
                    children: /* @__PURE__ */ jsx(MinusIcon, { className: "h-4 w-4" })
                  }
                )
              ] }, index2)),
              addNewUrlInput && /* @__PURE__ */ jsx(
                Button,
                {
                  type: "button",
                  variant: "outline",
                  onClick: () => setServerUrls([...serverUrls, ""]),
                  children: "Add Another URL"
                }
              )
            ] })
          ] }),
          /* @__PURE__ */ jsx(TabsContent, { value: "remote", className: "space-y-6", children: /* @__PURE__ */ jsxs("div", { className: "space-y-2", children: [
            /* @__PURE__ */ jsx(Label, { htmlFor: "remoteUrl", children: "Remote URL" }),
            /* @__PURE__ */ jsx(
              Input,
              {
                id: "remoteUrl",
                name: "remoteUrl",
                placeholder: "Enter remote URL",
                defaultValue: customAgent.remoteUrl || ""
              }
            )
          ] }) })
        ] }) }),
        /* @__PURE__ */ jsxs(CardFooter, { className: "flex justify-end space-x-2", children: [
          /* @__PURE__ */ jsx(Link, { to: `/agent/${customAgent.id}`, children: /* @__PURE__ */ jsx(Button, { variant: "outline", children: "Close" }) }),
          /* @__PURE__ */ jsx(Button, { type: "submit", disabled: !customAgent.id, children: "Save" })
        ] })
      ]
    }
  ) });
}
const agentEdit = UNSAFE_withComponentProps(function Page2() {
  const {
    chat: chat2,
    user
  } = useRouteLoaderData("routes/layout");
  chat2.agentEditor = true;
  return /* @__PURE__ */ jsxs("div", {
    className: "grid grid-cols-2",
    children: [/* @__PURE__ */ jsx("div", {
      className: "col-span-1 p-4 grid",
      children: /* @__PURE__ */ jsx(AgentEditor, {
        chat: chat2,
        customAgent: chat2.customAgent || {}
      })
    }), /* @__PURE__ */ jsx("div", {
      className: "col-span-1",
      children: /* @__PURE__ */ jsx(Chat, {
        chat: chat2,
        disableHeader: true,
        user
      })
    })]
  });
});
const route11 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  default: agentEdit
}, Symbol.toStringTag, { value: "Module" }));
async function action({
  request,
  params
}) {
  const formData = await request.formData();
  const servers = [];
  for (const [key, value] of formData.entries()) {
    if (key.startsWith("serverUrls[") && key.endsWith("]") && value) {
      servers.push(value);
    }
  }
  const customAgent = {
    id: params.agentId || "",
    remoteUrl: formData.get("remoteUrl"),
    name: formData.get("name"),
    description: formData.get("description"),
    icons: {
      light: formData.get("icon"),
      dark: formData.get("iconDark")
    },
    isPublic: formData.get("isPublic") === "true",
    instructions: formData.get("instructions"),
    introductionMessage: formData.get("introductionMessage"),
    baseAgent: formData.get("model"),
    mcpServers: servers.map((x) => {
      return {
        url: x
      };
    }).filter((x) => x.url)
  };
  if (request.method === "POST" && params.agentId) {
    return redirect(`/agent/${params.agentId}`);
  }
  return await updateCustomAgent(getContext(), params.id || "new", customAgent);
}
const chatAgentEdit = UNSAFE_withComponentProps(function Page3() {
  const {
    chat: chat2,
    user
  } = useRouteLoaderData("routes/layout");
  chat2.agentEditor = true;
  return /* @__PURE__ */ jsxs("div", {
    className: "grid grid-cols-2",
    children: [/* @__PURE__ */ jsx("div", {
      className: "col-span-1 p-4 grid",
      children: /* @__PURE__ */ jsx(AgentEditor, {
        chat: chat2,
        customAgent: chat2.customAgent || {}
      })
    }), /* @__PURE__ */ jsx("div", {
      className: "col-span-1",
      children: /* @__PURE__ */ jsx(Chat, {
        chat: chat2,
        disableHeader: true,
        user
      })
    })]
  });
});
const route12 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  action,
  default: chatAgentEdit
}, Symbol.toStringTag, { value: "Module" }));
const serverManifest = { "entry": { "module": "/assets/entry.client-B8NV4dut.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js", "/assets/index-DnWePWqS.js"], "css": [] }, "routes": { "root": { "id": "root", "parentId": void 0, "path": "", "index": void 0, "caseSensitive": void 0, "hasAction": false, "hasLoader": true, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": true, "module": "/assets/root-R8WEOGxs.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js", "/assets/index-DnWePWqS.js", "/assets/index-_SmGZh4S.js", "/assets/index-DxM-FJDF.js"], "css": ["/assets/root-CEcDwxn4.css"], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/action/create-resource": { "id": "routes/action/create-resource", "parentId": "root", "path": "chat/:id/resource", "index": void 0, "caseSensitive": void 0, "hasAction": true, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/create-resource-l0sNRNKZ.js", "imports": [], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/action/clone": { "id": "routes/action/clone", "parentId": "root", "path": "chat/:id/clone", "index": void 0, "caseSensitive": void 0, "hasAction": true, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/clone-l0sNRNKZ.js", "imports": [], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/action/delete-agent": { "id": "routes/action/delete-agent", "parentId": "root", "path": "chat/:id/delete-agent", "index": void 0, "caseSensitive": void 0, "hasAction": true, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/delete-agent-l0sNRNKZ.js", "imports": [], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/action/new-agent": { "id": "routes/action/new-agent", "parentId": "root", "path": "agent/new", "index": void 0, "caseSensitive": void 0, "hasAction": true, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/new-agent-l0sNRNKZ.js", "imports": [], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/login": { "id": "routes/login", "parentId": "root", "path": "login", "index": void 0, "caseSensitive": void 0, "hasAction": true, "hasLoader": true, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/login-Dsn1_hA9.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js"], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/logout": { "id": "routes/logout", "parentId": "root", "path": "logout", "index": void 0, "caseSensitive": void 0, "hasAction": false, "hasLoader": true, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/logout-C7YEhUxw.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js"], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/layout": { "id": "routes/layout", "parentId": "root", "path": void 0, "index": void 0, "caseSensitive": void 0, "hasAction": false, "hasLoader": true, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/layout-D448XXKX.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js", "/assets/use-chat-BHBTbUC_.js", "/assets/index-_SmGZh4S.js", "/assets/index-DxM-FJDF.js", "/assets/chevron-up-BrjaNDif.js", "/assets/index-DnWePWqS.js"], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/index": { "id": "routes/index", "parentId": "routes/layout", "path": void 0, "index": true, "caseSensitive": void 0, "hasAction": false, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/index-DXHSoTpT.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js", "/assets/chat-sfWFjsOZ.js", "/assets/chat-RWQutRVA.js", "/assets/use-chat-BHBTbUC_.js", "/assets/index-DnWePWqS.js", "/assets/index-_SmGZh4S.js"], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/chat": { "id": "routes/chat", "parentId": "routes/layout", "path": "chat/:id", "index": void 0, "caseSensitive": void 0, "hasAction": true, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/chat-sfWFjsOZ.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js", "/assets/chat-RWQutRVA.js", "/assets/use-chat-BHBTbUC_.js", "/assets/index-DnWePWqS.js", "/assets/index-_SmGZh4S.js"], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/agent": { "id": "routes/agent", "parentId": "routes/layout", "path": "agent/:agentId", "index": void 0, "caseSensitive": void 0, "hasAction": false, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/agent-DXHSoTpT.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js", "/assets/chat-sfWFjsOZ.js", "/assets/chat-RWQutRVA.js", "/assets/use-chat-BHBTbUC_.js", "/assets/index-DnWePWqS.js", "/assets/index-_SmGZh4S.js"], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/agent-edit": { "id": "routes/agent-edit", "parentId": "routes/layout", "path": "agent/:agentId/edit", "index": void 0, "caseSensitive": void 0, "hasAction": false, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/agent-edit-l0MG4KEo.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js", "/assets/agent-editor-Cuk1Kuqn.js", "/assets/chat-RWQutRVA.js", "/assets/use-chat-BHBTbUC_.js", "/assets/index-DnWePWqS.js", "/assets/chevron-up-BrjaNDif.js", "/assets/index-_SmGZh4S.js"], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 }, "routes/chat-agent-edit": { "id": "routes/chat-agent-edit", "parentId": "routes/layout", "path": "chat/:id/agent/:agentId/edit", "index": void 0, "caseSensitive": void 0, "hasAction": true, "hasLoader": false, "hasClientAction": false, "hasClientLoader": false, "hasClientMiddleware": false, "hasErrorBoundary": false, "module": "/assets/chat-agent-edit-l0MG4KEo.js", "imports": ["/assets/chunk-QMGIS6GS-DfTvyqh6.js", "/assets/agent-editor-Cuk1Kuqn.js", "/assets/chat-RWQutRVA.js", "/assets/use-chat-BHBTbUC_.js", "/assets/index-DnWePWqS.js", "/assets/chevron-up-BrjaNDif.js", "/assets/index-_SmGZh4S.js"], "css": [], "clientActionModule": void 0, "clientLoaderModule": void 0, "clientMiddlewareModule": void 0, "hydrateFallbackModule": void 0 } }, "url": "/assets/manifest-6fd88c32.js", "version": "6fd88c32", "sri": void 0 };
const assetsBuildDirectory = "build/client";
const basename = "/";
const future = { "unstable_middleware": false, "unstable_optimizeDeps": false, "unstable_splitRouteModules": false, "unstable_subResourceIntegrity": false, "unstable_viteEnvironmentApi": false };
const ssr = true;
const isSpaMode = false;
const prerender = [];
const routeDiscovery = { "mode": "lazy", "manifestPath": "/__manifest" };
const publicPath = "/";
const entry = { module: entryServer };
const routes = {
  "root": {
    id: "root",
    parentId: void 0,
    path: "",
    index: void 0,
    caseSensitive: void 0,
    module: route0
  },
  "routes/action/create-resource": {
    id: "routes/action/create-resource",
    parentId: "root",
    path: "chat/:id/resource",
    index: void 0,
    caseSensitive: void 0,
    module: route1
  },
  "routes/action/clone": {
    id: "routes/action/clone",
    parentId: "root",
    path: "chat/:id/clone",
    index: void 0,
    caseSensitive: void 0,
    module: route2
  },
  "routes/action/delete-agent": {
    id: "routes/action/delete-agent",
    parentId: "root",
    path: "chat/:id/delete-agent",
    index: void 0,
    caseSensitive: void 0,
    module: route3
  },
  "routes/action/new-agent": {
    id: "routes/action/new-agent",
    parentId: "root",
    path: "agent/new",
    index: void 0,
    caseSensitive: void 0,
    module: route4
  },
  "routes/login": {
    id: "routes/login",
    parentId: "root",
    path: "login",
    index: void 0,
    caseSensitive: void 0,
    module: route5
  },
  "routes/logout": {
    id: "routes/logout",
    parentId: "root",
    path: "logout",
    index: void 0,
    caseSensitive: void 0,
    module: route6
  },
  "routes/layout": {
    id: "routes/layout",
    parentId: "root",
    path: void 0,
    index: void 0,
    caseSensitive: void 0,
    module: route7
  },
  "routes/index": {
    id: "routes/index",
    parentId: "routes/layout",
    path: void 0,
    index: true,
    caseSensitive: void 0,
    module: route8
  },
  "routes/chat": {
    id: "routes/chat",
    parentId: "routes/layout",
    path: "chat/:id",
    index: void 0,
    caseSensitive: void 0,
    module: route9
  },
  "routes/agent": {
    id: "routes/agent",
    parentId: "routes/layout",
    path: "agent/:agentId",
    index: void 0,
    caseSensitive: void 0,
    module: route10
  },
  "routes/agent-edit": {
    id: "routes/agent-edit",
    parentId: "routes/layout",
    path: "agent/:agentId/edit",
    index: void 0,
    caseSensitive: void 0,
    module: route11
  },
  "routes/chat-agent-edit": {
    id: "routes/chat-agent-edit",
    parentId: "routes/layout",
    path: "chat/:id/agent/:agentId/edit",
    index: void 0,
    caseSensitive: void 0,
    module: route12
  }
};
export {
  serverManifest as assets,
  assetsBuildDirectory,
  basename,
  entry,
  future,
  isSpaMode,
  prerender,
  publicPath,
  routeDiscovery,
  routes,
  ssr
};
