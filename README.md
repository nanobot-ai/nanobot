![Nanobot](docs/header.svg)

# Nanobot - Build MCP Agents

Nanobot is a way to build agents with MCP and MCP-UI by providing a flexible MCP host.
Existing applications like VSCode, Claude, Cursor, ChatGPT, and Goose all include an MCP host.
Nanobot is designed to be a standalone, open-source MCP host that can be easily deployed or integrated into
your applications. You can use Nanobot to create your own dedicated MCP and MCP-UI powered chatbot.

## What is an MCP Host?

An [MCP host](https://modelcontextprotocol.io/specification/2025-06-18/architecture/index#host) is
the service that combines MCP servers with an LLM and context to present an agent experience to a
consumer. The primary experience today is a chat interface, but it can be many other interfaces such
as voice, SMS, e-mail, AR/VR, Slack, MCP, or any other interface that can be used to interact with
an agent.

![MCP host](docs/mcp-host.png)

## Examples

Here are some examples of Nanobots in action:
- [Example 1](https://example.com)
- [Example 2](https://example.com)
- [Example 3](https://example.com)

## Installation

Nanobot can be installed via [Homebrew](https://brew.sh/):

```bash
brew install nanobot-ai/tap/nanobot
```

This will give you the `nanobot` CLI, which you can use to run and manage your MCP host.

---

## Getting Started

Create a configuration file (e.g. `nanobot.yaml`) that defines your agents and MCP servers.

**Example:**

```yaml
agents:
  main:
    model: gpt-4
    instructions: "You are a helpful assistant."
    mcpServers: myserver

mcpServers:
  myserver:
    url: http://my-mcp-server.example.com/mcp
```

Start Nanobot with:

```bash
nanobot run ./nanobot.yaml
```

The UI will be available at [http://localhost:8080](http://localhost:8080).

---

## Configuration

Nanobot supports the following providers:

- **OpenAI** (e.g. `gpt-4`)
- **Anthropic** (e.g. `claude-3`)

To use them, set the corresponding API key:

```bash
# For OpenAI models
export OPENAI_API_KEY=sk-...

# For Anthropic models
export ANTHROPIC_API_KEY=sk-ant-...
```

Nanobot automatically selects the correct provider based on the model specified.

---

## Development & Contribution

Contributions are welcome! Nanobot is still in **alpha**, so expect active development and rapid changes.

### Build from Source

```bash
make
```

### Working on the UI

The Nanobot UI lives in the `./ui` directory. To develop against it:

1. Remove the old build artifacts:

   ```bash
   rm -rf ./ui/dist
   ```

2. Rebuild the Nanobot binary:

   ```bash
   make
   ```

3. Start the UI in development mode:

   ```bash
   cd ui
   npm run dev
   ```

4. The UI must be served from port **5173**.\
   Nanobot runs on port **8080** and will forward UI requests to `:5173`.

---

## Features & Roadmap

Nanobot aims to be a **fully compliant MCP Host** and support all MCP + MCP-UI features.

| Feature Category           | Feature | Status        |
| -------------------------- |---------| ------------- |
| **MCP Core**               | TODO    | ✅ Implemented |
|                            | TODO    | 🚧 Partial    |
|                            | TODO    | ❌ Not yet     |
|                            | TODO    | ✅ Implemented |
| **MCP-UI**                 | TODO    | 🚧 Partial    |
|                            | TODO    | ✅ Implemented |
|                            | TODO    | ❌ Not yet     |

✅ = Implemented 🚧 = Partial / WIP ❌ = Not yet ⏳ = Planned

### Roadmap

- Full MCP + MCP-UI compliance
- More robust multi-agent support
- Production-ready UI
- Expanded model provider support
- Expanded authentication and security features
- Frontend integrations (Slack, SMS, email, embedded web agents)
- Easy embedding into existing apps and websites

---

## License

Nanobot is licensed under the [Apache 2.0 License](LICENSE).

---

## Links

- Website: [nanobot.ai](https://nanobot.ai)
- GitHub: [github.com/nanobot-ai/nanobot](https://github.com/nanobot-ai/nanobot)
