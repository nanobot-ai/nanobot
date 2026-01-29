package system

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSkills(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	result, err := server.listSkills(ctx, struct{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// We should have at least the 3 skills we know about
	assert.GreaterOrEqual(t, len(result.Skills), 3)

	// Verify the skills have names and descriptions
	skillNames := make(map[string]bool)
	for _, skill := range result.Skills {
		assert.NotEmpty(t, skill.Name, "skill name should not be empty")
		assert.NotEmpty(t, skill.Description, "skill description should not be empty")
		skillNames[skill.Name] = true
	}

	// Check for expected skills
	assert.True(t, skillNames["python-scripts"], "should have python-scripts skill")
	assert.True(t, skillNames["learn"], "should have learn skill")
	assert.True(t, skillNames["mcp-curl"], "should have mcp-curl skill")
}

func TestGetSkill(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	tests := []struct {
		name          string
		skillName     string
		expectError   bool
		shouldContain string
	}{
		{
			name:          "get skill without extension",
			skillName:     "python-scripts",
			expectError:   false,
			shouldContain: "name: python-scripts",
		},
		{
			name:          "get skill with extension",
			skillName:     "python-scripts.md",
			expectError:   false,
			shouldContain: "name: python-scripts",
		},
		{
			name:          "get learn skill",
			skillName:     "learn",
			expectError:   false,
			shouldContain: "name: learn",
		},
		{
			name:          "get mcp-curl skill",
			skillName:     "mcp-skill",
			expectError:   false,
			shouldContain: "name: mcp-curl",
		},
		{
			name:        "nonexistent skill",
			skillName:   "nonexistent-skill",
			expectError: true,
		},
		{
			name:        "empty skill name",
			skillName:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := server.getSkill(ctx, GetSkillParams{Name: tt.skillName})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, content)
				if tt.shouldContain != "" {
					assert.Contains(t, content, tt.shouldContain)
				}
				// Verify it starts with frontmatter
				assert.True(t, len(content) > 0 && content[0:3] == "---", "content should start with frontmatter")
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expected    map[string]string
	}{
		{
			name: "valid frontmatter",
			content: `---
name: test-skill
description: A test skill
---

# Content here`,
			expectError: false,
			expected: map[string]string{
				"name":        "test-skill",
				"description": "A test skill",
			},
		},
		{
			name: "frontmatter with extra fields",
			content: `---
name: test-skill
description: A test skill
author: test
---

Content`,
			expectError: false,
			expected: map[string]string{
				"name":        "test-skill",
				"description": "A test skill",
				"author":      "test",
			},
		},
		{
			name:        "no frontmatter",
			content:     "Just regular content",
			expectError: true,
		},
		{
			name: "unclosed frontmatter",
			content: `---
name: test
description: test`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFrontmatter(tt.content)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
