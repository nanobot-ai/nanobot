package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp/auditlogs"
)

type staticHookRunner struct {
	response SessionMessageHook
}

func (r staticHookRunner) RunHook(_ context.Context, _, out any, _ string) (bool, error) {
	*(out.(*SessionMessageHook)) = r.response
	return true, nil
}

func TestCallAllHooksRecordsMutatedToolRequestBody(t *testing.T) {
	mutated := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"test","arguments":{"value":"mutated"}}`),
	}
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{
		HookRunner: staticHookRunner{response: SessionMessageHook{
			Accept:  true,
			Mutated: true,
			Message: mutated,
		}},
		hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{Target: "test-hook"}}}},
	}

	_, err := s.callAllHooks(WithAuditLog(context.Background(), auditLog), &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"test","arguments":{"value":"original"}}`),
	}, "request")
	if err != nil {
		t.Fatal(err)
	}

	assertJSONEqual(t, auditLog.MutatedRequestBody, mutated)
	if len(auditLog.OriginalResponseBody) != 0 {
		t.Fatalf("original response body was recorded for request mutation: %s", auditLog.OriginalResponseBody)
	}
}

func TestCallAllHooksRecordsOriginalToolResponseBody(t *testing.T) {
	mutated := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Result:  json.RawMessage(`{"content":[{"type":"text","text":"mutated"}]}`),
	}
	original := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Result:  json.RawMessage(`{"content":[{"type":"text","text":"original"}]}`),
	}
	originalBytes, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{
		HookRunner: staticHookRunner{response: SessionMessageHook{
			Accept:  true,
			Mutated: true,
			Message: mutated,
		}},
		hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{Target: "test-hook"}}}},
	}

	_, err = s.callAllHooks(WithAuditLog(context.Background(), auditLog), original, "response")
	if err != nil {
		t.Fatal(err)
	}

	assertJSONEqual(t, auditLog.OriginalResponseBody, json.RawMessage(originalBytes))
	if len(auditLog.MutatedRequestBody) != 0 {
		t.Fatalf("mutated request body was recorded for response mutation: %s", auditLog.MutatedRequestBody)
	}
}

func assertJSONEqual(t *testing.T, actual json.RawMessage, expected any) {
	t.Helper()

	expectedBytes, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}

	var actualValue any
	if err := json.Unmarshal(actual, &actualValue); err != nil {
		t.Fatalf("failed to unmarshal actual JSON: %v", err)
	}
	var expectedValue any
	if err := json.Unmarshal(expectedBytes, &expectedValue); err != nil {
		t.Fatalf("failed to unmarshal expected JSON: %v", err)
	}

	actualBytes, _ := json.Marshal(actualValue)
	expectedBytes, _ = json.Marshal(expectedValue)
	if string(actualBytes) != string(expectedBytes) {
		t.Fatalf("JSON mismatch\nactual:   %s\nexpected: %s", actualBytes, expectedBytes)
	}
}
