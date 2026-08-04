package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/obot-platform/nanobot/pkg/safehttp"
	"golang.org/x/oauth2"
)

func TestGetOAuthMetadata(t *testing.T) {
	var serverURL string
	var pathMetadataRequested, rootMetadataRequested atomic.Bool
	const (
		clientName  = "Test Client"
		redirectURL = "http://localhost/callback"
	)
	protectedResourceMetadata := json.RawMessage(`{"resource":"resource","authorization_servers":["issuer"],"scopes_supported":["read"]}`)
	authorizationServerMetadata := json.RawMessage(`{"issuer":"issuer","authorization_endpoint":"authorize","token_endpoint":"token","registration_endpoint":"register","response_types_supported":["code"],"client_id_metadata_document_supported":true}`)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/mcp":
			if req.Method != http.MethodPost {
				http.NotFound(w, req)
				return
			}
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case "/.well-known/oauth-protected-resource/mcp":
			pathMetadataRequested.Store(true)
			http.NotFound(w, req)
		case "/.well-known/oauth-protected-resource":
			rootMetadataRequested.Store(true)
			if req.Header.Get("X-Test") != "value" {
				http.Error(w, "missing test header", http.StatusBadRequest)
				return
			}
			metadata := map[string]any{}
			if err := json.Unmarshal(protectedResourceMetadata, &metadata); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			metadata["resource"] = serverURL
			metadata["authorization_servers"] = []string{serverURL + "/issuer"}
			_ = json.NewEncoder(w).Encode(metadata)
		case "/.well-known/oauth-authorization-server/issuer":
			if req.Header.Get("X-Test") != "value" {
				http.Error(w, "missing test header", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(authorizationServerMetadata)
		default:
			http.NotFound(w, req)
		}
	}))
	defer ts.Close()
	serverURL = ts.URL

	result, err := GetOAuthMetadataWithClient(t.Context(), safehttp.NewClient(false, true, true), Server{
		BaseURL: ts.URL + "/mcp",
		Headers: map[string]string{
			"X-Test": "value",
		},
	}, clientName, redirectURL)
	if err != nil {
		t.Fatal(err)
	}

	if result.ProtectedResourceMetadataURL != ts.URL+"/.well-known/oauth-protected-resource" {
		t.Fatalf("unexpected protected resource URL: %s", result.ProtectedResourceMetadataURL)
	}
	if !pathMetadataRequested.Load() || !rootMetadataRequested.Load() {
		t.Fatalf("expected path and root protected resource metadata URLs to be requested")
	}
	if result.AuthorizationServerMetadataURL != ts.URL+"/.well-known/oauth-authorization-server/issuer" {
		t.Fatalf("unexpected authorization server URL: %s", result.AuthorizationServerMetadataURL)
	}
	if len(result.ProtectedResourceMetadata) == 0 {
		t.Fatalf("expected protected resource metadata")
	}
	if string(result.AuthorizationServerMetadata) != string(authorizationServerMetadata) {
		t.Fatalf("unexpected authorization server metadata: %s", result.AuthorizationServerMetadata)
	}
	if !result.DynamicClientRegistration {
		t.Fatalf("expected dynamic client registration support")
	}
	if !result.ClientIDMetadataDocumentSupported {
		t.Fatalf("expected client ID metadata document support")
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal oauth metadata: %v", err)
	}
	var resultFields map[string]json.RawMessage
	if err := json.Unmarshal(resultJSON, &resultFields); err != nil {
		t.Fatalf("failed to parse oauth metadata: %v", err)
	}
	if string(resultFields["clientIdMetadataDocumentSupported"]) != "true" {
		t.Fatalf("expected clientIdMetadataDocumentSupported JSON field, got %s", resultFields["clientIdMetadataDocumentSupported"])
	}
	var clientRegistration ClientRegistrationMetadata
	if err := json.Unmarshal(result.ClientRegistration, &clientRegistration); err != nil {
		t.Fatalf("failed to parse client registration metadata: %v", err)
	}
	if clientRegistration.ClientName != clientName {
		t.Fatalf("unexpected client name: %s", clientRegistration.ClientName)
	}
	if len(clientRegistration.RedirectURIs) != 1 || clientRegistration.RedirectURIs[0] != redirectURL {
		t.Fatalf("unexpected redirect URIs: %v", clientRegistration.RedirectURIs)
	}
	if clientRegistration.Scope != "read" {
		t.Fatalf("unexpected scope: %s", clientRegistration.Scope)
	}
	if !slices.Equal(clientRegistration.GrantTypes, []string{"authorization_code"}) {
		t.Fatalf("unexpected grant types: %v", clientRegistration.GrantTypes)
	}
}

