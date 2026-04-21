package integration_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/agents"
	"github.com/nanobot-ai/nanobot/pkg/llm"
	"github.com/nanobot-ai/nanobot/pkg/llm/anthropic"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

// setupWorkflowListing creates a recorder that mocks the glob tool with a pair of
// fake workflow files, and a runtime context ready for agent execution.
func setupWorkflowListing(t *testing.T, completer types.Completer) (context.Context, *agents.Agents, *toolCallRecorder) {
	t.Helper()
	recorder := newRecorder(map[string]ToolHandler{
		"glob": func(_ mcp.CallToolRequest) *mcp.CallToolResult {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{Type: "text", Text: "workflows/code-review/SKILL.md\nworkflows/deploy/SKILL.md\n"}},
			}
		},
	})
	ctx, svc := newTestRuntime(t, completer, recorder)
	return ctx, svc, recorder
}

// getSkillDiagnostic returns a string describing whether the model called getSkill,
// used as additional context in glob assertion failures.
func getSkillDiagnostic(recorder *toolCallRecorder) string {
	call := recorder.find("getSkill")
	if call == nil {
		return "model did not call getSkill"
	}
	name, _ := call.Arguments["name"].(string)
	return fmt.Sprintf("model called getSkill(%q)", name)
}

// assertCorrectGlobCall verifies that the recorder captured a glob call with valid
// workflow-discovery parameters. Two forms are accepted:
//
//	pattern="**/SKILL.md"           path ends with "workflows"
//	pattern="workflows/**/SKILL.md" path does NOT contain "workflows" (already in pattern)
func assertCorrectGlobCall(t *testing.T, recorder *toolCallRecorder) {
	t.Helper()

	globCall := recorder.find("glob")
	if globCall == nil {
		t.Fatalf("LLM did not call glob (%s)\nall tool calls:\n%s",
			getSkillDiagnostic(recorder), recorder.summary())
	}

	pattern, _ := globCall.Arguments["pattern"].(string)
	path, _ := globCall.Arguments["path"].(string)
	cleanPath := strings.TrimRight(path, "/")

	patternWithPath := pattern == "**/SKILL.md" &&
		(cleanPath == "workflows" || strings.HasSuffix(cleanPath, "/workflows"))
	patternOnly := pattern == "workflows/**/SKILL.md" &&
		cleanPath != "workflows" && !strings.HasSuffix(cleanPath, "/workflows")

	if !patternWithPath && !patternOnly {
		t.Errorf("unexpected glob call: pattern=%q path=%q\n  want: pattern=\"**/SKILL.md\" with path ending in \"workflows\"\n     or: pattern=\"workflows/**/SKILL.md\" with path NOT containing \"workflows\" (would double-nest)\n  (%s)\nall tool calls:\n%s",
			pattern, path, getSkillDiagnostic(recorder), recorder.summary())
	}
}

// TestWorkflowListingGlobPattern runs the agent against a real LLM and asserts that
// the model calls glob with the correct pattern (**/SKILL.md) and path (workflows/).
// The glob tool is intercepted and not actually executed.
//
// Requires ANTHROPIC_API_KEY; skipped otherwise.
func TestWorkflowListingGlobPattern(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Fatal("ANTHROPIC_API_KEY must be set when running integration tests")
	}

	completer := llm.NewClient(llm.Config{
		DefaultModel: "claude-sonnet-4-6",
		Anthropic:    anthropic.Config{APIKey: apiKey},
	})

	prompts := []string{
		"list all workflows",
		"show my workflows",
		"what workflows do I have?",
		"show me my available workflows",
		"display all workflows",
	}

	for _, prompt := range prompts {
		t.Run(prompt, func(t *testing.T) {
			t.Parallel()
			for i := range *numRuns {
				t.Run(fmt.Sprintf("run%d", i+1), func(t *testing.T) {
					t.Parallel()
					ctx, svc, recorder := setupWorkflowListing(t, completer)

					err := runAgent(ctx, svc, prompt)
					if err != nil {
						t.Fatalf("Complete() failed: %v", err)
					}

					assertCorrectGlobCall(t, recorder)

					if call := recorder.find("searchArtifacts"); call != nil {
						t.Errorf("model should not call searchArtifacts for local listing prompts\nall tool calls:\n%s", recorder.summary())
					}
				})
			}
		})
	}
}
