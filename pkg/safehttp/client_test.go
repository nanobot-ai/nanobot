package safehttp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClientBlocksLoopbackLiteralIP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Fatal("request should have been blocked before reaching server")
	}))
	defer ts.Close()

	_, err := NewClient(true, false, false).Get(ts.URL)
	if err == nil {
		t.Fatal("expected loopback IP to be blocked")
	}
	if !strings.Contains(err.Error(), "blocked loopback IP") {
		t.Fatalf("expected loopback error, got %v", err)
	}
}

func TestClientBlocksLoopbackHostname(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Fatal("request should have been blocked before reaching server")
	}))
	defer ts.Close()

	_, err := NewClient(true, false, false).Get(strings.Replace(ts.URL, "127.0.0.1", "localhost", 1))
	if err == nil {
		t.Fatal("expected localhost to resolve to a blocked loopback IP")
	}
	if !strings.Contains(err.Error(), "blocked loopback IP") {
		t.Fatalf("expected loopback error, got %v", err)
	}
}

func TestClientBlocksPrivateIP(t *testing.T) {
	_, err := NewClient(false, true, false).Get("http://192.168.0.1/.well-known/oauth-protected-resource")
	if err == nil {
		t.Fatal("expected private IP to be blocked")
	}
	if !strings.Contains(err.Error(), "blocked private IP") {
		t.Fatalf("expected private IP error, got %v", err)
	}
}

func TestClientBlocksLinkLocalIP(t *testing.T) {
	_, err := NewClient(false, false, true).Get("http://169.254.169.254/latest/meta-data")
	if err == nil {
		t.Fatal("expected link-local IP to be blocked")
	}
	if !strings.Contains(err.Error(), "blocked link-local IP") {
		t.Fatalf("expected link-local IP error, got %v", err)
	}
}

func TestClientAllowsExplicitlyDisabledBlockedRanges(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	resp, err := NewClient(false, false, false).Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestClientAllowsBlockedIPWhenAllowListed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	serverURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := NewClientWithAllowList(true, false, false, []string{serverURL.Host}).Get(ts.URL)
	if err != nil {
		t.Fatalf("expected allow-listed loopback IP to be allowed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestClientAllowsBlockedHostnameWhenAllowListed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	serverURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	localURL := strings.Replace(ts.URL, "127.0.0.1", "localhost", 1)

	resp, err := NewClientWithAllowList(true, false, false, []string{"localhost:" + serverURL.Port()}).Get(localURL)
	if err != nil {
		t.Fatalf("expected allow-listed localhost to be allowed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestClientBlocksAllowListedHostnameWithMismatchedPort(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Fatal("request should have been blocked before reaching server")
	}))
	defer ts.Close()

	localURL := strings.Replace(ts.URL, "127.0.0.1", "localhost", 1)
	_, err := NewClientWithAllowList(true, false, false, []string{"localhost:1"}).Get(localURL)
	if err == nil {
		t.Fatal("expected localhost to be blocked when allow-list port mismatches")
	}
	if !strings.Contains(err.Error(), "blocked loopback IP") {
		t.Fatalf("expected loopback error, got %v", err)
	}
}

func TestAllowListMatchesExactAndSuffixHosts(t *testing.T) {
	dialer := safeDialer{
		allowList: parseAllowList([]string{
			"api.example.com",
			"*.internal.example.com",
			"db.example.com:8443",
		}),
	}

	tests := []struct {
		name string
		host string
		port string
		want bool
	}{
		{
			name: "exact host",
			host: "api.example.com",
			port: "443",
			want: true,
		},
		{
			name: "exact host is not suffix",
			host: "xapi.example.com",
			port: "443",
			want: false,
		},
		{
			name: "suffix host",
			host: "mcp.internal.example.com",
			port: "443",
			want: true,
		},
		{
			name: "suffix host nested",
			host: "a.b.internal.example.com",
			port: "443",
			want: true,
		},
		{
			name: "suffix does not match apex",
			host: "internal.example.com",
			port: "443",
			want: false,
		},
		{
			name: "port match",
			host: "db.example.com",
			port: "8443",
			want: true,
		},
		{
			name: "port mismatch",
			host: "db.example.com",
			port: "443",
			want: false,
		},
		{
			name: "case insensitive",
			host: "API.EXAMPLE.COM",
			port: "443",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dialer.isAllowed(tt.host, tt.port); got != tt.want {
				t.Fatalf("isAllowed(%q, %q) = %v, want %v", tt.host, tt.port, got, tt.want)
			}
		})
	}
}
