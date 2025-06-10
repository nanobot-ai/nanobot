package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/obot-platform/nanobot/pkg/complete"
	"github.com/obot-platform/nanobot/pkg/log"
	"github.com/obot-platform/nanobot/pkg/types"
)

type Client struct {
	Config
	config types.Config
}

type Config struct {
	APIKey  string
	BaseURL string
	Headers map[string]string
}

// NewClient creates a new OpenAI client with the provided API key and base URL.
func NewClient(cfg Config, config types.Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com/v1"
	}
	if cfg.Headers == nil {
		cfg.Headers = map[string]string{}
	}
	if _, ok := cfg.Headers["x-api-key"]; !ok && cfg.APIKey != "" {
		cfg.Headers["x-api-key"] = cfg.APIKey
	}
	if _, ok := cfg.Headers["anthropic-version"]; !ok {
		cfg.Headers["anthropic-version"] = "2023-06-01"
	}
	if _, ok := cfg.Headers["Content-Type"]; !ok {
		cfg.Headers["Content-Type"] = "application/json"
	}

	return &Client{
		Config: cfg,
		config: config,
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

func (c *Client) complete(ctx context.Context, req Request, opts ...types.CompletionOptions) (*Response, error) {
	var (
		opt = complete.Complete(opts...)
	)

	req.Stream = true

	data, _ := json.Marshal(req)
	log.Messages(ctx, "anthropic-api", true, data)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/messages", bytes.NewBuffer(data))
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

	var (
		lines       = bufio.NewScanner(httpResp.Body)
		resp        Response
		partialJSON = ""
	)

	for lines.Scan() {
		line := lines.Text()

		header, body, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(header) != "data" {
			continue
		}

		var delta DeltaEvent
		body = strings.TrimSpace(body)
		if err := json.Unmarshal([]byte(body), &delta); err != nil {
			log.Errorf(ctx, "failed to decode event: %v: %s", err, body)
			continue
		}
		if opt.Progress != nil {
			opt.Progress <- []byte(body)
		}
		contentIndex := len(resp.Content) - 1
		switch delta.Type {
		case "message_start":
			resp = delta.Message
		case "content_block_start":
			partialJSON = ""
			resp.Content = append(resp.Content, delta.ContentBlock)
		case "content_block_delta":
			switch delta.Delta.Type {
			case "text_delta":
				if contentIndex >= 0 {
					*resp.Content[contentIndex].Text += delta.Delta.Text
				}
			case "input_json_delta":
				partialJSON += delta.Delta.PartialJSON
			}
		case "content_block_stop":
			if contentIndex >= 0 && partialJSON != "" {
				args := map[string]any{}
				if err := json.Unmarshal([]byte(partialJSON), &args); err != nil {
					return nil, fmt.Errorf("failed to unmarshal function call arguments: %w", err)
				}
				resp.Content[contentIndex].Input = args
			}
		case "message_delta":
			err := json.Unmarshal([]byte(body), &struct {
				Delta *Response `json:"delta"`
			}{
				Delta: &resp,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal message delta: %w", err)
			}
		case "message_stop":
			// nothing to do, but here for completeness
		}
	}

	if err := lines.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	respData, err := json.Marshal(resp)
	if err == nil {
		log.Messages(ctx, "anthropic-api", false, respData)
	}

	return &resp, nil
}
