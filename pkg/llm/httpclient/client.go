package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RetryableClient wraps an http.Client with automatic retry logic
// using exponential backoff with jitter for failed requests.
type RetryableClient struct {
	client      *http.Client
	maxRetries  int
	baseDelay   time.Duration
	maxDelay    time.Duration
	shouldRetry func(resp *http.Response, err error) bool
}

// Options configures the RetryableClient behavior
type Options struct {
	MaxRetries  *int
	BaseDelay   *time.Duration
	MaxDelay    *time.Duration
	ShouldRetry func(resp *http.Response, err error) bool
}

// Merge combines two Options, with the other Options taking precedence
func (o Options) Merge(other Options) (result Options) {
	result.MaxRetries = complete.Last(o.MaxRetries, other.MaxRetries)
	result.BaseDelay = complete.Last(o.BaseDelay, other.BaseDelay)
	result.MaxDelay = complete.Last(o.MaxDelay, other.MaxDelay)
	result.ShouldRetry = o.ShouldRetry
	if other.ShouldRetry != nil {
		result.ShouldRetry = other.ShouldRetry
	}
	return
}

// Complete fills in default values for any unset options
func (o Options) Complete() Options {
	if o.MaxRetries == nil {
		o.MaxRetries = ptr(3)
	}
	if o.BaseDelay == nil {
		o.BaseDelay = ptr(time.Second)
	}
	if o.MaxDelay == nil {
		o.MaxDelay = ptr(30 * time.Second)
	}
	if o.ShouldRetry == nil {
		o.ShouldRetry = defaultShouldRetry
	}
	return o
}

// WithMaxRetries sets the maximum number of retry attempts (default: 3)
func WithMaxRetries(max int) Options {
	return Options{MaxRetries: ptr(max)}
}

// WithBaseDelay sets the initial delay for exponential backoff (default: 1s)
func WithBaseDelay(delay time.Duration) Options {
	return Options{BaseDelay: ptr(delay)}
}

// WithMaxDelay sets the maximum delay between retries (default: 30s)
func WithMaxDelay(delay time.Duration) Options {
	return Options{MaxDelay: ptr(delay)}
}

// WithShouldRetry sets a custom function to determine if a request should be retried
func WithShouldRetry(fn func(resp *http.Response, err error) bool) Options {
	return Options{ShouldRetry: fn}
}

// New creates a new RetryableClient with the given http.Client and options
func New(client *http.Client, opts ...Options) *RetryableClient {
	if client == nil {
		client = http.DefaultClient
	}

	opt := complete.Complete(opts...)

	return &RetryableClient{
		client:      client,
		maxRetries:  *opt.MaxRetries,
		baseDelay:   *opt.BaseDelay,
		maxDelay:    *opt.MaxDelay,
		shouldRetry: opt.ShouldRetry,
	}
}

// ptr is a helper function to create a pointer to a value
func ptr[T any](v T) *T {
	return &v
}

// defaultShouldRetry determines if a request should be retried based on
// the response status code and error. Retries on network errors, timeouts,
// and 5xx server errors.
func defaultShouldRetry(resp *http.Response, err error) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on 5xx server errors
	if resp != nil && resp.StatusCode >= 500 {
		return true
	}

	// Retry on 429 (rate limit)
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		return true
	}

	return false
}

// Do executes an HTTP request with automatic retry logic using exponential backoff.
// The request body is assumed to be idempotent and will be replayed on retries.
func (c *RetryableClient) Do(req *http.Request) (*http.Response, error) {
	// Create span for HTTP request
	tracer := otel.Tracer("llm.httpclient")
	ctx, span := tracer.Start(req.Context(), "llm.http.request",
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
			attribute.String("http.host", req.URL.Host),
		),
	)
	defer span.End()

	// Replace request context with traced context
	req = req.WithContext(ctx)

	var resp *http.Response
	var err error
	var bodyBytes []byte

	// Read the request body once so we can replay it on retries
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to read request body")
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()

		// Record request size
		span.SetAttributes(attribute.Int("http.request.body.size", len(bodyBytes)))
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Check if context is cancelled before attempting
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		default:
		}

		// Reset request body for retry
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Execute the request
		resp, err = c.client.Do(req)

		// Record status code and response metadata if we got a response
		if resp != nil {
			span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

			// Record Content-Length if present (may not be set for streaming responses)
			if contentLength := resp.ContentLength; contentLength > 0 {
				span.SetAttributes(attribute.Int64("http.response.body.size", contentLength))
			}
		}

		// Check if we should retry
		if !c.shouldRetry(resp, err) {
			// Success or non-retryable error
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "http request failed")
			} else {
				span.SetStatus(codes.Ok, "request completed")
			}
			return resp, err
		}

		// Don't retry if we've exhausted attempts
		if attempt >= c.maxRetries {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "max retries exceeded")
			} else {
				span.SetStatus(codes.Error, fmt.Sprintf("max retries exceeded with status %d", resp.StatusCode))
			}
			return resp, err
		}

		// Log the retry attempt
		ctx := req.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		delay := c.calculateBackoff(attempt)

		var reason string
		if err != nil {
			reason = fmt.Sprintf("network error: %v", err)
		} else if resp != nil {
			reason = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}

		// Record retry as span event
		span.AddEvent("retry",
			trace.WithAttributes(
				attribute.Int("retry.attempt", attempt+1),
				attribute.Int("retry.max_attempts", c.maxRetries),
				attribute.String("retry.reason", reason),
				attribute.String("retry.delay", delay.Round(time.Millisecond).String()),
			),
		)

		log.Debugf(ctx, "Retrying HTTP request (attempt %d/%d) after %v due to %s: %s %s",
			attempt+1, c.maxRetries, delay.Round(time.Millisecond), reason, req.Method, req.URL.String())

		// Close response body before retry to avoid resource leak
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		// Wait for the backoff period or until context is cancelled
		timer := time.NewTimer(delay)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
			// Continue to next attempt
		}
	}

	return resp, err
}

// calculateBackoff computes the delay before the next retry using
// exponential backoff with jitter: min(maxDelay, baseDelay * 2^attempt * (0.5-1.5))
func (c *RetryableClient) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff: baseDelay * 2^attempt
	backoff := float64(c.baseDelay) * math.Pow(2, float64(attempt))

	// Apply jitter: multiply by random factor between 0.5 and 1.5
	jitter := 0.5 + rand.Float64()
	backoff = backoff * jitter

	// Cap at maxDelay
	if backoff > float64(c.maxDelay) {
		backoff = float64(c.maxDelay)
	}

	return time.Duration(backoff)
}