func TestOAuthResourceMetadataURLs(t *testing.T) {
	tests := []struct {
		name               string
		baseURL            string
		authenticateHeader string
		wantURLs           []string
		wantResources      []string
		wantScope          string
	}{
		{
			name:    "defaults to path-specific then root metadata without an auth header",
			baseURL: "https://mcp.example.com/mcp",
			wantURLs: []string{
				"https://mcp.example.com/.well-known/oauth-protected-resource/mcp",
				"https://mcp.example.com/.well-known/oauth-protected-resource",
			},
			wantResources: []string{
				"https://mcp.example.com/mcp",
				"https://mcp.example.com",
			},
		},
		{
			name:               "retains challenge scope for default metadata URLs",
			baseURL:            "https://mcp.example.com/v1/mcp",
			authenticateHeader: `Bearer scope="read write"`,
			wantURLs: []string{
				"https://mcp.example.com/.well-known/oauth-protected-resource/v1/mcp",
				"https://mcp.example.com/.well-known/oauth-protected-resource",
			},
			wantResources: []string{
				"https://mcp.example.com/v1/mcp",
				"https://mcp.example.com",
			},
			wantScope: "read write",
		},
		{
			name:               "uses advertised resource metadata URL exclusively",
			baseURL:            "https://mcp.example.com/mcp",
			authenticateHeader: `Bearer resource_metadata="https://auth.example.com/resources/mcp" scope="read"`,
			wantURLs:           []string{"https://auth.example.com/resources/mcp"},
			wantResources:      []string{"https://mcp.example.com/mcp"},
			wantScope:          "read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls, scope, err := oauthResourceMetadataURLs(tt.baseURL, tt.authenticateHeader)
			if err != nil {
				t.Fatal(err)
			}

			gotURLs := make([]string, len(urls))
			gotResources := make([]string, len(urls))
			for i, u := range urls {
				gotURLs[i] = u.url.String()
				gotResources[i] = u.expectedResource
			}
			if !slices.Equal(gotURLs, tt.wantURLs) {
				t.Fatalf("unexpected resource metadata URLs: got %v, want %v", gotURLs, tt.wantURLs)
			}
			if !slices.Equal(gotResources, tt.wantResources) {
				t.Fatalf("unexpected expected resources: got %v, want %v", gotResources, tt.wantResources)
			}
			if scope != tt.wantScope {
				t.Fatalf("unexpected scope: got %q, want %q", scope, tt.wantScope)
			}
		})
	}
}

func TestDiscoverOAuthMetadataValidatesResourceIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		discovery string
		valid     bool
	}{
		{name: "path-specific valid", discovery: "path", valid: true},
		{name: "path-specific mismatch", discovery: "path"},
		{name: "root fallback valid", discovery: "root", valid: true},
		{name: "root fallback mismatch", discovery: "root"},
		{name: "WWW-Authenticate valid", discovery: "header", valid: true},
		{name: "WWW-Authenticate mismatch", discovery: "header"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverURL string
			var pathMetadataRequested, rootMetadataRequested, headerMetadataRequested, authorizationServerRequested atomic.Bool

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				baseURL := serverURL + "/mcp"
				resource := baseURL
				if tt.discovery == "root" {
					resource = serverURL
				}
				if !tt.valid {
					resource += "/mismatch"
				}

				switch req.URL.Path {
				case "/.well-known/oauth-protected-resource/mcp":
					pathMetadataRequested.Store(true)
					if tt.discovery == "root" {
						http.NotFound(w, req)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{
						"resource":              resource,
						"authorization_servers": []string{serverURL + "/issuer"},
					})
				case "/.well-known/oauth-protected-resource":
					rootMetadataRequested.Store(true)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"resource":              resource,
						"authorization_servers": []string{serverURL + "/issuer"},
					})
				case "/advertised-metadata":
					headerMetadataRequested.Store(true)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"resource":              resource,
						"authorization_servers": []string{serverURL + "/issuer"},
					})
				case "/.well-known/oauth-authorization-server/issuer":
					authorizationServerRequested.Store(true)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"issuer":                   serverURL + "/issuer",
						"response_types_supported": []string{"code"},
					})
				default:
					http.NotFound(w, req)
				}
			}))
			defer ts.Close()
			serverURL = ts.URL

			baseURL := serverURL + "/mcp"
			authenticateHeader := ""
			if tt.discovery == "header" {
				authenticateHeader = `Bearer resource_metadata="` + serverURL + `/advertised-metadata"`
			}

			discovery, ok, err := discoverOAuthMetadata(t.Context(), safehttp.NewClient(false, true, true), baseURL, authenticateHeader, "", "", nil)
			if tt.valid {
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal("expected OAuth metadata discovery to succeed")
				}
				if discovery.ProtectedResourceMetadata.Resource == "" {
					t.Fatal("expected protected resource metadata")
				}
			} else {
				if err == nil {
					t.Fatal("expected resource identifier mismatch")
				}
				if ok {
					t.Fatal("mismatched resource identifier must not produce metadata")
				}
			}

			if authorizationServerRequested.Load() != tt.valid {
				t.Fatalf("authorization server requested = %v, want %v", authorizationServerRequested.Load(), tt.valid)
			}
			if pathMetadataRequested.Load() != (tt.discovery != "header") {
				t.Fatalf("path metadata requested = %v for %s discovery", pathMetadataRequested.Load(), tt.discovery)
			}
			if rootMetadataRequested.Load() != (tt.discovery == "root") {
				t.Fatalf("root metadata requested = %v for %s discovery", rootMetadataRequested.Load(), tt.discovery)
			}
			if headerMetadataRequested.Load() != (tt.discovery == "header") {
				t.Fatalf("advertised metadata requested = %v for %s discovery", headerMetadataRequested.Load(), tt.discovery)
			}
		})
	}
}

