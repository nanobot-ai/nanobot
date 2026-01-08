package agent

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestAppendProgress(t *testing.T) {
	tests := []struct {
		name            string
		initialResponse types.CompletionResponse
		progressPayload progressPayload
		wantResponse    types.CompletionResponse
		wantNil         bool
	}{
		{
			name: "non-progress notification - return unchanged",
			initialResponse: types.CompletionResponse{
				InternalMessages: []types.Message{},
			},
			progressPayload: progressPayload{}, // No progress metadata
			wantResponse: types.CompletionResponse{
				InternalMessages: []types.Message{},
			},
			wantNil: false,
		},
		{
			name: "create new message with text content",
			initialResponse: types.CompletionResponse{
				InternalMessages: []types.Message{},
			},
			progressPayload: progressPayload{
				Meta: progressPayloadMeta{
					Progress: &types.CompletionProgress{
						MessageID: "msg1",
						Role:      "assistant",
						Item: types.CompletionItem{
							ID:      "item1",
							Partial: false,
							Content: &mcp.Content{Type: "text", Text: "Hello"},
						},
					},
				},
			},
			wantResponse: types.CompletionResponse{
				HasMore: true,
				InternalMessages: []types.Message{
					{
						ID:      "msg1",
						Role:    "assistant",
						HasMore: true,
						Items: []types.CompletionItem{
							{
								ID:      "item1",
								Partial: false,
								Content: &mcp.Content{Type: "text", Text: "Hello", ID: "item1"},
							},
						},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "merge partial text content",
			initialResponse: types.CompletionResponse{
				InternalMessages: []types.Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []types.CompletionItem{
							{
								ID:      "item1",
								Partial: true,
								Content: &mcp.Content{Type: "text", Text: "Hello "},
							},
						},
					},
				},
			},
			progressPayload: progressPayload{
				Meta: progressPayloadMeta{
					Progress: &types.CompletionProgress{
						MessageID: "msg1",
						Item: types.CompletionItem{
							ID:      "item1",
							Partial: true,
							Content: &mcp.Content{Type: "text", Text: "World"},
						},
					},
				},
			},
			wantResponse: types.CompletionResponse{
				HasMore: true,
				InternalMessages: []types.Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []types.CompletionItem{
							{
								ID:      "item1",
								Partial: true,
								Content: &mcp.Content{Type: "text", Text: "Hello World"},
							},
						},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "add tool call result to existing tool call",
			initialResponse: types.CompletionResponse{
				InternalMessages: []types.Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []types.CompletionItem{
							{
								ID: "item1",
								ToolCall: &types.ToolCall{
									CallID:    "call_123",
									Name:      "search",
									Arguments: `{"query":"test"}`,
								},
							},
						},
					},
				},
			},
			progressPayload: progressPayload{
				Meta: progressPayloadMeta{
					Progress: &types.CompletionProgress{
						MessageID: "msg1",
						Item: types.CompletionItem{
							ToolCallResult: &types.ToolCallResult{
								CallID: "call_123",
								Output: types.CallResult{
									Content: []mcp.Content{
										{Type: "text", Text: "Result"},
									},
								},
							},
						},
					},
				},
			},
			wantResponse: types.CompletionResponse{
				HasMore: true,
				InternalMessages: []types.Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []types.CompletionItem{
							{
								ID: "item1",
								ToolCall: &types.ToolCall{
									CallID:    "call_123",
									Name:      "search",
									Arguments: `{"query":"test"}`,
								},
								ToolCallResult: &types.ToolCallResult{
									CallID: "call_123",
									Output: types.CallResult{
										Content: []mcp.Content{
											{Type: "text", Text: "Result"},
										},
									},
								},
							},
						},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "streaming tool call arguments",
			initialResponse: types.CompletionResponse{
				InternalMessages: []types.Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []types.CompletionItem{
							{
								ID: "item1",
								ToolCall: &types.ToolCall{
									Arguments: `{"query":`,
									Name:      "search",
								},
							},
						},
					},
				},
			},
			progressPayload: progressPayload{
				Meta: progressPayloadMeta{
					Progress: &types.CompletionProgress{
						MessageID: "msg1",
						Item: types.CompletionItem{
							ID:      "item1",
							Partial: true,
							ToolCall: &types.ToolCall{
								Arguments: `"test"}`,
							},
						},
					},
				},
			},
			wantResponse: types.CompletionResponse{
				HasMore: true,
				InternalMessages: []types.Message{
					{
						ID:   "msg1",
						Role: "assistant",
						Items: []types.CompletionItem{
							{
								ID: "item1",
								ToolCall: &types.ToolCall{
									Arguments: `{"query":"test"}`,
									Name:      "search",
								},
							},
						},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "default role to assistant when empty",
			initialResponse: types.CompletionResponse{
				InternalMessages: []types.Message{},
			},
			progressPayload: progressPayload{
				Meta: progressPayloadMeta{
					Progress: &types.CompletionProgress{
						MessageID: "msg1",
						Role:      "", // Empty role
						Item: types.CompletionItem{
							ID:      "item1",
							Content: &mcp.Content{Type: "text", Text: "Hello"},
						},
					},
				},
			},
			wantResponse: types.CompletionResponse{
				HasMore: true,
				InternalMessages: []types.Message{
					{
						ID:      "msg1",
						Role:    "assistant", // Should default to assistant
						HasMore: true,
						Items: []types.CompletionItem{
							{
								ID:      "item1",
								Content: &mcp.Content{Type: "text", Text: "Hello", ID: "item1"},
							},
						},
					},
				},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create session and context
			session := mcp.TestSession(context.Background())
			ctx := mcp.WithSession(context.Background(), session)

			// Set initial response in session
			session.Set(progressSessionKey, &tt.initialResponse)

			// Create MCP message with progress notification
			params, _ := json.Marshal(tt.progressPayload)
			msg := &mcp.Message{
				Method: "notifications/progress",
				Params: params,
			}

			// Call appendProgress
			result, err := appendProgress(ctx, session, msg)

			if err != nil {
				t.Fatalf("appendProgress returned error: %v", err)
			}

			if tt.wantNil && result != nil {
				t.Errorf("Expected nil result, got %+v", result)
			}

			// Get updated response from session
			var gotResponse types.CompletionResponse
			session.Get(progressSessionKey, &gotResponse)

			// Compare responses (ignoring Created timestamps and HasMore on messages)
			if gotResponse.HasMore != tt.wantResponse.HasMore {
				t.Errorf("HasMore = %v, want %v", gotResponse.HasMore, tt.wantResponse.HasMore)
			}

			if len(gotResponse.InternalMessages) != len(tt.wantResponse.InternalMessages) {
				t.Fatalf("InternalMessages length = %d, want %d",
					len(gotResponse.InternalMessages), len(tt.wantResponse.InternalMessages))
			}

			for i := range gotResponse.InternalMessages {
				got := gotResponse.InternalMessages[i]
				want := tt.wantResponse.InternalMessages[i]

				if got.ID != want.ID {
					t.Errorf("Message[%d].ID = %v, want %v", i, got.ID, want.ID)
				}
				if got.Role != want.Role {
					t.Errorf("Message[%d].Role = %v, want %v", i, got.Role, want.Role)
				}
				if len(got.Items) != len(want.Items) {
					t.Errorf("Message[%d].Items length = %d, want %d", i, len(got.Items), len(want.Items))
					continue
				}

				for j := range got.Items {
					gotItem := got.Items[j]
					wantItem := want.Items[j]

					if gotItem.ID != wantItem.ID {
						t.Errorf("Message[%d].Items[%d].ID = %v, want %v", i, j, gotItem.ID, wantItem.ID)
					}

					// Compare content
					if !reflect.DeepEqual(gotItem.Content, wantItem.Content) {
						t.Errorf("Message[%d].Items[%d].Content = %+v, want %+v",
							i, j, gotItem.Content, wantItem.Content)
					}

					// Compare tool calls
					if !reflect.DeepEqual(gotItem.ToolCall, wantItem.ToolCall) {
						t.Errorf("Message[%d].Items[%d].ToolCall = %+v, want %+v",
							i, j, gotItem.ToolCall, wantItem.ToolCall)
					}

					// Compare tool call results
					if !reflect.DeepEqual(gotItem.ToolCallResult, wantItem.ToolCallResult) {
						t.Errorf("Message[%d].Items[%d].ToolCallResult = %+v, want %+v",
							i, j, gotItem.ToolCallResult, wantItem.ToolCallResult)
					}
				}
			}
		})
	}
}

func TestAppendProgress_NonProgressNotification(t *testing.T) {
	session := mcp.TestSession(context.Background())
	ctx := mcp.WithSession(context.Background(), session)

	session.Set(progressSessionKey, &types.CompletionResponse{})

	msg := &mcp.Message{
		Method: "some/other/method",
	}

	result, err := appendProgress(ctx, session, msg)
	if err != nil {
		t.Fatalf("appendProgress returned error: %v", err)
	}

	if result != msg {
		t.Errorf("Expected original message to be returned for non-progress notification")
	}
}
