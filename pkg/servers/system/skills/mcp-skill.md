---
name: mcp-curl
description: Interact with Model Context Protocol (MCP) servers using curl over Streamable HTTP transport. Use when debugging MCP servers, testing MCP endpoints, manually calling MCP tools, or when the user needs to interact with an MCP server without a dedicated client.
---

# MCP Server Interaction via curl

Interact with MCP servers using curl over the Streamable HTTP transport. MCP uses JSON-RPC 2.0 for message encoding with UTF-8.

## Required Headers

All requests must include:
- `Content-Type: application/json`
- `Accept: application/json, text/event-stream`
- `Authorization: Bearer <api-key>` (ask the user for their API key)

After initialization, also include:
- `MCP-Session-Id: <session-id>` (returned by server during init)
- `MCP-Protocol-Version: 2025-11-25`

If you still receive `401 Unauthorized` after providing a valid API key, the user may need to visit the server's root domain (e.g., `https://demo.obot.ai`) to complete server configuration.

## Session Lifecycle

### 1. Initialize Session

```bash
curl -i -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-11-25",
      "capabilities": {},
      "clientInfo": {"name": "curl-client", "version": "1.0.0"}
    }
  }'
```

Save the `MCP-Session-Id` header from the response.

### 2. Send Initialized Notification

```bash
curl -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc": "2.0", "method": "notifications/initialized"}'
```

Server responds with `202 Accepted` (no body).

### 3. List Tools

```bash
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | jq .
```

### 4. Call a Tool

```bash
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {"name": "tool_name", "arguments": {"key": "value"}}
  }' | jq .
```

### 5. End Session

```bash
curl -X DELETE "$MCP_ENDPOINT" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID"
```

## Complete Example Script

For a full working example with session management and error handling, see [reference.md](reference.md).

## Error Handling

| Error | Meaning |
|-------|---------|
| HTTP 400 | Missing or invalid `MCP-Session-Id` header |
| HTTP 404 | Session expired or terminated - reinitialize |
| JSON-RPC error `-32602` | Invalid params (e.g., unknown tool) |
| `isError: true` in result | Tool execution failed - check error message |

## Response Content Types

Tool responses contain `content` array with typed items:
- **Text**: `{"type": "text", "text": "..."}`
- **Image**: `{"type": "image", "data": "base64...", "mimeType": "image/png"}`
- **Structured**: Check `structuredContent` field for parsed JSON

## SSE Streaming

For long-running operations, use `-N` flag to disable buffering:

```bash
curl -N -X POST "$MCP_ENDPOINT" ...
```

Events arrive as:
```
id: event-id
data: {"jsonrpc":"2.0","id":6,"result":{...}}
```

## Pagination

If `tools/list` response includes `nextCursor`, fetch more:

```bash
-d '{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{"cursor":"<cursor>"}}'
```

# MCP curl Reference

Complete reference for interacting with MCP servers via curl.

## Complete Example Session

```bash
#!/bin/bash
# MCP curl session example

# Store the endpoint and API key
MCP_ENDPOINT="https://example.com/mcp"
API_KEY="your-api-key-here"

# 1. Initialize and capture session ID
RESPONSE=$(curl -s -i -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"curl-client","version":"1.0.0"}}}')

SESSION_ID=$(echo "$RESPONSE" | grep -i "mcp-session-id:" | cut -d' ' -f2 | tr -d '\r')
echo "Session ID: $SESSION_ID"

# 2. Send initialized notification
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'

# 3. List available tools
echo "Available tools:"
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | jq .

# 4. Call a tool
echo "Calling tool:"
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_weather","arguments":{"location":"Seattle"}}}' | jq .

# 5. Cleanup (optional)
curl -X DELETE "$MCP_ENDPOINT" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID"
```

## Initialize Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2025-11-25",
    "capabilities": {
      "tools": {
        "listChanged": true
      }
    },
    "serverInfo": {
      "name": "ExampleServer",
      "version": "1.0.0"
    }
  }
}
```

## tools/list Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "get_weather",
        "description": "Get current weather information for a location",
        "inputSchema": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "City name or zip code"
            }
          },
          "required": ["location"]
        }
      }
    ]
  }
}
```

## tools/call Response Format

### Successful Response

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Current weather in Seattle:\nTemperature: 72°F\nConditions: Partly cloudy"
      }
    ],
    "isError": false
  }
}
```

### Tool Execution Error

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Invalid date format: expected YYYY-MM-DD"
      }
    ],
    "isError": true
  }
}
```

### Protocol Error

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "error": {
    "code": -32602,
    "message": "Unknown tool: invalid_tool_name"
  }
}
```

## Structured Content

Some tools return structured JSON in `structuredContent`:

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "content": [
      {"type": "text", "text": "{\"temperature\": 22.5, \"humidity\": 65}"}
    ],
    "structuredContent": {
      "temperature": 22.5,
      "humidity": 65
    }
  }
}
```

Prefer `structuredContent` when available for easier parsing.

## Image Content

```json
{
  "content": [
    {
      "type": "image",
      "data": "base64-encoded-data",
      "mimeType": "image/png"
    }
  ]
}
```

## Tool with No Arguments

```bash
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "get_current_time",
      "arguments": {}
    }
  }'
```

## Pagination for tools/list

If the response includes `nextCursor`, fetch more tools:

```bash
curl -s -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/list",
    "params": {
      "cursor": "<next-page-cursor>"
    }
  }'
```

## SSE Streaming Details

For long-running operations, the server may respond with `Content-Type: text/event-stream`. Use the `-N` flag to disable buffering:

```bash
curl -N -X POST "$MCP_ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer $API_KEY" \
  -H "MCP-Session-Id: $SESSION_ID" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"long_running_tool","arguments":{}}}'
```

SSE events have format:
```
id: event-id
data: {"jsonrpc":"2.0","id":6,"result":{...}}
```
