package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"golang.org/x/oauth2"
)

var resourceMetadataRegex = regexp.MustCompile(`resource_metadata="([^"]*)"`)

type oauth struct {
	redirectURL, clientName string
	metadataClient          *http.Client
	callbackHandler         CallbackHandler
	clientLookup            ClientCredLookup
	tokenStorage            TokenStorage
}

func newOAuth(callbackHandler CallbackHandler, clientLookup ClientCredLookup, tokenStorage TokenStorage, clientName, redirectURL string) *oauth {
	return &oauth{
		clientName:      clientName,
		redirectURL:     redirectURL,
		callbackHandler: callbackHandler,
		metadataClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		clientLookup: clientLookup,
		tokenStorage: tokenStorage,
	}
}

func (o *oauth) loadFromStorage(ctx context.Context, connectURL string) (*http.Client, error) {
	if o.tokenStorage == nil {
		return nil, nil
	}

	// Read the token config from storage to see if we have valid auth
	conf, tok, err := o.tokenStorage.GetTokenConfig(ctx, connectURL)
	if err != nil {
		log.Infof(ctx, "failed to read token config: %v", err)
		log.Infof(ctx, "continuing with authentication")
	}

	if conf != nil && tok != nil {
		tok, err = conf.TokenSource(ctx, tok).Token()
		if err == nil && tok.Valid() {
			return conf.Client(ctx, tok), nil
		}
	}

	return nil, nil
}

func (o *oauth) oauthClient(ctx context.Context, c *HTTPClient, connectURL, authenticateHeader string) (*http.Client, error) {
	if o.callbackHandler == nil || o.redirectURL == "" {
		return nil, fmt.Errorf("oauth callback server is not configured")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MCP URL: %w", err)
	}

	var resourceMetadataURL string
	if authenticateHeader != "" {
		var err error
		resourceMetadataURL, err = parseResourceMetadata(authenticateHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse authenticate header: %w", err)
		}
	} else {
		// If the authenticate header was not sent back, then the spec says we should default to...
		u.Path = "/.well-known/oauth-protected-resource"
		resourceMetadataURL = u.String()
	}

	// Get the protected resource metadata
	protectedResourceResp, err := o.metadataClient.Get(resourceMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get protected resource metadata: %w", err)
	}
	defer protectedResourceResp.Body.Close()

	var protectedResourceMetadata protectedResourceMetadata
	if protectedResourceResp.StatusCode != http.StatusOK && protectedResourceResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(protectedResourceResp.Body)
		return nil, fmt.Errorf("unexpeted status getting protected resource metadata (%d): %s", protectedResourceResp.StatusCode, string(body))
	} else if protectedResourceResp.StatusCode == http.StatusOK {
		protectedResourceMetadata, err = parseProtectedResourceMetadata(protectedResourceResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse protected resource metadata: %w", err)
		}
	}

	if len(protectedResourceMetadata.AuthorizationServers) == 0 {
		protectedResourceMetadata.AuthorizationServers = []string{fmt.Sprintf("%s://%s", u.Scheme, u.Host)}
	}

	authorizationServerMetadata, err := o.getAuthServerMetadata(protectedResourceMetadata.AuthorizationServers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization server metadata: %w", err)
	}

	clientMetadata := authServerMetadataToClientRegistration(authorizationServerMetadata)
	clientMetadata.RedirectURIs = []string{o.redirectURL}
	clientMetadata.ClientName = o.clientName

	b, err := json.Marshal(clientMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal client metadata: %w", err)
	}

	req, err := http.NewRequest("POST", authorizationServerMetadata.RegistrationEndpoint, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := o.metadataClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to register client: %w", err)
	}
	defer resp.Body.Close()

	var clientInfo clientRegistrationResponse
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		// If the registration endpoint produces a not found, then look for static client credentials.
		clientInfo.ClientID, clientInfo.ClientSecret, err = o.clientLookup.Lookup(ctx, protectedResourceMetadata.AuthorizationServers[0])
		if err != nil {
			return nil, fmt.Errorf("failed to lookup client credentials: %w", err)
		}
		if clientInfo.ClientID == "" {
			return nil, fmt.Errorf("failed to lookup client credentials")
		}
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpeted status registering client (%d): %s", resp.StatusCode, string(body))
	} else {
		clientInfo, err = parseClientRegistrationResponse(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse client registration response: %w", err)
		}
	}

	conf := &oauth2.Config{
		ClientID:     clientInfo.ClientID,
		ClientSecret: clientInfo.ClientSecret,
		RedirectURL:  clientMetadata.RedirectURIs[0],
		Endpoint: oauth2.Endpoint{
			AuthURL:  authorizationServerMetadata.AuthorizationEndpoint,
			TokenURL: authorizationServerMetadata.TokenEndpoint,
		},
	}
	if clientMetadata.Scope != "" {
		conf.Scopes = strings.Split(clientMetadata.Scope, " ")
	}
	switch clientMetadata.TokenEndpointAuthMethod {
	case "client_secret_basic":
		conf.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	case "client_secret_post":
		conf.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	default:
		conf.Endpoint.AuthStyle = oauth2.AuthStyleAutoDetect
	}

	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()

	state, ch, err := o.callbackHandler.NewState(ctx, conf, verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create state: %w", err)
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	authURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	handled, err := o.callbackHandler.HandleAuthURL(ctx, c.serverName, authURL)
	if err != nil {
		return nil, fmt.Errorf("failed to handle auth url %s: %w", authURL, err)
	} else if !handled {
		return nil, nil
	}

	var cb CallbackPayload
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case cb = <-ch:
		if cb.Error != "" {
			return nil, fmt.Errorf("authorization failed: %s, %s", cb.Error, cb.ErrorDescription)
		}
		if cb.Code == "" {
			return nil, fmt.Errorf("authorization failed: no code returned")
		}
	}

	tok, err := conf.Exchange(ctx, cb.Code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	if o.tokenStorage != nil {
		if err = o.tokenStorage.SetTokenConfig(ctx, connectURL, conf, tok); err != nil {
			log.Infof(ctx, "failed to save token config: %v", err)
		}
	}

	return conf.Client(ctx, tok), nil
}

