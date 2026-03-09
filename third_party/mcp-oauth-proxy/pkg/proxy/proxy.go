package proxy

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"maps"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/db"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/encryption"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/handlerutils"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/authorize"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/callback"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/register"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/revoke"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/success"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/token"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/validate"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/providers"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/ratelimit"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/tokens"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"golang.org/x/oauth2"
)

type OAuthProxy struct {
	metadata      *types.OAuthMetadata
	db            *db.Store
	rateLimiter   *ratelimit.RateLimiter
	providers     *providers.Manager
	tokenManager  *tokens.TokenManager
	provider      string
	encryptionKey []byte
	resourceName  string
	config        *types.Config

	ctx    context.Context
	cancel context.CancelFunc
}

const (
	ModeProxy       = "proxy"
	ModeForwardAuth = "forward_auth"
	ModeMiddleware  = "middleware"
)

func NewOAuthProxy(config *types.Config) (*OAuthProxy, error) {
	databaseDSN := config.DatabaseDSN

	// Log database configuration
	if databaseDSN == "" {
		log.Println("DATABASE_DSN not set, using SQLite database at data/oauth_proxy.db")
	} else if strings.HasPrefix(databaseDSN, "postgres://") || strings.HasPrefix(databaseDSN, "postgresql://") {
		log.Println("Using PostgreSQL database")
	} else {
		log.Printf("Using SQLite database at: %s", databaseDSN)
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	switch config.Mode {
	case "":
		fmt.Println("Defaulting to proxy mode")
		config.Mode = ModeProxy
	case ModeProxy, ModeForwardAuth, ModeMiddleware:
	default:
		return nil, fmt.Errorf("invalid mode: %s", config.Mode)
	}

	if config.Mode == ModeProxy {
		if u, err := url.Parse(config.MCPServerURL); err != nil || u.Scheme != "http" && u.Scheme != "https" {
			return nil, fmt.Errorf("invalid MCP server URL: %w", err)
		} else if u.Path != "" && u.Path != "/" || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("MCP server URL must not contain a path, query, or fragment")
		}
	}

	// Initialize database
	db, err := db.New(databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize rate limiter
	rateLimiter := ratelimit.NewRateLimiter(
		time.Duration(15)*time.Minute,
		5000,
	)

	// Initialize provider manager
	providerManager := providers.NewManager()
	provider := ""

	// Register generic provider
	if config.OAuthClientID != "" && config.OAuthClientSecret != "" && config.OAuthAuthorizeURL != "" {
		genericProvider := providers.NewGenericProvider(config.OAuthAuthorizeURL)
		providerManager.RegisterProvider("generic", genericProvider)
		provider = "generic"
	}

	// Initialize token manager with JWKS and API key auth support
	tokenManager, err := tokens.NewTokenManagerWithJWKSURLAndAPIKeyAuth(db, config.OAuthJWKSURL, config.APIKeyAuthWebhookURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize token manager: %w", err)
	}

	encryptionKey, err := base64.StdEncoding.DecodeString(config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	// Split and trim scopes to handle whitespace
	scopesSupported := ParseScopesSupported(config.ScopesSupported)

	metadata := &types.OAuthMetadata{
		ResponseTypesSupported:                   []string{"code"},
		CodeChallengeMethodsSupported:            []string{"S256"},
		TokenEndpointAuthMethodsSupported:        []string{"client_secret_post", "none"},
		GrantTypesSupported:                      []string{"authorization_code", "refresh_token"},
		ScopesSupported:                          scopesSupported,
		RevocationEndpointAuthMethodsSupported:   []string{"client_secret_post", "none"},
		RegistrationEndpointAuthMethodsSupported: []string{"client_secret_post"},
	}

	return &OAuthProxy{
		metadata:      metadata,
		db:            db,
		rateLimiter:   rateLimiter,
		providers:     providerManager,
		tokenManager:  tokenManager,
		provider:      provider,
		resourceName:  "MCP Tools",
		encryptionKey: encryptionKey,
		config:        config,
	}, nil
}

// GetOAuthClientID returns the OAuth client ID from config
func (p *OAuthProxy) GetOAuthClientID() string {
	return p.config.OAuthClientID
}

// GetOAuthClientSecret returns the OAuth client secret from config
func (p *OAuthProxy) GetOAuthClientSecret() string {
	return p.config.OAuthClientSecret
}

// GetOAuthAuthorizeURL returns the OAuth authorization URL from config
func (p *OAuthProxy) GetOAuthAuthorizeURL() string {
	return p.config.OAuthAuthorizeURL
}

// GetMCPServerURL returns the MCP server URL from config
func (p *OAuthProxy) GetMCPServerURL() string {
	return p.config.MCPServerURL
}

func (p *OAuthProxy) Close() error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *OAuthProxy) Start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)

	// Setup cleanup goroutine for expired tokens
	go func() {
		ticker := time.NewTicker(time.Hour) // Cleanup every hour
		defer ticker.Stop()
		context.AfterFunc(p.ctx, ticker.Stop)
		for range ticker.C {
			if err := p.db.CleanupExpiredTokens(); err != nil {
				log.Printf("Failed to cleanup expired tokens: %v", err)
			}
		}
	}()

	return nil
}

