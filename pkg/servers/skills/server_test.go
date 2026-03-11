package skills

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)

func testContext(t *testing.T, env map[string]string) context.Context {
	t.Helper()

	ctx := context.Background()
	session := mcp.NewEmptySession(ctx)
	session.Set(mcp.SessionEnvMapKey, env)
	return mcp.WithSession(ctx, session)
}

func createSkillZIP(t *testing.T, fm skillformat.Frontmatter, body string, files map[string]string) []byte {
	t.Helper()

	content, err := skillformat.FormatSkillMD(fm, body)
	if err != nil {
		t.Fatalf("failed to format SKILL.md: %v", err)
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	entry, err := writer.Create(skillformat.SkillMainFile)
	if err != nil {
		t.Fatalf("failed to create SKILL.md entry: %v", err)
	}
	if _, err := entry.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	for name, fileContent := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
		if _, err := entry.Write([]byte(fileContent)); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close ZIP writer: %v", err)
	}

	return buf.Bytes()
}

func TestInstallSkill(t *testing.T) {
	zipData := createSkillZIP(t, skillformat.Frontmatter{
		Name:        "postgres-helper",
		Description: "Postgres utilities",
	}, "\n# Postgres\n", map[string]string{
		"scripts/run.sh": "echo hi\n",
	})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/skills/sk1":
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"id":"sk1","name":"postgres-helper","displayName":"Postgres Helper","description":"Postgres utilities","repoID":"repo-1","commitSHA":"abc123"}`))
		case "/api/skills/sk1/download":
			rw.Header().Set("Content-Type", "application/zip")
			_, _ = rw.Write(zipData)
		default:
			http.NotFound(rw, req)
		}
	}))
	defer server.Close()

	workDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(cwd)

	s := NewServer()
	ctx := testContext(t, map[string]string{"OBOT_URL": server.URL})
	result, err := s.installSkill(ctx, installSkillParams{SkillID: "sk1"})
	if err != nil {
		t.Fatalf("installSkill() failed: %v", err)
	}

	if result.CommitSHA != "abc123" {
		t.Fatalf("commitSHA = %q", result.CommitSHA)
	}
	if len(result.InstalledFiles) != 2 {
		t.Fatalf("installedFiles = %v", result.InstalledFiles)
	}
	if _, err := os.Stat(filepath.Join(workDir, "skills", "postgres-helper", "SKILL.md")); err != nil {
		t.Fatalf("expected installed SKILL.md: %v", err)
	}
}

func TestInstallSkillRejectsInvalidArchive(t *testing.T) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	entry, err := writer.Create("../evil.sh")
	if err != nil {
		t.Fatalf("failed to create ZIP entry: %v", err)
	}
	if _, err := entry.Write([]byte("echo bad\n")); err != nil {
		t.Fatalf("failed to write ZIP entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close ZIP writer: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/skills/sk1":
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"id":"sk1","name":"postgres-helper","displayName":"Postgres Helper","description":"Postgres utilities","repoID":"repo-1","commitSHA":"abc123"}`))
		case "/api/skills/sk1/download":
			rw.Header().Set("Content-Type", "application/zip")
			_, _ = rw.Write(buf.Bytes())
		default:
			http.NotFound(rw, req)
		}
	}))
	defer server.Close()

	workDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(cwd)

	s := NewServer()
	ctx := testContext(t, map[string]string{"OBOT_URL": server.URL})
	_, err = s.installSkill(ctx, installSkillParams{SkillID: "sk1"})
	if err == nil {
		t.Fatal("expected installSkill() to fail")
	}
	if !strings.Contains(err.Error(), "path traversal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstallSkillOverwriteCanceled(t *testing.T) {
	zipData := createSkillZIP(t, skillformat.Frontmatter{
		Name:        "postgres-helper",
		Description: "Postgres utilities",
	}, "\n# Updated\n", nil)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/skills/sk1":
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"id":"sk1","name":"postgres-helper","displayName":"Postgres Helper","description":"Postgres utilities","repoID":"repo-1","commitSHA":"abc123"}`))
		case "/api/skills/sk1/download":
			rw.Header().Set("Content-Type", "application/zip")
			_, _ = rw.Write(zipData)
		default:
			http.NotFound(rw, req)
		}
	}))
	defer server.Close()

	workDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(cwd)

	if err := os.MkdirAll(filepath.Join("skills", "postgres-helper"), 0o755); err != nil {
		t.Fatalf("failed to create existing skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join("skills", "postgres-helper", "SKILL.md"), []byte("original"), 0o644); err != nil {
		t.Fatalf("failed to write existing skill: %v", err)
	}

	s := NewServer()
	s.confirmOverwrite = func(context.Context, string) (bool, error) {
		return false, nil
	}

	ctx := testContext(t, map[string]string{"OBOT_URL": server.URL})
	result, err := s.installSkill(ctx, installSkillParams{SkillID: "sk1"})
	if err != nil {
		t.Fatalf("installSkill() failed: %v", err)
	}
	if !strings.Contains(result.Message, "canceled") {
		t.Fatalf("unexpected message: %s", result.Message)
	}

	content, err := os.ReadFile(filepath.Join("skills", "postgres-helper", "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read existing skill: %v", err)
	}
	if string(content) != "original" {
		t.Fatalf("expected existing skill to remain unchanged, got %q", string(content))
	}
}