func (o *oauth) getAuthServerMetadata(authURL string) (authorizationServerMetadata, error) {
	authServerURL := strings.TrimSuffix(authURL, "/")

	authServerMetadata := authServerURL
	// If the authServer URL has a path, then the well-known path is prepended to the path
	if u, err := url.Parse(authServerMetadata); err != nil {
		return authorizationServerMetadata{}, fmt.Errorf("failed to parse auth server URL: %w", err)
	} else if u.Path != "" {
		u.Path = "/.well-known/oauth-authorization-server" + u.Path
		authServerMetadata = u.String()
	} else {
		authServerMetadata = fmt.Sprintf("%s/.well-known/oauth-authorization-server/", authServerMetadata)
	}
	oauthMetadataResp, err := o.metadataClient.Get(authServerMetadata)
	if err != nil {
		return authorizationServerMetadata{}, fmt.Errorf("failed to get authorization server metadata: %w", err)
	}
	defer oauthMetadataResp.Body.Close()

	var authorizationServerMetadataContent authorizationServerMetadata
	if oauthMetadataResp.StatusCode != http.StatusOK && oauthMetadataResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(oauthMetadataResp.Body)
		return authorizationServerMetadata{}, fmt.Errorf("unexpeted status getting authorization server metadata (%d): %s", oauthMetadataResp.StatusCode, string(body))
	} else if oauthMetadataResp.StatusCode == http.StatusOK {
		authorizationServerMetadataContent, err = parseAuthorizationServerMetadata(oauthMetadataResp.Body)
		if err != nil {
			return authorizationServerMetadata{}, fmt.Errorf("failed to parse authorization server metadata: %w", err)
		}
	} else {
		// We couldn't find the oauth-authorization-server endpoint, so look for the openid-configuration endpoint.
		openIDConfigResp, err := o.metadataClient.Get(strings.Replace(authServerMetadata, "/.well-known/oauth-authorization-server", "/.well-known/openid-configuration", 1))
		if err != nil {
			return authorizationServerMetadata{}, fmt.Errorf("failed to get openid-configuration: %w", err)
		}
		defer openIDConfigResp.Body.Close()

		if openIDConfigResp.StatusCode != http.StatusOK {
			authorizationServerMetadataContent, err = parseAuthorizationServerMetadata(openIDConfigResp.Body)
			if err != nil {
				return authorizationServerMetadata{}, fmt.Errorf("failed to parse openid configuration: %w", err)
			}
		} else {
			// The last URL we check is appending the openid-configuration path to the end.
			openIDConfigResp, err := o.metadataClient.Get(strings.Replace(authServerMetadata, "/.well-known/oauth-authorization-server", "", 1) + "/.well-known/openid-configuration")
			if err != nil {
				return authorizationServerMetadata{}, fmt.Errorf("failed to get openid-configuration: %w", err)
			}
			defer openIDConfigResp.Body.Close()

			authorizationServerMetadataContent, err = parseAuthorizationServerMetadata(openIDConfigResp.Body)
			if err != nil {
				return authorizationServerMetadata{}, fmt.Errorf("failed to parse openid configuration: %w", err)
			}
		}

	}

	if authorizationServerMetadataContent.AuthorizationEndpoint == "" {
		authorizationServerMetadataContent.AuthorizationEndpoint = fmt.Sprintf("%s/authorize", authServerURL)
	}
	if authorizationServerMetadataContent.TokenEndpoint == "" {
		authorizationServerMetadataContent.TokenEndpoint = fmt.Sprintf("%s/token", authServerURL)
	}
	if authorizationServerMetadataContent.RegistrationEndpoint == "" {
		authorizationServerMetadataContent.RegistrationEndpoint = fmt.Sprintf("%s/register", authServerURL)
	}

	return authorizationServerMetadataContent, nil
}