func (p *OAuthProxy) SetupRoutes(mux *http.ServeMux, next http.Handler) {
	provider, err := p.providers.GetProvider(p.provider)
	if err != nil {
		log.Fatalf("Failed to get provider: %v", err)
	}

	if p.config.CookieNamePrefix == "" {
		p.config.CookieNamePrefix = "mcp_oauth_proxy_"
	} else if !strings.HasSuffix(p.config.CookieNamePrefix, "_") && !strings.HasSuffix(p.config.CookieNamePrefix, "-") {
		p.config.CookieNamePrefix += "_"
	}

	authorizeHandler := authorize.NewHandler(p.db, provider, p.metadata.ScopesSupported, p.GetOAuthClientID(), p.GetOAuthClientSecret(), p.config.RoutePrefix)
	tokenHandler := token.NewHandler(p.db)
	callbackHandler := callback.NewHandler(p.db, provider, p.encryptionKey, p.GetOAuthClientID(), p.GetOAuthClientSecret(), p.config.RoutePrefix, p.config.CookieNamePrefix)
	revokeHandler := revoke.NewHandler(p.db)
	tokenValidator := validate.NewTokenValidator(p.tokenManager, p.encryptionKey, p.db, provider, p.config.RoutePrefix, p.GetOAuthClientID(), p.GetOAuthClientSecret(), p.config.CookieNamePrefix, p.config.MCPServerID, p.metadata.ScopesSupported, p.config.MCPPaths)
	successHandler := success.NewHandler()

	// Get route prefix from config
	prefix := p.config.RoutePrefix

	mux.HandleFunc("GET "+prefix+"/health", p.withCORS(p.healthHandler))

	// OAuth endpoints
	mux.HandleFunc("GET "+prefix+"/authorize", p.withCORS(p.withRateLimit(authorizeHandler)))
	mux.HandleFunc("GET "+prefix+"/callback", p.withCORS(p.withRateLimit(callbackHandler)))
	mux.HandleFunc("POST "+prefix+"/token", p.withCORS(p.withRateLimit(tokenHandler)))
	mux.HandleFunc("POST "+prefix+"/revoke", p.withCORS(p.withRateLimit(revokeHandler)))
	mux.HandleFunc("POST "+prefix+"/register", p.withCORS(p.withRateLimit(register.NewHandler(p.db))))

	// Metadata endpoints
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", p.withCORS(p.oauthMetadataHandler))
	mux.HandleFunc("GET /.well-known/oauth-protected-resource", p.withCORS(p.protectedResourceMetadataHandler))
	mux.HandleFunc("GET /.well-known/oauth-protected-resource/{path...}", p.withCORS(p.protectedResourceMetadataHandler))

	mux.HandleFunc("GET "+prefix+"/auth/mcp-ui/success", p.withCORS(p.withRateLimit(successHandler)))

	// Protect everything else
	mux.HandleFunc(prefix+"/{path...}", p.withCORS(p.withRateLimit(tokenValidator.WithTokenValidation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.mcpProxyHandler(w, r, next)
	})))))
}

