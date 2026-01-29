package config

import (
	"context"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestLoadBuiltinAgents(t *testing.T) {
	// Create a minimal config
	cfg := &types.Config{
		Agents: make(map[string]types.Agent),
	}

	// Load builtin agents with a test schema
	testSchema := "# Test Workflow Schema\n\nThis is a test schema."
	err := loadBuiltinAgents(cfg, testSchema)
	if err != nil {
		t.Fatalf("unexpected error loading builtin agents: %v", err)
	}

	// Verify that executor and planner agents were loaded
	expectedAgents := []string{"executor", "planner"}

	for _, agentName := range expectedAgents {
		agent, exists := cfg.Agents[agentName]
		if !exists {
			t.Errorf("expected builtin agent %q to be loaded", agentName)
			continue
		}

		// Verify instructions contain the workflow schema wrapped in tags
		instructions := agent.Instructions.Instructions
		if !strings.Contains(instructions, "<workflow_schema>") {
			t.Errorf("agent %q instructions should contain <workflow_schema> tag", agentName)
		}
		if !strings.Contains(instructions, "</workflow_schema>") {
			t.Errorf("agent %q instructions should contain </workflow_schema> tag", agentName)
		}
		if !strings.Contains(instructions, testSchema) {
			t.Errorf("agent %q instructions should contain the workflow schema", agentName)
		}

		// Verify the original agent body is present (check for frontmatter values)
		if agent.Description == "" {
			t.Errorf("agent %q should have a description from frontmatter", agentName)
		}
	}
}

func TestLoadBuiltinAgents_ConflictError(t *testing.T) {
	// Create a config with an existing agent that conflicts with builtin
	cfg := &types.Config{
		Agents: map[string]types.Agent{
			"executor": {
				HookAgent: types.HookAgent{
					Name: "My Custom Executor",
				},
			},
		},
	}

	// Try to load builtin agents - should error
	testSchema := "# Test Schema"
	err := loadBuiltinAgents(cfg, testSchema)
	if err == nil {
		t.Fatal("expected error when builtin agent conflicts with existing agent")
	}

	if !strings.Contains(err.Error(), "cannot override built-in agent") {
		t.Errorf("expected error message about overriding builtin agent, got: %v", err)
	}
	if !strings.Contains(err.Error(), "executor") {
		t.Errorf("expected error message to mention 'executor', got: %v", err)
	}
}

func TestLoad_WithBuiltinAgents(t *testing.T) {
	ctx := context.Background()

	// Load a minimal config with includeDefaultAgents=true
	// This uses the testdata that already exists
	cfg, _, err := Load(ctx, "./testdata/directory-simple", true)
	if err != nil {
		t.Fatalf("unexpected error loading config with builtin agents: %v", err)
	}

	// Should have the main agent from testdata plus executor and planner
	if len(cfg.Agents) < 3 {
		t.Errorf("expected at least 3 agents (main + executor + planner), got %d", len(cfg.Agents))
	}

	// Check that executor and planner exist
	for _, agentName := range []string{"executor", "planner"} {
		agent, exists := cfg.Agents[agentName]
		if !exists {
			t.Errorf("expected builtin agent %q to be loaded", agentName)
			continue
		}

		// Verify workflow schema is in instructions
		if !strings.Contains(agent.Instructions.Instructions, "<workflow_schema>") {
			t.Errorf("builtin agent %q should have workflow schema in instructions", agentName)
		}
	}
}

func TestLoad_WithoutBuiltinAgents(t *testing.T) {
	ctx := context.Background()

	// Load a config with includeDefaultAgents=false
	cfg, _, err := Load(ctx, "./testdata/directory-simple", false)
	if err != nil {
		t.Fatalf("unexpected error loading config without builtin agents: %v", err)
	}

	// Should only have the main agent from testdata
	if len(cfg.Agents) != 1 {
		t.Errorf("expected exactly 1 agent (main), got %d", len(cfg.Agents))
	}

	// Verify executor and planner do NOT exist
	for _, agentName := range []string{"executor", "planner"} {
		if _, exists := cfg.Agents[agentName]; exists {
			t.Errorf("did not expect builtin agent %q when includeDefaultAgents=false", agentName)
		}
	}
}
