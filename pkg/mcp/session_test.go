package mcp

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	filtercontract "github.com/obot-platform/nanobot/pkg/filter"
	"github.com/obot-platform/nanobot/pkg/mcp/auditlogs"
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

type hookRunnerFunc func(context.Context, any, any, string) (bool, error)

func (f hookRunnerFunc) RunHook(ctx context.Context, in, out any, target string) (bool, error) {
	return f(ctx, in, out, target)
}

type recordingDeleteSessionCloser struct {
	calls []bool
}

func (r *recordingDeleteSessionCloser) Close(deleteSession bool) {
	r.calls = append(r.calls, deleteSession)
}

type closeCallbackWire struct {
	onClose func(bool)
}

func (c *closeCallbackWire) Close(deleteSession bool) {
	c.onClose(deleteSession)
}

func (*closeCallbackWire) Wait() {}

func (*closeCallbackWire) Start(context.Context, WireHandler) error {
	return nil
}

func (*closeCallbackWire) Send(context.Context, Message) error {
	return nil
}

func (*closeCallbackWire) SessionID() string {
	return "test-session"
}

func (r *sequenceHookRunner) RunHook(_ context.Context, _, out any, _ string) (bool, error) {
	*(out.(*SessionMessageHook)) = r.responses[r.next]
	r.next++
	return true, nil
}

func TestSessionClosePropagatesExplicitDeleteToOwnedClosers(t *testing.T) {
	session := NewEmptySession(context.Background())
	closer := &recordingDeleteSessionCloser{}
	session.Set("clients/downstream", closer)

	session.Close(true)
	session.Close(true)

	if !slices.Equal(closer.calls, []bool{true}) {
		t.Fatalf("close calls = %v, want [true]", closer.calls)
	}
}

func TestSessionCloseDoesNotPropagateIdleCloseToOwnedClosers(t *testing.T) {
	session := NewEmptySession(context.Background())
	closer := &recordingDeleteSessionCloser{}
	session.Set("clients/downstream", closer)

	session.Close(false)

	if len(closer.calls) != 0 {
		t.Fatalf("close calls = %v, want none", closer.calls)
	}
}

func TestSessionCloseCancelsBeforeCollectingOwnedClosers(t *testing.T) {
	session := NewEmptySession(context.Background())
	closer := &recordingDeleteSessionCloser{}
	session.wire = &closeCallbackWire{onClose: func(deleteSession bool) {
		if !deleteSession {
			t.Error("wire Close called without explicit deletion")
		}
		if session.IsActive() {
			t.Error("session still active while wire is closing")
		}
		// Inject a closer into the old race window. Close must take its snapshot
		// after the wire has closed so this closer is not missed.
		session.Set("clients/late", closer)
	}}

	session.Close(true)

	if !slices.Equal(closer.calls, []bool{true}) {
		t.Fatalf("close calls = %v, want [true]", closer.calls)
	}
}

