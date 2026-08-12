package filter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV1RequestFixtures(t *testing.T) {
	paths, err := filepath.Glob("testdata/v1/*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if strings.Contains(filepath.Base(path), "response-") {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var request Request
			if err := json.Unmarshal(data, &request); err != nil {
				t.Fatal(err)
			}
			if err := request.Validate(); err != nil {
				t.Fatal(err)
			}
			assertJSONEquivalent(t, data, request)
		})
	}
}

func TestV1ToolRequestWrapsEnvelope(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata/v1", "mcp-request.json"))
	if err != nil {
		t.Fatal(err)
	}
	var request Request
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}

	wrapped, err := json.Marshal(ToolRequest{Request: request})
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(wrapped, &value); err != nil {
		t.Fatal(err)
	}
	if len(value) != 1 {
		t.Fatalf("tool input fields = %v, want only request", value)
	}
	assertJSONEquivalent(t, data, json.RawMessage(value["request"]))
}

func TestV1ResponseFixtures(t *testing.T) {
	tests := []struct {
		name      string
		canMutate bool
	}{
		{name: "response-accept.json"},
		{name: "response-reject.json"},
		{name: "response-mutate.json", canMutate: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata/v1", tt.name))
			if err != nil {
				t.Fatal(err)
			}
			var response Response
			if err := json.Unmarshal(data, &response); err != nil {
				t.Fatal(err)
			}
			if err := response.Validate(Capabilities{CanReject: true, CanMutate: tt.canMutate}); err != nil {
				t.Fatal(err)
			}
			assertJSONEquivalent(t, data, response)
		})
	}
}

func TestRequestValidationRejectsBadDiscriminators(t *testing.T) {
	base := Request{
		APIVersion: APIVersionV1,
		Source:     SourceLocalAgent,
		Event: Event{
			Type:    EventTypeUserPrompt,
			Phase:   PhaseRequest,
			Surface: SurfaceUserPrompt,
		},
		Payload: json.RawMessage(`"hello"`),
	}
	tests := map[string]func(*Request){
		"version": func(r *Request) { r.APIVersion = "v2" },
		"source":  func(r *Request) { r.Source = "unknown" },
		"type":    func(r *Request) { r.Event.Type = EventTypeToolCall },
		"phase":   func(r *Request) { r.Event.Phase = PhaseResponse },
		"surface": func(r *Request) { r.Event.Surface = "unknown" },
		"payload": func(r *Request) { r.Payload = nil },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			request := base
			mutate(&request)
			if err := request.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestResponseValidation(t *testing.T) {
	tests := []Response{
		{},
		{Decision: "unknown"},
		{Decision: DecisionAccept, Payload: json.RawMessage(`{}`)},
		{Decision: DecisionReject, Payload: json.RawMessage(`null`)},
		{Decision: DecisionMutate},
		{Decision: DecisionMutate, Payload: json.RawMessage(`not-json`)},
	}
	for _, response := range tests {
		if err := response.Validate(Capabilities{CanReject: true, CanMutate: true}); err == nil {
			t.Fatalf("expected validation error for %#v", response)
		}
	}

	response := Response{Decision: DecisionMutate, Payload: json.RawMessage(`{}`)}
	if err := response.Validate(Capabilities{CanReject: true}); err == nil {
		t.Fatal("expected forbidden mutation to fail validation")
	}
}

func assertJSONEquivalent(t *testing.T, expected []byte, actual any) {
	t.Helper()
	actualData, err := json.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}
	var expectedValue, actualValue any
	if err := json.Unmarshal(expected, &expectedValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(actualData, &actualValue); err != nil {
		t.Fatal(err)
	}
	expectedData, _ := json.Marshal(expectedValue)
	actualData, _ = json.Marshal(actualValue)
	if string(expectedData) != string(actualData) {
		t.Fatalf("JSON mismatch\nexpected: %s\nactual:   %s", expectedData, actualData)
	}
}
