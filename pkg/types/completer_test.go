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
		Content: &mcp.Content{
			Type: "text",
			Text: "Hello, world!",
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","partial":false,"type":"text","text":"Hello, world!"}`).Equal(t, string(data))

	var unmarshalledItem CompletionItem
	if err := json.Unmarshal(data, &unmarshalledItem); err != nil {
		t.Fatalf("Failed to unmarshal CompletionItem: %v", err)
	}

	// Compare relevant fields individually since unmarshal may populate zero values
	if item.Content.Type != unmarshalledItem.Content.Type {
		t.Errorf("Expected content type %q but got %q", item.Content.Type, unmarshalledItem.Content.Type)
	}
	if item.Content.Text != unmarshalledItem.Content.Text {
		t.Errorf("Expected content text %q but got %q", item.Content.Text, unmarshalledItem.Content.Text)
	}
	if item.ID != unmarshalledItem.ID || item.Partial != unmarshalledItem.Partial {
		t.Errorf("Expected unmarshalled item fields to match original, but got: %+v", unmarshalledItem)
	}
}

func TestCompletionItem_Image(t *testing.T) {
	item := CompletionItem{
		ID: "test-id",
		Content: &mcp.Content{
			Type: "image",
			Data: "base64-image-data",
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","partial":false,"type":"image","data":"base64-image-data"}`).Equal(t, string(data))

	var unmarshalledItem CompletionItem
	if err := json.Unmarshal(data, &unmarshalledItem); err != nil {
		t.Fatalf("Failed to unmarshal CompletionItem: %v", err)
	}

	// Compare relevant fields individually since unmarshal may populate zero values
	if item.Content.Type != unmarshalledItem.Content.Type {
		t.Errorf("Expected content type %q but got %q", item.Content.Type, unmarshalledItem.Content.Type)
	}
	if item.Content.Data != unmarshalledItem.Content.Data {
		t.Errorf("Expected content data %q but got %q", item.Content.Data, unmarshalledItem.Content.Data)
	}
	if item.ID != unmarshalledItem.ID || item.Partial != unmarshalledItem.Partial {
		t.Errorf("Expected unmarshalled item fields to match original, but got: %+v", unmarshalledItem)
	}
}

func TestCompletionItem_Tool(t *testing.T) {
	item := CompletionItem{
		ID: "test-id",
		ToolCall: &ToolCall{
			Arguments: "test-arguments",
			CallID:    "test-call-id",
			Name:      "test-name",
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionItem: %v", err)
	}

	autogold.Expect(`{"id":"test-id","type":"tool","arguments":"test-arguments","callID":"test-call-id","name":"test-name"}`).Equal(t, string(data))

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
			CallID: "test-call-id",
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

	autogold.Expect(`{"id":"test-id","type":"tool","output":{"content":[{"type":"text","text":"This is a tool result"}]},"callID":"test-call-id"}`).Equal(t, string(data))

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
			Arguments: "test-arguments",
			CallID:    "test-call-id",
			Name:      "test-name",
		},
		ToolCallResult: &ToolCallResult{
			CallID: "test-call-id",
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

	autogold.Expect(`{"id":"test-id","type":"tool","output":{"content":[{"type":"text","text":"This is a tool result"}]},"arguments":"test-arguments","callID":"test-call-id","name":"test-name"}`).Equal(t, string(data))

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

func TestMergeContent(t *testing.T) {
	tests := []struct {
		name     string
		existing *mcp.Content
		new      *mcp.Content
		want     *mcp.Content
	}{
		{
			name:     "merge two text contents",
			existing: &mcp.Content{Type: "text", Text: "Hello "},
			new:      &mcp.Content{Type: "text", Text: "World"},
			want:     &mcp.Content{Type: "text", Text: "Hello World"},
		},
		{
			name:     "existing is nil",
			existing: nil,
			new:      &mcp.Content{Type: "text", Text: "Hello"},
			want:     &mcp.Content{Type: "text", Text: "Hello"},
		},
		{
			name:     "different types - use new",
			existing: &mcp.Content{Type: "text", Text: "Hello"},
			new:      &mcp.Content{Type: "image", Data: "base64"},
			want:     &mcp.Content{Type: "image", Data: "base64"},
		},
		{
			name:     "existing is image, new is text",
			existing: &mcp.Content{Type: "image", Data: "base64"},
			new:      &mcp.Content{Type: "text", Text: "Hello"},
			want:     &mcp.Content{Type: "text", Text: "Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeContent(tt.existing, tt.new)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeContent() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMergeToolCall(t *testing.T) {
	tests := []struct {
		name     string
		existing *ToolCall
		new      *ToolCall
		want     *ToolCall
	}{
		{
			name: "merge arguments",
			existing: &ToolCall{
				Arguments: `{"arg1":`,
				Name:      "test_tool",
				CallID:    "call_123",
			},
			new: &ToolCall{
				Arguments: `"value"}`,
			},
			want: &ToolCall{
				Arguments: `{"arg1":"value"}`,
				Name:      "test_tool",
				CallID:    "call_123",
			},
		},
		{
			name: "update name and callID",
			existing: &ToolCall{
				Arguments: `{"arg1":`,
			},
			new: &ToolCall{
				Arguments: `"value"}`,
				Name:      "test_tool",
				CallID:    "call_456",
			},
			want: &ToolCall{
				Arguments: `{"arg1":"value"}`,
				Name:      "test_tool",
				CallID:    "call_456",
			},
		},
		{
			name:     "existing is nil",
			existing: nil,
			new: &ToolCall{
				Arguments: `{"arg":"value"}`,
				Name:      "test_tool",
				CallID:    "call_789",
			},
			want: &ToolCall{
				Arguments: `{"arg":"value"}`,
				Name:      "test_tool",
				CallID:    "call_789",
			},
		},
		{
			name: "preserve existing name and callID if not provided",
			existing: &ToolCall{
				Arguments: `{"old":`,
				Name:      "existing_tool",
				CallID:    "call_existing",
			},
			new: &ToolCall{
				Arguments: `"data"}`,
			},
			want: &ToolCall{
				Arguments: `{"old":"data"}`,
				Name:      "existing_tool",
				CallID:    "call_existing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeToolCall(tt.existing, tt.new)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeToolCall() = %+v, want %+v", got, tt.want)
			}

			// Verify no side effects on input
			if tt.existing != nil {
				// Check that existing wasn't modified
				if tt.name == "merge arguments" {
					if tt.existing.Arguments != `{"arg1":` {
						t.Errorf("existing was modified: %+v", tt.existing)
					}
				}
			}
		})
	}
}

func TestMergeReasoning(t *testing.T) {
	tests := []struct {
		name     string
		existing *Reasoning
		new      *Reasoning
		want     *Reasoning
	}{
		{
			name: "merge encrypted content and summaries",
			existing: &Reasoning{
				EncryptedContent: "part1",
				Summary: []SummaryText{
					{Text: "Summary 1"},
				},
			},
			new: &Reasoning{
				EncryptedContent: "part2",
				Summary: []SummaryText{
					{Text: "Summary 2"},
				},
			},
			want: &Reasoning{
				EncryptedContent: "part1part2",
				Summary: []SummaryText{
					{Text: "Summary 1"},
					{Text: "Summary 2"},
				},
			},
		},
		{
			name:     "existing is nil",
			existing: nil,
			new: &Reasoning{
				EncryptedContent: "content",
				Summary: []SummaryText{
					{Text: "Summary"},
				},
			},
			want: &Reasoning{
				EncryptedContent: "content",
				Summary: []SummaryText{
					{Text: "Summary"},
				},
			},
		},
		{
			name: "empty summaries",
			existing: &Reasoning{
				EncryptedContent: "part1",
			},
			new: &Reasoning{
				EncryptedContent: "part2",
			},
			want: &Reasoning{
				EncryptedContent: "part1part2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeReasoning(tt.existing, tt.new)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeReasoning() = %+v, want %+v", got, tt.want)
			}

			// Verify no side effects on input
			if tt.existing != nil && tt.name == "merge encrypted content and summaries" {
				if len(tt.existing.Summary) != 1 {
					t.Errorf("existing summary was modified: %+v", tt.existing.Summary)
				}
			}
		})
	}
}

func TestMergeCompletionItems(t *testing.T) {
	tests := []struct {
		name     string
		existing CompletionItem
		new      CompletionItem
		want     CompletionItem
	}{
		{
			name: "new item is not partial - replace",
			existing: CompletionItem{
				ID:      "item1",
				Partial: false,
				Content: &mcp.Content{Type: "text", Text: "Old"},
			},
			new: CompletionItem{
				ID:      "item2",
				Partial: false,
				Content: &mcp.Content{Type: "text", Text: "New"},
			},
			want: CompletionItem{
				ID:      "item2",
				Partial: false,
				Content: &mcp.Content{Type: "text", Text: "New"},
			},
		},
		{
			name: "merge partial text content",
			existing: CompletionItem{
				ID:      "item1",
				Partial: true,
				Content: &mcp.Content{Type: "text", Text: "Hello "},
			},
			new: CompletionItem{
				ID:      "item1",
				Partial: true,
				Content: &mcp.Content{Type: "text", Text: "World"},
			},
			want: CompletionItem{
				ID:      "item1",
				Partial: true,
				Content: &mcp.Content{Type: "text", Text: "Hello World"},
			},
		},
		{
			name: "merge partial tool call",
			existing: CompletionItem{
				ID:      "item1",
				Partial: true,
				ToolCall: &ToolCall{
					Arguments: `{"key":`,
					Name:      "tool",
					CallID:    "call_1",
				},
			},
			new: CompletionItem{
				ID:      "item1",
				Partial: true,
				ToolCall: &ToolCall{
					Arguments: `"value"}`,
				},
			},
			want: CompletionItem{
				ID:      "item1",
				Partial: true,
				ToolCall: &ToolCall{
					Arguments: `{"key":"value"}`,
					Name:      "tool",
					CallID:    "call_1",
				},
			},
		},
		{
			name: "merge partial reasoning",
			existing: CompletionItem{
				ID:      "item1",
				Partial: true,
				Reasoning: &Reasoning{
					EncryptedContent: "part1",
					Summary:          []SummaryText{{Text: "Sum1"}},
				},
			},
			new: CompletionItem{
				ID:      "item1",
				Partial: true,
				Reasoning: &Reasoning{
					EncryptedContent: "part2",
					Summary:          []SummaryText{{Text: "Sum2"}},
				},
			},
			want: CompletionItem{
				ID:      "item1",
				Partial: true,
				Reasoning: &Reasoning{
					EncryptedContent: "part1part2",
					Summary:          []SummaryText{{Text: "Sum1"}, {Text: "Sum2"}},
				},
			},
		},
		{
			name: "partial content becomes non-partial",
			existing: CompletionItem{
				ID:      "item1",
				Partial: true,
				Content: &mcp.Content{Type: "text", Text: "Hello "},
			},
			new: CompletionItem{
				ID:      "item1",
				Partial: false,
				Content: &mcp.Content{Type: "text", Text: "World"},
			},
			want: CompletionItem{
				ID:      "item1",
				Partial: false,
				Content: &mcp.Content{Type: "text", Text: "World"},
			},
		},
		{
			name: "partial with no content fields - keep existing",
			existing: CompletionItem{
				ID:      "item1",
				Partial: false,
				Content: &mcp.Content{Type: "text", Text: "Existing content"},
			},
			new: CompletionItem{
				ID:      "item1",
				Partial: true,
				// No content fields set
			},
			want: CompletionItem{
				ID:      "item1",
				Partial: false,
				Content: &mcp.Content{Type: "text", Text: "Existing content"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeCompletionItems(tt.existing, tt.new)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeCompletionItems() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAppendProgress(t *testing.T) {
	tests := []struct {
		name     string
		resp     CompletionResponse
		progress CompletionProgress
		want     CompletionResponse
	}{
		{
			name: "update agent and model",
			resp: CompletionResponse{
				Agent: "old_agent",
				Model: "old_model",
			},
			progress: CompletionProgress{
				Agent: "new_agent",
				Model: "new_model",
			},
			want: CompletionResponse{
				Agent: "new_agent",
				Model: "new_model",
			},
		},
		{
			name: "no message ID - return unchanged",
			resp: CompletionResponse{
				InternalMessages: []Message{},
			},
			progress: CompletionProgress{
				Item: CompletionItem{
					Content: &mcp.Content{Type: "text", Text: "Test"},
				},
			},
			want: CompletionResponse{
				InternalMessages: []Message{},
			},
		},
		{
			name: "create new message",
			resp: CompletionResponse{
				InternalMessages: []Message{},
			},
			progress: CompletionProgress{
				MessageID: "msg1",
				Role:      "assistant",
				Item: CompletionItem{
					ID:      "item1",
					Content: &mcp.Content{Type: "text", Text: "Hello"},
				},
			},
			want: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Content: &mcp.Content{Type: "text", Text: "Hello"},
							},
						},
					},
				},
			},
		},
		{
			name: "append item to existing message",
			resp: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Content: &mcp.Content{Type: "text", Text: "Hello"},
							},
						},
					},
				},
			},
			progress: CompletionProgress{
				MessageID: "msg1",
				Item: CompletionItem{
					ID:      "item2",
					Content: &mcp.Content{Type: "text", Text: "World"},
				},
			},
			want: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Content: &mcp.Content{Type: "text", Text: "Hello"},
							},
							{
								ID:      "item2",
								Content: &mcp.Content{Type: "text", Text: "World"},
							},
						},
					},
				},
			},
		},
		{
			name: "merge partial item with matching ID",
			resp: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Partial: true,
								Content: &mcp.Content{Type: "text", Text: "Hello "},
							},
						},
					},
				},
			},
			progress: CompletionProgress{
				MessageID: "msg1",
				Item: CompletionItem{
					ID:      "item1",
					Partial: true,
					Content: &mcp.Content{Type: "text", Text: "World"},
				},
			},
			want: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Partial: true,
								Content: &mcp.Content{Type: "text", Text: "Hello World"},
							},
						},
					},
				},
			},
		},
		{
			name: "update role in existing message",
			resp: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "user",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Content: &mcp.Content{Type: "text", Text: "Hello"},
							},
						},
					},
				},
			},
			progress: CompletionProgress{
				MessageID: "msg1",
				Role:      "assistant",
				Item: CompletionItem{
					ID:      "item2",
					Content: &mcp.Content{Type: "text", Text: "World"},
				},
			},
			want: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Content: &mcp.Content{Type: "text", Text: "Hello"},
							},
							{
								ID:      "item2",
								Content: &mcp.Content{Type: "text", Text: "World"},
							},
						},
					},
				},
			},
		},
		{
			name: "streaming tool call - merge arguments",
			resp: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Partial: true,
								ToolCall: &ToolCall{
									Arguments: `{"query":`,
									Name:      "search",
								},
							},
						},
					},
				},
			},
			progress: CompletionProgress{
				MessageID: "msg1",
				Item: CompletionItem{
					ID:      "item1",
					Partial: true,
					ToolCall: &ToolCall{
						Arguments: `"test"}`,
						CallID:    "call_123",
					},
				},
			},
			want: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []CompletionItem{
							{
								ID:      "item1",
								Partial: true,
								ToolCall: &ToolCall{
									Arguments: `{"query":"test"}`,
									Name:      "search",
									CallID:    "call_123",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple messages - append to correct one",
			resp: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:    "msg1",
						Role:  "user",
						Items: []CompletionItem{{ID: "item1"}},
					},
					{
						ID:    "msg2",
						Role:  "assistant",
						Items: []CompletionItem{{ID: "item2"}},
					},
				},
			},
			progress: CompletionProgress{
				MessageID: "msg2",
				Item: CompletionItem{
					ID:      "item3",
					Content: &mcp.Content{Type: "text", Text: "New item"},
				},
			},
			want: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:    "msg1",
						Role:  "user",
						Items: []CompletionItem{{ID: "item1"}},
					},
					{
						ID:   "msg2",
						Role: "assistant",
						Items: []CompletionItem{
							{ID: "item2"},
							{
								ID:      "item3",
								Content: &mcp.Content{Type: "text", Text: "New item"},
							},
						},
					},
				},
			},
		},
		{
			name: "create new message with empty role - defaults to assistant",
			resp: CompletionResponse{
				InternalMessages: []Message{},
			},
			progress: CompletionProgress{
				MessageID: "msg1",
				Role:      "", // Empty role should default to "assistant"
				Item: CompletionItem{
					ID:      "item1",
					Content: &mcp.Content{Type: "text", Text: "Hello"},
				},
			},
			want: CompletionResponse{
				InternalMessages: []Message{
					{
						ID:   "msg1",
						Role: "assistant", // Should be defaulted
						Items: []CompletionItem{
							{
								ID:      "item1",
								Content: &mcp.Content{Type: "text", Text: "Hello"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendProgress(tt.resp, tt.progress)

			// Compare non-timestamp fields
			if got.Agent != tt.want.Agent {
				t.Errorf("Agent = %v, want %v", got.Agent, tt.want.Agent)
			}
			if got.Model != tt.want.Model {
				t.Errorf("Model = %v, want %v", got.Model, tt.want.Model)
			}
			if len(got.InternalMessages) != len(tt.want.InternalMessages) {
				t.Fatalf("InternalMessages length = %d, want %d", len(got.InternalMessages), len(tt.want.InternalMessages))
			}

			// Compare messages individually, ignoring Created timestamps
			for i := range got.InternalMessages {
				gotMsg := got.InternalMessages[i]
				wantMsg := tt.want.InternalMessages[i]

				if gotMsg.ID != wantMsg.ID {
					t.Errorf("Message[%d].ID = %v, want %v", i, gotMsg.ID, wantMsg.ID)
				}
				if gotMsg.Role != wantMsg.Role {
					t.Errorf("Message[%d].Role = %v, want %v", i, gotMsg.Role, wantMsg.Role)
				}
				if !reflect.DeepEqual(gotMsg.Items, wantMsg.Items) {
					t.Errorf("Message[%d].Items = %+v, want %+v", i, gotMsg.Items, wantMsg.Items)
				}
			}
		})
	}
}

func TestToolCallClone(t *testing.T) {
	t.Run("clone non-nil", func(t *testing.T) {
		original := &ToolCall{
			Arguments: "args",
			CallID:    "call_id",
			Name:      "name",
		}
		clone := original.Clone()

		if !reflect.DeepEqual(original, clone) {
			t.Errorf("Clone not equal to original")
		}

		// Verify it's a copy
		clone.Arguments = "modified"
		if original.Arguments == "modified" {
			t.Errorf("Modifying clone affected original")
		}
	})

	t.Run("clone nil", func(t *testing.T) {
		var original *ToolCall
		clone := original.Clone()
		if clone != nil {
			t.Errorf("Clone of nil should be nil, got %+v", clone)
		}
	})
}

func TestReasoningClone(t *testing.T) {
	t.Run("clone non-nil", func(t *testing.T) {
		original := &Reasoning{
			EncryptedContent: "encrypted",
			Summary: []SummaryText{
				{Text: "Summary 1"},
				{Text: "Summary 2"},
			},
		}
		clone := original.Clone()

		if !reflect.DeepEqual(original, clone) {
			t.Errorf("Clone not equal to original")
		}

		// Verify it's a copy
		clone.Summary[0].Text = "Modified"
		if original.Summary[0].Text == "Modified" {
			t.Errorf("Modifying clone affected original")
		}
	})

	t.Run("clone nil", func(t *testing.T) {
		var original *Reasoning
		clone := original.Clone()
		if clone != nil {
			t.Errorf("Clone of nil should be nil, got %+v", clone)
		}
	})
}

func TestAppendProgress_NoSideEffects(t *testing.T) {
	t.Run("does not modify input response", func(t *testing.T) {
		originalResp := CompletionResponse{
			Agent: "agent1",
			Model: "model1",
			InternalMessages: []Message{
				{
					ID:   "msg1",
					Role: "assistant",
					Items: []CompletionItem{
						{
							ID:      "item1",
							Partial: true,
							Content: &mcp.Content{Type: "text", Text: "Hello "},
						},
					},
				},
			},
		}

		// Create a deep copy to compare against
		expectedResp := CompletionResponse{
			Agent: "agent1",
			Model: "model1",
			InternalMessages: []Message{
				{
					ID:   "msg1",
					Role: "assistant",
					Items: []CompletionItem{
						{
							ID:      "item1",
							Partial: true,
							Content: &mcp.Content{Type: "text", Text: "Hello "},
						},
					},
				},
			},
		}

		progress := CompletionProgress{
			MessageID: "msg1",
			Item: CompletionItem{
				ID:      "item1",
				Partial: true,
				Content: &mcp.Content{Type: "text", Text: "World"},
			},
		}

		_ = AppendProgress(originalResp, progress)

		// Verify original response was not modified
		if originalResp.Agent != expectedResp.Agent {
			t.Errorf("Original response Agent was modified")
		}
		if originalResp.Model != expectedResp.Model {
			t.Errorf("Original response Model was modified")
		}
		if len(originalResp.InternalMessages) != len(expectedResp.InternalMessages) {
			t.Errorf("Original response InternalMessages length was modified")
		}
		if originalResp.InternalMessages[0].Items[0].Content.Text != "Hello " {
			t.Errorf("Original response item content was modified: got %q, want %q",
				originalResp.InternalMessages[0].Items[0].Content.Text, "Hello ")
		}
	})

	t.Run("does not modify input progress", func(t *testing.T) {
		resp := CompletionResponse{
			InternalMessages: []Message{
				{
					ID:   "msg1",
					Role: "assistant",
					Items: []CompletionItem{
						{
							ID: "item1",
							ToolCall: &ToolCall{
								Arguments: `{"key":`,
								Name:      "tool",
							},
						},
					},
				},
			},
		}

		originalProgress := CompletionProgress{
			Agent:     "agent1",
			Model:     "model1",
			MessageID: "msg1",
			Role:      "assistant",
			Item: CompletionItem{
				ID: "item1",
				ToolCall: &ToolCall{
					Arguments: `"value"}`,
					CallID:    "call_123",
				},
			},
		}

		// Store original values
		expectedAgent := originalProgress.Agent
		expectedModel := originalProgress.Model
		expectedArgs := originalProgress.Item.ToolCall.Arguments

		_ = AppendProgress(resp, originalProgress)

		// Verify progress was not modified
		if originalProgress.Agent != expectedAgent {
			t.Errorf("Progress Agent was modified")
		}
		if originalProgress.Model != expectedModel {
			t.Errorf("Progress Model was modified")
		}
		if originalProgress.Item.ToolCall.Arguments != expectedArgs {
			t.Errorf("Progress ToolCall Arguments was modified")
		}
	})

	t.Run("merging does not modify existing item in response", func(t *testing.T) {
		originalItem := CompletionItem{
			ID:      "item1",
			Partial: true,
			Content: &mcp.Content{Type: "text", Text: "Original "},
		}

		resp := CompletionResponse{
			InternalMessages: []Message{
				{
					ID:    "msg1",
					Role:  "assistant",
					Items: []CompletionItem{originalItem},
				},
			},
		}

		progress := CompletionProgress{
			MessageID: "msg1",
			Item: CompletionItem{
				ID:      "item1",
				Partial: true,
				Content: &mcp.Content{Type: "text", Text: "More"},
			},
		}

		result := AppendProgress(resp, progress)

		// Verify the original item in the input response was not modified
		if resp.InternalMessages[0].Items[0].Content.Text != "Original " {
			t.Errorf("Original item was modified: got %q, want %q",
				resp.InternalMessages[0].Items[0].Content.Text, "Original ")
		}

		// Verify the result has merged content
		if result.InternalMessages[0].Items[0].Content.Text != "Original More" {
			t.Errorf("Result item not merged correctly: got %q, want %q",
				result.InternalMessages[0].Items[0].Content.Text, "Original More")
		}
	})

	t.Run("merging reasoning does not modify original summaries", func(t *testing.T) {
		originalSummary := []SummaryText{{Text: "Summary 1"}}
		originalItem := CompletionItem{
			ID:      "item1",
			Partial: true,
			Reasoning: &Reasoning{
				EncryptedContent: "part1",
				Summary:          originalSummary,
			},
		}

		resp := CompletionResponse{
			InternalMessages: []Message{
				{
					ID:    "msg1",
					Role:  "assistant",
					Items: []CompletionItem{originalItem},
				},
			},
		}

		progress := CompletionProgress{
			MessageID: "msg1",
			Item: CompletionItem{
				ID:      "item1",
				Partial: true,
				Reasoning: &Reasoning{
					EncryptedContent: "part2",
					Summary:          []SummaryText{{Text: "Summary 2"}},
				},
			},
		}

		result := AppendProgress(resp, progress)

		// Verify original summary slice was not modified
		if len(originalSummary) != 1 {
			t.Errorf("Original summary slice was modified: length = %d, want 1", len(originalSummary))
		}
		if originalSummary[0].Text != "Summary 1" {
			t.Errorf("Original summary text was modified: got %q, want %q",
				originalSummary[0].Text, "Summary 1")
		}

		// Verify the original item's reasoning in the input response was not modified
		if len(resp.InternalMessages[0].Items[0].Reasoning.Summary) != 1 {
			t.Errorf("Original item reasoning summary was modified")
		}

		// Verify result has merged summaries
		if len(result.InternalMessages[0].Items[0].Reasoning.Summary) != 2 {
			t.Errorf("Result reasoning summary not merged correctly: length = %d, want 2",
				len(result.InternalMessages[0].Items[0].Reasoning.Summary))
		}
	})
}
