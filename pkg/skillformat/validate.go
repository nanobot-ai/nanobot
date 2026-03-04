package skillformat

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMainFile is the filename for the main skill definition file in a skill directory.
const SkillMainFile = "SKILL.md"

// Frontmatter represents the YAML frontmatter of a SKILL.md file,
// following the Agent Skills specification.
type Frontmatter struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty"`
}

var nameRegexp = regexp.MustCompile(`^[a-z0-9-]+$`)

// ValidateName checks that a skill name conforms to the Agent Skills spec:
// 1-64 characters, lowercase letters/numbers/hyphens only, no leading/trailing
// hyphens, no consecutive hyphens.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("name must be at most 64 characters, got %d", len(name))
	}
	if !nameRegexp.MatchString(name) {
		return fmt.Errorf("name must contain only lowercase letters, numbers, and hyphens (got %q)", name)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("name must not start with a hyphen (got %q)", name)
	}
	if strings.HasSuffix(name, "-") {
		return fmt.Errorf("name must not end with a hyphen (got %q)", name)
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("name must not contain consecutive hyphens (got %q)", name)
	}
	return nil
}

// ValidateDescription checks that a description is non-empty and at most 1024 characters.
func ValidateDescription(description string) error {
	if description == "" {
		return fmt.Errorf("description must not be empty")
	}
	if len(description) > 1024 {
		return fmt.Errorf("description must be at most 1024 characters, got %d", len(description))
	}
	return nil
}

// ValidateFrontmatter validates all fields of a parsed Frontmatter struct.
func ValidateFrontmatter(fm Frontmatter) error {
	var errs []error
	if err := ValidateName(fm.Name); err != nil {
		errs = append(errs, fmt.Errorf("invalid name: %w", err))
	}
	if err := ValidateDescription(fm.Description); err != nil {
		errs = append(errs, fmt.Errorf("invalid description: %w", err))
	}
	if fm.Compatibility != "" && len(fm.Compatibility) > 500 {
		errs = append(errs, fmt.Errorf("compatibility must be at most 500 characters, got %d", len(fm.Compatibility)))
	}
	return errors.Join(errs...)
}

// ValidateNameMatchesDir returns an error if the frontmatter name does not
// match the directory name.
func ValidateNameMatchesDir(name, dirName string) error {
	if name != dirName {
		return fmt.Errorf("skill name %q does not match directory name %q — they must be identical", name, dirName)
	}
	return nil
}

// ParseFrontmatter extracts YAML frontmatter and body content from markdown.
// If no frontmatter is found (no opening ---), returns zero-value Frontmatter,
// the full content as body, and a nil error.
func ParseFrontmatter(content string) (Frontmatter, string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return Frontmatter{}, content, nil
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return Frontmatter{}, "", fmt.Errorf("frontmatter missing closing delimiter")
	}

	fmYAML := strings.Join(lines[1:endIdx], "\n")
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(fmYAML), &fm); err != nil {
		return Frontmatter{}, "", fmt.Errorf("failed to parse frontmatter YAML: %w", err)
	}

	body := ""
	if endIdx+1 < len(lines) {
		body = strings.Join(lines[endIdx+1:], "\n")
	}

	return fm, body, nil
}

// ParseAndValidateFrontmatter parses YAML frontmatter from markdown content
// and validates the result. Returns the parsed frontmatter, body content, and
// any error from parsing or validation.
func ParseAndValidateFrontmatter(content string) (Frontmatter, string, error) {
	fm, body, err := ParseFrontmatter(content)
	if err != nil {
		return fm, body, err
	}
	if err := ValidateFrontmatter(fm); err != nil {
		return fm, body, err
	}
	return fm, body, nil
}

// FormatSkillMD serializes a Frontmatter and body back into a complete SKILL.md
// string with YAML frontmatter delimiters.
func FormatSkillMD(fm Frontmatter, body string) (string, error) {
	fmData, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmData)
	sb.WriteString("---\n")
	sb.WriteString(body)
	return sb.String(), nil
}

// DisplayName converts a skill slug (e.g., "code-review") to a
// human-readable display name (e.g., "Code Review").
func DisplayName(slug string) string {
	if slug == "" {
		return ""
	}
	words := strings.Split(slug, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// ValidateSkillDirectory validates a skill directory: checks that SKILL.md
// exists, parses and validates its frontmatter, and ensures the frontmatter
// name matches the directory name.
func ValidateSkillDirectory(dirPath string) error {
	skillFile := filepath.Join(dirPath, SkillMainFile)
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", SkillMainFile, err)
	}

	fm, _, err := ParseAndValidateFrontmatter(string(content))
	if err != nil {
		return fmt.Errorf("invalid %s: %w", SkillMainFile, err)
	}

	if err := ValidateNameMatchesDir(fm.Name, filepath.Base(dirPath)); err != nil {
		return err
	}

	return nil
}
