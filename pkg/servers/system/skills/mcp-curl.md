---
name: mcp-curl
description: Interact with MCP servers using curl over Streamable HTTP transport.
---

# MCP Server Interaction via curl

Use curl to interact with MCP servers over HTTP. MCP uses JSON-RPC 2.0.

## Required Headers

All requests need:
- `Content-Type: application/json`
- `Accept: application/json, text/event-stream`
- `Authorization: Bearer <api-key>`

After initialization, also include:
- `MCP-Session-Id: <session-id>` (from init response)
- `MCP-Protocol-Version: 2025-11-25`

## Authentication for Discovered Remote MCP Servers

When connecting to a remote MCP server that you've discovered, use the `MCP_API_KEY` environment variable as the bearer token in the Authorization header:

```bash
# Set the API key for discovered remote MCP servers
export MCP_API_KEY="your-api-key-here"

# Use it in the Authorization header
curl -X POST "$MCP_ENDPOINT" \
  -H "Authorization: Bearer $MCP_API_KEY" \
  ...
```

This convention ensures consistent authentication across discovered MCP servers.

## Session Lifecycle

### 1. Initialize

```bash
curl -i -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $MCP_API_KEY" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"curl-client","version":"1.0.0"}}}'
```

Save the `MCP-Session-Id` header from the response.

### 2. Send Initialized Notification

```bash
curl -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $MCP_API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'
```

### 3. List Tools

```bash
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $MCP_API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | jq .
```

### 4. Call a Tool

```bash
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $MCP_API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"tool_name","arguments":{"key":"value"}}}' | jq .
```

### 5. End Session

```bash
curl -X DELETE "$MCP_ENDPOINT" \
  -H "Authorization: Bearer $MCP_API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID"
```

## Error Handling

| Error | Meaning |
|-------|---------|
| HTTP 400 | Missing/invalid `MCP-Session-Id` |
| HTTP 401 | Bad API key (user may need to visit server's root domain to configure) |
| HTTP 404 | Session expired - reinitialize |
| `isError: true` in result | Tool execution failed |

## Streaming

For long-running operations, use `-N` to disable buffering:

```bash
curl -N -X POST "$MCP_ENDPOINT" ...
```
