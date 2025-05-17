package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

func CompleteCompletionOptions(opts ...types.CompletionOptions) types.CompletionOptions {
	var all types.CompletionOptions
	for _, opt := range opts {
		if opt.Progress != nil {
			if all.Progress != nil {
				panic("multiple progress handlers provided")
			}
			all.Progress = opt.Progress
		}
		if opt.Continue {
			all.Continue = true
		}
	}
	return all
}

func (c *Client) Complete(ctx context.Context, req types.Request, opts ...types.CompletionOptions) (*types.Response, error) {
	var (
		response types.Response
		opt      = CompleteCompletionOptions(opts...)
		event    = struct {
			Type     string         `json:"type"`
			Response types.Response `json:"response"`
		}{}
	)

	req.Stream = &[]bool{true}[0]
	req.Store = new(bool)

	data, _ := json.Marshal(req)
	log.Messages(ctx, "llm", true, data)
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

	lines := bufio.NewScanner(httpResp.Body)
	for lines.Scan() {
		line := lines.Text()

		header, body, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(header) {
		case "data":
			body = strings.TrimSpace(body)
			if err := json.Unmarshal([]byte(body), &event); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to decode event: %w\n%s", err, body)
				continue
			}
			if opt.Progress != nil {
				opt.Progress <- []byte(body)
			}
			if event.Type == "response.completed" || event.Type == "response.failed" || event.Type == "response.incomplete" {
				response = event.Response
			}
		}
	}

	if err := lines.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors in the response
	if response.Error != nil {
		return nil, fmt.Errorf("API error: %s %s", response.Error.Code, response.Error.Message)
	}

	return &response, nil
}