// parseAuthorizationServerMetadata parses OAuth 2.0 Authorization Server Metadata
// from a reader containing JSON data as defined in RFC 8414
func parseAuthorizationServerMetadata(reader io.Reader) (authorizationServerMetadata, error) {
	var metadata authorizationServerMetadata
	if err := json.NewDecoder(reader).Decode(&metadata); err != nil {
		return metadata, fmt.Errorf("failed to decode authorization server metadata: %w", err)
	}

	// Validate required fields
	if metadata.Issuer == "" {
		return metadata, fmt.Errorf("issuer is required but not provided")
	}

	if len(metadata.ResponseTypesSupported) == 0 {
		return metadata, fmt.Errorf("response_types_supported is required but not provided")
	}

	// Set default values for optional fields if not provided
	if len(metadata.ResponseModesSupported) == 0 {
		metadata.ResponseModesSupported = []string{"query", "fragment"}
	}

	if len(metadata.GrantTypesSupported) == 0 {
		metadata.GrantTypesSupported = []string{"authorization_code", "implicit"}
	}

	if len(metadata.TokenEndpointAuthMethodsSupported) == 0 {
		metadata.TokenEndpointAuthMethodsSupported = []string{"client_secret_basic"}
	}

	if len(metadata.RevocationEndpointAuthMethodsSupported) == 0 {
		metadata.RevocationEndpointAuthMethodsSupported = []string{"client_secret_basic"}
	}

	return metadata, nil
}

// parseProtectedResourceMetadata parses OAuth 2.0 Protected Resource Metadata
// from a reader containing JSON data as defined in RFC 8707
func parseProtectedResourceMetadata(reader io.Reader) (protectedResourceMetadata, error) {
	var metadata protectedResourceMetadata
	if err := json.NewDecoder(reader).Decode(&metadata); err != nil {
		return metadata, fmt.Errorf("failed to decode protected resource metadata: %w", err)
	}

	// Validate required fields
	if metadata.Resource == "" {
		return metadata, fmt.Errorf("resource is required but not provided")
	}

	// Set default values for optional fields if not provided
	// According to RFC 8707, if bearer_methods_supported is omitted, no default bearer methods are implied
	// The empty array [] can be used to indicate that no bearer methods are supported
	// We don't set defaults here as the absence has specific meaning

	// Validate that resource_signing_alg_values_supported does not contain "none"
	for _, alg := range metadata.ResourceSigningAlgValuesSupported {
		if alg == "none" {
			return metadata, fmt.Errorf("resource_signing_alg_values_supported must not contain 'none'")
		}
	}

	return metadata, nil
}

