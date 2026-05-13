package responses

import (
	"encoding/json"
	"testing"

	"github.com/obot-platform/nanobot/pkg/types"
)

func TestToRequestOmitsSamplingControls(t *testing.T) {
	temperature := json.Number("0.7")
	topP := json.Number("0.9")

	req, err := toRequest(&types.CompletionRequest{
		Model:       "gpt-4.1",
		Temperature: &temperature,
		TopP:        &topP,
	})
	if err != nil {
		t.Fatalf("toRequest failed: %v", err)
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}
	if _, ok := raw["temperature"]; ok {
		t.Fatalf("request includes temperature: %s", data)
	}
	if _, ok := raw["top_p"]; ok {
		t.Fatalf("request includes top_p: %s", data)
	}
}

func TestToRequestAlwaysIncludesReasoningDefaults(t *testing.T) {
	for _, model := range []string{"gpt-4.1", "gpt-5", "o1", "o3-mini"} {
		t.Run(model, func(t *testing.T) {
			req, err := toRequest(&types.CompletionRequest{
				Model: model,
			})
			if err != nil {
				t.Fatalf("toRequest failed: %v", err)
			}

			if req.Reasoning == nil {
				t.Fatal("expected reasoning")
			}
			if req.Reasoning.Summary == nil || *req.Reasoning.Summary != "auto" {
				t.Fatalf("expected summary auto, got %#v", req.Reasoning.Summary)
			}
			if req.Reasoning.Effort == nil || *req.Reasoning.Effort != "medium" {
				t.Fatalf("expected effort medium, got %#v", req.Reasoning.Effort)
			}
			if len(req.Include) != 1 || req.Include[0] != "reasoning.encrypted_content" {
				t.Fatalf("expected reasoning include, got %v", req.Include)
			}
		})
	}
}

func TestToRequestUsesExplicitReasoningValues(t *testing.T) {
	req, err := toRequest(&types.CompletionRequest{
		Model: "gpt-5",
		Reasoning: &types.AgentReasoning{
			Summary: "detailed",
			Effort:  "high",
		},
	})
	if err != nil {
		t.Fatalf("toRequest failed: %v", err)
	}

	if req.Reasoning == nil {
		t.Fatal("expected reasoning")
	}
	if req.Reasoning.Summary == nil || *req.Reasoning.Summary != "detailed" {
		t.Fatalf("expected summary detailed, got %#v", req.Reasoning.Summary)
	}
	if req.Reasoning.Effort == nil || *req.Reasoning.Effort != "high" {
		t.Fatalf("expected effort high, got %#v", req.Reasoning.Effort)
	}
	if len(req.Include) != 1 || req.Include[0] != "reasoning.encrypted_content" {
		t.Fatalf("expected reasoning include, got %v", req.Include)
	}
}
