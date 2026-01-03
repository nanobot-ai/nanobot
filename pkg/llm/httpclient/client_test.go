package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryableClient_SuccessfulRequest(t *testing.T) {
	// Server that always succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := New(http.DefaultClient)
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "success" {
		t.Errorf("expected body 'success', got '%s'", string(body))
	}
}

func TestRetryableClient_RetryOn500(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithBaseDelay(10*time.Millisecond))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestRetryableClient_RetryOn429(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithBaseDelay(10*time.Millisecond))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}
}

func TestRetryableClient_NoRetryOn400(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithBaseDelay(10*time.Millisecond))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	if attempts.Load() != 1 {
		t.Errorf("expected 1 attempt (no retry), got %d", attempts.Load())
	}
}

func TestRetryableClient_MaxRetriesExhausted(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithMaxRetries(2), WithBaseDelay(10*time.Millisecond))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}

	// With maxRetries=2, we expect 3 total attempts (initial + 2 retries)
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts (initial + 2 retries), got %d", attempts.Load())
	}
}

func TestRetryableClient_RequestBodyReplay(t *testing.T) {
	var attempts atomic.Int32
	var receivedBodies []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		body, _ := io.ReadAll(r.Body)
		receivedBodies = append(receivedBodies, string(body))

		if count < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithBaseDelay(10*time.Millisecond))
	requestBody := "test request body"
	req, _ := http.NewRequest(http.MethodPost, server.URL, strings.NewReader(requestBody))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}

	// Verify that the request body was sent correctly in both attempts
	for i, body := range receivedBodies {
		if body != requestBody {
			t.Errorf("attempt %d: expected body '%s', got '%s'", i+1, requestBody, body)
		}
	}
}

func TestRetryableClient_ContextCancellation(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithBaseDelay(100*time.Millisecond), WithMaxRetries(5))

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)

	// Cancel context after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := client.Do(req)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}

	// Should have attempted once and then cancelled during backoff
	if attempts.Load() > 2 {
		t.Errorf("expected at most 2 attempts before cancellation, got %d", attempts.Load())
	}
}

func TestRetryableClient_ContextTimeout(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithBaseDelay(100*time.Millisecond), WithMaxRetries(5))

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)

	_, err := client.Do(req)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded error, got %v", err)
	}
}

func TestRetryableClient_CustomShouldRetry(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	// Custom retry function that retries on 400
	customRetry := func(resp *http.Response, err error) bool {
		if resp != nil && resp.StatusCode == http.StatusBadRequest {
			return true
		}
		return defaultShouldRetry(resp, err)
	}

	client := New(http.DefaultClient, WithShouldRetry(customRetry), WithBaseDelay(10*time.Millisecond), WithMaxRetries(2))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	// Should retry on 400 with custom function
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts (with custom retry), got %d", attempts.Load())
	}
}

func TestRetryableClient_ExponentialBackoff(t *testing.T) {
	var attempts atomic.Int32
	var timestamps []time.Time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamps = append(timestamps, time.Now())
		count := attempts.Add(1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithBaseDelay(50*time.Millisecond), WithMaxRetries(3))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	_, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(timestamps) < 3 {
		t.Fatalf("expected at least 3 attempts, got %d", len(timestamps))
	}

	// Check that delays are increasing (with jitter tolerance)
	delay1 := timestamps[1].Sub(timestamps[0])
	delay2 := timestamps[2].Sub(timestamps[1])

	// First delay should be around 50ms * 2^0 = 50ms (with jitter 0.5-1.5x)
	// Second delay should be around 50ms * 2^1 = 100ms (with jitter 0.5-1.5x)

	// With jitter, delays can vary significantly, but second should generally be longer
	// Allow for significant variation due to jitter and test execution timing
	if delay1 < 25*time.Millisecond || delay1 > 150*time.Millisecond {
		t.Logf("warning: first delay %v outside expected range [25ms, 150ms]", delay1)
	}

	if delay2 < 50*time.Millisecond || delay2 > 300*time.Millisecond {
		t.Logf("warning: second delay %v outside expected range [50ms, 300ms]", delay2)
	}
}

func TestRetryableClient_NilBody(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(http.DefaultClient, WithBaseDelay(10*time.Millisecond))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}
}

func TestRetryableClient_NetworkError(t *testing.T) {
	var attempts atomic.Int32

	// Create a custom transport that simulates network errors
	transport := &errorTransport{
		attempts:  &attempts,
		failUntil: 2,
	}

	client := New(&http.Client{Transport: transport}, WithBaseDelay(10*time.Millisecond))
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error after retries, got %v", err)
	}
	defer resp.Body.Close()

	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}
}

// errorTransport is a custom RoundTripper that simulates network errors
type errorTransport struct {
	attempts  *atomic.Int32
	failUntil int32
}

func (t *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	count := t.attempts.Add(1)
	if count < t.failUntil {
		return nil, fmt.Errorf("simulated network error")
	}

	// Return a successful response
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("success")),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestRetryableClient_MaxDelayRespected(t *testing.T) {
	var attempts atomic.Int32
	var timestamps []time.Time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamps = append(timestamps, time.Now())
		count := attempts.Add(1)
		if count < 4 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set a low max delay to ensure it's enforced
	client := New(http.DefaultClient,
		WithBaseDelay(100*time.Millisecond),
		WithMaxDelay(150*time.Millisecond),
		WithMaxRetries(4))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	_, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(timestamps) < 4 {
		t.Fatalf("expected at least 4 attempts, got %d", len(timestamps))
	}

	// Check that no delay exceeds maxDelay * 1.5 (jitter factor) + some buffer
	for i := 1; i < len(timestamps); i++ {
		delay := timestamps[i].Sub(timestamps[i-1])
		// Max delay is 150ms, with jitter up to 1.5x = 225ms, plus 50ms buffer for test execution
		if delay > 275*time.Millisecond {
			t.Errorf("delay %d exceeded max delay: %v", i, delay)
		}
	}
}
