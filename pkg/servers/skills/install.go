package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/agent"
	"github.com/nanobot-ai/nanobot/pkg/servers/installzip"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
)

const skillsDir = "skills"

type installSkillParams struct {
	SkillID string `json:"skillID"`
}

type installSkillResult struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	DisplayName    string   `json:"displayName,omitempty"`
	RepoID         string   `json:"repoID,omitempty"`
	CommitSHA      string   `json:"commitSHA,omitempty"`
	Path           string   `json:"path"`
	InstalledFiles []string `json:"installedFiles,omitempty"`
	Message        string   `json:"message"`
}

func (s *Server) installSkill(ctx context.Context, params installSkillParams) (*installSkillResult, error) {
	if params.SkillID == "" {
		return nil, fmt.Errorf("skillID is required")
	}

	client, err := s.newClient(ctx)
	if err != nil {
		return nil, err
	}

	skill, err := client.GetSkill(ctx, params.SkillID)
	if err != nil {
		return nil, err
	}
	if skill.Name == "" {
		return nil, fmt.Errorf("skill detail response did not include a name")
	}
	if err := skillformat.ValidateName(skill.Name); err != nil {
		return nil, fmt.Errorf("skill detail contains an invalid name: %w", err)
	}

	zipData, err := client.DownloadSkill(ctx, params.SkillID)
	if err != nil {
		return nil, err
	}

	fm, err := installzip.ReadFrontmatter(zipData)
	if err != nil {
		return nil, fmt.Errorf("failed to validate skill archive: %w", err)
	}
	if fm.Name != skill.Name {
		return nil, fmt.Errorf("skill archive name %q does not match indexed skill name %q", fm.Name, skill.Name)
	}

	targetDir := filepath.Join(".", skillsDir, skill.Name)
	if _, err := os.Stat(targetDir); err == nil {
		overwrite, err := s.confirmOverwrite(ctx, skill.Name)
		if err != nil {
			return nil, err
		}
		if !overwrite {
			return &installSkillResult{
				ID:          skill.ID,
				Name:        skill.Name,
				DisplayName: skill.DisplayName,
				RepoID:      skill.RepoID,
				CommitSHA:   skill.CommitSHA,
				Path:        targetDir,
				Message:     fmt.Sprintf("Installation of %q was canceled by the user. The existing skill was not modified.", skill.Name),
			}, nil
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to inspect existing skill directory: %w", err)
	}

	installedFiles, err := installIntoDirectory(zipData, skill.Name, targetDir)
	if err != nil {
		return nil, err
	}

	return &installSkillResult{
		ID:             skill.ID,
		Name:           skill.Name,
		DisplayName:    skill.DisplayName,
		RepoID:         skill.RepoID,
		CommitSHA:      skill.CommitSHA,
		Path:           targetDir,
		InstalledFiles: installedFiles,
		Message:        fmt.Sprintf("Installed %s into %s (%d files)", skill.Name, targetDir, len(installedFiles)),
	}, nil
}

func installIntoDirectory(zipData []byte, skillName, targetDir string) ([]string, error) {
	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create skills directory: %w", err)
	}

	tempRoot, err := os.MkdirTemp(parentDir, ".skill-install-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary install directory: %w", err)
	}
	defer os.RemoveAll(tempRoot)

	tempDir := filepath.Join(tempRoot, skillName)

	installedFiles, err := installzip.Extract(zipData, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract skill archive: %w", err)
	}

	if err := skillformat.ValidateSkillDirectory(tempDir); err != nil {
		return nil, fmt.Errorf("failed to validate extracted skill: %w", err)
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.Rename(tempDir, targetDir); err != nil {
			return nil, fmt.Errorf("failed to finalize skill install: %w", err)
		}
		return installedFiles, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to inspect target skill directory: %w", err)
	}

	backupDir := filepath.Join(parentDir, "."+filepath.Base(targetDir)+".backup")
	_ = os.RemoveAll(backupDir)

	if err := os.Rename(targetDir, backupDir); err != nil {
		return nil, fmt.Errorf("failed to prepare existing skill backup: %w", err)
	}

	if err := os.Rename(tempDir, targetDir); err != nil {
		_ = os.Rename(backupDir, targetDir)
		return nil, fmt.Errorf("failed to finalize overwritten skill install: %w", err)
	}

	if err := os.RemoveAll(backupDir); err != nil {
		return nil, fmt.Errorf("failed to clean up previous skill version: %w", err)
	}

	return installedFiles, nil
}

func defaultConfirmOverwrite(ctx context.Context, name string) (bool, error) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return false, fmt.Errorf("no session found in context")
	}

	elicit := mcp.ElicitRequest{
		Message: fmt.Sprintf("A skill named %q already exists. Do you want to overwrite it?", name),
		RequestedSchema: mcp.PrimitiveSchema{
			Type:       "object",
			Properties: map[string]mcp.PrimitiveProperty{},
		},
	}

	var result mcp.ElicitResult
	if err := agent.ExchangeElicitation(ctx, session, elicit, &result); err != nil {
		return false, fmt.Errorf("failed to send overwrite confirmation: %w", err)
	}

	return result.Action == "accept", nil
}
