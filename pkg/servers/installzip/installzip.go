package installzip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)

const (
	MaxArchiveBytes      = 100 * 1024 * 1024
	maxUncompressedBytes = 100 * 1024 * 1024
	maxFileCount         = 100
)

func ReadAll(r io.Reader, label string) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxArchiveBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s data: %w", label, err)
	}
	if len(data) > MaxArchiveBytes {
		return nil, fmt.Errorf("%s exceeds maximum size of %d bytes", label, MaxArchiveBytes)
	}
	return data, nil
}

func ReadFrontmatter(data []byte) (skillformat.Frontmatter, error) {
	reader, err := open(data)
	if err != nil {
		return skillformat.Frontmatter{}, err
	}

	for _, file := range reader.File {
		name, err := sanitizeArchivePath(file.Name)
		if err != nil {
			return skillformat.Frontmatter{}, err
		}
		if name != skillformat.SkillMainFile {
			continue
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return skillformat.Frontmatter{}, fmt.Errorf("%s must not be a symbolic link", skillformat.SkillMainFile)
		}

		rc, err := file.Open()
		if err != nil {
			return skillformat.Frontmatter{}, fmt.Errorf("failed to open %s: %w", skillformat.SkillMainFile, err)
		}

		content, readErr := io.ReadAll(io.LimitReader(rc, MaxArchiveBytes+1))
		closeErr := rc.Close()
		if readErr != nil {
			return skillformat.Frontmatter{}, fmt.Errorf("failed to read %s: %w", skillformat.SkillMainFile, readErr)
		}
		if closeErr != nil {
			return skillformat.Frontmatter{}, fmt.Errorf("failed to close %s: %w", skillformat.SkillMainFile, closeErr)
		}
		if len(content) > MaxArchiveBytes {
			return skillformat.Frontmatter{}, fmt.Errorf("%s exceeds maximum size of %d bytes", skillformat.SkillMainFile, MaxArchiveBytes)
		}

		fm, _, err := skillformat.ParseAndValidateFrontmatter(string(content))
		if err != nil {
			return skillformat.Frontmatter{}, fmt.Errorf("invalid %s: %w", skillformat.SkillMainFile, err)
		}
		return fm, nil
	}

	return skillformat.Frontmatter{}, fmt.Errorf("%s not found in ZIP", skillformat.SkillMainFile)
}

func Extract(data []byte, targetDir string) ([]string, error) {
	reader, err := open(data)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	baseDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target directory: %w", err)
	}

	if len(reader.File) > maxFileCount {
		return nil, fmt.Errorf("archive contains %d entries, exceeding maximum of %d", len(reader.File), maxFileCount)
	}

	var (
		installed      []string
		foundSkillMain bool
		totalWritten   uint64
	)

	for _, file := range reader.File {
		name, err := sanitizeArchivePath(file.Name)
		if err != nil {
			return nil, err
		}
		if name == "." {
			continue
		}

		if file.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("symbolic links are not allowed in ZIP contents: %s", name)
		}

		destPath := filepath.Join(baseDir, filepath.FromSlash(name))
		if err := ensureWithinBase(baseDir, destPath); err != nil {
			return nil, err
		}

		if strings.HasSuffix(file.Name, "/") || file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", name, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create parent directory for %s: %w", name, err)
		}

		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open ZIP entry %s: %w", name, err)
		}

		out, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, file.Mode().Perm())
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("failed to create %s: %w", name, err)
		}

		remaining := maxUncompressedBytes - totalWritten
		written, copyErr := io.Copy(out, io.LimitReader(rc, int64(remaining)+1))
		totalWritten += uint64(written)
		closeOutErr := out.Close()
		closeInErr := rc.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", name, copyErr)
		}
		if totalWritten > maxUncompressedBytes {
			return nil, fmt.Errorf("archive exceeds maximum uncompressed size of %d bytes", maxUncompressedBytes)
		}
		if closeOutErr != nil {
			return nil, fmt.Errorf("failed to close %s: %w", name, closeOutErr)
		}
		if closeInErr != nil {
			return nil, fmt.Errorf("failed to close ZIP entry %s: %w", name, closeInErr)
		}

		if name == skillformat.SkillMainFile {
			foundSkillMain = true
		}
		installed = append(installed, name)
	}

	if !foundSkillMain {
		return nil, fmt.Errorf("%s not found in ZIP", skillformat.SkillMainFile)
	}

	sort.Strings(installed)
	return installed, nil
}

func open(data []byte) (*zip.Reader, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid ZIP archive: %w", err)
	}
	return reader, nil
}

func sanitizeArchivePath(name string) (string, error) {
	name = normalizeArchiveSlashes(name)
	cleaned := path.Clean(strings.TrimPrefix(name, "./"))
	switch {
	case cleaned == "":
		return ".", nil
	case cleaned == ".":
		return ".", nil
	case strings.HasPrefix(cleaned, "/"):
		return "", fmt.Errorf("absolute paths are not allowed in ZIP contents: %s", name)
	case cleaned == ".." || strings.HasPrefix(cleaned, "../"):
		return "", fmt.Errorf("path traversal is not allowed in ZIP contents: %s", name)
	}
	if len(cleaned) >= 2 && cleaned[1] == ':' &&
		((cleaned[0] >= 'a' && cleaned[0] <= 'z') || (cleaned[0] >= 'A' && cleaned[0] <= 'Z')) {
		return "", fmt.Errorf("absolute paths are not allowed in ZIP contents: %s", name)
	}
	if volume := filepath.VolumeName(filepath.FromSlash(cleaned)); volume != "" {
		return "", fmt.Errorf("absolute paths are not allowed in ZIP contents: %s", name)
	}
	return cleaned, nil
}

func ensureWithinBase(baseDir, targetPath string) error {
	if !filepath.IsAbs(baseDir) {
		return fmt.Errorf("base extraction directory must be absolute: %s", baseDir)
	}
	if !filepath.IsAbs(targetPath) {
		return fmt.Errorf("extracted path must be absolute: %s", targetPath)
	}

	rel, err := filepath.Rel(baseDir, targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve extracted path %s: %w", targetPath, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path traversal is not allowed in ZIP contents: %s", targetPath)
	}
	return nil
}

func normalizeArchiveSlashes(name string) string {
	name = filepath.ToSlash(name)
	return strings.ReplaceAll(name, "\\", "/")
}
