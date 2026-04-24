package mcp

import (
	"context"
	"encoding/json"
	"slices"
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

type sequenceHookRunner struct {
	responses []SessionMessageHook
	next      int
}

func (r *sequenceHookRunner) RunHook(_ context.Context, _, out any, _ string) (bool, error) {
	*(out.(*SessionMessageHook)) = r.responses[r.next]
	r.next++
	return true, nil
}

func TestCallAllHooksRecordsMutatedToolRequestBody(t *testing.T) {
	mutated := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"test","arguments":{"value":"mutated"}}`),
	}
	reason := "request was normalized"
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{
		HookRunner: staticHookRunner{response: SessionMessageHook{
			Accept:  true,
			Mutated: true,
			Reason:  reason,
			Message: mutated,
		}},
		hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{Target: "test-hook"}}}},
	}

	msg, err := s.callAllHooks(WithAuditLog(context.Background(), auditLog), &Message{
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
	assertHookMutation(t, msg.HookMutations, "request", reason)
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
	reason := "response was filtered"
	originalBytes, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{
		HookRunner: staticHookRunner{response: SessionMessageHook{
			Accept:  true,
			Mutated: true,
			Reason:  reason,
			Message: mutated,
		}},
		hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{Target: "test-hook"}}}},
	}

	msg, err := s.callAllHooks(WithAuditLog(context.Background(), auditLog), original, "response")
	if err != nil {
		t.Fatal(err)
	}

	assertJSONEqual(t, auditLog.OriginalResponseBody, json.RawMessage(originalBytes))
	if len(auditLog.MutatedRequestBody) != 0 {
		t.Fatalf("mutated request body was recorded for response mutation: %s", auditLog.MutatedRequestBody)
	}
	assertHookMutation(t, msg.HookMutations, "response", reason)
}

func TestAddHookMutationsMeta(t *testing.T) {
	resp := &Message{
		Result: json.RawMessage(`{"content":[{"type":"text","text":"ok"}],"_meta":{"existing":true}}`),
		HookMutations: map[string]HookMutation{
			"request":  {Mutated: true, Reasons: []string{"request was normalized"}},
			"response": {Mutated: true, Reasons: []string{"response was filtered"}},
		},
	}

	if err := addHookMutationsMeta(resp); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Content []Content      `json:"content"`
		Meta    map[string]any `json:"_meta"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatal(err)
	}
	if result.Meta["existing"] != true {
		t.Fatalf("existing metadata was not preserved: %#v", result.Meta)
	}
	if len(result.Content) != 1 || result.Content[0].Text != "ok" {
		t.Fatalf("content should not be modified: %#v", result.Content)
	}

	mutations, ok := result.Meta[HookMutationsMetaKey].(map[string]any)
	if !ok {
		t.Fatalf("missing hook mutations metadata: %#v", result.Meta)
	}
	assertMutationMeta(t, mutations, "request", "request was normalized")
	assertMutationMeta(t, mutations, "response", "response was filtered")
	if _, ok := mutations["original"]; ok {
		t.Fatalf("original value leaked into metadata: %#v", mutations)
	}
}

