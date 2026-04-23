package config

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestLoadBuiltinAgents(t *testing.T) {
	// Create a minimal config
	cfg := &types.Config{
		Agents: make(map[string]types.Agent),
	}

	// Load builtin agents
	err := loadBuiltinAgents(cfg)
	if err != nil {
		t.Fatalf("unexpected error loading builtin agents: %v", err)
	}

	// Verify that nanobot agent was loaded
	expectedAgents := []string{"nanobot"}

	for _, agentName := range expectedAgents {
		agent, exists := cfg.Agents[agentName]
		if !exists {
			t.Errorf("expected builtin agent %q to be loaded", agentName)
			continue
		}

		// Verify instructions are present
		instructions := agent.Instructions.Instructions
		if instructions == "" {
			t.Errorf("agent %q should have instructions", agentName)
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
			"nanobot": {
				HookAgent: types.HookAgent{
					Name: "My Custom Explorer",
				},
			},
		},
	}

	// Try to load builtin agents - should error
	err := loadBuiltinAgents(cfg)
	if err == nil {
		t.Fatal("expected error when builtin agent conflicts with existing agent")
	}

	if !strings.Contains(err.Error(), "cannot override built-in agent") {
		t.Errorf("expected error message about overriding builtin agent, got: %v", err)
	}
	if !strings.Contains(err.Error(), "nanobot") {
		t.Errorf("expected error message to mention 'nanobot', got: %v", err)
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

	// Should have the main agent from testdata plus nanobot
	if len(cfg.Agents) < 2 {
		t.Errorf("expected at least 2 agents (main + nanobot), got %d", len(cfg.Agents))
	}

	// Check that nanobot exists
	agent, exists := cfg.Agents["nanobot"]
	if !exists {
		t.Errorf("expected builtin agent 'nanobot' to be loaded")
	} else {
		// Verify instructions are present
		if agent.Instructions.Instructions == "" {
			t.Errorf("builtin agent 'nanobot' should have instructions")
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

	// Verify nanobot does NOT exist
	if _, exists := cfg.Agents["nanobot"]; exists {
		t.Errorf("did not expect builtin agent 'nanobot' when includeDefaultAgents=false")
	}
}

func TestLoadMany_MergesInOrder(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	basePath := dir + "/base.yaml"
	overlayPath := dir + "/overlay.yaml"

	base := []byte(`agents:
  assistant:
    name: Base Assistant
    instructions: Base instructions
publish:
  entrypoint:
    - assistant
`)
	overlay := []byte(`agents:
  assistant:
    name: Overlay Assistant
  helper:
    name: Helper
    instructions: Helper instructions
publish:
  entrypoint:
    - helper
`)

	if err := os.WriteFile(basePath, base, 0o644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatalf("failed to write overlay config: %v", err)
	}

	cfg, _, err := LoadMany(ctx, []string{basePath, overlayPath}, false)
	if err != nil {
		t.Fatalf("unexpected error loading merged configs: %v", err)
	}

	if got := cfg.Agents["assistant"].Name; got != "Overlay Assistant" {
		t.Fatalf("expected later config to override assistant name, got %q", got)
	}
	if got := cfg.Agents["assistant"].Instructions.Instructions; got != "Base instructions" {
		t.Fatalf("expected overlay merge to preserve existing instructions, got %q", got)
	}
	if _, ok := cfg.Agents["helper"]; !ok {
		t.Fatalf("expected helper agent from overlay config to be present")
	}
	if got := cfg.Publish.Entrypoint; len(got) != 2 || got[0] != "assistant" || got[1] != "helper" {
		t.Fatalf("expected entrypoint arrays to concatenate in order, got %v", got)
	}
}
