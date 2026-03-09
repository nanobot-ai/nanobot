package validate

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/encryption"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/handlerutils"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/providers"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/tokens"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var validateTracer = otel.Tracer("mcp-oauth-proxy/validate")

type TokenValidator struct {
	tokenManager           *tokens.TokenManager
	encryptionKey          []byte
	db                     TokenStore         // Database for refresh operations
	provider               providers.Provider // OAuth provider for generating auth URLs
	routePrefix            string
	clientID               string // OAuth client ID
	clientSecret           string // OAuth client secret
	accessTokenCookieName  string
	refreshTokenCookieName string
	mcpServerID            string
	scopesSupported        []string // Supported OAuth scopes
	mcpPaths               []string
}

// TokenStore interface for database operations needed by validator
type TokenStore interface {
	GetToken(accessToken string) (*types.TokenData, error)
	GetTokenByRefreshToken(refreshToken string) (*types.TokenData, error)
	StoreToken(token *types.TokenData) error
	RevokeToken(token string) error
	StoreAuthRequest(key string, data map[string]any) error
}

func NewTokenValidator(tokenManager *tokens.TokenManager, encryptionKey []byte, db TokenStore, provider providers.Provider, routePrefix, clientID, clientSecret, cookieNamePrefix, mcpServerID string, scopesSupported, mcpPaths []string) *TokenValidator {
	return &TokenValidator{
		tokenManager:           tokenManager,
		encryptionKey:          encryptionKey,
		db:                     db,
		provider:               provider,
		routePrefix:            routePrefix,
		clientID:               clientID,
		clientSecret:           clientSecret,
		accessTokenCookieName:  cookieNamePrefix + types.AccessTokenCookieName,
		refreshTokenCookieName: cookieNamePrefix + types.RefreshTokenCookieName,
		mcpServerID:            mcpServerID,
		scopesSupported:        scopesSupported,
		mcpPaths:               mcpPaths,
	}
}

func (p *TokenValidator) SetOAuthConfig(provider providers.Provider, clientID, clientSecret string, scopesSupported []string) {
	p.provider = provider
	p.clientID = clientID
	p.clientSecret = clientSecret
	p.scopesSupported = scopesSupported
}

// generatePKCE generates PKCE code verifier and challenge
func generatePKCE() (codeVerifier, codeChallenge string) {
	// Generate a random code verifier (43-128 characters, base64url)
	codeVerifier = encryption.GenerateRandomString(32) // This generates 32 bytes -> ~43 chars in base64

	// Generate code challenge using S256 method
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return codeVerifier, codeChallenge
}

func (p *TokenValidator) WithTokenValidation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := validateTracer.Start(r.Context(), "auth.token_validation", trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()
		r = r.WithContext(ctx)

		var (
			token      string
			fromCookie bool
			authHeader = r.Header.Get("Authorization")
		)
		if authHeader == "" {
			cookie, err := r.Cookie(p.accessTokenCookieName)
			if err == nil && cookie.Value != "" {
				// Decrypt the cookie value
				token, err = encryption.DecryptCookie(p.encryptionKey, cookie.Value)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					p.sendUnauthorizedResponse(w, r, "Failed to decrypt token")
					return
				}
				fromCookie = true
			}
		} else {
			// Parse Authorization header
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || parts[1] == "" {
				span.SetStatus(codes.Error, "invalid authorization header")
				p.sendUnauthorizedResponse(w, r, "Invalid Authorization header format, expected 'Bearer TOKEN'")
				return
			}

			token = parts[1]
			fromCookie = false
		}
		span.SetAttributes(attribute.Bool("auth.token.from_cookie", fromCookie))

		// Validate token (handles both API keys and JWT/simple tokens)
		tokenInfo, err := p.tokenManager.GetTokenInfoWithContext(r.Context(), token, p.mcpServerID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			p.sendUnauthorizedResponse(w, r, fmt.Sprintf("Invalid or expired token: %v", err))
			return
		}

		// Check if token is within 15 minutes of expiring and attempt refresh if from cookie
		if fromCookie && time.Until(tokenInfo.ExpiresAt) < 15*time.Minute {
			newToken, refreshErr := p.refreshAccessToken(w, r)
			if refreshErr != nil {
				// If token is already expired, refresh is required
				if time.Now().After(tokenInfo.ExpiresAt) {
					span.RecordError(refreshErr)
					span.SetStatus(codes.Error, refreshErr.Error())
					p.sendUnauthorizedResponse(w, r, "Token expired and refresh failed")
					return
				}
				// Token not yet expired, log warning and continue with current token
				fmt.Printf("Token refresh failed but token still valid, continuing: %v\n", refreshErr)
			} else {
				// Refresh succeeded, use new token and get updated token info
				token = newToken
				tokenInfo, err = p.tokenManager.GetTokenInfoWithContext(r.Context(), token, p.mcpServerID)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					p.sendUnauthorizedResponse(w, r, "Failed to validate refreshed token")
					return
				}
			}
		}

		// Decrypt props if needed before storing in context
		if tokenInfo.Props != nil {
			decryptedProps, err := encryption.DecryptPropsIfNeeded(p.encryptionKey, tokenInfo.Props)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				p.sendUnauthorizedResponse(w, r, "Failed to decrypt token data")
				return
			}
			tokenInfo.Props = decryptedProps
		}

		// Store both tokenInfo and the original bearer token string
		ctx = context.WithValue(r.Context(), tokenInfoKey{}, tokenInfo)
		ctx = context.WithValue(ctx, bearerTokenKey{}, token)
		span.SetStatus(codes.Ok, "")
		next(w, r.WithContext(ctx))
	}
}

