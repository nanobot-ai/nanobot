package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"golang.org/x/oauth2"
)

func TestGetOAuthMetadata(t *testing.T) {
	var serverURL string
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
			w.Header().Set("WWW-Authenticate", `Bearer resource_metadata="`+serverURL+`/.well-known/oauth-protected-resource/mcp"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case "/.well-known/oauth-protected-resource/mcp":
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

	result, err := GetOAuthMetadata(context.Background(), Server{
		BaseURL: ts.URL + "/mcp",
		Headers: map[string]string{
			"X-Test": "value",
		},
	}, clientName, redirectURL)
	if err != nil {
		t.Fatal(err)
	}

	if result.ProtectedResourceMetadataURL != ts.URL+"/.well-known/oauth-protected-resource/mcp" {
		t.Fatalf("unexpected protected resource URL: %s", result.ProtectedResourceMetadataURL)
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
}

func TestGetOAuthMetadataMissingProtectedResource(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	result, err := GetOAuthMetadata(context.Background(), Server{BaseURL: ts.URL}, "", "")
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

	result, err := GetOAuthMetadata(context.Background(), Server{BaseURL: ts.URL + "/mcp"}, "", "")
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

	result, err := GetOAuthMetadata(context.Background(), Server{BaseURL: ts.URL}, "", "")
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

	clientInfo, err := o.resolveClientInfo(context.Background(), "test-server", oauthMetadataDiscovery{
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

	clientInfo, err := o.resolveClientInfo(context.Background(), "test-server", oauthMetadataDiscovery{
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
