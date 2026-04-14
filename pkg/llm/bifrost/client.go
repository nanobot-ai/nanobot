package bifrost

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/nanobot-ai/nanobot/pkg/complete"
	llmProgress "github.com/nanobot-ai/nanobot/pkg/llm/progress"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type Client struct {
	Config
}

type Config struct {
	APIKey   string
	BaseURL  string
	Headers  map[string]string
	Provider string
}

func NewClient(cfg Config) *Client {
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
	if cfg.Headers == nil {
		cfg.Headers = map[string]string{}
	}
	if _, ok := cfg.Headers["Authorization"]; !ok && cfg.APIKey != "" {
		cfg.Headers["Authorization"] = "Bearer " + cfg.APIKey
	}
	if _, ok := cfg.Headers["Content-Type"]; !ok {
		cfg.Headers["Content-Type"] = "application/json"
	}
	return &Client{Config: cfg}
}

func (c *Client) Complete(ctx context.Context, completionRequest types.CompletionRequest, opts ...types.CompletionOptions) (*types.CompletionResponse, error) {
	req, err := toRequest(c.Provider, &completionRequest)
	if err != nil {
		return nil, err
	}

	resp, err := c.complete(ctx, completionRequest.Agent, req, opts...)
	if err != nil {
		return nil, err
	}

	return toResponse(&completionRequest, resp)
}

var httpClient = &http.Client{Timeout: 10 * time.Minute}

func (c *Client) complete(ctx context.Context, agentName string, req *schemas.BifrostResponsesRequest, opts ...types.CompletionOptions) (*schemas.BifrostResponsesResponse, error) {
	opt := complete.Complete(opts...)

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bifrost request: %w", err)
	}
	log.Messages(ctx, "bifrost-request", true, data)

	url := fmt.Sprintf("%s/v1/responses", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(mcp.UserContext(ctx), http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	for key, value := range c.Headers {
		httpReq.Header.Set(key, value)
	}
	if requestType := types.InternalLLMRequestType(ctx); requestType != "" {
		httpReq.Header.Set(types.InternalLLMRequestTypeHeader, requestType)
	}

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("bifrost request failed: %s %q", httpResp.Status, string(body))
	}

	return c.parseStream(ctx, agentName, httpResp.Body, opt.ProgressToken)
}

func (c *Client) parseStream(ctx context.Context, agentName string, body io.Reader, progressToken any) (*schemas.BifrostResponsesResponse, error) {
	lines := bufio.NewScanner(body)
	// Use 1 MiB buffer — the response.completed event carries the full response body.
	lines.Buffer(make([]byte, 0, 4096), 1024*1024)

	var (
		resp     *schemas.BifrostResponsesResponse
		progress = types.CompletionProgress{Agent: agentName}
	)

	for lines.Scan() {
		line := lines.Text()

		header, body, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(header) != "data" {
			continue
		}
		body = strings.TrimSpace(body)
		if body == "[DONE]" {
			break
		}

		var event schemas.BifrostResponsesStreamResponse
		if err := json.Unmarshal([]byte(body), &event); err != nil {
			slog.Error("bifrost: failed to decode stream event", "error", err, "body", body)
			continue
		}

		switch event.Type {
		case schemas.ResponsesStreamResponseTypeCreated:
			if event.Response != nil {
				progress.Model = event.Response.Model
				if event.Response.ID != nil {
					progress.MessageID = *event.Response.ID
				}
			}

		case schemas.ResponsesStreamResponseTypeOutputItemAdded:
			if event.Item == nil || event.Item.Type == nil {
				continue
			}
			itemID := ""
			if event.Item.ID != nil {
				itemID = *event.Item.ID
			}
			switch *event.Item.Type {
			case schemas.ResponsesMessageTypeMessage:
				progress.Item = types.CompletionItem{
					Partial: true,
					HasMore: true,
					ID:      itemID,
					Content: &mcp.Content{Type: "text"},
				}
			case schemas.ResponsesMessageTypeFunctionCall:
				tc := &types.ToolCall{}
				if event.Item.ResponsesToolMessage != nil {
					if event.Item.ResponsesToolMessage.Name != nil {
						tc.Name = *event.Item.ResponsesToolMessage.Name
					}
					if event.Item.ResponsesToolMessage.CallID != nil {
						tc.CallID = *event.Item.ResponsesToolMessage.CallID
					}
				}
				progress.Item = types.CompletionItem{
					Partial:  true,
					HasMore:  true,
					ID:       itemID,
					ToolCall: tc,
				}
			}

		case schemas.ResponsesStreamResponseTypeOutputTextDelta:
			if event.Delta != nil && progress.Item.Content != nil {
				progress.Item.Content.Text = *event.Delta
				llmProgress.Send(ctx, &progress, progressToken)
			}

		case schemas.ResponsesStreamResponseTypeFunctionCallArgumentsDelta:
			if event.Delta != nil && progress.Item.ToolCall != nil {
				progress.Item.ToolCall.Arguments = *event.Delta
				llmProgress.Send(ctx, &progress, progressToken)
			}

		case schemas.ResponsesStreamResponseTypeOutputItemDone:
			if progress.Item.ID != "" {
				llmProgress.Send(ctx, &types.CompletionProgress{
					Agent:     agentName,
					Model:     progress.Model,
					MessageID: progress.MessageID,
					Item:      types.CompletionItem{Partial: true, ID: progress.Item.ID},
				}, progressToken)
			}
			progress.Item = types.CompletionItem{}

		case schemas.ResponsesStreamResponseTypeCompleted:
			if event.Response != nil {
				resp = event.Response
				data, _ := json.Marshal(resp)
				log.Messages(ctx, "bifrost-request", false, data)
			}

		case schemas.ResponsesStreamResponseTypeFailed, schemas.ResponsesStreamResponseTypeIncomplete:
			if event.Response != nil && event.Response.Error != nil {
				return nil, fmt.Errorf("bifrost stream error: %s %s", event.Response.Error.Code, event.Response.Error.Message)
			}
			return nil, fmt.Errorf("bifrost stream ended with status: %s", event.Type)
		}
	}

	if err := lines.Err(); err != nil {
		return nil, fmt.Errorf("bifrost stream read error: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("bifrost stream ended without a completed response")
	}
	return resp, nil
}