// refreshAccessToken attempts to refresh an access token using the refresh token
// Returns the new access token string, or an error if refresh fails
func (p *TokenValidator) refreshAccessToken(w http.ResponseWriter, r *http.Request) (string, error) {
	// Get refresh token from cookie
	refreshCookie, err := r.Cookie(p.refreshTokenCookieName)
	if err != nil || refreshCookie.Value == "" {
		return "", fmt.Errorf("no refresh token cookie found")
	}

	// Decrypt refresh token
	refreshToken, err := encryption.DecryptCookie(p.encryptionKey, refreshCookie.Value)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	// Validate refresh token and get token data
	tokenData, err := p.db.GetTokenByRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if token is revoked
	if tokenData.Revoked {
		return "", fmt.Errorf("refresh token has been revoked")
	}

	// Generate new access token: userId:grantId:accessTokenSecret
	accessTokenSecret := encryption.GenerateRandomString(32)
	newAccessToken := fmt.Sprintf("%s:%s:%s", tokenData.UserID, tokenData.GrantID, accessTokenSecret)

	// Generate new refresh token
	refreshTokenSecret := encryption.GenerateRandomString(32)
	newRefreshToken := fmt.Sprintf("%s:%s:%s", tokenData.UserID, tokenData.GrantID, refreshTokenSecret)

	// Store new token data in database
	newTokenData := &types.TokenData{
		AccessToken:           newAccessToken,
		RefreshToken:          newRefreshToken,
		ClientID:              tokenData.ClientID,
		UserID:                tokenData.UserID,
		GrantID:               tokenData.GrantID,
		Scope:                 tokenData.Scope,
		ExpiresAt:             time.Now().Add(time.Hour),           // 1 hour
		RefreshTokenExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt:             time.Now(),
		Revoked:               false,
	}

	if err := p.db.StoreToken(newTokenData); err != nil {
		return "", fmt.Errorf("failed to store new token: %w", err)
	}

	// Determine if request is secure for cookie Secure flag
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	// Encrypt and set access token cookie (1 hour = 3600 seconds)
	encryptedAccessToken, err := encryption.EncryptCookie(p.encryptionKey, newAccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt access token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     p.accessTokenCookieName,
		Value:    encryptedAccessToken,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Encrypt and set refresh token cookie (30 days = 2592000 seconds)
	encryptedRefreshToken, err := encryption.EncryptCookie(p.encryptionKey, newRefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     p.refreshTokenCookieName,
		Value:    encryptedRefreshToken,
		Path:     "/",
		MaxAge:   30 * 24 * 3600,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	return newAccessToken, nil
}

func (p *TokenValidator) handleOauthFlow(w http.ResponseWriter, r *http.Request) {
	// Check if OAuth provider is configured
	if p.provider == nil || p.clientID == "" || p.clientSecret == "" {
		handlerutils.JSON(w, http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "OAuth provider not configured",
		})
		return
	}

	// Generate PKCE parameters
	codeVerifier, codeChallenge := generatePKCE()

	// Generate a random state key for this auth request
	stateKey := encryption.GenerateRandomString(32)

	// Get current path for post-auth redirect
	currentPath := r.URL.Path
	if r.URL.RawQuery != "" {
		currentPath += "?" + r.URL.RawQuery
	}

	// Store the auth request data in the database with PKCE parameters
	authData := map[string]any{
		"response_type":         "code",
		"client_id":             p.clientID,
		"redirect_uri":          fmt.Sprintf("%s/callback", handlerutils.GetBaseURL(r)),
		"scope":                 strings.Join(p.scopesSupported, " "),
		"state":                 stateKey,
		"code_challenge":        codeChallenge,
		"code_challenge_method": "S256",
		"code_verifier":         codeVerifier, // Store for later use in token exchange
		"rd":                    currentPath,  // Store original path for post-auth redirect
	}

	if err := p.db.StoreAuthRequest(stateKey, authData); err != nil {
		handlerutils.JSON(w, http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to store authorization request",
		})
		return
	}

	// Build the authorization URL directly to OAuth provider (not our /authorize)
	redirectURI := fmt.Sprintf("%s%s/callback", handlerutils.GetBaseURL(r), p.routePrefix)
	scope := strings.Join(p.scopesSupported, " ")

	// Generate authorization URL with PKCE
	authURL := p.provider.GetAuthorizationURL(p.clientID, redirectURI, scope, stateKey)

	w.Header().Set("X-Redirect-URL", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (p *TokenValidator) sendUnauthorizedResponse(w http.ResponseWriter, r *http.Request, message string) {
	if !slices.Contains(p.mcpPaths, r.URL.Path) && strings.Contains(strings.ToLower(r.UserAgent()), "mozilla") {
		p.handleOauthFlow(w, r)
		return
	}

	w.Header().Set(
		"WWW-Authenticate",
		fmt.Sprintf(
			`Bearer error="invalid_token", error_description="%s", resource_metadata="%s"`,
			message,
			fmt.Sprintf("%s/.well-known/oauth-protected-resource%s", handlerutils.GetBaseURL(r), r.URL.Path),
		),
	)

	handlerutils.JSON(w, http.StatusUnauthorized, map[string]string{
		"error":             "invalid_token",
		"error_description": message,
	})
}

func GetTokenInfo(r *http.Request) *tokens.TokenInfo {
	v, _ := r.Context().Value(tokenInfoKey{}).(*tokens.TokenInfo)
	return v
}

type tokenInfoKey struct{}
type bearerTokenKey struct{}
