package safehttp

import (
	"net/http"
	"net/http/httptest"
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
