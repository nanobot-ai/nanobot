package responses

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/obot-platform/nanobot/pkg/types"
)

func TestCompleteUnsupportedReasoningEffortError(t *testing.T) {
	client := newErrorClient(t, `{
		"error": {
			"message": "Unsupported parameter: 'reasoning.effort' is not supported with this model.",
			"type": "invalid_request_error",
			"param": "reasoning.effort",
			"code": "unsupported_parameter"
		}
	}`)

	_, err := client.Complete(context.Background(), types.CompletionRequest{Model: "gpt-4.1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "non-reasoning models are currently unsupported") {
		t.Fatalf("expected non-reasoning model error, got %v", err)
	}
}

func TestCompleteUnsupportedEncryptedContentError(t *testing.T) {
	client := newErrorClient(t, `{
		"error": {
			"message": "Encrypted content is not supported with this model.",
			"type": "invalid_request_error",
			"param": "include",
			"code": null
		}
	}`)

	_, err := client.Complete(context.Background(), types.CompletionRequest{Model: "gpt-4.1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "non-reasoning models are currently unsupported") {
		t.Fatalf("expected non-reasoning model error, got %v", err)
	}
}

func newErrorClient(t *testing.T, body string) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return NewClient(Config{BaseURL: server.URL})
}
