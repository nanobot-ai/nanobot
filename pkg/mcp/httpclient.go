package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
)

type HTTPClient struct {
	ctx         context.Context
	cancel      context.CancelFunc
	handler     wireHandler
	baseURL     string
	messageURL  string
	headers     map[string]string
	input       io.ReadCloser
	waiter      *waiter
	sse         bool
	initialized bool
}

func NewHTTPClient(baseURL string, headers map[string]string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		messageURL: baseURL,
		headers:    maps.Clone(headers),
		waiter:     newWaiter(),
	}
}

func (s *HTTPClient) Close() {
	s.waiter.Close()
	_ = s.input.Close()
}

func (s *HTTPClient) Wait() {
	s.waiter.Wait()
}

func (s *HTTPClient) newRequest(ctx context.Context, method string, in any) (*http.Request, error) {
	var (
		body io.Reader
	)
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, s.messageURL, body)
	if err != nil {
		return nil, err
	}
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "text/event-stream")
	if method != http.MethodGet {
		req.Header.Add("Accept", "application/json")
	}
	return req, nil
}

func (s *HTTPClient) startSSE(ctx context.Context, msg *Message) error {
	gotResponse := make(chan error, 1)
	req, err := s.newRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to connect to SSE server: %s", resp.Status)
	}

	go func() {
		defer s.waiter.Close()
		defer resp.Body.Close()

		messages := newSSEStream(resp.Body)

		if msg == nil {
			s.messageURL = s.baseURL
		} else {
			data, ok := messages.readNextMessage()
			if !ok {
				gotResponse <- fmt.Errorf("failed to read SSE message: %w", messages.err())
				return
			}

			baseURL, err := url.Parse(s.baseURL)
			if err != nil {
				gotResponse <- fmt.Errorf("failed to parse SSE URL: %w", err)
				return
			}

			u, err := url.Parse(data)
			if err != nil {
				gotResponse <- fmt.Errorf("failed to parse returned SSE URL: %w", err)
				return
			}

			baseURL.Path = u.Path
			baseURL.RawQuery = u.RawQuery
			s.messageURL = baseURL.String()
			s.sse = true

			initReq, err := s.newRequest(ctx, http.MethodPost, msg)
			if err != nil {
				gotResponse <- fmt.Errorf("failed to create initialize message req: %w", err)
				return
			}

			initResp, err := http.DefaultClient.Do(initReq)
			if err != nil {
				gotResponse <- fmt.Errorf("failed to POST initialize message: %w", err)
				return
			}
			_ = initResp.Body.Close()

			if initResp.StatusCode != http.StatusOK && initResp.StatusCode != http.StatusAccepted {
				gotResponse <- fmt.Errorf("failed to POST initialize message got status: %s", initResp.Status)
				return
			}
		}

		close(gotResponse)

		for {
			message, ok := messages.readNextMessage()
			if !ok {
				return
			}
			var msg Message
			if err := json.Unmarshal([]byte(message), &msg); err != nil {
				continue
			}
			go s.handler(msg)
		}
	}()

	return <-gotResponse
}

func (s *HTTPClient) Start(ctx context.Context, handler wireHandler) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.handler = handler
	return nil
}

func (s *HTTPClient) initialize(ctx context.Context, msg Message) (err error) {
	req, err := s.newRequest(ctx, http.MethodPost, msg)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return s.startSSE(ctx, &msg)
	}

	sessionID := resp.Header.Get("Mcp-Session-Id")
	if sessionID != "" {
		if s.headers == nil {
			s.headers = make(map[string]string)
		}
		s.headers["Mcp-Session-Id"] = sessionID
	}

	var initResp Message
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	s.handler(initResp)

	defer func() {
		if err == nil {
			s.initialized = true
		}
	}()

	return s.startSSE(ctx, nil)
}

func (s *HTTPClient) Send(ctx context.Context, msg Message) error {
	if !s.initialized {
		if msg.Method != "initialize" {
			return fmt.Errorf("client not initialized, must send InitializeRequest first")
		}
		if err := s.initialize(ctx, msg); err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}
		s.initialized = true
		return nil
	}

	req, err := s.newRequest(ctx, http.MethodPost, msg)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to send message: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if s.sse {
		return nil
	}

	if len(data) > 0 {
		var result Message
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to unmarshal mcp send message response: %w", err)

		}
		go s.handler(result)
	}
	return nil
}

type SSEStream struct {
	lines *bufio.Scanner
}

func newSSEStream(input io.Reader) *SSEStream {
	return &SSEStream{
		lines: bufio.NewScanner(input),
	}
}

func (s *SSEStream) err() error {
	return s.lines.Err()
}

func (s *SSEStream) readNextMessage() (string, bool) {
	var (
		eventName string
	)
	for s.lines.Scan() {
		line := s.lines.Text()
		if len(line) == 0 {
			eventName = ""
			continue
		}
		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(line[6:])
			continue
		} else if strings.HasPrefix(line, "data:") && (eventName == "message" || eventName == "") {
			data := strings.TrimSpace(line[5:])
			return data, true
		}
	}

	return "", false
}
