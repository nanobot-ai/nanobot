package integration_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/agents"
	"github.com/nanobot-ai/nanobot/pkg/config"
	"github.com/nanobot-ai/nanobot/pkg/llm"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/servers/system"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

var numRuns = flag.Int("runs", 5, "number of times to run each prompt")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// toolCallRecorder wraps an MCP handler, records every tools/call invocation, and
// ToolHandler is a function that handles a mocked tool call and returns a result.
type ToolHandler func(mcp.CallToolRequest) *mcp.CallToolResult

// toolCallRecorder wraps an MCP handler and records every tools/call invocation.
// By default, "config" and "getSkill" pass through to the real server; all other
// tools return an error unless a custom ToolHandler is registered.
type toolCallRecorder struct {
	mu       sync.Mutex
	calls    []mcp.CallToolRequest
	handlers map[string]ToolHandler
}

func newRecorder(handlers map[string]ToolHandler) *toolCallRecorder {
	return &toolCallRecorder{handlers: handlers}
}

func (r *toolCallRecorder) wrap(inner mcp.MessageHandler) mcp.MessageHandler {
	return mcp.MessageHandlerFunc(func(ctx context.Context, msg mcp.Message) {
		if msg.Method != "tools/call" {
			inner.OnMessage(ctx, msg)
			return
		}

		var call mcp.CallToolRequest
		_ = json.Unmarshal(msg.Params, &call)
		r.mu.Lock()
		r.calls = append(r.calls, call)
		r.mu.Unlock()

		switch call.Name {
		case "config", "getSkill":
			// Always pass through: config drives agent setup, getSkill loads skills.
			inner.OnMessage(ctx, msg)
		default:
			if handler, ok := r.handlers[call.Name]; ok {
				mcp.Invoke(ctx, msg, func(_ context.Context, _ mcp.Message, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					return handler(req), nil
				})
			} else {
				mcp.Invoke(ctx, msg, func(_ context.Context, _ mcp.Message, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					return &mcp.CallToolResult{
						IsError: true,
						Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("tool %q is not available in the test environment", call.Name)}},
					}, nil
				})
			}
		}
	})
}

func (r *toolCallRecorder) find(name string) *mcp.CallToolRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.calls {
		if r.calls[i].Name == name {
			return &r.calls[i]
		}
	}
	return nil
}

// summary returns a multi-line description of all recorded tool calls with their
// arguments, for use in test failure output.
func (r *toolCallRecorder) summary() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.calls) == 0 {
		return "  (no tool calls recorded)"
	}
	var sb strings.Builder
	for i, c := range r.calls {
		args, _ := json.MarshalIndent(c.Arguments, "    ", "  ")
		fmt.Fprintf(&sb, "  [%d] %s\n    args: %s\n", i+1, c.Name, args)
	}
	return sb.String()
}

// stubMCPHandler returns a minimal, no-tool MCP server used as a placeholder for
// servers that aren't needed in the test but are referenced by the config hook.
func stubMCPHandler() mcp.MessageHandler {
	stubTools := mcp.NewServerTools()
	return mcp.MessageHandlerFunc(func(ctx context.Context, msg mcp.Message) {
		switch msg.Method {
		case "initialize":
			mcp.Invoke(ctx, msg, func(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
				return &mcp.InitializeResult{
					ProtocolVersion: params.ProtocolVersion,
					Capabilities:    mcp.ServerCapabilities{Tools: &mcp.ToolsServerCapability{}},
					ServerInfo:      mcp.ServerInfo{Name: "stub", Version: "0"},
				}, nil
			})
		case "notifications/initialized", "notifications/cancelled":
		case "tools/list":
			mcp.Invoke(ctx, msg, stubTools.List)
		case "tools/call":
			mcp.Invoke(ctx, msg, stubTools.Call)
		default:
			msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%s", msg.Method))
		}
	})
}

// newTestRuntime creates a Runtime and installs a toolCallRecorder on nanobot.system.
// The agent has no pre-loaded instructions; it discovers the workflows skill naturally
// via the getSkill tool that the config hook injects.
func newTestRuntime(t *testing.T, completer types.Completer, recorder *toolCallRecorder) context.Context {
	t.Helper()

	rt, err := runtime.NewRuntime(context.Background(), llm.Config{})
	if err != nil {
		t.Fatalf("NewRuntime failed: %v", err)
	}

	// Wrap the real system server so we can record and intercept tool calls.
	rt.AddServer("nanobot.system", func(string) mcp.MessageHandler {
		return recorder.wrap(system.NewServer("", ""))
	})
	// nanobot.tasks is only registered when LoopbackURL+Store are set; stub it so
	// the config hook doesn't fail when it adds the server to MCPServers.
	rt.AddServer("nanobot.tasks", func(string) mcp.MessageHandler { return stubMCPHandler() })

	// Load the production config (builtin nanobot agent) from the standard location.
	// config.Load handles a missing .nanobot/ directory gracefully.
	cfg, _, err := config.Load(context.Background(), ".nanobot/", true)
	if err != nil {
		t.Fatalf("config.Load failed: %v", err)
	}

	agentSvc := agents.New(completer, rt.Service)
	ctx := rt.WithTempSession(context.Background(), cfg)
	return context.WithValue(ctx, agentSvcKey{}, agentSvc)
}

type agentSvcKey struct{}

func runAgent(ctx context.Context, prompt string) error {
	svc := ctx.Value(agentSvcKey{}).(*agents.Agents)
	_, err := svc.Complete(ctx, types.CompletionRequest{
		Agent: "nanobot",
		Input: []types.Message{{
			Role: "user",
			Items: []types.CompletionItem{{
				Content: &mcp.Content{Type: "text", Text: prompt},
			}},
		}},
	})
	return err
}
