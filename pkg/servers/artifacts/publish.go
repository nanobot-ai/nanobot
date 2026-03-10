package artifacts

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)


type publishArtifactParams struct {
	WorkflowName string `json:"workflowName"`
}

type publishResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	Message string `json:"message"`
}

func (s *Server) publishArtifact(ctx context.Context, params publishArtifactParams) (*publishResult, error) {
	if params.WorkflowName == "" {
		return nil, fmt.Errorf("workflowName is required")
	}

	cfg, err := getObotConfig(ctx)
	if err != nil {
		return nil, err
	}

	workflowDir := filepath.Join(".", skillformat.WorkflowsDir, params.WorkflowName)
	mainFile := filepath.Join(workflowDir, skillformat.SkillMainFile)

	content, err := os.ReadFile(mainFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", skillformat.SkillMainFile, err)
	}

	fm, _, err := skillformat.ParseAndValidateFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", skillformat.SkillMainFile, err)
	}

	if err := skillformat.ValidateNameMatchesDir(fm.Name, filepath.Base(workflowDir)); err != nil {
		return nil, err
	}

	zipData, err := createZIP(workflowDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZIP: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.baseURL+"/api/published-artifacts", bytes.NewReader(zipData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if cfg.authHeader != "" {
		req.Header.Set("Authorization", cfg.authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish artifact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("publish failed (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		LatestVersion int    `json:"latestVersion"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &publishResult{
		ID:      apiResp.ID,
		Name:    apiResp.Name,
		Version: apiResp.LatestVersion,
		Message: fmt.Sprintf("Published %s v%d", apiResp.Name, apiResp.LatestVersion),
	}, nil
}

// createZIP creates a ZIP archive containing all files in the workflow directory.
// No manifest.yaml is generated — the SKILL.md frontmatter is the source of truth.
func createZIP(workflowDir string) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	if err := filepath.Walk(workflowDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(workflowDir, path)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", relPath, err)
		}

		fw, err := w.Create(filepath.ToSlash(relPath))
		if err != nil {
			return fmt.Errorf("failed to create ZIP entry %s: %w", relPath, err)
		}
		if _, err := fw.Write(data); err != nil {
			return fmt.Errorf("failed to write ZIP entry %s: %w", relPath, err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk workflow directory: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP: %w", err)
	}

	return buf.Bytes(), nil
}
