package bifrost

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
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

// errReader wraps an io.Reader and substitutes a given error for io.EOF, allowing
// tests to simulate a mid-stream connection error (e.g. context cancellation).
type errReader struct {
	r   io.Reader
	err error
}

func (r *errReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if err == io.EOF {
		return n, r.err
	}
	return n, err
}

// TestParseStream_Cancellation verifies that when lines.Err() is set and the
// context was cancelled with a *mcp.RequestCancelledError, parseStream returns
// the partial response (with the error appended as text) rather than an error.
func TestParseStream_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancelErr := &mcp.RequestCancelledError{Reason: "user stopped"}
	cancel(cancelErr)

	// Partial SSE stream: created + one text delta, no completed event.
	partialSSE := strings.Join([]string{
		`data: {"type":"response.created","sequence_number":0,"response":{"id":"msg_cancel_001","model":"test-model","output":null}}`,
		``,
		`data: {"type":"response.output_item.added","sequence_number":1,"item":{"id":"item_001","type":"message","role":"assistant"},"output_index":0}`,
		``,
		`data: {"type":"response.output_text.delta","sequence_number":2,"delta":"Hello","item_id":"item_001","output_index":0,"content_index":0}`,
		``,
	}, "\n")

	reader := &errReader{r: strings.NewReader(partialSSE), err: context.Canceled}

	c := &Client{}
	got, err := c.parseStream(ctx, "test-agent", reader, nil)
	if err != nil {
		t.Fatalf("expected no error for cancellation, got: %v", err)
	}
	if got == nil {
		t.Fatal("expected a non-nil result")
	}
	if got.Output.ID != "msg_cancel_001" {
		t.Errorf("output ID: got %q, want %q", got.Output.ID, "msg_cancel_001")
	}
	if len(got.Output.Items) == 0 {
		t.Fatal("expected at least one output item")
	}
	lastItem := got.Output.Items[len(got.Output.Items)-1]
	if lastItem.Content == nil {
		t.Fatal("last item content is nil")
	}
	wantSuffix := "\n\n" + strings.ToUpper(cancelErr.Error())
	if !strings.HasSuffix(lastItem.Content.Text, wantSuffix) {
		t.Errorf("text: got %q, want suffix %q", lastItem.Content.Text, wantSuffix)
	}
}

// TestParseStream_Cancellation_EmptyStream verifies cancellation handling when
// no items have been accumulated yet (cancelled before any output item).
func TestParseStream_Cancellation_EmptyStream(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancelErr := &mcp.RequestCancelledError{Reason: "aborted"}
	cancel(cancelErr)

	// Only the created event — no output items yet.
	partialSSE := `data: {"type":"response.created","sequence_number":0,"response":{"id":"msg_cancel_002","model":"test-model","output":null}}` + "\n\n"

	reader := &errReader{r: strings.NewReader(partialSSE), err: context.Canceled}

	c := &Client{}
	got, err := c.parseStream(ctx, "test-agent", reader, nil)
	if err != nil {
		t.Fatalf("expected no error for cancellation, got: %v", err)
	}
	if got == nil {
		t.Fatal("expected a non-nil result")
	}
	if len(got.Output.Items) != 1 {
		t.Fatalf("items: got %d, want 1", len(got.Output.Items))
	}
	wantText := "\n\n" + strings.ToUpper(cancelErr.Error())
	if got.Output.Items[0].Content == nil || got.Output.Items[0].Content.Text != wantText {
		t.Errorf("text: got %q, want %q", got.Output.Items[0].Content.Text, wantText)
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
