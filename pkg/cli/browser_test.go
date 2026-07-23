package cli

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMuxBrowserRoutesRequireExplicitOptIn(t *testing.T) {
	fallback := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusTeapot)
	})
	apiHandler := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNoContent)
	})

	for _, test := range []struct {
		name          string
		enableBrowser bool
		wantStatus    int
	}{
		{name: "disabled", wantStatus: http.StatusTeapot},
		{name: "enabled", enableBrowser: true, wantStatus: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := newHTTPMux(fallback, nil, apiHandler, test.enableBrowser)
			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/browser/healthz", nil))
			if rw.Code != test.wantStatus {
				t.Fatalf("got browser status %d, want %d", rw.Code, test.wantStatus)
			}
		})
	}
}

func TestHTTPMuxAlwaysRoutesAPIAndMCP(t *testing.T) {
	fallback := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusTeapot)
	})
	apiHandler := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNoContent)
	})
	handler := newHTTPMux(fallback, nil, apiHandler, false)

	for _, test := range []struct {
		path       string
		wantStatus int
	}{
		{path: "/api/events/thread", wantStatus: http.StatusNoContent},
		{path: "/mcp", wantStatus: http.StatusTeapot},
	} {
		t.Run(test.path, func(t *testing.T) {
			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, test.path, nil))
			if rw.Code != test.wantStatus {
				t.Fatalf("got status %d, want %d", rw.Code, test.wantStatus)
			}
		})
	}
}
