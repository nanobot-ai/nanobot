package mcp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestHTTPClientInitializeDoesNotShareSSEErrorState(t *testing.T) {
	readStarted := make(chan struct{})
	releaseRead := make(chan struct{})

	client, err := newHTTPClient(
		"race-test",
		Server{BaseURL: "http://mcp.example/mcp"},
		HTTPClientOptions{},
		nil,
		nil,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(context.Background(), func(context.Context, Message) {}); err != nil {
		t.Fatal(err)
	}
	client.httpClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case http.MethodPost:
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     http.StatusText(http.StatusOK),
					Header: http.Header{
						"Content-Type":  []string{"application/json"},
						SessionIDHeader: []string{"race-session"},
					},
					Body: &synchronizedReadCloser{
						Reader:      strings.NewReader(`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-06-18","capabilities":{},"serverInfo":{"name":"test","version":"1"}}}`),
						readStarted: readStarted,
						release:     releaseRead,
					},
					Request: req,
				}, nil
			case http.MethodGet:
				<-readStarted
				close(releaseRead)
				return nil, errors.New("event stream unavailable")
			default:
				return nil, errors.New("unexpected HTTP method")
			}
		}),
	}

	if err := client.initialize(context.Background(), Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  []byte(`{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}`),
	}); err != nil {
		t.Fatal(err)
	}
	client.Close(false)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type synchronizedReadCloser struct {
	io.Reader
	once        sync.Once
	readStarted chan struct{}
	release     <-chan struct{}
}

func (r *synchronizedReadCloser) Read(p []byte) (int, error) {
	r.once.Do(func() {
		close(r.readStarted)
		<-r.release
	})
	return r.Reader.Read(p)
}

func (*synchronizedReadCloser) Close() error {
	return nil
}
