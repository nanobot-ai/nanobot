# Integration Tests

End-to-end tests that run against a real LLM. Excluded from `go test ./...` by default via the `//go:build integration` build tag.

## Running

```bash
ANTHROPIC_API_KEY=... go test -tags integration ./integration_test/ -runs 5
```

- `-tags integration` enables compilation of the tests (required; without it they are completely skipped).
- `-runs` controls how many times each prompt is run per test (default: 5).

## Agent Configuration

The harness loads the builtin `nanobot` agent via `config.Load(".nanobot/", true)`. If a `.nanobot/` directory exists in the working directory when the tests run, its config is merged in — allowing you to override the agent's instructions, add MCP servers, or adjust other settings without changing test code.

If no `.nanobot/` directory exists, the tests run against the default builtin agent.

## Adding a Test

Tests use shared tooling:

- **`newTestRuntime(t, completer, recorder)`** — creates a `Runtime` wired with a recording/intercepting system server and returns a context and agent service ready for execution.
- **`runAgent(ctx, svc, prompt)`** — runs the agent with the given user prompt.
- **`newRecorder(handlers)`** — creates a `toolCallRecorder` with custom `ToolHandler`s for specific tools. By default, `config` and `getSkill` always pass through to the real server; all other tools return an error unless a handler is registered.
- **`recorder.find(name)`** — returns the first recorded call for a tool, or nil.
- **`recorder.summary()`** — returns a formatted list of all tool calls with arguments, useful in failure messages.

### Example

```go
func TestMySkillBehavior(t *testing.T) {
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        t.Fatal("ANTHROPIC_API_KEY must be set when running integration tests")
    }

    completer := llm.NewClient(llm.Config{
        DefaultModel: "claude-sonnet-4-6",
        Anthropic:    anthropic.Config{APIKey: apiKey},
    })

    recorder := newRecorder(map[string]ToolHandler{
        // Register tools the agent is expected to call with mock responses.
        "bash": func(req mcp.CallToolRequest) *mcp.CallToolResult {
            return &mcp.CallToolResult{
                Content: []mcp.Content{{Type: "text", Text: "mock output"}},
            }
        },
    })

    ctx, svc := newTestRuntime(t, completer, recorder)
    if err := runAgent(ctx, svc, "my prompt"); err != nil {
        t.Fatalf("Complete() failed: %v", err)
    }

    call := recorder.find("bash")
    if call == nil {
        t.Fatalf("expected bash call\nall tool calls:\n%s", recorder.summary())
    }
    // assert call.Arguments as needed
}
```