func TestAuthServerMetadataToClientRegistrationFiltersGrantTypes(t *testing.T) {
	tests := []struct {
		name      string
		supported []string
		want      []string
	}{
		{
			name:      "keeps only authorization code and refresh token",
			supported: []string{"client_credentials", "refresh_token", "authorization_code", "implicit"},
			want:      []string{"authorization_code", "refresh_token"},
		},
		{
			name:      "omits unsupported grant types",
			supported: []string{"client_credentials", "implicit"},
		},
		{
			name:      "keeps refresh token when advertised",
			supported: []string{"refresh_token"},
			want:      []string{"refresh_token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientRegistration := AuthServerMetadataToClientRegistration(AuthorizationServerMetadata{
				GrantTypesSupported: tt.supported,
			}, "", "", "")
			if !slices.Equal(clientRegistration.GrantTypes, tt.want) {
				t.Fatalf("unexpected grant types: got %v, want %v", clientRegistration.GrantTypes, tt.want)
			}
		})
	}
}

func TestGetOAuthMetadataMissingProtectedResource(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	result, err := GetOAuthMetadataWithClient(t.Context(), safehttp.NewClient(false, true, true), Server{BaseURL: ts.URL}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.ProtectedResourceMetadataURL != "" || len(result.ProtectedResourceMetadata) != 0 {
		t.Fatalf("expected empty result for missing protected resource metadata: %#v", result)
	}
}

func TestGetOAuthMetadataInitializeSuccessDeletesSession(t *testing.T) {
	var deleted, metadataFetched atomic.Bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/mcp":
			w.Header().Set(SessionIDHeader, "session-1")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  map[string]any{},
			})
		case req.Method == http.MethodDelete && req.URL.Path == "/mcp":
			if req.Header.Get(SessionIDHeader) != "session-1" {
				http.Error(w, "missing session id", http.StatusBadRequest)
				return
			}
			deleted.Store(true)
			w.WriteHeader(http.StatusAccepted)
		case req.URL.Path == "/.well-known/oauth-protected-resource":
			metadataFetched.Store(true)
			http.Error(w, "metadata should not be fetched after successful initialize", http.StatusInternalServerError)
		default:
			http.NotFound(w, req)
		}
	}))
	defer ts.Close()

	result, err := GetOAuthMetadataWithClient(t.Context(), safehttp.NewClient(false, true, true), Server{BaseURL: ts.URL + "/mcp"}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.ProtectedResourceMetadataURL != "" || len(result.ProtectedResourceMetadata) != 0 {
		t.Fatalf("expected empty result after successful initialize: %#v", result)
	}
	if !deleted.Load() {
		t.Fatalf("expected successful initialize session to be deleted")
	}
	if metadataFetched.Load() {
		t.Fatalf("metadata should not be fetched after successful initialize")
	}
}

