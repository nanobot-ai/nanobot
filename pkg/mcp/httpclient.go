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
	"sync"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

const SessionIDHeader = "Mcp-Session-Id"

type HTTPClient struct {
	ctx           context.Context
	cancel        context.CancelFunc
	clientLock    sync.RWMutex
	httpClient    *http.Client
	handler       wireHandler
	oauthHandler  *oauth
	baseURL       string
	messageURL    string
	serverName    string
	headers       map[string]string
	waiter        *waiter
	sse           bool
	initialized   bool
	sessionID     string
	sseLock       sync.RWMutex
	needReconnect bool
}

func NewHTTPClient(ctx context.Context, serverName, baseURL, oauthClientName, oauthRedirectURL string, callbackHandler CallbackHandler, clientCredLookup ClientCredLookup, tokenStorage TokenStorage, headers map[string]string) *HTTPClient {
	_, initialized := headers[SessionIDHeader]
	h := &HTTPClient{
		httpClient:    http.DefaultClient,
		oauthHandler:  newOAuth(callbackHandler, clientCredLookup, tokenStorage, oauthClientName, oauthRedirectURL),
		baseURL:       baseURL,
		messageURL:    baseURL,
		serverName:    serverName,
		headers:       maps.Clone(headers),
		waiter:        newWaiter(),
		needReconnect: true,
		sessionID:     headers[SessionIDHeader],
		initialized:   initialized,
	}

	httpClient, err := h.oauthHandler.loadFromStorage(ctx, baseURL)
	if err == nil && httpClient != nil {
		h.httpClient = httpClient
	}

	return h
}

func (s *HTTPClient) SetOAuthCallbackHandler(handler CallbackHandler) {
	s.oauthHandler.callbackHandler = handler
}

func (s *HTTPClient) SessionID() string {
	return s.sessionID
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
	if s.sessionID != "" {
		req.Header.Set(SessionIDHeader, s.sessionID)
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

func (s *HTTPClient) ensureSSE(ctx context.Context, msg *Message, lastEventID any) error {
	s.sseLock.RLock()
	if !s.needReconnect {
		s.sseLock.RUnlock()
		return nil
	}
	s.sseLock.RUnlock()

	// Hold the lock while we try to start the SSE endpoint.
	// We need to make sure that the message URL is set before continuing.
	s.sseLock.Lock()
	defer s.sseLock.Unlock()

	if !s.needReconnect {
		// Check again in case SSE was started while we were waiting for the lock.
		return nil
	}

	gotResponse := make(chan error, 1)
	// Start the SSE stream with the managed context.
	req, err := s.newRequest(s.ctx, http.MethodGet, nil)
	if err != nil {
		return err
	}

	if lastEventID != nil {
		req.Header.Set("Last-Event-ID", fmt.Sprintf("%v", lastEventID))
	}

	s.clientLock.RLock()
	httpClient := s.httpClient
	s.clientLock.RUnlock()

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return AuthRequiredErr{
			ProtectedResourceValue: resp.Header.Get("WWW-Authenticate"),
			Err:                    fmt.Errorf("failed to connect to SSE server: %s: %s", resp.Status, body),
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		_ = resp.Body.Close()
		// If msg is nil, then this is an SSE request for HTTP streaming.
		// If the server doesn't support a separate SSE endpoint, then we can just return.
		if msg == nil && resp.StatusCode == http.StatusMethodNotAllowed {
			s.needReconnect = false
			return nil
		}
		return fmt.Errorf("failed to connect to SSE server: %s", resp.Status)
	}

	s.needReconnect = false

	go func() (err error, send bool) {
		defer func() {
			if err != nil {
				s.sseLock.Lock()
				s.needReconnect = true
				s.sseLock.Unlock()

				// If we get an error, then we aren't reconnecting to the SSE endpoint.
				if send {
					gotResponse <- err
				}
			}

			resp.Body.Close()
		}()

		messages := newSSEStream(resp.Body)

		if !s.sse {
			s.messageURL = s.baseURL
			msg = nil
		} else if msg.Method == "initialize" {
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

			initReq, err := s.newRequest(ctx, http.MethodPost, msg)
			if err != nil {
				return fmt.Errorf("failed to create initialize message req: %w", err), true
			}

			initResp, err := httpClient.Do(initReq)
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
					s.sseLock.Lock()
					if !s.needReconnect {
						s.needReconnect = true
					}
					s.sseLock.Unlock()
				}

				if err := s.ensureSSE(ctx, msg, lastEventID); err != nil {
					return fmt.Errorf("failed to reconnect to SSE server: %v", err), false
				}

				return nil, false
			}

			var msg Message
			if err := json.Unmarshal([]byte(message), &msg); err != nil {
				continue
			}

			if msg.ID == nil {
				msg.ID = uuid.String()
			} else {
				lastEventID = msg.ID
			}

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

	s.clientLock.RLock()
	httpClient := s.httpClient
	s.clientLock.RUnlock()

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		streamingErrorMessage, _ := io.ReadAll(resp.Body)
		return AuthRequiredErr{
			ProtectedResourceValue: resp.Header.Get("WWW-Authenticate"),
			Err:                    fmt.Errorf("failed to initialize HTTP Streaming client: %s: %s", resp.Status, streamingErrorMessage),
		}
	}

	if resp.StatusCode != http.StatusOK {
		streamingErrorMessage, _ := io.ReadAll(resp.Body)
		streamError := fmt.Errorf("failed to initialize HTTP Streaming client: %s: %s", resp.Status, streamingErrorMessage)
		if err := s.ensureSSE(ctx, &msg, nil); err != nil {
			return errors.Join(streamError, err)
		}

		s.sse = true
		return nil
	}

	sessionID := resp.Header.Get(SessionIDHeader)
	if sessionID != "" {
		if s.headers == nil {
			s.headers = make(map[string]string)
		}
		s.headers[SessionIDHeader] = sessionID
	}

	seen, err := s.readResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to decode mcp initialize response: %w", err)
	} else if !seen {
		return fmt.Errorf("no response from server, expected an initialize response")
	}

	defer func() {
		if err == nil {
			s.initialized = true
		}
	}()

	return s.ensureSSE(ctx, nil, nil)
}

