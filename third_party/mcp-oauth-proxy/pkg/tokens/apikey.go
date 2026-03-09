package tokens

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tokenTracer = otel.Tracer("mcp-oauth-proxy/tokens")

// APIKeyValidator validates API keys by calling the authentication webhook.
type APIKeyValidator struct {
	authURL    string
	httpClient *http.Client
}

// NewAPIKeyValidator creates a new API key validator.
func NewAPIKeyValidator(authURL string) *APIKeyValidator {
	return &APIKeyValidator{
		authURL: authURL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

// ValidateAPIKey validates an API key by calling the authentication webhook.
// Returns TokenInfo on success, or an error if authentication/authorization fails.
func (v *APIKeyValidator) ValidateAPIKey(ctx context.Context, apiKey, mcpID string) (*TokenInfo, error) {
	ctx, span := tokenTracer.Start(ctx, "auth.api_key.validate", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	span.SetAttributes(
		attribute.String("auth.api_key.url", v.authURL),
		attribute.String("auth.mcp_id", mcpID),
	)

	if v.authURL == "" {
		err := fmt.Errorf("API key auth URL not configured")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	reqBody := types.APIKeyAuthRequest{
		MCPID: mcpID,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		err = fmt.Errorf("failed to marshal request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.authURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to call auth webhook: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	var authResp types.APIKeyAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		err = fmt.Errorf("failed to decode response: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if !authResp.Allowed {
		errMsg := "authentication failed"
		if authResp.Reason != "" {
			errMsg = authResp.Reason
		}
		err := fmt.Errorf("%s", errMsg)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	userInfo, _ := json.Marshal(map[string]any{
		"sub":      authResp.Subject,
		"username": authResp.PreferredUsername,
	})

	span.SetStatus(codes.Ok, "")
	return &TokenInfo{
		UserID: authResp.Subject,
		Props: map[string]any{
			"info":         string(userInfo),
			"api_key":      true,
			"access_token": apiKey,
		},
	}, nil
}