// GetHandler returns an http.Handler for the OAuth proxy
func (p *OAuthProxy) GetHandler() http.Handler {
	mux := http.NewServeMux()
	p.SetupRoutes(mux, nil)

	// Wrap with logging middleware
	loggedHandler := handlers.LoggingHandler(os.Stdout, mux)

	return loggedHandler
}

// withCORS wraps a handler with CORS headers
func (p *OAuthProxy) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, mcp-protocol-version")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(int((12 * time.Hour).Seconds())))

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// withRateLimit wraps a handler with rate limiting
func (p *OAuthProxy) withRateLimit(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p.rateLimiter != nil {
			clientIP := handlerutils.GetClientIP(r)
			if !p.rateLimiter.Allow(clientIP) {
				handlerutils.JSON(w, http.StatusTooManyRequests, types.OAuthError{
					Error:            "too_many_requests",
					ErrorDescription: "Rate limit exceeded",
				})
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}

func (p *OAuthProxy) healthHandler(w http.ResponseWriter, r *http.Request) {
	handlerutils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (p *OAuthProxy) oauthMetadataHandler(w http.ResponseWriter, r *http.Request) {
	baseURL := handlerutils.GetBaseURL(r)
	prefix := p.config.RoutePrefix

	// Create dynamic metadata based on the request
	metadata := &types.OAuthMetadata{
		Issuer:                                   baseURL,
		ServiceDocumentation:                     p.metadata.ServiceDocumentation,
		AuthorizationEndpoint:                    fmt.Sprintf("%s%s/authorize", baseURL, prefix),
		ResponseTypesSupported:                   p.metadata.ResponseTypesSupported,
		CodeChallengeMethodsSupported:            p.metadata.CodeChallengeMethodsSupported,
		TokenEndpoint:                            fmt.Sprintf("%s%s/token", baseURL, prefix),
		TokenEndpointAuthMethodsSupported:        p.metadata.TokenEndpointAuthMethodsSupported,
		GrantTypesSupported:                      p.metadata.GrantTypesSupported,
		ScopesSupported:                          p.metadata.ScopesSupported,
		RevocationEndpoint:                       fmt.Sprintf("%s%s/revoke", baseURL, prefix),
		RevocationEndpointAuthMethodsSupported:   p.metadata.RevocationEndpointAuthMethodsSupported,
		RegistrationEndpoint:                     fmt.Sprintf("%s%s/register", baseURL, prefix),
		RegistrationEndpointAuthMethodsSupported: p.metadata.RegistrationEndpointAuthMethodsSupported,
	}

	handlerutils.JSON(w, http.StatusOK, metadata)
}

func (p *OAuthProxy) protectedResourceMetadataHandler(w http.ResponseWriter, r *http.Request) {
	baseURL := handlerutils.GetBaseURL(r)
	prefix := p.config.RoutePrefix
	resourceURL := strings.TrimSuffix(baseURL+prefix, "/")

	metadata := types.OAuthProtectedResourceMetadata{
		Resource:              resourceURL,
		AuthorizationServers:  []string{baseURL + prefix},
		Scopes:                p.metadata.ScopesSupported,
		ResourceName:          p.resourceName,
		ResourceDocumentation: p.metadata.ServiceDocumentation,
	}

	handlerutils.JSON(w, http.StatusOK, metadata)
}

func (p *OAuthProxy) mcpProxyHandler(w http.ResponseWriter, r *http.Request, next http.Handler) {
	tokenInfo := validate.GetTokenInfo(r)
	path := r.PathValue("path")

	// Check if the access token is expired and refresh if needed
	if tokenInfo != nil && tokenInfo.Props != nil {
		if _, ok := tokenInfo.Props["access_token"].(string); ok {
			// Check if token is expired (with a 5-minute buffer)
			expiresAt, ok := tokenInfo.Props["expires_at"].(float64)
			if ok && expiresAt > 0 {
				if time.Now().Add(5 * time.Minute).After(time.Unix(int64(expiresAt), 0)) {
					log.Printf("Access token is expired or will expire soon, attempting to refresh")

					// Get the refresh token
					refreshToken, ok := tokenInfo.Props["refresh_token"].(string)
					if !ok || refreshToken == "" {
						log.Printf("No refresh token available, cannot refresh access token")
						handlerutils.JSON(w, http.StatusUnauthorized, map[string]string{
							"error":             "invalid_token",
							"error_description": "Access token expired and no refresh token available",
						})
						return
					}

					// Get the provider
					provider, err := p.providers.GetProvider(p.provider)
					if err != nil {
						log.Printf("Failed to get provider for token refresh: %v", err)
						handlerutils.JSON(w, http.StatusInternalServerError, map[string]string{
							"error":             "server_error",
							"error_description": "Failed to refresh token",
						})
						return
					}

					// Get provider credentials
					clientID := p.GetOAuthClientID()
					clientSecret := p.GetOAuthClientSecret()
					if clientID == "" || clientSecret == "" {
						log.Printf("OAuth credentials not configured for token refresh")
						handlerutils.JSON(w, http.StatusInternalServerError, map[string]string{
							"error":             "server_error",
							"error_description": "OAuth credentials not configured",
						})
						return
					}

					// Refresh the token
					newTokenInfo, err := provider.RefreshToken(r.Context(), refreshToken, clientID, clientSecret)
					if err != nil {
						log.Printf("Failed to refresh token: %v", err)
						handlerutils.JSON(w, http.StatusUnauthorized, map[string]string{
							"error":             "invalid_token",
							"error_description": "Failed to refresh access token",
						})
						return
					}

					// Update the grant with new token information
					if err := p.updateGrant(tokenInfo.GrantID, tokenInfo.UserID, tokenInfo, newTokenInfo); err != nil {
						log.Printf("Failed to update grant: %v", err)
						handlerutils.JSON(w, http.StatusInternalServerError, map[string]string{
							"error":             "server_error",
							"error_description": "Failed to update grant with new token",
						})
						return
					}

					// Update the token info with the new access token for the current request
					tokenInfo.Props["access_token"] = newTokenInfo.AccessToken
					log.Printf("Successfully refreshed access token")
				}
			}
		}
	}

	switch p.config.Mode {
	case ModeMiddleware:
		next.ServeHTTP(w, r)
	case ModeForwardAuth:
		setHeaders(w.Header(), tokenInfo.Props)
	case ModeProxy:
		// Create target URL
		targetURL := p.GetMCPServerURL() + "/" + path
		// Log the proxy request for debugging
		log.Printf("Proxying request: %s %s -> %s", r.Method, r.URL.Path, targetURL)

		// Create reverse proxy
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.Header.Del("Authorization")
				req.Header.Set("X-Forwarded-Host", req.Host)
				req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)

				newURL, _ := url.Parse(targetURL)
				req.URL.Scheme = newURL.Scheme
				req.URL.Host = newURL.Host
				req.Host = newURL.Host

				// Add forwarded headers from token props
				setHeaders(req.Header, tokenInfo.Props)
			},
			ModifyResponse: func(resp *http.Response) error {
				// Rewrite Location header to use proxy host instead of downstream server host
				if location := resp.Header.Get("Location"); location != "" {
					if locationURL, err := url.Parse(location); err == nil {
						// Get the original request to extract proxy host
						proxyHost := resp.Request.Header.Get("X-Forwarded-Host")
						if proxyHost != "" {
							// Parse downstream server URL to get scheme
							downstreamURL, _ := url.Parse(p.GetMCPServerURL())

							// Only rewrite if the location points to the downstream server
							if locationURL.Host == downstreamURL.Host {
								locationURL.Scheme = resp.Request.URL.Scheme
								locationURL.Host = proxyHost
								resp.Header.Set("Location", locationURL.String())
							}
						}
					}
				}
				return nil
			},
			ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
				log.Printf("Proxy error: %v", err)
				rw.WriteHeader(http.StatusBadGateway)
			},
		}

		// Serve the proxied request
		proxy.ServeHTTP(w, r)
	}
}

