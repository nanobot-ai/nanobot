package types

import "time"

// AuthRequest represents parsed OAuth authorization request parameters
type AuthRequest struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge,omitempty"`
	CodeChallengeMethod string `json:"code_challenge_method,omitempty"`
}

// ClientInfo represents OAuth client registration information
type ClientInfo struct {
	ClientID                string      `gorm:"primaryKey" json:"client_id"`
	ClientSecret            string      `json:"client_secret,omitempty"`
	RedirectUris            StringSlice `gorm:"type:text" json:"redirect_uris"`
	ClientName              string      `json:"client_name,omitempty"`
	LogoURI                 string      `json:"logo_uri,omitempty"`
	ClientURI               string      `json:"client_uri,omitempty"`
	PolicyURI               string      `json:"policy_uri,omitempty"`
	TosURI                  string      `json:"tos_uri,omitempty"`
	JwksURI                 string      `json:"jwks_uri,omitempty"`
	Contacts                StringSlice `gorm:"type:text" json:"contacts,omitempty"`
	GrantTypes              StringSlice `gorm:"type:text" json:"grant_types,omitempty"`
	ResponseTypes           StringSlice `gorm:"type:text" json:"response_types,omitempty"`
	RegistrationDate        int64       `json:"registration_date,omitempty"`
	TokenEndpointAuthMethod string      `gorm:"default:client_secret_basic" json:"token_endpoint_auth_method"`
	CreatedAt               time.Time   `gorm:"autoCreateTime" json:"-"`
	UpdatedAt               time.Time   `gorm:"autoUpdateTime" json:"-"`
}

// OAuthMetadata represents OAuth authorization server metadata
type OAuthMetadata struct {
	Issuer                                   string   `json:"issuer"`
	ServiceDocumentation                     string   `json:"service_documentation,omitempty"`
	AuthorizationEndpoint                    string   `json:"authorization_endpoint"`
	ResponseTypesSupported                   []string `json:"response_types_supported"`
	CodeChallengeMethodsSupported            []string `json:"code_challenge_methods_supported"`
	TokenEndpoint                            string   `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported        []string `json:"token_endpoint_auth_methods_supported"`
	GrantTypesSupported                      []string `json:"grant_types_supported"`
	ScopesSupported                          []string `json:"scopes_supported,omitempty"`
	RevocationEndpoint                       string   `json:"revocation_endpoint,omitempty"`
	RevocationEndpointAuthMethodsSupported   []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`
	RegistrationEndpoint                     string   `json:"registration_endpoint,omitempty"`
	RegistrationEndpointAuthMethodsSupported []string `json:"registration_endpoint_auth_methods_supported,omitempty"`
	UserinfoEndpoint                         string   `json:"userinfo_endpoint,omitempty"`
}

// OAuthProtectedResourceMetadata represents protected resource metadata
type OAuthProtectedResourceMetadata struct {
	Resource              string   `json:"resource"`
	AuthorizationServers  []string `json:"authorization_servers"`
	Scopes                []string `json:"scopes,omitempty"`
	ResourceName          string   `json:"resource_name,omitempty"`
	ResourceDocumentation string   `json:"resource_documentation,omitempty"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// OAuthError represents OAuth error response
type OAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}
