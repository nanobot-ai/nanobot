package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

func TestCompletionItem_Text(t *testing.T) {
	item := CompletionItem{
		ID: "test-id",
		Message: &mcp.SamplingMessage{
			Role: "user",
			Content: mcp.Content{
				Type: "text",
				Text: "Hello, world!",
			},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","role":"user","type":"text","text":"Hello, world!"}`).Equal(t, string(data))

	var unmarshalledItem CompletionItem
	if err := json.Unmarshal(data, &unmarshalledItem); err != nil {
		t.Fatalf("Failed to unmarshal CompletionItem: %v", err)
	}

	if !reflect.DeepEqual(item, unmarshalledItem) {
		t.Errorf("Expected unmarshalled item to be equal to original item, but got: %+v", unmarshalledItem)
	}
}

func TestCompletionItem_Image(t *testing.T) {
	item := CompletionItem{
		ID: "test-id",
		Message: &mcp.SamplingMessage{
			Role: "user",
			Content: mcp.Content{
				Type: "image",
				Data: "base64-image-data",
			},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","role":"user","type":"image","data":"base64-image-data"}`).Equal(t, string(data))

	var unmarshalledItem CompletionItem
	if err := json.Unmarshal(data, &unmarshalledItem); err != nil {
		t.Fatalf("Failed to unmarshal CompletionItem: %v", err)
	}

	if !reflect.DeepEqual(item, unmarshalledItem) {
		t.Errorf("Expected unmarshalled item to be equal to original item, but got: %+v", unmarshalledItem)
	}
}

func TestCompletionItem_Tool(t *testing.T) {
	item := CompletionItem{
		ID: "test-id",
		ToolCall: &ToolCall{
			Arguments:  "test-arguments",
			CallID:     "test-call-id",
			Name:       "test-name",
			ID:         "test-id",
			Target:     "test-target",
			TargetType: "test-target-type",
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","type":"tool","arguments":"test-arguments","callID":"test-call-id","name":"test-name","target":"test-target","targetType":"test-target-type"}`).Equal(t, string(data))

	var unmarshalledItem CompletionItem
	if err := json.Unmarshal(data, &unmarshalledItem); err != nil {
		t.Fatalf("Failed to unmarshal CompletionItem: %v", err)
	}

	if !reflect.DeepEqual(item, unmarshalledItem) {
		t.Errorf("Expected unmarshalled item to be equal to original item, but got: %+v", unmarshalledItem)
	}
}

func TestCompletionItem_ToolResult(t *testing.T) {
	item := CompletionItem{
		ID: "test-id",
		ToolCallResult: &ToolCallResult{
			OutputRole: "assistant",
			CallID:     "test-call-id",
			Output: CallResult{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: "This is a tool result",
					},
				},
			},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","type":"tool","output":{"content":[{"type":"text","text":"This is a tool result"}]},"outputRole":"assistant","callID":"test-call-id"}`).Equal(t, string(data))

	var unmarshalledItem CompletionItem
	if err := json.Unmarshal(data, &unmarshalledItem); err != nil {
		t.Fatalf("Failed to unmarshal CompletionItem: %v", err)
	}

	if !reflect.DeepEqual(item, unmarshalledItem) {
		t.Errorf("Expected unmarshalled item to be equal to original item, but got: %+v", unmarshalledItem)
	}
}

func TestCompletionItem_ToolBoth(t *testing.T) {
	item := CompletionItem{
		ID: "test-id",
		ToolCall: &ToolCall{
			Arguments:  "test-arguments",
			CallID:     "test-call-id",
			Name:       "test-name",
			ID:         "test-id",
			Target:     "test-target",
			TargetType: "test-target-type",
		},
		ToolCallResult: &ToolCallResult{
			OutputRole: "assistant",
			CallID:     "test-call-id",
			Output: CallResult{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: "This is a tool result",
					},
				},
			},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","type":"tool","output":{"content":[{"type":"text","text":"This is a tool result"}]},"outputRole":"assistant","arguments":"test-arguments","callID":"test-call-id","name":"test-name","target":"test-target","targetType":"test-target-type"}`).Equal(t, string(data))

	var unmarshalledItem CompletionItem
	if err := json.Unmarshal(data, &unmarshalledItem); err != nil {
		t.Fatalf("Failed to unmarshal CompletionItem: %v", err)
	}

	if !reflect.DeepEqual(item, unmarshalledItem) {
		t.Errorf("Expected unmarshalled item to be equal to original item, but got: %+v", unmarshalledItem)
	}
}

func TestCompletionItem_Reasoning(t *testing.T) {
	item := CompletionItem{
		ID: "test-id",
		Reasoning: &Reasoning{
			EncryptedContent: "encrypted-content",
			Summary: []SummaryText{
				{Text: "This is a summary of the reasoning."},
			},
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","type":"reasoning","encryptedContent":"encrypted-content","summary":[{"text":"This is a summary of the reasoning."}]}`).Equal(t, string(data))

	var unmarshalledItem CompletionItem
	if err := json.Unmarshal(data, &unmarshalledItem); err != nil {
		t.Fatalf("Failed to unmarshal CompletionItem: %v", err)
	}

	if !reflect.DeepEqual(item, unmarshalledItem) {
		t.Errorf("Expected unmarshalled item to be equal to original item, but got: %+v", unmarshalledItem)
	}
}
