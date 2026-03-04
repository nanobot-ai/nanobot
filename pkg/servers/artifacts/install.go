package artifacts

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/agent"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)

const maxDownloadBytes = 100 * 1024 * 1024 // 100 MB

type installArtifactParams struct {
	ID      string `json:"id"`
	Version *int   `json:"version,omitempty"`
}

type installResult struct {
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	InstalledFiles []string `json:"installedFiles"`
	Message        string   `json:"message"`
}

func (s *Server) installArtifact(ctx context.Context, params installArtifactParams) (*installResult, error) {
	// We rely on the `unzip` command to extract the artifact onto the local filesystem.
	// People should never be using this part of Nanobot outside of the container anyway,
	// so just block Windows.
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("artifact installation is not supported on Windows")
	}

	if params.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	cfg, err := getObotConfig(ctx)
	if err != nil {
		return nil, err
	}

	downloadURL := cfg.baseURL + "/api/published-artifacts/" + params.ID + "/download"
	if params.Version != nil {
		downloadURL += "?version=" + strconv.Itoa(*params.Version)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if cfg.authHeader != "" {
		req.Header.Set("Authorization", cfg.authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(body))
	}

	zipData, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact data: %w", err)
	}
	if len(zipData) > maxDownloadBytes {
		return nil, fmt.Errorf("artifact exceeds maximum size of %d bytes", maxDownloadBytes)
	}

	fm, err := readFrontmatterFromZIP(zipData)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s from ZIP: %w", skillformat.SkillMainFile, err)
	}

	if filepath.Base(fm.Name) != fm.Name || fm.Name == "." || fm.Name == ".." {
		return nil, fmt.Errorf("invalid artifact name: %s", fm.Name)
	}

	// All artifacts are currently workflows.
	targetDir := filepath.Join(".", workflowsDir, fm.Name)

	// If the workflow already exists, ask the user for confirmation before overwriting.
	if _, err := os.Stat(targetDir); err == nil {
		session := mcp.SessionFromContext(ctx)
		if session == nil {
			return nil, fmt.Errorf("no session found in context")
		}

		elicit := mcp.ElicitRequest{
			Message: fmt.Sprintf("A workflow named %q already exists. Do you want to overwrite it?", fm.Name),
			RequestedSchema: mcp.PrimitiveSchema{
				Type:       "object",
				Properties: map[string]mcp.PrimitiveProperty{},
			},
		}

		var result mcp.ElicitResult
		if err := agent.ExchangeElicitation(ctx, session, elicit, &result); err != nil {
			return nil, fmt.Errorf("failed to send overwrite confirmation: %w", err)
		}

		if result.Action != "accept" {
			return &installResult{
				Name:    fm.Name,
				Path:    targetDir,
				Message: fmt.Sprintf("Installation of %q was canceled by the user. The existing workflow was not modified.", fm.Name),
			}, nil
		}
	}

	// Remove existing directory to allow overwrite.
	if err := os.RemoveAll(targetDir); err != nil {
		return nil, fmt.Errorf("failed to remove existing directory: %w", err)
	}

	installedFiles, err := extractZIP(ctx, zipData, targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract artifact: %w", err)
	}

	return &installResult{
		Name:           fm.Name,
		Path:           targetDir,
		InstalledFiles: installedFiles,
		Message:        fmt.Sprintf("Installed %s into %s (%d files)", fm.Name, targetDir, len(installedFiles)),
	}, nil
}

func readFrontmatterFromZIP(data []byte) (skillformat.Frontmatter, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return skillformat.Frontmatter{}, fmt.Errorf("invalid ZIP archive: %w", err)
	}

	for _, f := range r.File {
		if f.Name == skillformat.SkillMainFile {
			rc, err := f.Open()
			if err != nil {
				return skillformat.Frontmatter{}, fmt.Errorf("failed to open %s: %w", skillformat.SkillMainFile, err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return skillformat.Frontmatter{}, fmt.Errorf("failed to read %s: %w", skillformat.SkillMainFile, err)
			}

			fm, _, err := skillformat.ParseAndValidateFrontmatter(string(content))
			if err != nil {
				return skillformat.Frontmatter{}, fmt.Errorf("invalid %s: %w", skillformat.SkillMainFile, err)
			}
			return fm, nil
		}
	}

	return skillformat.Frontmatter{}, fmt.Errorf("%s not found in ZIP", skillformat.SkillMainFile)
}

// extractZIP writes the ZIP data to a temp file and uses the system `unzip`
// command to extract it into targetDir.
func extractZIP(ctx context.Context, data []byte, targetDir string) ([]string, error) {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "artifact-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	cmd := exec.CommandContext(ctx, "unzip", "-o", tmpFile.Name(), "-d", targetDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("unzip failed: %w\n%s", err, string(output))
	}

	// Walk the extracted directory to build the installed files list.
	var installed []string
	if err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(targetDir, path)
		if err != nil {
			return err
		}
		installed = append(installed, filepath.ToSlash(relPath))
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list extracted files: %w", err)
	}

	return installed, nil
}
