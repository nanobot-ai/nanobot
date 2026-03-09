package register

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/encryption"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/handlerutils"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
)

type ClientStore interface {
	StoreClient(client *types.ClientInfo) error
}

func NewHandler(db ClientStore) http.Handler {
	return &Handler{
		db: db,
	}
}

type Handler struct {
	db ClientStore
}

func (p *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Client registration is always enabled when this endpoint is accessible

	// Only allow POST method
	if r.Method != "POST" {
		handlerutils.JSON(w, http.StatusMethodNotAllowed, types.OAuthError{
			Error:            "invalid_request",
			ErrorDescription: "Method not allowed",
		})
		return
	}

	// Check content length (max 1MB)
	if r.ContentLength > 1024*1024 {
		handlerutils.JSON(w, http.StatusRequestEntityTooLarge, types.OAuthError{
			Error:            "invalid_request",
			ErrorDescription: "Request payload too large, must be under 1 MiB",
		})
		return
	}

	// Parse request.JSON body
	var clientMetadata map[string]any
	if err := json.NewDecoder(io.LimitReader(r.Body, 1024*1024)).Decode(&clientMetadata); err != nil {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_request",
			ErrorDescription: "Invalid request.JSON payload",
		})
		return
	}

	// Validate and extract client metadata
	clientInfo, err := p.validateClientMetadata(clientMetadata)
	if err != nil {
		handlerutils.JSON(w, http.StatusBadRequest, types.OAuthError{
			Error:            "invalid_client_metadata",
			ErrorDescription: err.Error(),
		})
		return
	}

	// Generate client ID and secret
	clientID := encryption.GenerateRandomString(16)
	clientSecret := ""

	// Generate client secret for confidential clients
	if clientInfo.TokenEndpointAuthMethod != "none" {
		clientSecret = encryption.GenerateRandomString(32)
	}

	// Set registration date
	clientInfo.ClientID = clientID
	clientInfo.ClientSecret = clientSecret
	clientInfo.RegistrationDate = time.Now().Unix()

	// Convert to database.ClientInfo
	dbClientInfo := &types.ClientInfo{
		ClientID:                clientInfo.ClientID,
		ClientSecret:            clientInfo.ClientSecret,
		RedirectUris:            clientInfo.RedirectUris,
		ClientName:              clientInfo.ClientName,
		LogoURI:                 clientInfo.LogoURI,
		ClientURI:               clientInfo.ClientURI,
		PolicyURI:               clientInfo.PolicyURI,
		TosURI:                  clientInfo.TosURI,
		JwksURI:                 clientInfo.JwksURI,
		Contacts:                clientInfo.Contacts,
		GrantTypes:              clientInfo.GrantTypes,
		ResponseTypes:           clientInfo.ResponseTypes,
		RegistrationDate:        clientInfo.RegistrationDate,
		TokenEndpointAuthMethod: clientInfo.TokenEndpointAuthMethod,
	}

	// Store client in database
	if err := p.db.StoreClient(dbClientInfo); err != nil {
		log.Printf("Failed to store client: %v", err)
		handlerutils.JSON(w, http.StatusInternalServerError, types.OAuthError{
			Error:            "server_error",
			ErrorDescription: "Failed to register client",
		})
		return
	}

	// Build response
	baseURL := handlerutils.GetBaseURL(r)
	response := map[string]any{
		"client_id":                  clientInfo.ClientID,
		"redirect_uris":              clientInfo.RedirectUris,
		"client_name":                clientInfo.ClientName,
		"logo_uri":                   clientInfo.LogoURI,
		"client_uri":                 clientInfo.ClientURI,
		"policy_uri":                 clientInfo.PolicyURI,
		"tos_uri":                    clientInfo.TosURI,
		"jwks_uri":                   clientInfo.JwksURI,
		"contacts":                   clientInfo.Contacts,
		"grant_types":                clientInfo.GrantTypes,
		"response_types":             clientInfo.ResponseTypes,
		"token_endpoint_auth_method": clientInfo.TokenEndpointAuthMethod,
		"registration_client_uri":    fmt.Sprintf("%s/register/%s", baseURL, clientID),
		"client_id_issued_at":        clientInfo.RegistrationDate,
	}

	// Include client secret in response for confidential clients
	if clientSecret != "" {
		response["client_secret"] = clientSecret
		response["client_secret_expires_at"] = 0 // Never expires
	}

	// Drop empty values
	for k, v := range response {
		if v == "" || v == nil {
			delete(response, k)
		}
	}

	handlerutils.JSON(w, http.StatusOK, response)
}

