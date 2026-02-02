package system

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/types"
)

// Helper function to create AgentPermissions from a map
func createPermissions(t *testing.T, perms map[string]string) *types.AgentPermissions {
	t.Helper()
	data, err := json.Marshal(perms)
	if err != nil {
		t.Fatalf("failed to marshal permissions: %v", err)
	}
	var ap types.AgentPermissions
	if err := json.Unmarshal(data, &ap); err != nil {
		t.Fatalf("failed to unmarshal permissions: %v", err)
	}
	return &ap
}

func TestConfigSkillsPermissionAppendsInstructions(t *testing.T) {
	server := NewServer("")
	ctx := context.Background()

	agent := &types.HookAgent{
		Name: "test-agent",
		Permissions: createPermissions(t, map[string]string{
			"skills": "allow",
		}),
		Instructions: types.DynamicInstructions{
			Instructions: "You are a helpful assistant.",
		},
	}

	result, err := server.config(ctx, types.AgentConfigHook{
		Agent: agent,
	})

	if err != nil {
		t.Fatalf("config() failed: %v", err)
	}

	if result.Agent == nil {
		t.Fatal("expected Agent to be set in result")
	}

	instructions := result.Agent.Instructions.Instructions
	if instructions == "You are a helpful assistant." {
		t.Error("expected instructions to be modified, but they were unchanged")
	}

	// Verify the original instructions are still there
	if !strings.Contains(instructions, "You are a helpful assistant.") {
		t.Error("original instructions should be preserved")
	}

	// Verify skills section was added
	if !strings.Contains(instructions, "## Available Skills") {
		t.Error("expected '## Available Skills' header in instructions")
	}

	if !strings.Contains(instructions, "getSkill") {
		t.Error("expected mention of 'getSkill' tool in instructions")
	}
}

func TestConfigSkillsPermissionIncludesSkillDetails(t *testing.T) {
	server := NewServer("")
	ctx := context.Background()

	agent := &types.HookAgent{
		Name: "test-agent",
		Permissions: createPermissions(t, map[string]string{
			"skills": "allow",
		}),
		Instructions: types.DynamicInstructions{
			Instructions: "Initial instructions.",
		},
	}

	result, err := server.config(ctx, types.AgentConfigHook{
		Agent: agent,
	})

	if err != nil {
		t.Fatalf("config() failed: %v", err)
	}

	instructions := result.Agent.Instructions.Instructions

	// Check for specific skills we know exist
	expectedSkills := []string{"python-scripts", "learn", "mcp-curl"}
	for _, skillName := range expectedSkills {
		if !strings.Contains(instructions, skillName) {
			t.Errorf("expected skill '%s' to be listed in instructions", skillName)
		}
	}

	// Verify the format includes skill names in bold markdown
	if !strings.Contains(instructions, "**python-scripts**") {
		t.Error("expected skills to be formatted with markdown bold")
	}

	// Verify descriptions are included
	if !strings.Contains(instructions, "Python") {
		t.Error("expected skill descriptions to be included")
	}
}

func TestConfigNoSkillsPermission(t *testing.T) {
	server := NewServer("")
	ctx := context.Background()

	originalInstructions := "You are a helpful assistant."
	agent := &types.HookAgent{
		Name: "test-agent",
		Permissions: createPermissions(t, map[string]string{
			"*":     "deny",
			"read":  "allow",
			"write": "allow",
		}),
		Instructions: types.DynamicInstructions{
			Instructions: originalInstructions,
		},
	}

	result, err := server.config(ctx, types.AgentConfigHook{
		Agent: agent,
	})

	if err != nil {
		t.Fatalf("config() failed: %v", err)
	}

	if result.Agent == nil {
		t.Fatal("expected Agent to be set in result")
	}

	instructions := result.Agent.Instructions.Instructions
	if instructions != originalInstructions {
		t.Errorf("instructions should not be modified when skills permission is not present\ngot: %s\nwant: %s", instructions, originalInstructions)
	}

	// Verify skills section was NOT added
	if strings.Contains(instructions, "## Available Skills") {
		t.Error("skills section should not be added without skills permission")
	}
}

