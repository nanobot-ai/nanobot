package token

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/encryption"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/handlerutils"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
)

type TokenStore interface {
	GetClient(clientID string) (*types.ClientInfo, error)
	StoreToken(token *types.TokenData) error
	ValidateAuthCode(code string) (string, string, error)
	GetGrant(grantID string, userID string) (*types.Grant, error)
	DeleteAuthCode(code string) error
	GetTokenByRefreshToken(refreshToken string) (*types.TokenData, error)
	RevokeToken(token string) error
}

type Handler struct {
	db TokenStore
}

func NewHandler(db TokenStore) http.Handler {
	return &Handler{
		db: db,
	}
}

func (p *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_request",
			ErrorDescription: "Invalid request body",
		})
		return
	}

	grantType := r.FormValue("grant_type")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	// Check for client ID
	if clientID == "" {
		handlerutils.JSON(w, http.StatusUnauthorized, types.OAuthError{
			Error:            "invalid_client",
			ErrorDescription: "Client ID is required",
		})
		return
	}

	// Get client info to determine authentication method
	clientInfo, err := p.db.GetClient(clientID)
	if err != nil || clientInfo == nil {
		handlerutils.JSON(w, http.StatusUnauthorized, types.OAuthError{
			Error:            "invalid_client",
			ErrorDescription: "Client not found",
		})
		return
	}

	// Check if this is a public client (auth method "none")
	isPublicClient := clientInfo.TokenEndpointAuthMethod == "none"

	// For public clients, client_secret is not required
	// For confidential clients, client_secret is required
	if !isPublicClient {
		if clientSecret == "" {
			handlerutils.JSON(w, http.StatusUnauthorized, types.OAuthError{
				Error:            "invalid_client",
				ErrorDescription: "Client secret is required for confidential clients",
			})
			return
		}

		// Validate client secret for confidential clients
		if clientInfo.ClientSecret != clientSecret {
			handlerutils.JSON(w, http.StatusUnauthorized, types.OAuthError{
				Error:            "invalid_client",
				ErrorDescription: "Invalid client secret",
			})
			return
		}
	}

	switch grantType {
	case "authorization_code":
		p.handleAuthorizationCodeGrant(w, r, clientID)
	case "refresh_token":
		p.handleRefreshTokenGrant(w, r, clientID)
	default:
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "unsupported_grant_type",
			ErrorDescription: "The grant type is not supported by this authorization server",
		})
	}
}

func (p *Handler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request, clientID string) {
	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")
	redirectURI := r.FormValue("redirect_uri")

	// Validate authorization code
	grantID, userID, err := p.db.ValidateAuthCode(code)
	if err != nil {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Invalid authorization code",
		})
		return
	}

	// Get the grant
	grant, err := p.db.GetGrant(grantID, userID)
	if err != nil {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Grant not found",
		})
		return
	}

	// Verify client ID matches
	if grant.ClientID != clientID {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Client ID mismatch",
		})
		return
	}

	// Validate redirect_uri if provided
	if redirectURI != "" {
		// Get client info to validate redirect URI
		clientInfo, err := p.db.GetClient(clientID)
		if err != nil || clientInfo == nil {
			handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
				Error:            "invalid_client",
				ErrorDescription: "Client not found",
			})
			return
		}

		// Check if redirect URI is registered for this client
		if !slices.Contains(clientInfo.RedirectUris, redirectURI) {
			handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
				Error:            "invalid_grant",
				ErrorDescription: "Invalid redirect URI",
			})
			return
		}
	}

	// Check if PKCE is being used
	isPkceEnabled := grant.CodeChallenge != ""

	// OAuth 2.1 requires redirect_uri parameter unless PKCE is used
	if redirectURI == "" && !isPkceEnabled {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_request",
			ErrorDescription: "redirect_uri is required when not using PKCE",
		})
		return
	}

	// Check PKCE if code_verifier is provided
	if codeVerifier != "" {
		if !isPkceEnabled {
			handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
				Error:            "invalid_request",
				ErrorDescription: "code_verifier provided for a flow that did not use PKCE",
			})
			return
		}

		// Verify PKCE code_verifier
		calculatedChallenge := ""
		if grant.CodeChallengeMethod == "S256" {
			hash := sha256.Sum256([]byte(codeVerifier))
			calculatedChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
		} else {
			calculatedChallenge = codeVerifier
		}

		if calculatedChallenge != grant.CodeChallenge {
			handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
				Error:            "invalid_grant",
				ErrorDescription: "Invalid PKCE code_verifier",
			})
			return
		}
	}

	// Props are stored in the grant and will be accessed when needed
	// For simple string token generation, we don't need to decrypt them here

	// Generate access token in format: userId:grantId:accessTokenSecret
	accessTokenSecret := encryption.GenerateRandomString(32)
	accessToken := fmt.Sprintf("%s:%s:%s", userID, grantID, accessTokenSecret)

	// Generate refresh token in format: userId:grantId:refreshTokenSecret
	refreshTokenSecret := encryption.GenerateRandomString(32)
	refreshToken := fmt.Sprintf("%s:%s:%s", userID, grantID, refreshTokenSecret)

	// Store tokens in database
	tokenData := &types.TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ClientID:     clientID,
		UserID:       userID,
		GrantID:      grantID,
		Scope:        strings.Join(grant.Scope, " "),
		ExpiresAt:    time.Now().Add(time.Duration(3600) * time.Second),
		CreatedAt:    time.Now(),
	}

	if err := p.db.StoreToken(tokenData); err != nil {
		log.Printf("Failed to store token: %v", err)
		handlerutils.JSON(w, http.StatusInternalServerError, types.OAuthError{
			Error:            "server_error",
			ErrorDescription: "Failed to store token",
		})
		return
	}

	// Delete the authorization code (single-use)
	if err := p.db.DeleteAuthCode(code); err != nil {
		log.Printf("Error deleting authorization code: %v", err)
	}

	response := types.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		Scope:        strings.Join(grant.Scope, " "),
	}

	handlerutils.JSON(w, http.StatusOK, response)
}

