package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestTruncateToolResultArchivesOutput(t *testing.T) {
	a := &Agents{}
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	largeText := strings.Repeat("data ", toolOutputPreviewLimit)
	result := &types.CallResult{
		Content: []mcp.Content{
			{Type: "text", Text: largeText},
		},
	}

	preview := a.truncateToolResult("call-id", result)
	if preview == result {
		t.Fatalf("expected truncated result to return new instance")
	}

	if len(preview.Content) != 1 {
		t.Fatalf("expected single preview content, got %d", len(preview.Content))
	}

	text := preview.Content[0].Text
	if !strings.Contains(text, "Full output saved to") {
		t.Fatalf("preview text missing archive instructions: %s", text)
	}

	meta, ok := preview.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("expected structured content metadata")
	}
	nanobotMeta, ok := meta["nanobot"].(map[string]any)
	if !ok {
		t.Fatalf("expected nanobot metadata entry")
	}
	archivePath, ok := nanobotMeta["toolOutputArchive"].(string)
	if !ok || archivePath == "" {
		t.Fatalf("expected archive path metadata, got %v", nanobotMeta["toolOutputArchive"])
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("failed to read archive file: %v", err)
	}
	var archived types.CallResult
	if err := json.Unmarshal(data, &archived); err != nil {
		t.Fatalf("failed to decode archive json: %v", err)
	}

	if len(archived.Content) == 0 || archived.Content[0].Text != largeText {
		t.Fatalf("archive content mismatch")
	}

	expectedDir := filepath.Join(tmpHome, ".nanobot", "tool-output")
	if _, err := os.Stat(expectedDir); err != nil {
		t.Fatalf("expected tool output directory: %v", err)
	}
}

func TestTruncateToolResultPassthrough(t *testing.T) {
	a := &Agents{}
	t.Setenv("HOME", t.TempDir())
	small := &types.CallResult{
		Content: []mcp.Content{{Type: "text", Text: "small"}},
	}
	preview := a.truncateToolResult("", small)
	if preview != small {
		t.Fatalf("expected passthrough when under limit")
	}
}