func setHeaders(header http.Header, props map[string]any) {
	if userID, ok := props["user_id"].(string); ok {
		header.Set("X-Forwarded-User", userID)
	} else {
		header.Del("X-Forwarded-User")
	}
	if email, ok := props["email"].(string); ok {
		header.Set("X-Forwarded-Email", email)
	} else {
		header.Del("X-Forwarded-Email")
	}
	if name, ok := props["name"].(string); ok {
		header.Set("X-Forwarded-Name", name)
	} else {
		header.Del("X-Forwarded-Name")
	}
	if accessToken, ok := props["access_token"].(string); ok {
		header.Set("X-Forwarded-Access-Token", accessToken)
	} else {
		header.Del("X-Forwarded-Access-Token")
	}
}

// updateGrant updates a grant with new token information
func (p *OAuthProxy) updateGrant(grantID, userID string, oldTokenInfo *tokens.TokenInfo, newTokenInfo *oauth2.Token) error {
	// Get the existing grant
	grant, err := p.db.GetGrant(grantID, userID)
	if err != nil {
		return fmt.Errorf("failed to get grant: %w", err)
	}

	sensitiveProps := map[string]any{}
	if oldTokenInfo.Props != nil {
		// keep all the old props, that include a lot of the user info
		maps.Copy(sensitiveProps, oldTokenInfo.Props)
	}

	// Prepare sensitive props data
	sensitiveProps["access_token"] = newTokenInfo.AccessToken
	sensitiveProps["refresh_token"] = newTokenInfo.RefreshToken
	sensitiveProps["expires_at"] = newTokenInfo.Expiry.Unix()

	// Add existing user info if available
	if grant.Props != nil {
		if email, ok := grant.Props["email"].(string); ok {
			sensitiveProps["email"] = email
		}
		if name, ok := grant.Props["name"].(string); ok {
			sensitiveProps["name"] = name
		}
		if userID, ok := grant.Props["user_id"].(string); ok {
			sensitiveProps["user_id"] = userID
		}
	}

	// use old refresh token in case new one is not provided
	if sensitiveProps["refresh_token"] == "" {
		sensitiveProps["refresh_token"] = oldTokenInfo.Props["refresh_token"]
	}

	// Initialize props map
	props := make(map[string]any)

	// Encrypt the sensitive props data
	encryptedProps, err := encryption.EncryptData(sensitiveProps, p.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt props data: %w", err)
	}

	// Store encrypted data
	props["encrypted_data"] = encryptedProps.Data
	props["iv"] = encryptedProps.IV
	props["algorithm"] = encryptedProps.Algorithm
	props["encrypted"] = true

	// Update the grant with new props
	grant.Props = props

	// Update the grant in the database
	if err := p.db.UpdateGrant(grant); err != nil {
		return fmt.Errorf("failed to update grant: %w", err)
	}

	return nil
}

// ParseScopesSupported parses a comma-separated scopes string and trims whitespace from each scope.
func ParseScopesSupported(envScopes string) []string {
	scopesRaw := strings.Split(envScopes, ",")
	scopesSupported := make([]string, 0, len(scopesRaw))
	for _, scope := range scopesRaw {
		if trimmed := strings.TrimSpace(scope); trimmed != "" {
			scopesSupported = append(scopesSupported, trimmed)
		}
	}
	return scopesSupported
}