func (p *Handler) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request, clientID string) {
	refreshToken := r.FormValue("refresh_token")

	// Validate refresh token from database
	tokenData, err := p.db.GetTokenByRefreshToken(refreshToken)
	if err != nil {
		handlerutils.JSON(w, http.StatusUnauthorized, types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Invalid refresh token",
		})
		return
	}

	// Check if token is revoked
	if tokenData.Revoked {
		handlerutils.JSON(w, http.StatusUnauthorized, types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Token has been revoked",
		})
		return
	}

	// Check if refresh token is expired
	if time.Now().After(tokenData.RefreshTokenExpiresAt) {
		handlerutils.JSON(w, http.StatusUnauthorized, types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Refresh token has expired",
		})
		return
	}

	// Check if token belongs to the requesting client
	if tokenData.ClientID != clientID {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Token does not belong to the requesting client",
		})
		return
	}

	// Get the grant to access props
	_, err = p.db.GetGrant(tokenData.GrantID, tokenData.UserID)
	if err != nil {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Grant not found",
		})
		return
	}

	// Generate new access token in format: userId:grantId:accessTokenSecret
	accessTokenSecret := encryption.GenerateRandomString(32)
	accessToken := fmt.Sprintf("%s:%s:%s", tokenData.UserID, tokenData.GrantID, accessTokenSecret)

	// Generate new refresh token
	refreshTokenSecret := encryption.GenerateRandomString(32)
	refreshToken = fmt.Sprintf("%s:%s:%s", tokenData.UserID, tokenData.GrantID, refreshTokenSecret)
	refreshTokenExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days from now

	// Store new token in database (replaces the old one)
	newTokenData := &types.TokenData{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		ClientID:              clientID,
		UserID:                tokenData.UserID,
		GrantID:               tokenData.GrantID,
		Scope:                 tokenData.Scope,
		ExpiresAt:             time.Now().Add(3600 * time.Second),
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
		CreatedAt:             time.Now(),
	}

	if err := p.db.StoreToken(newTokenData); err != nil {
		log.Printf("Failed to store new token: %v", err)
		handlerutils.JSON(w, http.StatusInternalServerError, types.OAuthError{
			Error:            "server_error",
			ErrorDescription: "Failed to store new token",
		})
		return
	}

	// Revoke the old refresh token
	if err := p.db.RevokeToken(refreshToken); err != nil {
		log.Printf("Failed to revoke old refresh token: %v", err)
	}

	response := types.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		Scope:        tokenData.Scope,
	}

	handlerutils.JSON(w, http.StatusOK, response)
}
