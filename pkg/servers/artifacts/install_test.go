package artifacts

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/servers/installzip"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)

func createTestZIP(t *testing.T, fm skillformat.Frontmatter, body string, files map[string]string) []byte {
	t.Helper()
	content, err := skillformat.FormatSkillMD(fm, body)
	if err != nil {
		t.Fatalf("failed to format SKILL.md: %v", err)
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	fw, err := w.Create(skillformat.SkillMainFile)
	if err != nil {
		t.Fatalf("failed to create SKILL.md entry: %v", err)
	}
	if _, err := fw.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	for name, fileContent := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create entry %s: %v", name, err)
		}
		if _, err := fw.Write([]byte(fileContent)); err != nil {
			t.Fatalf("failed to write entry %s: %v", name, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close ZIP: %v", err)
	}
	return buf.Bytes()
}

func TestExtractZIP(t *testing.T) {
	fm := skillformat.Frontmatter{
		Name:        "test-wf",
		Description: "A test workflow.",
	}
	zipData := createTestZIP(t, fm, "# Test WF\n", map[string]string{
		"scripts/run.sh": "#!/bin/bash\necho hi\n",
	})

	targetDir := filepath.Join(t.TempDir(), "extracted")
	installed, err := installzip.Extract(zipData, targetDir)
	if err != nil {
		t.Fatalf("installzip.Extract() error: %v", err)
	}

	if len(installed) != 2 {
		t.Errorf("installed = %v, want 2 files (SKILL.md + scripts/run.sh)", installed)
	}

	content, err := os.ReadFile(filepath.Join(targetDir, skillformat.SkillMainFile))
	if err != nil {
		t.Fatalf("failed to read extracted SKILL.md: %v", err)
	}
	if !strings.Contains(string(content), "test-wf") {
		t.Errorf("SKILL.md content should contain skill name, got %q", string(content))
	}
}

func TestExtractZIP_NestedDirectories(t *testing.T) {
	fm := skillformat.Frontmatter{
		Name:        "nested-wf",
		Description: "A nested workflow.",
	}
	zipData := createTestZIP(t, fm, "# Nested\n", map[string]string{
		"scripts/analyze.py":  "print('hi')\n",
		"scripts/lib/util.py": "pass\n",
	})

	targetDir := filepath.Join(t.TempDir(), "extracted")
	installed, err := installzip.Extract(zipData, targetDir)
	if err != nil {
		t.Fatalf("installzip.Extract() error: %v", err)
	}

	if len(installed) != 3 {
		t.Errorf("installed count = %d, want 3", len(installed))
	}

	content, err := os.ReadFile(filepath.Join(targetDir, "scripts", "lib", "util.py"))
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}
	if string(content) != "pass\n" {
		t.Errorf("content = %q, want %q", string(content), "pass\n")
	}
}

func TestExtractZIP_IncludesSkillMD(t *testing.T) {
	fm := skillformat.Frontmatter{
		Name:        "with-skill",
		Description: "A workflow with SKILL.md.",
	}
	zipData := createTestZIP(t, fm, "# Test\n", nil)

	targetDir := filepath.Join(t.TempDir(), "extracted")
	installed, err := installzip.Extract(zipData, targetDir)
	if err != nil {
		t.Fatalf("installzip.Extract() error: %v", err)
	}

	var foundSkill bool
	for _, f := range installed {
		if f == skillformat.SkillMainFile {
			foundSkill = true
		}
	}
	if !foundSkill {
		t.Error("SKILL.md should be in installed files")
	}

	if _, err := os.Stat(filepath.Join(targetDir, skillformat.SkillMainFile)); err != nil {
		t.Errorf("SKILL.md should exist on disk: %v", err)
	}
}

func TestReadFrontmatterFromZIP(t *testing.T) {
	fm := skillformat.Frontmatter{
		Name:        "read-test",
		Description: "A test workflow.",
	}
	zipData := createTestZIP(t, fm, "# Test\n", nil)

	got, err := installzip.ReadFrontmatter(zipData)
	if err != nil {
		t.Fatalf("installzip.ReadFrontmatter() error: %v", err)
	}
	if got.Name != "read-test" {
		t.Errorf("name = %q, want %q", got.Name, "read-test")
	}
	if got.Description != "A test workflow." {
		t.Errorf("description = %q, want %q", got.Description, "A test workflow.")
	}
}

func TestReadFrontmatterFromZIP_MissingSKILLMD(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fw, _ := w.Create("other.txt")
	fw.Write([]byte("hello"))
	w.Close()

	_, err := installzip.ReadFrontmatter(buf.Bytes())
	if err == nil {
		t.Fatal("expected error for missing SKILL.md")
	}
	if !strings.Contains(err.Error(), "SKILL.md not found") {
		t.Errorf("error = %q, want containing %q", err.Error(), "SKILL.md not found")
	}
}

func TestReadFrontmatterFromZIP_InvalidFrontmatter(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fw, _ := w.Create(skillformat.SkillMainFile)
	fw.Write([]byte("---\nbad yaml: [\n---\n"))
	w.Close()

	_, err := installzip.ReadFrontmatter(buf.Bytes())
	if err == nil {
		t.Fatal("expected error for invalid frontmatter")
	}
}

func TestReadFrontmatterFromZIP_InvalidZIP(t *testing.T) {
	_, err := installzip.ReadFrontmatter([]byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid ZIP")
	}
	if !strings.Contains(err.Error(), "invalid ZIP archive") {
		t.Errorf("error = %q, want containing %q", err.Error(), "invalid ZIP archive")
	}
}
