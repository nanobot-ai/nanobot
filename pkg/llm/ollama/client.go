package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type Client struct {
	Config
}

type Config struct {
	BaseURL string
	Headers map[string]string
}

// NewClient creates a new Ollama client with the provided base URL.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434"
	}
	if cfg.Headers == nil {
		cfg.Headers = map[string]string{}
	}
	if _, ok := cfg.Headers["Content-Type"]; !ok {
		cfg.Headers["Content-Type"] = "application/json"
	}

	return &Client{
		Config: cfg,
	}
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

func (c *Client) complete(ctx context.Context, req Request, opts ...types.CompletionOptions) (*Response, error) {
	var (
		opt = complete.Complete(opts...)
	)

	// Try streaming first, fallback to non-streaming if needed
	req.Stream = true

	data, _ := json.Marshal(req)
	log.Messages(ctx, "ollama-api", true, data)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/chat", bytes.NewBuffer(data))
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

	// Check if response is streaming
	contentType := httpResp.Header.Get("Content-Type")
	if contentType == "application/json" {
		// Non-streaming response
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var resp Response
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// Send progress notification for non-streaming response
		if resp.Message.Content != "" {
			send(ctx, &types.CompletionProgress{
				Model:   resp.Model,
				Partial: false,
				HasMore: false,
				Item: types.CompletionItem{
					ID: resp.Model,
					Message: &mcp.SamplingMessage{
						Role: "assistant",
						Content: mcp.Content{
							Type: "text",
							Text: resp.Message.Content + "\n",
						},
					},
				},
			}, opt.ProgressToken)
		}

		log.Messages(ctx, "ollama-api", false, body)
		return &resp, nil
	}

	// Handle streaming response
	var (
		lines = bufio.NewScanner(httpResp.Body)
		resp  Response
	)

	// Initialize the response message
	resp.Message = Message{
		Role: "assistant",
	}

	for lines.Scan() {
		line := lines.Text()
		if line == "" {
			continue
		}

		var streamResp Response
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			log.Errorf(ctx, "failed to decode stream response: %v: %s", err, line)
			continue
		}

		// Send progress for any content we receive
		if streamResp.Message.Content != "" {
			send(ctx, &types.CompletionProgress{
				Model:   streamResp.Model,
				Partial: true,
				HasMore: !streamResp.Done,
				Item: types.CompletionItem{
					ID: streamResp.Model,
					Message: &mcp.SamplingMessage{
						Role: "assistant",
						Content: mcp.Content{
							Type: "text",
							Text: streamResp.Message.Content,
						},
					},
				},
			}, opt.ProgressToken)
		}

		// If this is the first response and it has complete content with done=true,
		// use it directly (single-shot response)
		if resp.Message.Content == "" && streamResp.Message.Content != "" && streamResp.Done {
			resp = streamResp
			break
		}

		// Otherwise accumulate the response (streaming response)
		if streamResp.Message.Content != "" {
			resp.Message.Content += streamResp.Message.Content
		}
		if streamResp.Model != "" {
			resp.Model = streamResp.Model
		}

		// Copy role if not set
		if resp.Message.Role == "" && streamResp.Message.Role != "" {
			resp.Message.Role = streamResp.Message.Role
		}

		// Handle tool calls
		if len(streamResp.Message.ToolCalls) > 0 {
			resp.Message.ToolCalls = streamResp.Message.ToolCalls
		}

		// When done, preserve accumulated data
		if streamResp.Done {
			resp.Done = streamResp.Done
			resp.DoneReason = streamResp.DoneReason
			// Ensure we have the model set
			if resp.Model == "" && streamResp.Model != "" {
				resp.Model = streamResp.Model
			}
			break
		}
	}

	if err := lines.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// If we didn't get a done signal but have content, still return the response
	if !resp.Done && resp.Message.Content == "" && len(resp.Message.ToolCalls) == 0 {
		return nil, fmt.Errorf("no response content received from Ollama")
	}

	// Ensure we have a model set
	if resp.Model == "" {
		resp.Model = req.Model
	}

	// Send final newline for streaming responses
	if resp.Message.Content != "" {
		send(ctx, &types.CompletionProgress{
			Model:   resp.Model,
			Partial: false,
			HasMore: false,
			Item: types.CompletionItem{
				ID: resp.Model,
				Message: &mcp.SamplingMessage{
					Role: "assistant",
					Content: mcp.Content{
						Type: "text",
						Text: "\n",
					},
				},
			},
		}, opt.ProgressToken)
	}

	respData, err := json.Marshal(resp)
	if err == nil {
		log.Messages(ctx, "ollama-api", false, respData)
	}

	return &resp, nil
}
