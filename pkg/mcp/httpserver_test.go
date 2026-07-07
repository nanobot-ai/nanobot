package mcp

import (
	"encoding/json"
	"testing"
)

func TestExtractObotAuditCorrelationID(t *testing.T) {
	tests := []struct {
		name   string
		params any
		want   string
	}{
		{
			name: "tool arguments meta",
			params: map[string]any{
				"name": "search",
				"arguments": map[string]any{
					"query": "obot",
					"_meta": map[string]any{
						"obotAuditCorrelationID": "corr-arguments",
					},
				},
			},
			want: "corr-arguments",
		},
		{
			name: "top-level meta fallback",
			params: map[string]any{
				"name":      "search",
				"arguments": map[string]any{"query": "obot"},
				"_meta": map[string]any{
					"obotAuditCorrelationID": "corr-top",
				},
			},
			want: "corr-top",
		},
		{
			name: "missing",
			params: map[string]any{
				"name":      "search",
				"arguments": map[string]any{"query": "obot"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatal(err)
			}
			got := extractObotAuditCorrelationID(Message{
				Method: "tools/call",
				Params: params,
			})
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestExtractObotAuditCorrelationIDIgnoresOtherMethods(t *testing.T) {
	got := extractObotAuditCorrelationID(Message{
		Method: "prompts/get",
		Params: json.RawMessage(`{"_meta":{"obotAuditCorrelationID":"corr"}}`),
	})
	if got != "" {
		t.Fatalf("expected no correlation ID for prompts/get, got %q", got)
	}
}