func TestGetOAuthMetadataAuthorizationServerNoRegistration(t *testing.T) {
	var serverURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/.well-known/oauth-protected-resource":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resource": serverURL,
			})
		case "/.well-known/oauth-authorization-server":
			http.NotFound(w, req)
		case "/.well-known/openid-configuration":
			_, _ = w.Write([]byte(`{"issuer":"issuer","authorization_endpoint":"authorize","token_endpoint":"token","response_types_supported":["code"]}`))
		default:
			http.NotFound(w, req)
		}
	}))
	defer ts.Close()
	serverURL = ts.URL

	result, err := GetOAuthMetadataWithClient(t.Context(), safehttp.NewClient(false, true, true), Server{BaseURL: ts.URL}, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if result.AuthorizationServerMetadataURL != ts.URL+"/.well-known/openid-configuration" {
		t.Fatalf("unexpected authorization server fallback URL: %s", result.AuthorizationServerMetadataURL)
	}
	if result.DynamicClientRegistration {
		t.Fatalf("expected no dynamic client registration support")
	}
}

type testClientCredLookup struct {
	clientID     string
	clientSecret string
	calls        int
}

func (l *testClientCredLookup) Lookup(context.Context, string) (string, string, error) {
	l.calls++
	return l.clientID, l.clientSecret, nil
}

func TestResolveClientInfoUsesClientIDMetadataDocument(t *testing.T) {
	lookup := &testClientCredLookup{
		clientID:     "static-client-id",
		clientSecret: "static-client-secret",
	}
	o := &oauth{
		clientIDMetadataDocument: "https://client.example/oauth-client-metadata.json",
		clientLookup:             lookup,
	}

	clientInfo, err := o.resolveClientInfo(t.Context(), "test-server", oauthMetadataDiscovery{
		ProtectedResourceMetadata: protectedResourceMetadata{
			AuthorizationServers: []string{"https://issuer.example"},
		},
		AuthorizationServerMetadata: AuthorizationServerMetadata{
			ClientIDMetadataDocumentSupported: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if clientInfo.ClientID != o.clientIDMetadataDocument {
		t.Fatalf("expected metadata document client ID %q, got %q", o.clientIDMetadataDocument, clientInfo.ClientID)
	}
	if clientInfo.ClientSecret != "" {
		t.Fatalf("expected empty client secret, got %q", clientInfo.ClientSecret)
	}
	if lookup.calls != 0 {
		t.Fatalf("static client lookup should not be called, got %d calls", lookup.calls)
	}
}

func TestResolveClientInfoFallsBackWhenClientIDMetadataDocumentUnsupported(t *testing.T) {
	lookup := &testClientCredLookup{
		clientID:     "static-client-id",
		clientSecret: "static-client-secret",
	}
	o := &oauth{
		clientIDMetadataDocument: "https://client.example/oauth-client-metadata.json",
		clientLookup:             lookup,
	}

	clientInfo, err := o.resolveClientInfo(t.Context(), "test-server", oauthMetadataDiscovery{
		ProtectedResourceMetadata: protectedResourceMetadata{
			AuthorizationServers: []string{"https://issuer.example"},
		},
		AuthorizationServerMetadata: AuthorizationServerMetadata{
			ClientIDMetadataDocumentSupported: false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if clientInfo.ClientID != lookup.clientID {
		t.Fatalf("expected static client ID %q, got %q", lookup.clientID, clientInfo.ClientID)
	}
	if clientInfo.ClientSecret != lookup.clientSecret {
		t.Fatalf("expected static client secret %q, got %q", lookup.clientSecret, clientInfo.ClientSecret)
	}
	if lookup.calls != 1 {
		t.Fatalf("expected one static client lookup, got %d calls", lookup.calls)
	}
}

func TestTokenEndpointAuthStyleUsesParamsWithoutClientSecret(t *testing.T) {
	if got := tokenEndpointAuthStyle("client_secret_basic", false); got != oauth2.AuthStyleInParams {
		t.Fatalf("expected params auth style without client secret, got %v", got)
	}
}

func TestTokenEndpointAuthStyleHonorsClientSecretMethods(t *testing.T) {
	tests := []struct {
		method string
		want   oauth2.AuthStyle
	}{
		{method: "client_secret_basic", want: oauth2.AuthStyleInHeader},
		{method: "client_secret_post", want: oauth2.AuthStyleInParams},
		{method: "", want: oauth2.AuthStyleAutoDetect},
	}

	for _, tt := range tests {
		if got := tokenEndpointAuthStyle(tt.method, true); got != tt.want {
			t.Fatalf("expected auth style %v for method %q, got %v", tt.want, tt.method, got)
		}
	}
}
