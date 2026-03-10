package revoke

import (
	"log"
	"net/http"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/handlerutils"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
)

type Store interface {
	RevokeToken(token string) error
	GetToken(token string) (*types.TokenData, error)
	GetTokenByRefreshToken(token string) (*types.TokenData, error)
	GetClient(clientID string) (*types.ClientInfo, error)
}

type Handler struct {
	db Store
}

func NewHandler(db Store) http.Handler {
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

	token := r.FormValue("token")
	tokenTypeHint := r.FormValue("token_type_hint") // RFC 7009: token_type_hint parameter
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	// Validate client ID
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

	// Validate token parameter
	if token == "" {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_request",
			ErrorDescription: "Token parameter is required",
		})
		return
	}

	// Optional: Validate that the token belongs to the requesting client
	// This is a security best practice but not required by RFC 7009
	var tokenData *types.TokenData
	var tokenErr error

	// Use token_type_hint to optimize lookup if provided
	switch tokenTypeHint {
	case "refresh_token":
		tokenData, tokenErr = p.db.GetTokenByRefreshToken(token)
	case "access_token":
		tokenData, tokenErr = p.db.GetToken(token)
	default:
		// No hint or invalid hint, try both
		tokenData, tokenErr = p.db.GetToken(token)
		if tokenErr != nil {
			tokenData, tokenErr = p.db.GetTokenByRefreshToken(token)
		}
	}

	if tokenErr == nil && tokenData != nil {
		// Token found, check client ownership
		if tokenData.ClientID != clientID {
			log.Printf("Revocation attempt by wrong client: token belongs to %s, requested by %s", tokenData.ClientID, clientID)
			// Still return 200 OK as per RFC 7009
		}
	}

	// Revoke the token in database
	if err := p.db.RevokeToken(token); err != nil {
		log.Printf("Failed to revoke token: %v", err)
		// Don't return error to client for security reasons
		// RFC 7009: "The authorization server responds with HTTP status code 200 if the token
		// has been revoked successfully or if the client submitted an invalid token."
	}

	// Always return 200 OK as per RFC 7009
	w.WriteHeader(http.StatusOK)
}
