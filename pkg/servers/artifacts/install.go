package artifacts

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/agent"
	"github.com/nanobot-ai/nanobot/pkg/servers/installzip"
	obotconfig "github.com/nanobot-ai/nanobot/pkg/servers/obot"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)

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
	obotconfig.ApplyAuth(req, obotconfig.Config{AuthHeader: cfg.authHeader})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(body))
	}

	zipData, err := installzip.ReadAll(resp.Body, "artifact")
	if err != nil {
		return nil, err
	}

	fm, err := installzip.ReadFrontmatter(zipData)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s from ZIP: %w", skillformat.SkillMainFile, err)
	}

	if filepath.Base(fm.Name) != fm.Name || fm.Name == "." || fm.Name == ".." {
		return nil, fmt.Errorf("invalid artifact name: %s", fm.Name)
	}

	// All artifacts are currently workflows.
	targetDir := filepath.Join(".", skillformat.WorkflowsDir, fm.Name)

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

	installedFiles, err := installzip.Extract(zipData, targetDir)
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
