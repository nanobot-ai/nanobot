package responses

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/types"
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

	return toResponse(&completionRequest, resp)
}

func (c *Client) complete(ctx context.Context, req Request, opts ...types.CompletionOptions) (*Response, error) {
	var (
		response Response
		opt      = complete.Complete(opts...)
	)

	req.Stream = &[]bool{true}[0]
	req.Store = new(bool)

	data, _ := json.Marshal(req)
	log.Messages(ctx, "responses-api", true, data)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/responses", bytes.NewBuffer(data))
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
	if response.Error != nil {
		return nil, fmt.Errorf("responses API error: %s %s", response.Error.Code, response.Error.Message)
	}

	return &response, nil
}
