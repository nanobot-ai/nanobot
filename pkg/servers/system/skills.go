package system

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/system/skills"
	"github.com/nanobot-ai/nanobot/pkg/skillformat"
	"gopkg.in/yaml.v3"
)

// Skill represents a skill with its metadata
type Skill struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// SkillList is the response type for list_skills
type SkillList struct {
	Skills []Skill `json:"skills"`
}

// GetSkillParams is the input type for get_skill
type GetSkillParams struct {
	Name string `json:"name"`
}

// parseFrontmatter extracts YAML frontmatter from markdown content
func parseFrontmatter(content string) (map[string]string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("no frontmatter found")
	}

	// Find the closing ---
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return nil, fmt.Errorf("frontmatter not properly closed")
	}

	// Parse YAML from frontmatter
	frontmatterYAML := strings.Join(lines[1:endIdx], "\n")
	var frontmatter map[string]string
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return frontmatter, nil
}

func (s *Server) listSkills(ctx context.Context, _ struct{}) (*SkillList, error) {
	// Use a map to dedupe by name, with user skills taking precedence
	skillMap := make(map[string]Skill)

	// First, load built-in skills from embedded FS
	err := fs.WalkDir(skills.Skills, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.md files
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// Read the file
		content, err := fs.ReadFile(skills.Skills, path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Parse frontmatter
		frontmatter, err := parseFrontmatter(string(content))
		if err != nil {
			// Skip files without valid frontmatter
			return nil
		}

		name := strings.TrimSuffix(filepath.Base(path), ".md")
		skillMap[name] = Skill{
			Name:        name,
			DisplayName: frontmatter["name"],
			Description: frontmatter["description"],
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Then, load user skills from config directory (if it exists)
	// User skills override built-in skills with the same name
	if s.configDir != "" {
		userSkillsDir := filepath.Join(s.configDir, "skills")
		entries, err := os.ReadDir(userSkillsDir)
		if err == nil {
			// Load flat legacy skills first.
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
					continue
				}

				content, err := os.ReadFile(filepath.Join(userSkillsDir, entry.Name()))
				if err != nil {
					// Skip files we can't read
					continue
				}

				frontmatter, err := parseFrontmatter(string(content))
				if err != nil {
					// Skip files without valid frontmatter
					continue
				}

				// User skills override built-in skills
				name := strings.TrimSuffix(entry.Name(), ".md")
				skillMap[name] = Skill{
					Name:        name,
					DisplayName: frontmatter["name"],
					Description: frontmatter["description"],
				}
			}

			// Directory-based Agent Skills override both flat user skills and built-ins.
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				skill, present, err := loadDirectorySkill(userSkillsDir, entry.Name())
				if err != nil {
					delete(skillMap, entry.Name())
					continue
				}
				if present {
					skillMap[skill.Name] = skill
				}
			}
		}
		// If directory doesn't exist or can't be read, silently continue
	}

	// Convert map to slice
	result := make([]Skill, 0, len(skillMap))
	for _, skill := range skillMap {
		result = append(result, skill)
	}

	return &SkillList{
		Skills: result,
	}, nil
}

func (s *Server) getSkill(ctx context.Context, params GetSkillParams) (string, error) {
	if params.Name == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("skill name is required")
	}

	skillName := strings.TrimSuffix(params.Name, ".md")

	if s.configDir != "" {
		userSkillsDir := filepath.Join(s.configDir, "skills")

		dirSkillPath := filepath.Join(userSkillsDir, skillName, skillformat.SkillMainFile)
		if info, err := os.Stat(filepath.Join(userSkillsDir, skillName)); err == nil && info.IsDir() {
			content, err := os.ReadFile(dirSkillPath)
			if err != nil {
				return "", mcp.ErrRPCInvalidParams.WithMessage("skill '%s' not found", params.Name)
			}

			fm, _, err := skillformat.ParseAndValidateFrontmatter(string(content))
			if err != nil {
				return "", mcp.ErrRPCInvalidParams.WithMessage("skill '%s' not found", params.Name)
			}
			if err := skillformat.ValidateNameMatchesDir(fm.Name, skillName); err != nil {
				return "", mcp.ErrRPCInvalidParams.WithMessage("skill '%s' not found", params.Name)
			}

			return string(content), nil
		}

		userSkillPath := filepath.Join(userSkillsDir, skillName+".md")
		content, err := os.ReadFile(userSkillPath)
		if err == nil {
			return string(content), nil
		}
	}

	content, err := fs.ReadFile(skills.Skills, skillName+".md")
	if err != nil {
		return "", mcp.ErrRPCInvalidParams.WithMessage("skill '%s' not found", params.Name)
	}

	return string(content), nil
}

func loadDirectorySkill(userSkillsDir, dirName string) (Skill, bool, error) {
	skillPath := filepath.Join(userSkillsDir, dirName, skillformat.SkillMainFile)
	content, err := os.ReadFile(skillPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Skill{}, true, fmt.Errorf("missing %s", skillformat.SkillMainFile)
		}
		return Skill{}, true, err
	}

	fm, _, err := skillformat.ParseAndValidateFrontmatter(string(content))
	if err != nil {
		return Skill{}, true, err
	}
	if err := skillformat.ValidateNameMatchesDir(fm.Name, dirName); err != nil {
		return Skill{}, true, err
	}

	return Skill{
		Name:        fm.Name,
		DisplayName: skillformat.DisplayName(fm.Name),
		Description: fm.Description,
	}, true, nil
}