// parseResourceMetadata extracts the resource_metadata URL from a Bearer authenticate header
func parseResourceMetadata(authenticateHeader string) (string, error) {
	// Use regex to find resource_metadata parameter
	// Pattern matches: resource_metadata="<URL>"
	matches := resourceMetadataRegex.FindStringSubmatch(authenticateHeader)

	if len(matches) < 2 {
		return "", fmt.Errorf("resource_metadata not found in authenticate header")
	}

	return matches[1], nil
}

// protectedResourceMetadata represents OAuth 2.0 Protected Resource Metadata
// as defined in RFC 8707
type protectedResourceMetadata struct {
	// REQUIRED. The protected resource's resource identifier
	Resource string `json:"resource"`

	// OPTIONAL. JSON array containing a list of OAuth authorization server issuer identifiers
	AuthorizationServers []string `json:"authorization_servers,omitempty"`

	// OPTIONAL. URL of the protected resource's JSON Web Key (JWK) Set document
	JwksURI string `json:"jwks_uri,omitempty"`

	// RECOMMENDED. JSON array containing a list of scope values
	ScopesSupported []string `json:"scopes_supported,omitempty"`

	// OPTIONAL. JSON array containing a list of the supported methods of sending an OAuth 2.0 bearer token
	BearerMethodsSupported []string `json:"bearer_methods_supported,omitempty"`

	// OPTIONAL. JSON array containing a list of the JWS signing algorithms supported by the protected resource
	ResourceSigningAlgValuesSupported []string `json:"resource_signing_alg_values_supported,omitempty"`

	// OPTIONAL. Human-readable name of the protected resource intended for display to the end user
	ResourceName string `json:"resource_name,omitempty"`

	// OPTIONAL. URL of a page containing human-readable information that developers might want or need to know
	ResourceDocumentation string `json:"resource_documentation,omitempty"`

	// OPTIONAL. URL of a page containing human-readable information about the protected resource's requirements
	ResourcePolicyURI string `json:"resource_policy_uri,omitempty"`

	// OPTIONAL. URL of a page containing human-readable information about the protected resource's terms of service
	ResourceTosURI string `json:"resource_tos_uri,omitempty"`

	// OPTIONAL. Boolean value indicating protected resource support for mutual-TLS client certificate-bound access tokens
	TLSClientCertificateBoundAccessTokens bool `json:"tls_client_certificate_bound_access_tokens,omitempty"`

	// OPTIONAL. JSON array containing a list of the authorization details type values supported by the resource server
	AuthorizationDetailsTypesSupported []string `json:"authorization_details_types_supported,omitempty"`

	// OPTIONAL. JSON array containing a list of the JWS alg values supported by the resource server for validating DPoP proof JWTs
	DPoPSigningAlgValuesSupported []string `json:"dpop_signing_alg_values_supported,omitempty"`

	// OPTIONAL. Boolean value specifying whether the protected resource always requires the use of DPoP-bound access tokens
	DPoPBoundAccessTokensRequired bool `json:"dpop_bound_access_tokens_required,omitempty"`
}