func TestConfigSkillsPermissionDenied(t *testing.T) {
	server := NewServer("")
	ctx := context.Background()

	originalInstructions := "You are a helpful assistant."
	agent := &types.HookAgent{
		Name: "test-agent",
		Permissions: createPermissions(t, map[string]string{
			"skills": "deny",
		}),
		Instructions: types.DynamicInstructions{
			Instructions: originalInstructions,
		},
	}

	result, err := server.config(ctx, types.AgentConfigHook{
		Agent: agent,
	})

	if err != nil {
		t.Fatalf("config() failed: %v", err)
	}

	if result.Agent == nil {
		t.Fatal("expected Agent to be set in result")
	}

	instructions := result.Agent.Instructions.Instructions
	if instructions != originalInstructions {
		t.Errorf("instructions should not be modified when skills permission is denied\ngot: %s\nwant: %s", instructions, originalInstructions)
	}

	// Verify skills section was NOT added
	if strings.Contains(instructions, "## Available Skills") {
		t.Error("skills section should not be added when skills permission is denied")
	}
}

func TestConfigWithUserSkills(t *testing.T) {
	// Use test data directory with user skills
	server := NewServer(testdataDir(t, "with-user-skills"))
	ctx := context.Background()

	agent := &types.HookAgent{
		Name: "test-agent",
		Permissions: createPermissions(t, map[string]string{
			"skills": "allow",
		}),
		Instructions: types.DynamicInstructions{
			Instructions: "Initial instructions.",
		},
	}

	result, err := server.config(ctx, types.AgentConfigHook{
		Agent: agent,
	})

	if err != nil {
		t.Fatalf("config() failed: %v", err)
	}

	instructions := result.Agent.Instructions.Instructions

	// Check for built-in skills
	if !strings.Contains(instructions, "python-scripts") {
		t.Error("expected built-in skills to be listed")
	}

	// Check for user-defined skills
	if !strings.Contains(instructions, "my-custom-skill") {
		t.Error("expected user-defined skill 'my-custom-skill' to be listed")
	}

	if !strings.Contains(instructions, "user-skill") {
		t.Error("expected user-defined skill 'user-skill' to be listed")
	}
}

func TestConfigEmptyInstructions(t *testing.T) {
	server := NewServer("")
	ctx := context.Background()

	agent := &types.HookAgent{
		Name: "test-agent",
		Permissions: createPermissions(t, map[string]string{
			"skills": "allow",
		}),
		Instructions: types.DynamicInstructions{
			Instructions: "",
		},
	}

	result, err := server.config(ctx, types.AgentConfigHook{
		Agent: agent,
	})

	if err != nil {
		t.Fatalf("config() failed: %v", err)
	}

	if result.Agent == nil {
		t.Fatal("expected Agent to be set in result")
	}

	instructions := result.Agent.Instructions.Instructions

	// Even with empty initial instructions, skills should be added
	if !strings.Contains(instructions, "## Available Skills") {
		t.Error("expected skills section to be added even with empty initial instructions")
	}
}

func TestConfigNilAgent(t *testing.T) {
	server := NewServer("")
	ctx := context.Background()

	result, err := server.config(ctx, types.AgentConfigHook{
		Agent: nil,
	})

	if err != nil {
		t.Fatalf("config() should not error with nil agent: %v", err)
	}

	// Should just return the params unchanged
	if result.Agent != nil {
		t.Error("expected nil agent to remain nil")
	}
}

func TestConfigAddsToolsForPermissions(t *testing.T) {
	server := NewServer("")
	ctx := context.Background()

	agent := &types.HookAgent{
		Name: "test-agent",
		Permissions: createPermissions(t, map[string]string{
			"read":   "allow",
			"write":  "allow",
			"skills": "allow",
		}),
		MCPServers: []string{},
	}

	result, err := server.config(ctx, types.AgentConfigHook{
		Agent: agent,
	})

	if err != nil {
		t.Fatalf("config() failed: %v", err)
	}

	if result.Agent == nil {
		t.Fatal("expected Agent to be set in result")
	}

	// Check that appropriate tools were added to MCPServers
	expectedTools := []string{"nanobot.system/read", "nanobot.system/write", "nanobot.system/getSkill"}
	for _, tool := range expectedTools {
		found := false
		for _, mcpServer := range result.Agent.MCPServers {
			if mcpServer == tool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected tool '%s' to be added to MCPServers", tool)
		}
	}
}