func (s *HTTPClient) Send(ctx context.Context, msg Message) (err error) {
	defer func() {
		if err == nil {
			return
		}

		var oauthErr AuthRequiredErr
		if !errors.As(err, &oauthErr) {
			return
		}

		var httpClient *http.Client
		httpClient, err = s.oauthHandler.oauthClient(s.ctx, s, s.baseURL, oauthErr.ProtectedResourceValue)
		if err != nil || httpClient == nil {
			streamError := fmt.Errorf("failed to initialize HTTP Streaming client: %w", oauthErr)
			err = errors.Join(streamError, err)
			return
		}

		s.clientLock.Lock()
		s.httpClient = httpClient
		s.clientLock.Unlock()

		err = s.send(ctx, msg)
	}()

	return s.send(ctx, msg)
}

func (s *HTTPClient) send(ctx context.Context, msg Message) error {
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

	if err := s.ensureSSE(ctx, &msg, nil); err != nil {
		return fmt.Errorf("failed to restart SSE: %w", err)
	}

	req, err := s.newRequest(ctx, http.MethodPost, msg)
	if err != nil {
		return err
	}

	s.clientLock.RLock()
	httpClient := s.httpClient
	s.clientLock.RUnlock()

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		streamingErrorMessage, _ := io.ReadAll(resp.Body)
		return AuthRequiredErr{
			ProtectedResourceValue: resp.Header.Get("WWW-Authenticate"),
			Err:                    fmt.Errorf("failed to send message: %s: %s", resp.Status, streamingErrorMessage),
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to send message: %s", resp.Status)
	}

	// It is possible for the ContentLength here to be -1.
	if s.sse || resp.StatusCode == http.StatusAccepted {
		return nil
	}

	_, err = s.readResponse(resp)
	return err
}

func (s *HTTPClient) readResponse(resp *http.Response) (bool, error) {
	var seen bool
	handle := func(message *Message) {
		seen = true
		log.Messages(s.ctx, s.serverName, false, message.Result)
		go s.handler(*message)
	}

	if resp.Header.Get("Content-Type") == "text/event-stream" {
		stream := newSSEStream(resp.Body)
		for {
			data, ok := stream.readNextMessage()
			if !ok {
				return seen, nil
			}

			var message Message
			if err := json.Unmarshal([]byte(data), &message); err != nil {
				return seen, fmt.Errorf("failed to decode response: %w", err)
			}

			handle(&message)
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return seen, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(data) == 0 {
		return false, nil
	}

	if data[0] != '{' {
		return false, fmt.Errorf("invalid response format, expected JSON object, got: %s", data)
	}

	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	handle(&message)
	return seen, nil
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
