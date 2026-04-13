package bifrost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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

	resp, err := c.complete(ctx, req)
	if err != nil {
		return nil, err
	}

	cResp, err := toResponse(&completionRequest, resp)
	if err != nil {
		return nil, err
	}

	opt := complete.Complete(opts...)
	if opt.ProgressToken != nil {
		progress := types.CompletionProgress{
			Agent:     completionRequest.Agent,
			Model:     cResp.Model,
			MessageID: cResp.Output.ID,
			Role:      cResp.Output.Role,
		}
		for _, item := range cResp.Output.Items {
			progress.Item = types.CompletionItem{
				Partial:  true,
				HasMore:  true,
				ID:       item.ID,
				Content:  item.Content,
				ToolCall: item.ToolCall,
			}
			llmProgress.Send(ctx, &progress, opt.ProgressToken)

			// Signal end of item
			progress.Item = types.CompletionItem{
				Partial: true,
				ID:      item.ID,
			}
			llmProgress.Send(ctx, &progress, opt.ProgressToken)
		}
	}

	return cResp, nil
}

var httpClient = &http.Client{Timeout: 10 * time.Minute}

func (c *Client) complete(ctx context.Context, req *schemas.BifrostResponsesRequest) (*schemas.BifrostResponsesResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bifrost request: %w", err)
	}
	log.Messages(ctx, "bifrost-request", true, data)

	path := fmt.Sprintf("%s/v1/responses", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(mcp.UserContext(ctx), http.MethodPost, path, bytes.NewBuffer(data))
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

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read bifrost response body: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bifrost request failed: %s %q", httpResp.Status, string(body))
	}

	log.Messages(ctx, "bifrost-request", false, body)

	var resp schemas.BifrostResponsesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bifrost response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("bifrost response error: %s %s", resp.Error.Code, resp.Error.Message)
	}

	return &resp, nil
}
