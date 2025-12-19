import "clsx";
import { g as getContext, i as setContext } from "./index.js";
const WorkspaceMimeType = "application/vnd.nanobot.workspace+json";
const UIPath = "/mcp?ui";
const SvelteDate = globalThis.Date;
const SvelteSet = globalThis.Set;
const NOTIFICATIONS_KEY = Symbol("notifications");
function setNotificationContext(notifications) {
  setContext(NOTIFICATIONS_KEY, notifications);
}
function getNotificationContext() {
  return getContext(NOTIFICATIONS_KEY);
}
function logError(error) {
  try {
    const notifications = getNotificationContext();
    notifications.error("API Error", error?.toString());
  } catch {
    console.error("MCP Tool Error:", error);
  }
  console.error("Error:", error);
}
class SimpleClient {
  #url;
  #fetcher;
  #sessionId;
  #initializeResult;
  #initializationPromise;
  #externalSession;
  #sseConnection;
  #sseSubscriptions = /* @__PURE__ */ new Map();
  constructor(opts) {
    const baseUrl = opts?.baseUrl || "";
    const path = opts?.path || UIPath;
    this.#url = `${baseUrl}${path}`;
    this.#fetcher = opts?.fetcher || fetch;
    if (opts?.sessionId) {
      this.#sessionId = opts.sessionId === "new" ? void 0 : opts.sessionId;
      this.#externalSession = true;
    } else {
      const stored = this.#getStoredSession();
      if (stored) {
        this.#sessionId = stored.sessionId;
        this.#initializeResult = stored.initializeResult;
      }
      this.#externalSession = false;
    }
  }
  async deleteSession() {
    try {
      if (!this.#sessionId) {
        return;
      }
      await this.#fetcher(this.#url, {
        method: "DELETE",
        headers: {
          "Mcp-Session-Id": this.#sessionId
        }
      });
    } finally {
      this.#clearSession();
    }
  }
  #getStoredSession() {
    if (typeof window === "undefined" || typeof localStorage === "undefined") {
      return void 0;
    }
    const stored = localStorage.getItem(`mcp-session-${this.#url}`);
    if (!stored) {
      return void 0;
    }
    try {
      return JSON.parse(stored);
    } catch (e) {
      console.error("[SimpleClient] Failed to parse stored session:", e);
      return void 0;
    }
  }
  #storeSession(sessionId, initializeResult) {
    if (typeof window === "undefined" || typeof localStorage === "undefined") {
      return;
    }
    const session = {
      sessionId,
      initializeResult
    };
    localStorage.setItem(`mcp-session-${this.#url}`, JSON.stringify(session));
  }
  #clearSession() {
    if (typeof window === "undefined" || typeof localStorage === "undefined") {
      return;
    }
    localStorage.removeItem(`mcp-session-${this.#url}`);
    this.#sessionId = void 0;
    this.#initializeResult = void 0;
    if (this.#sseConnection) {
      this.#sseConnection.close();
      this.#sseConnection = void 0;
    }
    this.#sseSubscriptions.clear();
  }
  async getSessionDetails() {
    return {
      id: await this.#ensureSession(),
      initializeResult: this.#initializeResult
    };
  }
  async #initialize() {
    if (this.#initializationPromise) {
      return this.#initializationPromise;
    }
    this.#initializationPromise = (async () => {
      try {
        const initRequest = {
          jsonrpc: "2.0",
          id: crypto.randomUUID(),
          method: "initialize",
          params: {
            protocolVersion: "2024-11-05",
            capabilities: {},
            clientInfo: {
              name: "nanobot-ui",
              version: "0.0.1"
            }
          }
        };
        const initResp = await this.#fetcher(this.#url, {
          method: "POST",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify(initRequest)
        });
        if (!initResp.ok) {
          throw new Error(`Initialize failed: ${initResp.status} ${initResp.statusText}`);
        }
        const sessionId = initResp.headers.get("Mcp-Session-Id");
        if (!sessionId) {
          throw new Error("No Mcp-Session-Id header in initialize response");
        }
        const initData = await initResp.json();
        if (initData.error) {
          throw new Error(`Initialize error: ${initData.error.message}`);
        }
        this.#sessionId = sessionId;
        this.#initializeResult = initData.result;
        if (!this.#externalSession) {
          this.#storeSession(sessionId, this.#initializeResult);
        }
        const initializedRequest = {
          jsonrpc: "2.0",
          id: crypto.randomUUID(),
          method: "notifications/initialized",
          params: {}
        };
        const initializedResp = await this.#fetcher(this.#url, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Mcp-Session-Id": sessionId
          },
          body: JSON.stringify(initializedRequest)
        });
        if (!initializedResp.ok) {
          throw new Error(
            `Initialized notification failed: ${initializedResp.status} ${initializedResp.statusText}`
          );
        }
      } finally {
        this.#initializationPromise = void 0;
      }
    })();
    return this.#initializationPromise;
  }
  async #ensureSession() {
    if (!this.#sessionId) {
      await this.#initialize();
    }
    if (!this.#sessionId) {
      throw new Error("Failed to establish session");
    }
    return this.#sessionId;
  }
  async reply(id, result) {
    const sessionId = await this.#ensureSession();
    const resp = await this.#fetcher(this.#url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Mcp-Session-Id": sessionId
      },
      body: JSON.stringify({
        jsonrpc: "2.0",
        id,
        result
      })
    });
    if (resp.status === 204 || resp.status === 202) {
      return;
    }
    if (!resp.ok) {
      const text = await resp.text();
      logError(`reply: ${resp.status}: ${resp.statusText}: ${text}`);
      throw new Error(text);
    }
    try {
      const data = await resp.json();
      if (data.error?.message) {
        logError(data.error.message);
        throw new Error(data.error.message);
      }
    } catch (e) {
      if (e instanceof Error && e.message !== "Unexpected end of JSON input") {
        throw e;
      }
      console.debug("[SimpleClient] Error parsing JSON in reply:", e);
    }
  }
  async exchange(method, params, opts) {
    const sessionId = await this.#ensureSession();
    const request = {
      jsonrpc: "2.0",
      id: crypto.randomUUID(),
      method,
      params
    };
    const [basePath, existingQuery] = this.#url.split("?");
    const queryParams = new URLSearchParams(existingQuery || "");
    queryParams.set("method", method);
    if (method === "tools/call" && params && typeof params === "object" && "name" in params) {
      queryParams.set("toolcallname", String(params.name));
    }
    const url = `${basePath}?${queryParams.toString()}`;
    const resp = await this.#fetcher(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Mcp-Session-Id": sessionId
      },
      signal: opts?.abort?.signal,
      body: JSON.stringify(request)
    });
    if (resp.status === 404) {
      if (this.#externalSession) {
        throw new Error("Session not found (404). External sessions cannot be recreated.");
      }
      this.#clearSession();
      return this.exchange(method, params, { abort: opts?.abort });
    }
    if (!resp.ok) {
      const text = await resp.text();
      logError(`exchange: ${resp.status}: ${resp.statusText}: ${text}`);
      throw new Error(text);
    }
    const data = await resp.json();
    if (data.error) {
      logError(data.error.message);
      throw new Error(data.error.message);
    }
    return data.result;
  }
  async callMCPTool(name, opts) {
    const result = await this.exchange(
      "tools/call",
      {
        name,
        arguments: opts?.payload || {},
        ...opts?.async && {
          _meta: {
            "ai.nanobot.async": true,
            progressToken: opts?.progressToken
          }
        }
      },
      { abort: opts?.abort }
    );
    if (result && typeof result === "object" && "structuredContent" in result) {
      return result.structuredContent;
    }
    return result;
  }
  async listResources(opts) {
    const result = await this.exchange(
      "resources/list",
      {
        ...opts?.prefix && {
          _meta: {
            "ai.nanobot": {
              prefix: opts.prefix
            }
          }
        }
      },
      { abort: opts?.abort }
    );
    if (opts?.prefix) {
      return {
        ...result,
        resources: result.resources.filter(
          ({ name }) => opts?.prefix && name.startsWith(opts.prefix)
        )
      };
    }
    return result;
  }
  /**
   * Ensure SSE connection is established
   */
  async #ensureSSEConnection() {
    if (this.#sseConnection && this.#sseConnection.readyState !== EventSource.CLOSED) {
      return;
    }
    await this.#ensureSession();
    const [basePath, existingQuery] = this.#url.split("?");
    const queryParams = new URLSearchParams(existingQuery || "");
    queryParams.set("stream", "true");
    const sseUrl = `${basePath}?${queryParams.toString()}`;
    this.#sseConnection = new EventSource(sseUrl);
    this.#sseConnection.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.method === "notifications/resources/updated" && message.params?.uri) {
          const uri = message.params.uri;
          for (const [prefix, callbacks] of this.#sseSubscriptions.entries()) {
            if (uri.startsWith(prefix)) {
              this.#fetchResourceDetails(uri).then((resource) => {
                if (resource) {
                  callbacks.forEach((callback) => callback(resource));
                }
              });
            }
          }
        }
      } catch (e) {
        console.error("[SimpleClient] Failed to parse SSE message:", e);
      }
    };
    this.#sseConnection.onerror = (error) => {
      console.error("[SimpleClient] SSE connection error:", error);
    };
    this.#sseConnection.onopen = () => {
      console.log("[SimpleClient] SSE connection established");
    };
  }
  /**
   * Fetch resource details for a given URI
   */
  async #fetchResourceDetails(uri) {
    try {
      const result = await this.exchange("resources/read", { uri });
      if (result.resources?.length) {
        return result.resources[0];
      }
    } catch (e) {
      logError(
        `[SimpleClient] Failed to fetch resource ${uri}: ${e instanceof Error ? e.message : String(e)}`
      );
    }
    return null;
  }
  /**
   * Watch for resource changes with a given prefix.
   * Returns a cleanup function to stop watching.
   *
   * @param prefix - URI prefix to watch (e.g., 'workspace://')
   * @param callback - Called when a resource changes with the updated resource
   * @returns Cleanup function to stop watching
   */
  watchResource(prefix, callback) {
    if (!this.#sseSubscriptions.has(prefix)) {
      this.#sseSubscriptions.set(prefix, /* @__PURE__ */ new Set());
    }
    this.#sseSubscriptions.get(prefix).add(callback);
    this.#ensureSSEConnection().then(async () => {
      try {
        await this.exchange("resources/subscribe", { uri: prefix });
        console.log(`[SimpleClient] Subscribed to resources with prefix: ${prefix}`);
      } catch (e) {
        console.error(`[SimpleClient] Failed to subscribe to ${prefix}:`, e);
      }
    });
    return () => {
      const callbacks = this.#sseSubscriptions.get(prefix);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.#sseSubscriptions.delete(prefix);
          this.exchange("resources/unsubscribe", { uri: prefix }).catch((e) => {
            console.error(`[SimpleClient] Failed to unsubscribe from ${prefix}:`, e);
          });
        }
      }
      console.log(`[SimpleClient] Stopped watching resources with prefix: ${prefix}`);
      if (this.#sseSubscriptions.size === 0 && this.#sseConnection) {
        this.#sseConnection.close();
        this.#sseConnection = void 0;
      }
    };
  }
}
export {
  SvelteSet as S,
  UIPath as U,
  WorkspaceMimeType as W,
  SimpleClient as a,
  SvelteDate as b,
  getNotificationContext as g,
  setNotificationContext as s
};
