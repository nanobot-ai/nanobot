package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

type HTTPClient struct {
	ctx         context.Context
	cancel      context.CancelFunc
	handler     wireHandler
	baseURL     string
	messageURL  string
	serverName  string
	headers     map[string]string
	waiter      *waiter
	sse         bool
	initialized bool
}

func NewHTTPClient(serverName, baseURL string, headers map[string]string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		messageURL: baseURL,
		serverName: serverName,
		headers:    maps.Clone(headers),
		waiter:     newWaiter(),
	}
}

func (s *HTTPClient) Close() {
	if s.cancel != nil {
		s.cancel()
	}
	s.waiter.Close()
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
		log.Messages(ctx, s.serverName, true, data)
	}

	u := s.messageURL
	if method == http.MethodGet || u == "" {
		// If this is a GET request, then it is starting the SSE stream.
		// In this case, we need to use the base URL instead.
		u = s.baseURL
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "text/event-stream")
	if method != http.MethodGet {
		// Don't add because some *cough* CloudFront *cough* proxies don't like it
		req.Header.Set("Accept", "application/json, text/event-stream")
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (s *HTTPClient) startSSE(ctx context.Context, msg *Message, lastEventID any) error {
	gotResponse := make(chan error, 1)
	// Start the SSE stream with the managed context.
	req, err := s.newRequest(s.ctx, http.MethodGet, nil)
	if err != nil {
		return err
	}

	if lastEventID != nil {
		req.Header.Set("Last-Event-ID", fmt.Sprintf("%v", lastEventID))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		// If msg is nil, then this is an SSE request for HTTP streaming.
		// If the server doesn't support a separate SSE endpoint, then we can just return.
		if msg == nil && resp.StatusCode == http.StatusMethodNotAllowed {
			return nil
		}
		return fmt.Errorf("failed to connect to SSE server: %s", resp.Status)
	}

	go func() (err error, send bool) {
		defer func() {
			if err != nil {
				// If we get an error, then we aren't reconnecting to the SSE endpoint.
				// Therefore, close the waiter to indicate that we're done.
				s.waiter.Close()
				if send {
					gotResponse <- err
				}
			}

			resp.Body.Close()
		}()

		messages := newSSEStream(resp.Body)

		if msg == nil {
			s.messageURL = s.baseURL
		} else {
			data, ok := messages.readNextMessage()
			if !ok {
				return fmt.Errorf("failed to read SSE message: %w", messages.err()), true
			}

			baseURL, err := url.Parse(s.baseURL)
			if err != nil {
				return fmt.Errorf("failed to parse SSE URL: %w", err), true
			}

			u, err := url.Parse(data)
			if err != nil {
				return fmt.Errorf("failed to parse returned SSE URL: %w", err), true
			}

			baseURL.Path = u.Path
			baseURL.RawQuery = u.RawQuery
			s.messageURL = baseURL.String()
			s.sse = true

			initReq, err := s.newRequest(ctx, http.MethodPost, msg)
			if err != nil {
				return fmt.Errorf("failed to create initialize message req: %w", err), true
			}

			initResp, err := http.DefaultClient.Do(initReq)
			if err != nil {
				return fmt.Errorf("failed to POST initialize message: %w", err), true
			}
			body, _ := io.ReadAll(initResp.Body)
			_ = initResp.Body.Close()

			if initResp.StatusCode != http.StatusOK && initResp.StatusCode != http.StatusAccepted {
				return fmt.Errorf("failed to POST initialize message got status: %s: %s", initResp.Status, body), true
			}
		}

		close(gotResponse)

		for {
			message, ok := messages.readNextMessage()
			if !ok {
				if err := messages.err(); err != nil {
					if errors.Is(err, context.Canceled) {
						log.Debugf(ctx, "context canceled reading SSE message: %v", messages.err())
					} else {
						log.Errorf(ctx, "failed to read SSE message: %v", messages.err())
					}
				}

				select {
				case <-s.ctx.Done():
					// If the context is done, then we don't need to reconnect.
					// Returning the error here will close the waiter, indicating that
					// the client is done.
					return s.ctx.Err(), false
				default:
					if msg != nil {
						msg.ID = uuid.String()
					}
				}

				if err := s.startSSE(ctx, msg, lastEventID); err != nil {
					return fmt.Errorf("failed to reconnect to SSE server: %v", err), false
				}

				return nil, false
			}

			var msg Message
			if err := json.Unmarshal([]byte(message), &msg); err != nil {
				continue
			}

			lastEventID = msg.ID

			log.Messages(ctx, s.serverName, false, []byte(message))
			s.handler(msg)
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
		streamingErrorMessage, _ := io.ReadAll(resp.Body)
		streamError := fmt.Errorf("failed to initialize HTTP Streaming client: %s: %s", resp.Status, streamingErrorMessage)
		if err := s.startSSE(ctx, &msg, nil); err != nil {
			return errors.Join(streamError, err)
		}

		return nil
	}

	sessionID := resp.Header.Get("Mcp-Session-Id")
	if sessionID != "" {
		if s.headers == nil {
			s.headers = make(map[string]string)
		}
		s.headers["Mcp-Session-Id"] = sessionID
	}

	initResp, err := readResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to decode mcp initialize response: %w", err)
	}

	if initResp != nil {
		s.handler(*initResp)
	}

	defer func() {
		if err == nil {
			s.initialized = true
		}
	}()

	return s.startSSE(ctx, nil, nil)
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

	if s.sse || resp.ContentLength == 0 {
		return nil
	}

	result, err := readResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to decode mcp send message response: %w", err)
	}

	if result != nil {
		log.Messages(ctx, s.serverName, false, result.Result)
		go s.handler(*result)
	}

	return nil
}

func readResponse(resp *http.Response) (*Message, error) {
	if resp.ContentLength == 0 {
		return nil, nil
	}
	var init io.Reader
	if resp.Header.Get("Content-Type") == "application/json" {
		init = resp.Body
	} else {
		stream := newSSEStream(resp.Body)
		initEvent, ok := stream.readNextMessage()
		if !ok {
			return nil, fmt.Errorf("failed to read stream response: %w", stream.err())
		}

		init = strings.NewReader(initEvent)
	}

	var message Message
	if err := json.NewDecoder(init).Decode(&message); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &message, nil
}

type SSEStream struct {
	lines *bufio.Scanner
}

func newSSEStream(input io.Reader) *SSEStream {
	lines := bufio.NewScanner(input)
	lines.Buffer(make([]byte, 0, 1024), 10*1024*1024)
	return &SSEStream{
		lines: lines,
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
		} else if strings.HasPrefix(line, "data:") && (eventName == "message" || eventName == "" || eventName == "endpoint") {
			data := strings.TrimSpace(line[5:])
			return data, true
		}
	}

	return "", false
}
