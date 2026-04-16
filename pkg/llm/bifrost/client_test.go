package bifrost

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// loadFixture reads testdata to "replay" SSE responses and test parsing. Fixtures are recorded by simply using
// an io.TeeReader on the response body used for parseStream:
//
//	f, _ := os.Create(path)
//	streamBody := io.TeeReader(r.Body, f)
//	defer f.Close()
//	return c.parseStream(ctx, agentName, streamBody, opt.ProgressToken)
func loadFixture(t *testing.T, name string) *bytes.Reader {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("load fixture %s: %v", name, err)
	}
	return bytes.NewReader(data)
}

// TestParseStream_BedrockText covers the Bedrock path where the completed event
// carries output:null and content is delivered only through delta events.
func TestParseStream_BedrockText(t *testing.T) {
	c := &Client{}
	got, err := c.parseStream(context.Background(), "test-agent", loadFixture(t, "bedrock_text.sse"), nil)
	if err != nil {
		t.Fatalf("parseStream failed: %v", err)
	}

	if got.Model != "us.anthropic.claude-sonnet-4-6" {
		t.Errorf("model: got %q, want %q", got.Model, "us.anthropic.claude-sonnet-4-6")
	}
	if got.Output.ID != "msg_bedrock_001" {
		t.Errorf("output ID: got %q, want %q", got.Output.ID, "msg_bedrock_001")
	}
	if got.Output.Role != "assistant" {
		t.Errorf("role: got %q, want %q", got.Output.Role, "assistant")
	}
	if len(got.Output.Items) != 1 {
		t.Fatalf("items: got %d, want 1", len(got.Output.Items))
	}
	item := got.Output.Items[0]
	if item.Content == nil {
		t.Fatal("item.Content is nil")
	}
	if item.Content.Text != "Hello there!" {
		t.Errorf("text: got %q, want %q", item.Content.Text, "Hello there!")
	}
}

// TestParseStream_CompletedWithOutput covers providers that include the full
// output array in the response.completed event (e.g. OpenAI). The completed
// event's output takes precedence over what was accumulated from deltas.
func TestParseStream_CompletedWithOutput(t *testing.T) {
	c := &Client{}
	got, err := c.parseStream(context.Background(), "test-agent", loadFixture(t, "completed_with_output.sse"), nil)
	if err != nil {
		t.Fatalf("parseStream failed: %v", err)
	}

	if got.Model != "gpt-4o" {
		t.Errorf("model: got %q, want %q", got.Model, "gpt-4o")
	}
	if len(got.Output.Items) != 1 {
		t.Fatalf("items: got %d, want 1", len(got.Output.Items))
	}
	if got.Output.Items[0].Content == nil || got.Output.Items[0].Content.Text != "Hello there!" {
		t.Errorf("text: got %q, want %q", got.Output.Items[0].Content.Text, "Hello there!")
	}
}

// TestParseStream_FunctionCall covers tool-use responses where function call
// arguments are assembled from delta events.
func TestParseStream_FunctionCall(t *testing.T) {
	c := &Client{}
	got, err := c.parseStream(context.Background(), "test-agent", loadFixture(t, "function_call.sse"), nil)
	if err != nil {
		t.Fatalf("parseStream failed: %v", err)
	}

	if len(got.Output.Items) != 1 {
		t.Fatalf("items: got %d, want 1", len(got.Output.Items))
	}
	item := got.Output.Items[0]
	if item.ToolCall == nil {
		t.Fatal("item.ToolCall is nil")
	}
	if item.ToolCall.Name != "my_tool" {
		t.Errorf("name: got %q, want %q", item.ToolCall.Name, "my_tool")
	}
	if item.ToolCall.CallID != "call_abc123" {
		t.Errorf("call_id: got %q, want %q", item.ToolCall.CallID, "call_abc123")
	}
	if item.ToolCall.Arguments != `{"key":"value"}` {
		t.Errorf("arguments: got %q, want %q", item.ToolCall.Arguments, `{"key":"value"}`)
	}
}

func TestParseStream_Echo(t *testing.T) {
	c := &Client{}
	got, err := c.parseStream(context.Background(), "test-agent", loadFixture(t, "echo.sse"), nil)
	if err != nil {
		t.Fatalf("parseStream failed: %v", err)
	}

	if got.Model != "us.anthropic.claude-sonnet-4-6" {
		t.Errorf("model: got %q, want %q", got.Model, "us.anthropic.claude-sonnet-4-6")
	}
	if len(got.Output.Items) != 1 {
		t.Fatalf("items: got %d, want 1", len(got.Output.Items))
	}
	if got.Output.Items[0].Content == nil || got.Output.Items[0].Content.Text != "Echo: **1. Echo this message**" {
		t.Errorf("text: got %q, want %q", got.Output.Items[0].Content.Text, "Echo: **1. Echo this message**")
	}
}
