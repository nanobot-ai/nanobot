package installzip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)

func createZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close ZIP: %v", err)
	}
	return buf.Bytes()
}

func TestNormalizeArchiveSlashes(t *testing.T) {
	if got := normalizeArchiveSlashes(`nested\dir/SKILL.md`); got != "nested/dir/SKILL.md" {
		t.Fatalf("normalizeArchiveSlashes() = %q", got)
	}
}

func TestSanitizeArchivePathRejectsAbsoluteAndVolumePaths(t *testing.T) {
	tests := []string{
		"/tmp/SKILL.md",
		`C:\tmp\SKILL.md`,
		`C:/tmp/SKILL.md`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := sanitizeArchivePath(input)
			if err == nil {
				t.Fatal("expected sanitizeArchivePath() to fail")
			}
			if !strings.Contains(err.Error(), "absolute paths") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExtractRejectsWindowsStyleTraversal(t *testing.T) {
	zipData := createZIP(t, map[string]string{
		`..\evil.sh`: "echo bad\n",
	})

	_, err := Extract(zipData, t.TempDir())
	if err == nil {
		t.Fatal("expected Extract() to fail")
	}
	if !strings.Contains(err.Error(), "path traversal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureWithinBaseRejectsNonAbsoluteTargets(t *testing.T) {
	baseDir := t.TempDir()
	err := ensureWithinBase(baseDir, filepath.Join("relative", skillformat.SkillMainFile))
	if err == nil {
		t.Fatal("expected ensureWithinBase() to fail")
	}
	if !strings.Contains(err.Error(), "must be absolute") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractRejectsExcessiveFileCount(t *testing.T) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	// Create a SKILL.md entry so the archive is otherwise valid.
	entry, err := writer.Create(skillformat.SkillMainFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("---\nname: test\n---\n")); err != nil {
		t.Fatal(err)
	}

	// Add entries beyond the limit.
	for i := 0; i <= maxFileCount; i++ {
		entry, err := writer.Create(fmt.Sprintf("file_%d.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = Extract(buf.Bytes(), t.TempDir())
	if err == nil {
		t.Fatal("expected Extract() to fail for excessive file count")
	}
	if !strings.Contains(err.Error(), "exceeding maximum") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractEnforcesActualUncompressedSize(t *testing.T) {
	// Create a zip with a single large file that exceeds maxUncompressedBytes.
	// We use a store (no compression) method with content just over the limit.
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	// Add SKILL.md first.
	header := &zip.FileHeader{
		Name:   skillformat.SkillMainFile,
		Method: zip.Store,
	}
	entry, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("---\nname: test\n---\n")); err != nil {
		t.Fatal(err)
	}

	// Add a file whose actual content exceeds the limit.
	bigHeader := &zip.FileHeader{
		Name:   "big.bin",
		Method: zip.Store,
	}
	bigEntry, err := writer.CreateHeader(bigHeader)
	if err != nil {
		t.Fatal(err)
	}
	// Write in chunks to avoid allocating a huge slice.
	chunk := make([]byte, 1024*1024) // 1 MB
	for written := 0; written <= maxUncompressedBytes; written += len(chunk) {
		if _, err := bigEntry.Write(chunk); err != nil {
			t.Fatal(err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = Extract(buf.Bytes(), t.TempDir())
	if err == nil {
		t.Fatal("expected Extract() to fail for oversized uncompressed content")
	}
	if !strings.Contains(err.Error(), "maximum uncompressed size") {
		t.Fatalf("unexpected error: %v", err)
	}
}
