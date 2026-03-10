package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"golang.org/x/oauth2"
)

// GenericProvider implements a generic OAuth provider
type GenericProvider struct {
	authorizeURL string
	metadata     *types.OAuthMetadata
	httpClient   *http.Client
}

// NewGenericProvider creates a new generic OAuth provider
func NewGenericProvider(authorizeURL string) *GenericProvider {
	return &GenericProvider{
		authorizeURL: authorizeURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// discoverEndpoints attempts to discover OAuth endpoints using well-known paths
func (p *GenericProvider) discoverEndpoints() error {
	if p.metadata != nil {
		return nil // Already discovered
	}

	// Parse the authorize URL to get the base URL
	parsedURL, err := url.Parse(p.authorizeURL)
	if err != nil {
		return fmt.Errorf("invalid authorize URL: %w", err)
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Try different well-known paths
	wellKnownPaths := []string{
		"/.well-known/oauth-authorization-server" + parsedURL.Path,
		fmt.Sprintf("%s/.well-known/oauth-authorization-server", strings.TrimSuffix(parsedURL.Path, "/")),
		"/.well-known/openid-configuration" + parsedURL.Path,
		fmt.Sprintf("%s/.well-known/openid-configuration", strings.TrimSuffix(parsedURL.Path, "/")),
	}

	for _, path := range wellKnownPaths {
		metadata, err := p.fetchMetadata(baseURL + path)
		if err == nil && metadata != nil {
			p.metadata = metadata

			// for github, there is no userinfo endpoint, so we need to provide a hardcoded one
			if p.metadata.UserinfoEndpoint == "" && parsedURL.Host == "github.com" {
				p.metadata.UserinfoEndpoint = "https://api.github.com/user"
			}
			return nil
		}
	}

	// If no metadata found, create a basic metadata structure
	p.metadata = &types.OAuthMetadata{
		Issuer:                 baseURL,
		AuthorizationEndpoint:  p.authorizeURL,
		TokenEndpoint:          baseURL + "/token",
		UserinfoEndpoint:       baseURL + "/userinfo",
		ScopesSupported:        []string{"openid", "profile", "email"},
		ResponseTypesSupported: []string{"code"},
		GrantTypesSupported:    []string{"authorization_code", "refresh_token"},
	}

	return nil
}

// fetchMetadata fetches OAuth metadata from a URL
func (p *GenericProvider) fetchMetadata(metadataURL string) (*types.OAuthMetadata, error) {
	resp, err := p.httpClient.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the function
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch metadata: %s", resp.Status)
	}

	var metadata types.OAuthMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}

// GetAuthorizationURL returns the authorization URL for the provider
func (p *GenericProvider) GetAuthorizationURL(clientID, redirectURI, scope, state string) string {
	return p.GetAuthorizationURLWithPKCE(clientID, redirectURI, scope, state, "")
}

// GetAuthorizationURLWithPKCE returns the authorization URL with PKCE support
func (p *GenericProvider) GetAuthorizationURLWithPKCE(clientID, redirectURI, scope, state, codeChallenge string) string {
	// Fallback to basic URL construction
	authURL := p.authorizeURL
	if err := p.discoverEndpoints(); err == nil {
		authURL = p.metadata.AuthorizationEndpoint
	}

	o := p.buildOAuth2Config(authURL, clientID, "", redirectURI, scope)

	opts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	}
	if codeChallenge != "" {
		opts = append(opts, oauth2.S256ChallengeOption(codeChallenge))
	}
	return o.AuthCodeURL(state, opts...)
}

// any exchanges authorization code for tokens
func (p *GenericProvider) ExchangeCodeForToken(ctx context.Context, code, clientID, clientSecret, redirectURI string) (*oauth2.Token, error) {
	if err := p.discoverEndpoints(); err != nil {
		return nil, fmt.Errorf("failed to discover endpoints: %w", err)
	}

	return p.buildOAuth2Config(p.metadata.AuthorizationEndpoint, clientID, clientSecret, redirectURI, "").Exchange(ctx, code)
}

// GetUserInfo retrieves user information using the access token
func (p *GenericProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	if err := p.discoverEndpoints(); err != nil {
		return nil, fmt.Errorf("failed to discover endpoints: %w", err)
	}

	if p.metadata.UserinfoEndpoint == "" {
		return nil, fmt.Errorf("userinfo endpoint not available")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.metadata.UserinfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the function
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed: %s", resp.Status)
	}

	var userInfoResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&userInfoResp); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	userInfo := &UserInfo{
		ID:            getString(userInfoResp, "id"),
		Sub:           getString(userInfoResp, "sub"),
		Login:         getString(userInfoResp, "login"),
		Email:         getString(userInfoResp, "email"),
		EmailVerified: getBool(userInfoResp, "email_verified"),
		Name:          getString(userInfoResp, "name"),
		Picture:       getString(userInfoResp, "picture"),
		GivenName:     getString(userInfoResp, "given_name"),
		FamilyName:    getString(userInfoResp, "family_name"),
		Locale:        getString(userInfoResp, "locale"),
	}

	if userInfo.ID == "" {
		userInfo.ID = userInfo.Sub
	}

	if userInfo.ID == "" && p.metadata.UserinfoEndpoint == "https://api.github.com/user" {
		userInfo.ID = userInfo.Login
	}

	return userInfo, nil
}

// RefreshToken refreshes an access token using a refresh token
func (p *GenericProvider) RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret string) (*oauth2.Token, error) {
	if err := p.discoverEndpoints(); err != nil {
		return nil, fmt.Errorf("failed to discover endpoints: %w", err)
	}

	return p.buildOAuth2Config(p.metadata.AuthorizationEndpoint, clientID, clientSecret, "", "").
		TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).
		Token()
}

func (p *GenericProvider) buildOAuth2Config(authURL, clientID, clientSecret, redirectURI, scope string) *oauth2.Config {
	var scopes []string
	if scope != "" {
		scopes = strings.Fields(scope)
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: p.metadata.TokenEndpoint,
		},
	}
}

// GetName returns the provider name
func (p *GenericProvider) GetName() string {
	return "generic"
}

func getBool(m map[string]any, key string) bool {
	b, _ := m[key].(bool)
	return b
}

// Helper functions
func getString(m map[string]any, key string) string {
	str, _ := m[key].(string)
	return str
}