func (p *Handler) validateClientMetadata(metadata map[string]any) (*types.ClientInfo, error) {
	// Helper function to validate string fields
	validateStringField := func(field any, name string) (string, error) {
		if field == nil {
			return "", nil
		}
		if str, ok := field.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("field %s must be a string", name)
	}

	// Helper function to validate string arrays
	validateStringArray := func(arr any, name string) ([]string, error) {
		if arr == nil {
			return nil, nil
		}
		if array, ok := arr.([]any); ok {
			result := make([]string, len(array))
			for i, item := range array {
				if str, ok := item.(string); ok {
					result[i] = str
				} else {
					return nil, fmt.Errorf("all elements in %s must be strings", name)
				}
			}
			return result, nil
		}
		return nil, fmt.Errorf("field %s must be an array", name)
	}

	// Validate token_endpoint_auth_method
	authMethod, err := validateStringField(metadata["token_endpoint_auth_method"], "token_endpoint_auth_method")
	if err != nil {
		return nil, err
	}
	if authMethod == "" {
		authMethod = "client_secret_basic"
	}

	// Validate that the auth method is supported
	validAuthMethods := []string{
		"none",
		"client_secret_post",
	}

	isValidAuthMethod := false
	for _, validMethod := range validAuthMethods {
		if authMethod == validMethod {
			isValidAuthMethod = true
			break
		}
	}

	if !isValidAuthMethod {
		return nil, fmt.Errorf("unsupported token_endpoint_auth_method: %s. Supported methods: %v", authMethod, validAuthMethods)
	}

	// Validate redirect_uris (required)
	redirectUris, err := validateStringArray(metadata["redirect_uris"], "redirect_uris")
	if err != nil {
		return nil, err
	}
	if len(redirectUris) == 0 {
		return nil, fmt.Errorf("at least one redirect URI is required")
	}

	// Validate other fields
	clientName, err := validateStringField(metadata["client_name"], "client_name")
	if err != nil {
		return nil, err
	}

	logoURI, err := validateStringField(metadata["logo_uri"], "logo_uri")
	if err != nil {
		return nil, err
	}

	clientURI, err := validateStringField(metadata["client_uri"], "client_uri")
	if err != nil {
		return nil, err
	}

	policyURI, err := validateStringField(metadata["policy_uri"], "policy_uri")
	if err != nil {
		return nil, err
	}

	tosURI, err := validateStringField(metadata["tos_uri"], "tos_uri")
	if err != nil {
		return nil, err
	}

	jwksURI, err := validateStringField(metadata["jwks_uri"], "jwks_uri")
	if err != nil {
		return nil, err
	}

	contacts, err := validateStringArray(metadata["contacts"], "contacts")
	if err != nil {
		return nil, err
	}
	// inspector will check the schema and see if it is null, so this is a workaround
	if len(contacts) == 0 {
		contacts = []string{}
	}

	grantTypes, err := validateStringArray(metadata["grant_types"], "grant_types")
	if err != nil {
		return nil, err
	}
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code", "refresh_token"}
	}

	responseTypes, err := validateStringArray(metadata["response_types"], "response_types")
	if err != nil {
		return nil, err
	}
	if len(responseTypes) == 0 {
		responseTypes = []string{"code"}
	}

	return &types.ClientInfo{
		RedirectUris:            redirectUris,
		ClientName:              clientName,
		LogoURI:                 logoURI,
		ClientURI:               clientURI,
		PolicyURI:               policyURI,
		TosURI:                  tosURI,
		JwksURI:                 jwksURI,
		Contacts:                contacts,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: authMethod,
	}, nil
}
