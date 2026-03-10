package artifacts

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)

func withWorkingDir(t *testing.T, dir string) func() {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to directory %s: %v", dir, err)
	}
	return func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}
}

func createTestWorkflow(t *testing.T, baseDir, name string, files map[string]string) {
	t.Helper()
	dir := filepath.Join(baseDir, skillformat.WorkflowsDir, name)
	for relPath, content := range files {
		fullPath := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", relPath, err)
		}
	}
}

func TestCreateZIP(t *testing.T) {
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	createTestWorkflow(t, tempDir, "my-wf", map[string]string{
		skillformat.SkillMainFile: "---\nname: my-wf\ndescription: desc\n---\n# Steps\n",
		"scripts/analyze.py":      "print('hello')\n",
	})

	workflowDir := filepath.Join(tempDir, skillformat.WorkflowsDir, "my-wf")
	zipData, err := createZIP(workflowDir)
	if err != nil {
		t.Fatalf("createZIP() error: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("failed to open ZIP: %v", err)
	}

	expectedFiles := map[string]bool{
		"SKILL.md":           false,
		"scripts/analyze.py": false,
	}

	for _, f := range r.File {
		if _, ok := expectedFiles[f.Name]; !ok {
			t.Errorf("unexpected file in ZIP: %s", f.Name)
		}
		expectedFiles[f.Name] = true
	}

	for name, found := range expectedFiles {
		if !found {
			t.Errorf("expected file not found in ZIP: %s", name)
		}
	}

	// Verify no manifest.yaml in the ZIP
	for _, f := range r.File {
		if f.Name == "manifest.yaml" {
			t.Error("ZIP should not contain manifest.yaml")
		}
	}
}

func TestPublishArtifact_MissingWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	s := NewServer()
	_, err := s.publishArtifact(nil, publishArtifactParams{WorkflowName: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing workflow, got nil")
	}
}

func TestPublishArtifact_EmptyName(t *testing.T) {
	s := NewServer()
	_, err := s.publishArtifact(nil, publishArtifactParams{})
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}
