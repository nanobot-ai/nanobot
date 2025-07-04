package completions

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type Client struct {
	Config
}

type Config struct {
	APIKey  string
	BaseURL string
	Headers map[string]string
}

// NewClient creates a new OpenAI-compatible client with the provided API key and base URL.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Headers == nil {
		cfg.Headers = map[string]string{}
	}
	if _, ok := cfg.Headers["Authorization"]; !ok && cfg.APIKey != "" {
		cfg.Headers["Authorization"] = "Bearer " + cfg.APIKey
	}
	if _, ok := cfg.Headers["Content-Type"]; !ok {
		cfg.Headers["Content-Type"] = "application/json"
	}

	return &Client{
		Config: cfg,
	}
}

func (c *Client) Complete(ctx context.Context, completionRequest types.CompletionRequest, opts ...types.CompletionOptions) (*types.CompletionResponse, error) {
	req, err := toRequest(&completionRequest)
	if err != nil {
		return nil, err
	}

	resp, err := c.complete(ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return toResponse(resp)
}

func send(ctx context.Context, progress *types.CompletionProgress, progressToken any) {
	if progressToken == "" || progressToken == nil {
		return
	}
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return
	}

	_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
		ProgressToken: progressToken,
		Meta: map[string]any{
			types.CompletionProgressMetaKey: progress,
		},
	})
}

func (c *Client) complete(ctx context.Context, req *ChatCompletionRequest, opts ...types.CompletionOptions) (*ChatCompletionResponse, error) {
	var (
		opt = complete.Complete(opts...)
	)

	req.Stream = true

	data, _ := json.Marshal(req)
	log.Messages(ctx, "openai-api", true, data)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	for key, value := range c.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("failed to get response: %s %q", httpResp.Status, string(body))
	}

	response, ok, err := progressResponse(ctx, httpResp, opt.ProgressToken)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("failed to get response from stream")
	}

	// Check for errors in the response
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return response, nil
}

