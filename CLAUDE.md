# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Nanobot is a standalone MCP (Model Context Protocol) host that enables building agents with MCP and MCP-UI. Unlike built-in MCP hosts in applications like VSCode, Claude, or ChatGPT, Nanobot is designed to be an open-source, deployable solution that combines MCP servers with LLMs to create agent experiences through external interfaces such as Obot. The project is written in Go.

**Technology Stack:**
- Go 1.26.0 with GORM (SQLite, MySQL, PostgreSQL) and goja (JavaScript runtime for hooks)

## Build and Development Commands

### Backend (Go)

```bash
# Build the nanobot binary
make

# Run nanobot with a configuration file
./bin/nanobot run ./nanobot.yaml

# Run Go tests
go test ./...

# Run a specific test
go test ./pkg/agents -run TestName

# Format Go code
gofmt -w .
```

## Architecture Overview

### Backend Architecture

**Core Components:**

- **Runtime (`pkg/runtime/`)** - Main orchestration layer that initializes the system. Creates and wires together the LLM client, tool service, agents, and sampling components. Manages the overall execution environment.

- **Agents (`pkg/agents/`)** - Agent execution engine that handles tool mapping, request population, and agent interactions. Responsible for running agents backed by LLMs with access to tools.

- **Tools Service (`pkg/tools/`)** - Central registry for tools and agents. Handles tool discovery, mapping, and execution delegation. Manages MCP server connections.

- **MCP Layer (`pkg/mcp/`)** - MCP protocol implementation including sessions, clients, servers, and wire protocols. Handles both stdio and HTTP transports. Key types:
  - `Session` - Manages MCP session lifecycle and message routing
  - `Client` - MCP client implementation for connecting to servers
  - `Wire` - Protocol transport abstraction (stdio/HTTP)

- **LLM Integration (`pkg/llm/`)** - Abstraction over different LLM providers (OpenAI, Anthropic). Routes requests to appropriate providers based on model names. Handles both completion and response APIs.

- **Session Management (`pkg/session/`, `pkg/sessiondata/`)** - Manages user sessions, conversation state, and session-scoped data. Handles agent context, tool mappings, and resource mappings within sessions. Supports parent-child session relationships and database-backed OAuth token storage.

- **Server Layer (`pkg/server/`)** - HTTP server handling MCP protocol over HTTP. Routes requests for initialize, tools/list, tools/call, prompts/*, resources/*, etc. Manages session creation and request routing.

- **Built-in MCP Servers (`pkg/servers/`)** - Nanobot includes several built-in MCP servers:
  - `agent/` - Exposes individual agents as MCP servers with chat capabilities
  - `capabilities/` - Session initialization and capability management (workspace setup)
  - `meta/` - Metadata and introspection tools (list_chats, update_chat, list_agents)
  - `resources/` - Database-backed resource management (create_resource, delete_resource) with automatic mimetype detection
  - `workspace/` - Workspace and session management (create/update/delete workspaces, session reading)

- **Configuration (`pkg/config/`)** - YAML-based configuration loading and validation. Supports profiles, extends (inheritance), and environment variables. See `pkg/config/schema.yaml` for the complete schema.

**Key Architectural Patterns:**

- **Tool Mappings** - Tools from MCP servers are mapped to agent-accessible tools. The `BuildToolMappings` method creates this mapping by resolving tool references from agents and MCP servers.

- **Hooks** - Lifecycle hooks for agents and MCP servers (config, request, response). Hooks are TypeScript/JavaScript functions that can modify configuration and messages. See `hooks.ts` for type definitions.

- **Sandboxing** - MCP servers can run in Docker containers for isolation. The `pkg/mcp/sandbox/` handles containerization and port mapping.

Nanobot does not bundle a frontend. External UIs communicate through MCP endpoints, `/api/events`, and the browser proxy routes.

## Configuration

Configuration is YAML-based. Key top-level sections:

- `agents` - Define agents with their models, tools, instructions, and behaviors
- `mcp-servers` - MCP server configurations (command, URL, Docker image, etc.)
- `prompts` - Template definitions
- `publish` - Defines what to expose when Nanobot itself acts as an MCP server
- `env` - Environment variable definitions with descriptions and defaults
- `auth` - Authentication configuration (OAuth, remote headers)
- `profiles` - Configuration profiles for different environments
- `extends` - Inherit from other configuration files

Example minimal configuration:

```yaml
agents:
  myagent:
    name: My Agent
    model: gpt-4
    mcpServers: my-mcp-server

mcpServers:
  my-mcp-server:
    url: https://example.com/mcp
```

## Important Go Packages

- `pkg/types/` - Core type definitions shared across the system (Config, Agent, Message, ToolCall, etc.)
- `pkg/complete/` - Utility package for handling option completion and merging
- `pkg/expr/` - Expression evaluation for dynamic values in configurations
- `pkg/schema/` - JSON Schema validation and manipulation
- `pkg/supervise/` - Process supervision for running MCP server subprocesses
- `pkg/sampling/` - Handles MCP sampling requests (LLM-in-the-loop)
- `pkg/envvar/` - Environment variable handling with descriptions and defaults
- `pkg/cmd/` - CLI command handling (routes from `main.go`)

## Entry Points and Special Modes

**Main Entry Point:** `main.go` routes commands to `pkg/cmd`

**Special `_exec` Mode:** Nanobot can act as a daemon wrapper for MCP server subprocesses. When invoked with `_exec` as the first argument, it handles stdio piping and process lifecycle management for MCP servers. This enables Nanobot to supervise and manage MCP server processes.

## Testing

Go tests follow standard Go conventions:
- Test files are named `*_test.go`
- Run all tests: `go test ./...`
- Run specific package tests: `go test ./pkg/agents`
- Run specific test: `go test ./pkg/agents -run TestName`

## Common Patterns

### Adding a New Agent Hook

1. Define TypeScript types in `hooks.ts` (root level)
2. Update corresponding Go types in `pkg/types/hooks.go`
3. Implement hook handling in `pkg/agents/` or relevant package
4. Hook execution is managed through `pkg/mcp/hooks.go`

### Adding a New Tool

1. Tools come from MCP servers (external or built-in servers in `pkg/servers/`)
2. Tool resolution happens in `pkg/tools/service.go`

### Working with Sessions

- Sessions are scoped to MCP connections
- Use `mcp.SessionFromContext(ctx)` to get current session
- Session state includes tool mappings, current agent, and custom attributes
- Parent sessions can be accessed via `Session.Parent`
- Root session can be accessed via `session.Root()`

## MCP Protocol Notes

Nanobot supports both MCP standard and MCP-UI extensions:
- Standard MCP: tools, prompts, resources, sampling
- MCP-UI: Elicitations (user input prompts), progress notifications, structured UI elements

When implementing MCP features, refer to:
- MCP types in `pkg/mcp/types.go`
- Message handling in `pkg/mcp/message.go`
- Protocol reference at `https://modelcontextprotocol.io`

## Code Style

- Go: Follow standard Go conventions, use `gofmt`
