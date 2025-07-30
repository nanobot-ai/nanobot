import {
  type RouteConfig,
  index,
  route,
  layout,
} from "@react-router/dev/routes";

export default [
  route("chat/:id/resource", "routes/action/create-resource.ts"),
  route("chat/:id/clone", "routes/action/clone.ts"),
  route("chat/:id/delete-agent", "routes/action/delete-agent.ts"),
  route("agent/new", "routes/action/new-agent.ts"),
  route("login", "routes/login.tsx"),
  route("logout", "routes/logout.tsx"),
  layout("routes/layout.tsx", [
    index("routes/index.tsx"),
    route("chat/:id", "routes/chat.tsx"),
    route("agent/:agentId", "routes/agent.tsx"),
    route("agent/:agentId/edit", "routes/agent-edit.tsx"),
    route("chat/:id/agent/:agentId/edit", "routes/chat-agent-edit.tsx"),
  ]),
] satisfies RouteConfig;