func progressResponse(ctx context.Context, resp *http.Response, progressToken any) (*ChatCompletionResponse, bool, error) {
	scanner := bufio.NewScanner(resp.Body)
	var (
		response           *ChatCompletionResponse
		activeToolCalls    = make(map[string]string) // call_id -> partial_args
		completedToolCalls = make(map[string]bool)   // call_id -> completed
		toolCallOrder      []string                  // track order of tool calls
	)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var streamResp ChatCompletionStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue // Skip malformed lines
		}

		// Initialize response on first chunk
		if response == nil {
			response = &ChatCompletionResponse{
				ID:                streamResp.ID,
				Object:            "chat.completion",
				Created:           streamResp.Created,
				Model:             streamResp.Model,
				SystemFingerprint: streamResp.SystemFingerprint,
				Choices:           make([]Choice, len(streamResp.Choices)),
			}
			for i := range response.Choices {
				response.Choices[i] = Choice{
					Index:   i,
					Message: ChatMessage{Role: "assistant"},
				}
			}
		}

		// Process deltas and accumulate content
		for i, choice := range streamResp.Choices {
			if i >= len(response.Choices) {
				continue
			}

			if choice.Delta != nil {
				// Handle text content - send deltas immediately like Anthropic
				if choice.Delta.Content != nil {
					if deltaContent, ok := choice.Delta.Content.(string); ok && deltaContent != "" {
						// Accumulate content
						currentContent := ""
						if response.Choices[i].Message.Content != nil {
							if current, ok := response.Choices[i].Message.Content.(string); ok {
								currentContent = current
							}
						}
						response.Choices[i].Message.Content = currentContent + deltaContent

						// Send text delta immediately
						if progressToken != nil {
							progress := &types.CompletionProgress{
								Model:   response.Model,
								Partial: true,
								HasMore: true,
								Item: types.CompletionItem{
									ID: fmt.Sprintf("%s-text-%d", response.ID, i),
									Message: &mcp.SamplingMessage{
										Role: "assistant",
										Content: mcp.Content{
											Type: "text",
											Text: deltaContent,
										},
									},
								},
							}
							send(ctx, progress, progressToken)
						}
					}
				}

				// Handle tool calls - accumulate and send on completion
				if choice.Delta.ToolCalls != nil {
					// Initialize tool calls slice if needed
					if response.Choices[i].Message.ToolCalls == nil {
						response.Choices[i].Message.ToolCalls = []ToolCall{}
					}

					for _, deltaToolCall := range choice.Delta.ToolCalls {
						// Find existing tool call by index or create new one
						toolCallIndex := deltaToolCall.Index
						if toolCallIndex == nil {
							idx := len(response.Choices[i].Message.ToolCalls)
							toolCallIndex = &idx
						}

						// Ensure we have enough tool calls in the slice
						for len(response.Choices[i].Message.ToolCalls) <= *toolCallIndex {
							response.Choices[i].Message.ToolCalls = append(response.Choices[i].Message.ToolCalls, ToolCall{
								Type:     "function",
								Function: FunctionCall{},
							})
						}

						existingToolCall := &response.Choices[i].Message.ToolCalls[*toolCallIndex]

						// Handle tool call ID and type
						if deltaToolCall.ID != "" {
							existingToolCall.ID = deltaToolCall.ID
							// Track order of tool calls as they appear
							if existingToolCall.ID != "" {
								found := false
								for _, id := range toolCallOrder {
									if id == existingToolCall.ID {
										found = true
										break
									}
								}
								if !found {
									toolCallOrder = append(toolCallOrder, existingToolCall.ID)
								}
							}
						}
						if deltaToolCall.Type != "" {
							existingToolCall.Type = deltaToolCall.Type
						}

						// Handle function name - send "tool call added" event once
						if deltaToolCall.Function.Name != "" && existingToolCall.ID != "" {
							if existingToolCall.Function.Name == "" {
								existingToolCall.Function.Name = deltaToolCall.Function.Name

								log.Messages(ctx, "openai-api", true, []byte(fmt.Sprintf("Tool call started: %s(%s)", deltaToolCall.Function.Name, existingToolCall.ID)))

								if progressToken != nil {
									progress := &types.CompletionProgress{
										Model:   response.Model,
										Partial: true,
										HasMore: true,
										Item: types.CompletionItem{
											ID: existingToolCall.ID,
											ToolCall: &types.ToolCall{
												CallID: existingToolCall.ID,
												Name:   existingToolCall.Function.Name,
											},
										},
									}
									send(ctx, progress, progressToken)
								}
							} else {
								existingToolCall.Function.Name = deltaToolCall.Function.Name
							}
						}

						// Handle function arguments - send deltas like responses API
						if deltaToolCall.Function.Arguments != "" && existingToolCall.ID != "" {
							existingToolCall.Function.Arguments += deltaToolCall.Function.Arguments
							if _, exists := activeToolCalls[existingToolCall.ID]; !exists {
								activeToolCalls[existingToolCall.ID] = ""
							}
							activeToolCalls[existingToolCall.ID] += deltaToolCall.Function.Arguments

							// Send argument delta immediately like we were doing before
							if progressToken != nil && existingToolCall.Function.Name != "" {
								progress := &types.CompletionProgress{
									Model:   response.Model,
									Partial: true,
									HasMore: true,
									Item: types.CompletionItem{
										ID: existingToolCall.ID,
										ToolCall: &types.ToolCall{
											CallID:    existingToolCall.ID,
											Name:      existingToolCall.Function.Name,
											Arguments: deltaToolCall.Function.Arguments, // Send just the delta
										},
									},
								}
								send(ctx, progress, progressToken)
							}
						}
					}
				}
			}

			// Handle completion - send tool call completion events
			if choice.FinishReason != nil {
				response.Choices[i].FinishReason = choice.FinishReason

				// Send completion events for all tool calls when done, in order
				if len(response.Choices[i].Message.ToolCalls) > 0 && *choice.FinishReason == "tool_calls" {
					// Process tool calls in the order they were received
					for _, toolCallID := range toolCallOrder {
						// Find the tool call by ID
						for _, toolCall := range response.Choices[i].Message.ToolCalls {
							if toolCall.ID == toolCallID && !completedToolCalls[toolCall.ID] {
								log.Messages(ctx, "openai-api", true, []byte(fmt.Sprintf("Tool call completed: %s(%s)", toolCall.Function.Name, toolCall.ID)))

								if progressToken != nil {
									progress := &types.CompletionProgress{
										Model:   response.Model,
										Partial: false,
										HasMore: false,
										Item: types.CompletionItem{
											ID: toolCall.ID,
											ToolCall: &types.ToolCall{
												CallID: toolCall.ID,
											},
										},
									}
									send(ctx, progress, progressToken)
								}
								completedToolCalls[toolCall.ID] = true
								break
							}
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, false, err
	}

	return response, response != nil, nil
}

// generateToolCallID generates a random tool call ID
func generateToolCallID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "call_" + hex.EncodeToString(bytes)
}
