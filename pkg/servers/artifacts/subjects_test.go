package artifacts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListSubjectsUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"items":[{"id":"1","displayName":"Alice Smith","username":"alice","email":"alice@example.com"},{"id":"2","displayName":"Bob Jones","username":"bob","email":"bob@example.com"}]}`)
	}))
	defer ts.Close()

	s := NewServer()
	result, err := s.listSubjects(artifactTestContext(ts.URL, nil), listSubjectsParams{
		Type:  "user",
		Query: "alice",
	})
	if err != nil {
		t.Fatalf("listSubjects() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}
	if result.Items[0].ID != "1" {
		t.Fatalf("unexpected user id: %q", result.Items[0].ID)
	}
}

func TestListSubjectsGroups(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/groups" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"id":"okta:eng","name":"Engineering"},{"id":"okta:sales","name":"Sales"}]`)
	}))
	defer ts.Close()

	s := NewServer()
	result, err := s.listSubjects(artifactTestContext(ts.URL, nil), listSubjectsParams{
		Type:  "group",
		Query: "eng",
	})
	if err != nil {
		t.Fatalf("listSubjects() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}
	if result.Items[0].ID != "okta:eng" {
		t.Fatalf("unexpected group id: %q", result.Items[0].ID)
	}
}

func TestListSubjectsRejectsInvalidType(t *testing.T) {
	s := NewServer()
	_, err := s.listSubjects(artifactTestContext("https://example.com", nil), listSubjectsParams{Type: "team"})
	if err == nil {
		t.Fatal("expected error for invalid type, got nil")
	}
}
