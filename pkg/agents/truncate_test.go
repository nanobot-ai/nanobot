package agents

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func TestTruncateToolOutput(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "truncate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir for test
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	tests := []struct {
		name                  string
		toolName              string
		content               []mcp.Content
		maxLines              int
		maxBytes              int
		shouldTrunc           bool
		expectRemoved         int
		expectUnit            string
		expectFirstType       string
		expectPreviewNonEmpty bool
	}{
		{
			name:     "no truncation needed - small content",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: "line 1\nline 2\nline 3"},
			},
			maxLines:    10,
			maxBytes:    1000,
			shouldTrunc: false,
		},
		{
			name:     "truncate by lines",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: strings.Join(makeLines(100), "\n")},
			},
			maxLines:      50,
			maxBytes:      1000000,
			shouldTrunc:   true,
			expectRemoved: 50,
			expectUnit:    "lines",
		},
		{
			name:     "truncate by bytes",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: strings.Repeat("a", 100000)},
			},
			maxLines:              10000,
			maxBytes:              50000,
			shouldTrunc:           true,
			expectUnit:            "bytes",
			expectPreviewNonEmpty: true,
		},
		{
			name:     "preserve non-text content",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: strings.Join(makeLines(100), "\n")},
				{Type: "image", Data: "base64data"},
			},
			maxLines:    50,
			maxBytes:    1000000,
			shouldTrunc: true,
		},
		{
			name:     "preserve leading non-text order",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "image", Data: "base64data"},
				{Type: "text", Text: strings.Join(makeLines(100), "\n")},
			},
			maxLines:        50,
			maxBytes:        1000000,
			shouldTrunc:     true,
			expectFirstType: "image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &types.CallResult{
				Content: tt.content,
			}

			result, err := truncateToolOutput(ctx, tt.toolName, response, tt.maxLines, tt.maxBytes)
			if err != nil {
				t.Fatalf("truncateToolOutput failed: %v", err)
			}

			if result.Truncated != tt.shouldTrunc {
				t.Errorf("expected truncated=%v, got %v", tt.shouldTrunc, result.Truncated)
			}

			if tt.shouldTrunc {
				// Verify file was created
				if result.FilePath == "" {
					t.Error("expected file path to be set when truncated")
				}

				if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
					t.Errorf("truncation file not created at %s", result.FilePath)
				}

				// Verify truncation message is in the content
				foundTruncMsg := false
				var truncationText string
				for _, content := range result.Content {
					if content.Type == "text" && strings.Contains(content.Text, "Output truncated") {
						foundTruncMsg = true
						truncationText = content.Text
						if tt.expectUnit != "" && !strings.Contains(content.Text, tt.expectUnit) {
							t.Errorf("expected unit %s in truncation message, got: %s", tt.expectUnit, content.Text)
						}
					}
				}
				if !foundTruncMsg {
					t.Error("truncation message not found in content")
				}

				if tt.expectPreviewNonEmpty {
					preview := strings.Split(truncationText, "\n\n(Output truncated:")[0]
					if strings.TrimSpace(preview) == "" {
						t.Error("expected truncated preview to contain original content, got empty preview")
					}
				}

				// Verify non-text content is preserved
				originalNonText := 0
				resultNonText := 0
				for _, c := range tt.content {
					if c.Type != "text" {
						originalNonText++
					}
				}
				for _, c := range result.Content {
					if c.Type != "text" {
						resultNonText++
					}
				}
				if originalNonText != resultNonText {
					t.Errorf("non-text content count mismatch: original=%d, result=%d", originalNonText, resultNonText)
				}
			}

			if tt.expectFirstType != "" {
				if len(result.Content) == 0 || result.Content[0].Type != tt.expectFirstType {
					t.Errorf("expected first content type %s, got %v", tt.expectFirstType, result.Content)
				}
			}
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with spaces", "with-spaces"},
		{"with/slash", "with_slash"},
		{"with:colon", "with_colon"},
		{"with*asterisk", "with_asterisk"},
		{"with?question", "with_question"},
		{"with\"quote", "with_quote"},
		{"with<greater>", "with_greater_"},
		{"with|pipe", "with_pipe"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func makeLines(count int) []string {
	lines := make([]string, count)
	for i := 0; i < count; i++ {
		lines[i] = "line " + string(rune('0'+i%10))
	}
	return lines
}
