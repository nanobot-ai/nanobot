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
	var response *ChatCompletionResponse

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

		// Accumulate content from delta and send progress updates
		for i, choice := range streamResp.Choices {
			if i >= len(response.Choices) {
				continue
			}

			if choice.Delta != nil {
				if choice.Delta.Content != nil {
					// Handle content accumulation
					if deltaContent, ok := choice.Delta.Content.(string); ok && deltaContent != "" {
						currentContent := ""
						if response.Choices[i].Message.Content != nil {
							if current, ok := response.Choices[i].Message.Content.(string); ok {
								currentContent = current
							}
						}
						response.Choices[i].Message.Content = currentContent + deltaContent

						// Send progress update with only the delta content
						if progressToken != nil {
							progress := &types.CompletionProgress{
								Model:   response.Model,
								Partial: true,
								HasMore: true,
								Item: types.CompletionItem{
									Message: &mcp.SamplingMessage{
										Role: "assistant",
										Content: mcp.Content{
											Type: "text",
											Text: deltaContent, // Send only the delta, not accumulated content
										},
									},
								},
							}
							send(ctx, progress, progressToken)
						}
					}
				}

				// Handle tool calls
				if choice.Delta.ToolCalls != nil {
					// Append tool calls rather than replace them, as OpenAI may send them in multiple chunks
					if response.Choices[i].Message.ToolCalls == nil {
						response.Choices[i].Message.ToolCalls = []ToolCall{}
					}
					for _, toolCall := range choice.Delta.ToolCalls {
						// Only add tool calls with valid names to avoid empty tool calls
						if toolCall.Function.Name != "" {
							response.Choices[i].Message.ToolCalls = append(response.Choices[i].Message.ToolCalls, toolCall)
						}
					}
				}
			}

			if choice.FinishReason != nil {
				response.Choices[i].FinishReason = choice.FinishReason
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