// authorizationServerMetadata represents OAuth 2.0 Authorization Server Metadata
// as defined in RFC 8414
type authorizationServerMetadata struct {
	// REQUIRED. The authorization server's issuer identifier
	Issuer string `json:"issuer"`

	// URL of the authorization server's authorization endpoint
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`

	// URL of the authorization server's token endpoint
	TokenEndpoint string `json:"token_endpoint,omitempty"`

	// OPTIONAL. URL of the authorization server's JWK Set document
	JwksURI string `json:"jwks_uri,omitempty"`

	// OPTIONAL. URL of the authorization server's OAuth 2.0 Dynamic Client Registration endpoint
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`

	// RECOMMENDED. JSON array containing a list of the OAuth 2.0 scope values
	ScopesSupported []string `json:"scopes_supported,omitempty"`

	// REQUIRED. JSON array containing a list of the OAuth 2.0 response_type values
	ResponseTypesSupported []string `json:"response_types_supported"`

	// OPTIONAL. JSON array containing a list of the OAuth 2.0 response_mode values
	ResponseModesSupported []string `json:"response_modes_supported,omitempty"`

	// OPTIONAL. JSON array containing a list of the OAuth 2.0 grant type values
	GrantTypesSupported []string `json:"grant_types_supported,omitempty"`

	// OPTIONAL. JSON array containing a list of client authentication methods
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`

	// OPTIONAL. JSON array containing a list of the JWS signing algorithms
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`

	// OPTIONAL. URL of a page containing human-readable information
	ServiceDocumentation string `json:"service_documentation,omitempty"`

	// OPTIONAL. Languages and scripts supported for the user interface
	UILocalesSupported []string `json:"ui_locales_supported,omitempty"`

	// OPTIONAL. URL for authorization server's requirements on client data usage
	OpPolicyURI string `json:"op_policy_uri,omitempty"`

	// OPTIONAL. URL for authorization server's terms of service
	OpTosURI string `json:"op_tos_uri,omitempty"`

	// OPTIONAL. URL of the authorization server's OAuth 2.0 revocation endpoint
	RevocationEndpoint string `json:"revocation_endpoint,omitempty"`

	// OPTIONAL. JSON array containing client authentication methods for revocation endpoint
	RevocationEndpointAuthMethodsSupported []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`

	// OPTIONAL. JSON array containing JWS signing algorithms for revocation endpoint
	RevocationEndpointAuthSigningAlgValuesSupported []string `json:"revocation_endpoint_auth_signing_alg_values_supported,omitempty"`

	// OPTIONAL. URL of the authorization server's OAuth 2.0 introspection endpoint
	IntrospectionEndpoint string `json:"introspection_endpoint,omitempty"`

	// OPTIONAL. JSON array containing client authentication methods for introspection endpoint
	IntrospectionEndpointAuthMethodsSupported []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`

	// OPTIONAL. JSON array containing JWS signing algorithms for introspection endpoint
	IntrospectionEndpointAuthSigningAlgValuesSupported []string `json:"introspection_endpoint_auth_signing_alg_values_supported,omitempty"`

	// OPTIONAL. JSON array containing PKCE code challenge methods
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported,omitempty"`
}

// clientRegistrationMetadata represents OAuth 2.0 Dynamic Client Registration metadata
// as defined in RFC 7591, merged from protected resource and authorization server metadata
type clientRegistrationMetadata struct {
	// Array of redirection URI strings for use in redirect-based flows
	RedirectURIs []string `json:"redirect_uris,omitempty"`

	// String indicator of the requested authentication method for the token endpoint
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method,omitempty"`

	// Array of OAuth 2.0 grant type strings that the client can use at the token endpoint
	GrantTypes []string `json:"grant_types,omitempty"`

	// Array of the OAuth 2.0 response type strings that the client can use at the authorization endpoint
	ResponseTypes []string `json:"response_types,omitempty"`

	// Human-readable string name of the client to be presented to the end-user during authorization
	ClientName string `json:"client_name,omitempty"`

	// URL string of a web page providing information about the client
	ClientURI string `json:"client_uri,omitempty"`

	// URL string that references a logo for the client
	LogoURI string `json:"logo_uri,omitempty"`

	// String containing a space-separated list of scope values
	Scope string `json:"scope,omitempty"`

	// Array of strings representing ways to contact people responsible for this client
	Contacts []string `json:"contacts,omitempty"`

	// URL string that points to a human-readable terms of service document for the client
	TosURI string `json:"tos_uri,omitempty"`

	// URL string that points to a human-readable privacy policy document
	PolicyURI string `json:"policy_uri,omitempty"`

	// URL string referencing the client's JSON Web Key (JWK) Set document
	JwksURI string `json:"jwks_uri,omitempty"`

	// Client's JSON Web Key Set document value
	Jwks interface{} `json:"jwks,omitempty"`

	// A unique identifier string assigned by the client developer or software publisher
	SoftwareID string `json:"software_id,omitempty"`

	// A version identifier string for the client software identified by "software_id"
	SoftwareVersion string `json:"software_version,omitempty"`
}

