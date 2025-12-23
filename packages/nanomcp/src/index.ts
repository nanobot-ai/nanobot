import * as hooks from "./hooks/hooks.ts";

export type { CorsOptions } from "./lib/cors.ts";
export { cors } from "./lib/cors.ts";
export { toolResult } from "./lib/result.ts";
export { Server } from "./lib/router.ts";
export * from "./lib/types.ts";
export { hooks };
export { SimpleClient as Client } from "./lib/mcpclient.ts";
