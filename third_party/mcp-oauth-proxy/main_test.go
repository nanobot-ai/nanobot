package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/proxy"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationFlow(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Create OAuth proxy
	config := &types.Config{
		Mode:              proxy.ModeProxy,
		MCPServerURL:      "http://localhost:8081/",
		OAuthClientID:     "test_client_id",
		OAuthClientSecret: "test_client_secret",
		OAuthAuthorizeURL: "https://accounts.google.com",
		ScopesSupported:   "openid,profile,email",
	}
	_, err := proxy.NewOAuthProxy(config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	oauthProxy, err := proxy.NewOAuthProxy(config)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
	}
	defer func() {
		if err := oauthProxy.Close(); err != nil {
			t.Logf("Error closing OAuth proxy: %v", err)
		}
	}()

	// Get HTTP handler
	handler := oauthProxy.GetHandler()

	// Test health endpoint
	t.Run("HealthEndpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/health", nil)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")
	})

	// Test OAuth metadata endpoints
	t.Run("OAuthMetadata", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/.well-known/oauth-authorization-server", nil)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "authorization_endpoint")
		assert.Contains(t, w.Body.String(), "token_endpoint")
	})

	// Test protected resource metadata
	t.Run("ProtectedResourceMetadata", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/.well-known/oauth-protected-resource", nil)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "resource")
		assert.Contains(t, w.Body.String(), "authorization_servers")
	})

	// Test MCP endpoint requires authorization
	t.Run("MCPEndpointRequiresAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/mcp", nil)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Bearer")
	})

	// Test protected resource metadata with path parameter
	t.Run("ProtectedResourceMetadataWithPath", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/.well-known/oauth-protected-resource/test/path", nil)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "resource")
		assert.Contains(t, w.Body.String(), "authorization_servers")
	})

	// Test authorization endpoint redirects (basic validation)
	t.Run("AuthorizationEndpointRedirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/authorize?response_type=code&client_id=test&redirect_uri=http://localhost:8080/callback&scope=openid", nil)
		handler.ServeHTTP(w, req)

		// Should get a redirect to the OAuth provider or an error about invalid client
		assert.True(t, w.Code == http.StatusFound || w.Code == http.StatusBadRequest)
	})
}

func TestOAuthProxyCreation(t *testing.T) {
	// Create OAuth proxy
	config := &types.Config{
		Mode:              proxy.ModeProxy,
		MCPServerURL:      "http://localhost:8081/",
		OAuthClientID:     "test_client_id",
		OAuthClientSecret: "test_client_secret",
		OAuthAuthorizeURL: "https://accounts.google.com",
		ScopesSupported:   "openid,profile,email",
	}
	_, err := proxy.NewOAuthProxy(config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	oauthProxy, err := proxy.NewOAuthProxy(config)
	require.NoError(t, err, "Should be able to create OAuth proxy with valid environment")
	require.NotNil(t, oauthProxy, "OAuth proxy should not be nil")

	// Test that we can get handler without error
	require.NotPanics(t, func() {
		handler := oauthProxy.GetHandler()
		require.NotNil(t, handler)
	}, "GetHandler should not panic and should return non-nil handler")

	// Clean up
	err = oauthProxy.Close()
	assert.NoError(t, err, "Should be able to close OAuth proxy")
}

func TestOAuthProxyStart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OAuth proxy start test in short mode")
	}

	// Create OAuth proxy
	config := &types.Config{
		Mode:              proxy.ModeProxy,
		MCPServerURL:      "http://localhost:8081/",
		OAuthClientID:     "test_client_id",
		OAuthClientSecret: "test_client_secret",
		OAuthAuthorizeURL: "https://accounts.google.com",
		ScopesSupported:   "openid,profile,email",
	}
	_, err := proxy.NewOAuthProxy(config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	oauthProxy, err := proxy.NewOAuthProxy(config)
	require.NoError(t, err)
	defer func() {
		_ = oauthProxy.Close()
	}()

	// Test starting and stopping the proxy
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the proxy in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- oauthProxy.Start(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to stop the server
	cancel()

	// Wait for the server to stop
	select {
	case err := <-errCh:
		// Server stopped, which is expected when context is cancelled
		if err != nil {
			assert.Contains(t, err.Error(), "context canceled", "Expected context cancellation error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

func TestForwardAuthIntegrationFlow(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	config := &types.Config{
		Mode:              proxy.ModeForwardAuth,
		MCPServerURL:      "http://localhost:8081/",
		OAuthClientID:     "test_client_id",
		OAuthClientSecret: "test_client_secret",
		OAuthAuthorizeURL: "https://accounts.google.com",
		ScopesSupported:   "openid,profile,email",
		Port:              "8082",
	}
	_, err := proxy.NewOAuthProxy(config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	require.Equal(t, "forward_auth", config.Mode)
	require.Equal(t, "8082", config.Port)

	oauthProxy, err := proxy.NewOAuthProxy(config)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
	}
	defer func() {
		if err := oauthProxy.Close(); err != nil {
			t.Logf("Error closing OAuth proxy: %v", err)
		}
	}()

	// Get HTTP handler
	handler := oauthProxy.GetHandler()

	// Test health endpoint works in forward auth mode
	t.Run("ForwardAuthHealthEndpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/health", nil)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")
	})

	// Test OAuth metadata endpoints work in forward auth mode
	t.Run("ForwardAuthOAuthMetadata", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/.well-known/oauth-authorization-server", nil)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "authorization_endpoint")
		assert.Contains(t, w.Body.String(), "token_endpoint")
	})

	// Test protected resource metadata with path works in forward auth mode
	t.Run("ForwardAuthProtectedResourceMetadataWithPath", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/.well-known/oauth-protected-resource/api/v1", nil)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "resource")
		assert.Contains(t, w.Body.String(), "authorization_servers")
	})

	// Test that forward auth mode requires authorization for protected endpoints
	t.Run("ForwardAuthRequiresAuth", func(t *testing.T) {
		testPaths := []string{"/api", "/data", "/protected", "/mcp", "/test"}

		for _, path := range testPaths {
			t.Run("Path_"+path, func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", path, nil)
				handler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
				assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Bearer")
			})
		}
	})

	// Test that forward auth mode doesn't proxy to MCP server (no proxying behavior)
	t.Run("ForwardAuthNoProxying", func(t *testing.T) {
		// In forward auth mode, there should be no attempt to proxy to an MCP server
		// Instead, the proxy should validate tokens and set headers for downstream services
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		handler.ServeHTTP(w, req)

		// Should get unauthorized (no proxying attempt)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Bearer")

		// Should not have any proxy-related error messages
		assert.NotContains(t, w.Body.String(), "proxy")
		assert.NotContains(t, w.Body.String(), "502")
		assert.NotContains(t, w.Body.String(), "Bad Gateway")
	})

	// Test authorization endpoint redirects work in forward auth mode
	t.Run("ForwardAuthAuthorizationEndpointRedirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/authorize?response_type=code&client_id=test&redirect_uri=http://localhost:8082/callback&scope=openid", nil)
		handler.ServeHTTP(w, req)

		// Should get a redirect to the OAuth provider or an error about invalid client
		assert.True(t, w.Code == http.StatusFound || w.Code == http.StatusBadRequest)
	})

	// Test CORS headers work in forward auth mode
	t.Run("ForwardAuthCORSHeaders", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("OPTIONS", "/api", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		handler.ServeHTTP(w, req)

		// Should handle CORS preflight
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
	})
}
