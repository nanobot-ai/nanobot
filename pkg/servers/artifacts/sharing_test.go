package artifacts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

func artifactTestContext(baseURL string, env map[string]string) context.Context {
	session := &mcp.Session{}
	session.SetEnv(map[string]string{
		"OBOT_URL": baseURL,
	})
	session.AddEnv(env)
	return mcp.WithSession(context.Background(), session)
}

func TestNormalizeArtifactSubjects(t *testing.T) {
	subjects, err := normalizeArtifactSubjects([]artifactSubject{
		{Type: " USER ", ID: " alice "},
		{Type: "user", ID: "alice"},
		{Type: "group", ID: "eng"},
	})
	if err != nil {
		t.Fatalf("normalizeArtifactSubjects() error = %v", err)
	}

	if len(subjects) != 2 {
		t.Fatalf("expected 2 normalized subjects, got %d", len(subjects))
	}
	if subjects[0] != (artifactSubject{Type: "user", ID: "alice"}) {
		t.Fatalf("unexpected first subject: %+v", subjects[0])
	}
	if subjects[1] != (artifactSubject{Type: "group", ID: "eng"}) {
		t.Fatalf("unexpected second subject: %+v", subjects[1])
	}
}

func TestNormalizeArtifactSubjectsRejectsMixedSelector(t *testing.T) {
	_, err := normalizeArtifactSubjects([]artifactSubject{
		{Type: "selector", ID: "*"},
		{Type: "user", ID: "alice"},
	})
	if err == nil {
		t.Fatal("expected error for mixed selector subjects, got nil")
	}
}

func TestSetArtifactSubjects(t *testing.T) {
	var (
		gotAuthHeader string
		gotBody       map[string]any
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/published-artifacts/pa1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		gotAuthHeader = r.Header.Get("Authorization")
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed reading body: %v", err)
		}
		if err := json.Unmarshal(data, &gotBody); err != nil {
			t.Fatalf("failed decoding body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"pa1","name":"code-review","subjects":[{"type":"group","id":"eng"},{"type":"user","id":"alice"}]}`)
	}))
	defer ts.Close()

	s := NewServer()
	result, err := s.setArtifactSubjects(artifactTestContext(ts.URL, map[string]string{
		"MCP_API_KEY": "secret-token",
	}), setArtifactSubjectsParams{
		ID: "pa1",
		Subjects: []artifactSubject{
			{Type: "group", ID: "eng"},
			{Type: "user", ID: "alice"},
		},
	})
	if err != nil {
		t.Fatalf("setArtifactSubjects() error = %v", err)
	}

	if gotAuthHeader != "Bearer secret-token" {
		t.Fatalf("unexpected auth header: %q", gotAuthHeader)
	}
	if result.Message != "Updated sharing for code-review with 2 subject(s)." {
		t.Fatalf("unexpected message: %q", result.Message)
	}

	rawSubjects, ok := gotBody["subjects"].([]any)
	if !ok || len(rawSubjects) != 2 {
		t.Fatalf("unexpected request subjects payload: %#v", gotBody["subjects"])
	}
}

func TestSetArtifactSubjectsOwnerOnlyMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"pa1","name":"code-review","subjects":[]}`)
	}))
	defer ts.Close()

	s := NewServer()
	result, err := s.setArtifactSubjects(artifactTestContext(ts.URL, nil), setArtifactSubjectsParams{ID: "pa1"})
	if err != nil {
		t.Fatalf("setArtifactSubjects() error = %v", err)
	}

	if !strings.Contains(result.Message, "owner-only") {
		t.Fatalf("expected owner-only message, got %q", result.Message)
	}
}