func TestAddHookMutationsMetaOmitsUnmutatedDirections(t *testing.T) {
	resp := &Message{
		Result: json.RawMessage(`{"content":[{"type":"text","text":"ok"}]}`),
		HookMutations: map[string]HookMutation{
			"request": {Mutated: true, Reasons: []string{"request was normalized"}},
		},
	}

	if err := addHookMutationsMeta(resp); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Content []Content                          `json:"content"`
		Meta    map[string]map[string]HookMutation `json:"_meta"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 1 || result.Content[0].Text != "ok" {
		t.Fatalf("content should not be modified: %#v", result.Content)
	}
	mutations := result.Meta[HookMutationsMetaKey]
	assertHookMutation(t, mutations, "request", "request was normalized")
	if _, ok := mutations["response"]; ok {
		t.Fatalf("unmutated response metadata should be omitted: %#v", mutations)
	}
}

func TestResponseTypesPreserveHookMutationMeta(t *testing.T) {
	tests := []struct {
		name   string
		result json.RawMessage
		out    any
	}{
		{name: "elicit", result: json.RawMessage(`{"action":"accept"}`), out: &ElicitResult{}},
		{name: "list roots", result: json.RawMessage(`{"roots":[]}`), out: &ListRootsResult{}},
		{name: "create message", result: json.RawMessage(`{"content":{"type":"text","text":"ok"},"role":"assistant"}`), out: &CreateMessageResult{}},
		{name: "call tool", result: json.RawMessage(`{"content":[]}`), out: &CallToolResult{}},
		{name: "list tools", result: json.RawMessage(`{"tools":[]}`), out: &ListToolsResult{}},
		{name: "get prompt", result: json.RawMessage(`{"messages":[]}`), out: &GetPromptResult{}},
		{name: "read resource", result: json.RawMessage(`{"contents":[]}`), out: &ReadResourceResult{}},
		{name: "list resource templates", result: json.RawMessage(`{"resourceTemplates":[]}`), out: &ListResourceTemplatesResult{}},
		{name: "list resources", result: json.RawMessage(`{"resources":[]}`), out: &ListResourcesResult{}},
		{name: "list prompts", result: json.RawMessage(`{"prompts":[]}`), out: &ListPromptsResult{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Message{
				Result: tt.result,
				HookMutations: map[string]HookMutation{
					"request": {Mutated: true, Reasons: []string{"request was normalized"}},
				},
			}
			if err := addHookMutationsMeta(resp); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(resp.Result, tt.out); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(tt.out)
			if err != nil {
				t.Fatal(err)
			}

			var result struct {
				Meta map[string]any `json:"_meta"`
			}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatal(err)
			}
			mutations, ok := result.Meta[HookMutationsMetaKey].(map[string]any)
			if !ok {
				t.Fatalf("missing hook mutation metadata after typed round trip: %s", data)
			}
			assertMutationMeta(t, mutations, "request", "request was normalized")
		})
	}
}

func TestCallAllHooksAccumulatesMutationReasons(t *testing.T) {
	first := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"test","arguments":{"value":"first"}}`),
	}
	second := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"test","arguments":{"value":"second"}}`),
	}
	runner := &sequenceHookRunner{responses: []SessionMessageHook{
		{Accept: true, Mutated: true, Reason: "first mutation", Message: first},
		{Accept: true, Mutated: true, Reason: "second mutation", Message: second},
	}}
	s := &Session{
		HookRunner: runner,
		hooks: Hooks{
			{Name: "tools/call", Targets: []HookTarget{{Target: "first-hook"}}},
			{Name: "tools/call", Targets: []HookTarget{{Target: "second-hook"}}},
		},
	}

	msg, err := s.callAllHooks(context.Background(), &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"test","arguments":{"value":"original"}}`),
	}, "request")
	if err != nil {
		t.Fatal(err)
	}

	assertHookMutation(t, msg.HookMutations, "request", "first mutation", "second mutation")
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

func assertHookMutation(t *testing.T, mutations map[string]HookMutation, direction string, reasons ...string) {
	t.Helper()
	mutation, ok := mutations[direction]
	if !ok {
		t.Fatalf("missing %s hook mutation: %#v", direction, mutations)
	}
	if !mutation.Mutated || !slices.Equal(mutation.Reasons, reasons) {
		t.Fatalf("unexpected %s hook mutation: %#v", direction, mutation)
	}
}

func assertMutationMeta(t *testing.T, mutations map[string]any, direction string, reasons ...string) {
	t.Helper()
	mutation, ok := mutations[direction].(map[string]any)
	if !ok {
		t.Fatalf("missing %s hook mutation metadata: %#v", direction, mutations)
	}
	actualReasons, ok := mutation["reasons"].([]any)
	if !ok {
		t.Fatalf("missing %s hook mutation reasons: %#v", direction, mutation)
	}
	if len(actualReasons) != len(reasons) {
		t.Fatalf("unexpected %s hook mutation reasons: %#v", direction, mutation)
	}
	for i, reason := range reasons {
		if actualReasons[i] != reason {
			t.Fatalf("unexpected %s hook mutation reasons: %#v", direction, mutation)
		}
	}
	if mutation["mutated"] != true {
		t.Fatalf("unexpected %s hook mutation metadata: %#v", direction, mutation)
	}
}