func TestSessionRejectsCloserAfterExplicitDelete(t *testing.T) {
	session := NewEmptySession(context.Background())
	session.Close(true)

	closer := &recordingDeleteSessionCloser{}
	session.Set("clients/late", closer)

	if !slices.Equal(closer.calls, []bool{true}) {
		t.Fatalf("close calls = %v, want [true]", closer.calls)
	}
	if session.Get("clients/late", nil) {
		t.Fatal("closer was attached after explicit session deletion")
	}
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

func TestCallAllHooksUsesWrappedV1EnvelopeOnlyWhenMarked(t *testing.T) {
	var calls int
	runner := hookRunnerFunc(func(_ context.Context, in, out any, target string) (bool, error) {
		calls++
		if target != "filter/tool" {
			t.Fatalf("target = %q", target)
		}
		toolRequest, ok := in.(*filtercontract.ToolRequest)
		if !ok {
			t.Fatalf("input type = %T, want *filter.ToolRequest", in)
		}
		inputJSON, err := json.Marshal(toolRequest)
		if err != nil {
			t.Fatal(err)
		}
		var inputFields map[string]json.RawMessage
		if err := json.Unmarshal(inputJSON, &inputFields); err != nil {
			t.Fatal(err)
		}
		if len(inputFields) != 1 || len(inputFields["request"]) == 0 {
			t.Fatalf("v1 tool arguments = %s, want only a request wrapper", inputJSON)
		}
		request := &toolRequest.Request
		if request.APIVersion != filtercontract.APIVersionV1 || request.Source != filtercontract.SourceMCP {
			t.Fatalf("unexpected envelope: %#v", request)
		}
		if request.Event.Method != "tools/call" || request.Event.Identifier != "search" || request.Event.Phase != filtercontract.PhaseRequest {
			t.Fatalf("unexpected event: %#v", request.Event)
		}
		if !request.Capabilities.CanReject || !request.Capabilities.CanMutate {
			t.Fatalf("unexpected capabilities: %#v", request.Capabilities)
		}
		if request.Context.MCP == nil || request.Context.MCP.ServerName != "server-id" || request.Context.MCP.ServerShortName != "Search Server" {
			t.Fatalf("unexpected MCP context: %#v", request.Context.MCP)
		}
		if request.Context.Trace == nil || request.Context.Trace.SessionID != "test-session" {
			t.Fatalf("unexpected trace context: %#v", request.Context.Trace)
		}
		response, ok := out.(*filtercontract.Response)
		if !ok {
			t.Fatalf("output type = %T, want *filter.Response", out)
		}
		*response = filtercontract.Response{Decision: filtercontract.DecisionAccept, Reason: "allowed"}
		return true, nil
	})
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{
		HookRunner: runner,
		Parent:     &Session{wire: &closeCallbackWire{}},
		hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{
			Target:          "filter/tool",
			ContractVersion: filtercontract.ContractVersionV1,
		}}}},
	}
	original := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"search","arguments":{"query":"test"}}`),
	}
	ctx := WithMCPServerConfig(WithAuditLog(t.Context(), auditLog), Server{Name: "server-id", ShortName: "Search Server"})
	message, err := s.callAllHooks(ctx, original, "request")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if message != original {
		t.Fatal("acceptance replaced the original message")
	}
	if len(auditLog.WebhookStatuses) != 1 || auditLog.WebhookStatuses[0].Status != "ok" || auditLog.WebhookStatuses[0].Message != "allowed" {
		t.Fatalf("unexpected statuses: %#v", auditLog.WebhookStatuses)
	}
}

func TestCallAllHooksV1MutationChainingAndAggregateRejection(t *testing.T) {
	mutated := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"search","arguments":{"query":"redacted"}}`),
	}
	mutatedPayload, err := json.Marshal(mutated)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	runner := hookRunnerFunc(func(_ context.Context, in, out any, target string) (bool, error) {
		toolRequest := in.(*filtercontract.ToolRequest)
		request := &toolRequest.Request
		response := out.(*filtercontract.Response)
		calls = append(calls, target)
		switch target {
		case "first/filter":
			*response = filtercontract.Response{Decision: filtercontract.DecisionMutate, Reason: "redacted", Payload: mutatedPayload}
		case "second/filter":
			if !strings.Contains(string(request.Payload), `"query":"redacted"`) {
				t.Fatalf("second Filter did not receive mutation: %s", request.Payload)
			}
			*response = filtercontract.Response{Decision: filtercontract.DecisionReject, Reason: "blocked"}
		case "third/filter":
			if !strings.Contains(string(request.Payload), `"query":"redacted"`) {
				t.Fatalf("third Filter did not receive effective payload: %s", request.Payload)
			}
			*response = filtercontract.Response{Decision: filtercontract.DecisionAccept}
		}
		return true, nil
	})
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{
		HookRunner: runner,
		hooks: Hooks{{
			Name: "tools/call",
			Targets: []HookTarget{
				{Target: "first/filter", ContractVersion: filtercontract.ContractVersionV1},
				{Target: "second/filter", ContractVersion: filtercontract.ContractVersionV1},
				{Target: "third/filter", ContractVersion: filtercontract.ContractVersionV1},
			},
		}},
	}
	original := &Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"search","arguments":{"query":"original"}}`),
	}
	message, err := s.callAllHooks(WithAuditLog(t.Context(), auditLog), original, "request")
	if err == nil || !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("error = %v, want aggregate rejection", err)
	}
	if !slices.Equal(calls, []string{"first/filter", "second/filter", "third/filter"}) {
		t.Fatalf("calls = %v", calls)
	}
	if !strings.Contains(string(message.Params), `"query":"redacted"`) {
		t.Fatalf("effective message = %s", message.Params)
	}
	if len(auditLog.WebhookStatuses) != 3 {
		t.Fatalf("statuses = %#v, want 3", auditLog.WebhookStatuses)
	}
	if got := []string{auditLog.WebhookStatuses[0].Status, auditLog.WebhookStatuses[1].Status, auditLog.WebhookStatuses[2].Status}; !slices.Equal(got, []string{"mutated", "rejected", "ok"}) {
		t.Fatalf("statuses = %v", got)
	}
}

func TestCallAllHooksRejectsForbiddenV1Mutation(t *testing.T) {
	original := &Message{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: json.RawMessage(`{"name":"search"}`)}
	payload, err := json.Marshal(&Message{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: json.RawMessage(`{"name":"search","arguments":{"query":"changed"}}`)})
	if err != nil {
		t.Fatal(err)
	}
	runner := hookRunnerFunc(func(_ context.Context, in, out any, _ string) (bool, error) {
		toolRequest := in.(*filtercontract.ToolRequest)
		request := &toolRequest.Request
		if request.Capabilities.CanMutate {
			t.Fatal("mutation capability should be false")
		}
		*(out.(*filtercontract.Response)) = filtercontract.Response{Decision: filtercontract.DecisionMutate, Reason: "redact", Payload: payload}
		return true, nil
	})
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{HookRunner: runner, hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{
		Target: "filter/tool", MutateDisallowed: true, ContractVersion: filtercontract.ContractVersionV1,
	}}}}}
	message, err := s.callAllHooks(WithAuditLog(t.Context(), auditLog), original, "request")
	if err == nil || !strings.Contains(err.Error(), mutationDisallowedReason) {
		t.Fatalf("error = %v", err)
	}
	if message != original || original.HookMutations != nil {
		t.Fatal("forbidden mutation changed the message")
	}
	if len(auditLog.WebhookStatuses) != 1 || auditLog.WebhookStatuses[0].Status != "rejected" {
		t.Fatalf("statuses = %#v", auditLog.WebhookStatuses)
	}
}

func TestCallAllHooksRejectsInvalidV1DecisionWithoutRetry(t *testing.T) {
	var calls int
	runner := hookRunnerFunc(func(_ context.Context, _, out any, _ string) (bool, error) {
		calls++
		*(out.(*filtercontract.Response)) = filtercontract.Response{Decision: "maybe"}
		return true, nil
	})
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{HookRunner: runner, hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{
		Target: "filter/tool", ContractVersion: filtercontract.ContractVersionV1,
	}}}}}
	message := &Message{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: json.RawMessage(`{"name":"search"}`)}
	_, err := s.callAllHooks(WithAuditLog(t.Context(), auditLog), message, "request")
	if err == nil || !strings.Contains(err.Error(), `unknown filter decision "maybe"`) {
		t.Fatalf("error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want no fallback or retry", calls)
	}
	if len(auditLog.WebhookStatuses) != 1 || auditLog.WebhookStatuses[0].Status != "failed" {
		t.Fatalf("statuses = %#v", auditLog.WebhookStatuses)
	}
}

func TestCallAllHooksRejectsInvalidV1MutationPayload(t *testing.T) {
	runner := hookRunnerFunc(func(_ context.Context, _, out any, _ string) (bool, error) {
		*(out.(*filtercontract.Response)) = filtercontract.Response{
			Decision: filtercontract.DecisionMutate,
			Payload:  json.RawMessage(`"not-an-mcp-message"`),
		}
		return true, nil
	})
	auditLog := &auditlogs.MCPAuditLog{}
	s := &Session{HookRunner: runner, hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{
		Target: "filter/tool", ContractVersion: filtercontract.ContractVersionV1,
	}}}}}
	message := &Message{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: json.RawMessage(`{"name":"search"}`)}
	_, err := s.callAllHooks(WithAuditLog(t.Context(), auditLog), message, "request")
	if err == nil || !strings.Contains(err.Error(), "mutation payload must be an MCP message object") {
		t.Fatalf("error = %v", err)
	}
	if len(auditLog.WebhookStatuses) != 1 || auditLog.WebhookStatuses[0].Status != "failed" {
		t.Fatalf("statuses = %#v", auditLog.WebhookStatuses)
	}
}

func TestCallAllHooksRejectsInvalidV1MutationShape(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		message   *Message
		wantError string
	}{
		{
			name:      "request with result",
			direction: "request",
			message: &Message{
				JSONRPC: "2.0",
				ID:      float64(1),
				Method:  "tools/call",
				Params:  json.RawMessage(`{"name":"search"}`),
				Result:  json.RawMessage(`{}`),
			},
			wantError: "request mutation must not contain result or error",
		},
		{
			name:      "response with result and error",
			direction: "response",
			message: &Message{
				JSONRPC: "2.0",
				ID:      float64(1),
				Method:  "tools/call",
				Params:  json.RawMessage(`{"name":"search"}`),
				Result:  json.RawMessage(`{}`),
				Error:   NewRPCError(-32603, "failed"),
			},
			wantError: "response mutation must contain exactly one of result or error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatal(err)
			}
			runner := hookRunnerFunc(func(_ context.Context, _, out any, _ string) (bool, error) {
				*(out.(*filtercontract.Response)) = filtercontract.Response{Decision: filtercontract.DecisionMutate, Payload: payload}
				return true, nil
			})
			s := &Session{HookRunner: runner, hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{
				Target: "filter/tool", ContractVersion: filtercontract.ContractVersionV1,
			}}}}}
			_, err = s.callAllHooks(t.Context(), &Message{
				JSONRPC: "2.0",
				ID:      float64(1),
				Method:  "tools/call",
				Params:  json.RawMessage(`{"name":"search"}`),
			}, tt.direction)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("error = %v, want %q", err, tt.wantError)
			}
		})
	}
}

func TestCallAllHooksRejectsV1MutationOfMessageIdentity(t *testing.T) {
	mutated, err := json.Marshal(&Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"different-tool"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	runner := hookRunnerFunc(func(_ context.Context, _, out any, _ string) (bool, error) {
		*(out.(*filtercontract.Response)) = filtercontract.Response{Decision: filtercontract.DecisionMutate, Payload: mutated}
		return true, nil
	})
	s := &Session{HookRunner: runner, hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{
		Target: "filter/tool", ContractVersion: filtercontract.ContractVersionV1,
	}}}}}
	message := &Message{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: json.RawMessage(`{"name":"search"}`)}
	_, err = s.callAllHooks(t.Context(), message, "request")
	if err == nil || !strings.Contains(err.Error(), "mutation changed immutable MCP message identity") {
		t.Fatalf("error = %v", err)
	}
}

func TestCallAllHooksUnmarkedTargetUsesLegacyContract(t *testing.T) {
	runner := hookRunnerFunc(func(_ context.Context, in, out any, _ string) (bool, error) {
		if _, ok := in.(*SessionMessageHook); !ok {
			t.Fatalf("input type = %T, want legacy *SessionMessageHook", in)
		}
		*(out.(*SessionMessageHook)) = SessionMessageHook{Accept: true}
		return true, nil
	})
	s := &Session{HookRunner: runner, hooks: Hooks{{Name: "tools/call", Targets: []HookTarget{{Target: "filter/tool"}}}}}
	message := &Message{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: json.RawMessage(`{"name":"search"}`)}
	if _, err := s.callAllHooks(t.Context(), message, "request"); err != nil {
		t.Fatal(err)
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