func authServerMetadataToClientRegistration(authServer authorizationServerMetadata) clientRegistrationMetadata {
	merged := clientRegistrationMetadata{}

	// Set default values based on OAuth 2.0 specifications

	// token_endpoint_auth_method: default is "client_secret_basic" if not specified
	if len(authServer.TokenEndpointAuthMethodsSupported) > 0 {
		merged.TokenEndpointAuthMethod = authServer.TokenEndpointAuthMethodsSupported[0]
	} else {
		merged.TokenEndpointAuthMethod = "client_secret_basic"
	}

	// grant_types: default is "authorization_code" if not specified
	if len(authServer.GrantTypesSupported) > 0 {
		merged.GrantTypes = authServer.GrantTypesSupported
	} else {
		merged.GrantTypes = []string{"authorization_code"}
	}

	// response_types: default is "code" if not specified
	if len(authServer.ResponseTypesSupported) > 0 {
		merged.ResponseTypes = authServer.ResponseTypesSupported
	} else {
		merged.ResponseTypes = []string{"code"}
	}

	// scope: combine scopes from both sources, preferring protected resource
	if len(authServer.ScopesSupported) > 0 {
		merged.Scope = strings.Join(authServer.ScopesSupported, " ")
	}
	// Note: redirect_uris, logo_uri, contacts, jwks, software_id, and software_version
	// are typically client-specific and would need to be provided by the client application
	// These fields are left empty as they cannot be derived from server metadata

	return merged
}

// clientRegistrationResponse represents OAuth 2.0 Dynamic Client Registration Response
// as defined in RFC 7591
type clientRegistrationResponse struct {
	// REQUIRED. OAuth 2.0 client identifier string. It SHOULD NOT be
	// currently valid for any other registered client, though an
	// authorization server MAY issue the same client identifier to
	// multiple instances of a registered client at its discretion.
	ClientID string `json:"client_id"`

	// OPTIONAL. OAuth 2.0 client secret string. If issued, this MUST
	// be unique for each "client_id" and SHOULD be unique for multiple
	// instances of a client using the same "client_id". This value is
	// used by confidential clients to authenticate to the token
	// endpoint, as described in OAuth 2.0 [RFC6749], Section 2.3.1.
	ClientSecret string `json:"client_secret,omitempty"`

	// OPTIONAL. Time at which the client identifier was issued. The
	// time is represented as the number of seconds from
	// 1970-01-01T00:00:00Z as measured in UTC until the date/time of
	// issuance.
	ClientIDIssuedAt *int64 `json:"client_id_issued_at,omitempty"`

	// REQUIRED if "client_secret" is issued. Time at which the client
	// secret will expire or 0 if it will not expire. The time is
	// represented as the number of seconds from 1970-01-01T00:00:00Z as
	// measured in UTC until the date/time of expiration.
	ClientSecretExpiresAt *int64 `json:"client_secret_expires_at,omitempty"`
}

// parseClientRegistrationResponse parses OAuth 2.0 Dynamic Client Registration Response
// from a reader containing JSON data as defined in RFC 7591
func parseClientRegistrationResponse(reader io.Reader) (clientRegistrationResponse, error) {
	var response clientRegistrationResponse
	if err := json.NewDecoder(reader).Decode(&response); err != nil {
		return response, fmt.Errorf("failed to decode client registration response: %w", err)
	}

	// Validate required fields
	if response.ClientID == "" {
		return response, fmt.Errorf("client_id is required but not provided")
	}

	return response, nil
}
