package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

// parseSkillURI extracts the skill name from a skill:///name URI.
func parseSkillURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, "skill:///") {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid skill URI format, expected skill:///name")
	}

	skillName := strings.TrimPrefix(uri, "skill:///")
	if skillName == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("skill name is required")
	}

	skillName = strings.TrimSuffix(skillName, ".md")

	return skillName, nil
}

func (s *Server) deleteSkill(_ context.Context, data struct {
	URI string `json:"uri"`
}) (string, error) {
	skillName, err := parseSkillURI(data.URI)
	if err != nil {
		return "", fmt.Errorf("failed to parse skill URI: %w", err)
	}

	skillPath := filepath.Join(s.configDir, "skills", skillName)
	if err := os.RemoveAll(skillPath); err != nil {
		return "", fmt.Errorf("failed to delete skill: %w", err)
	}

	return fmt.Sprintf("%s deleted", data.URI), nil
}
