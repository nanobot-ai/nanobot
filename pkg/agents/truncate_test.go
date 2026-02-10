package agents

import (
	"context"
	"encoding/json"
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
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory %q: %v", tmpDir, err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory to %q: %v", origDir, err)
		}
	}()

	tests := []struct {
		name                  string
		toolName              string
		content               []mcp.Content
		maxBytes              int
		shouldTrunc           bool
		expectFirstType       string
		expectPreviewNonEmpty bool
	}{
		{
			name:     "no truncation needed - small content",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: "line 1\nline 2\nline 3"},
			},
			maxBytes:    1000,
			shouldTrunc: false,
		},
		{
			name:     "truncate by bytes",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: strings.Repeat("a", 100000)},
			},
			maxBytes:              50000,
			shouldTrunc:           true,
			expectPreviewNonEmpty: true,
		},
		{
			name:     "preserve non-text content when it fits",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: strings.Repeat("a", 100000)},
				{Type: "image", Data: "base64data"},
			},
			maxBytes:    50000,
			shouldTrunc: true,
		},
		{
			name:     "preserve leading non-text order",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "image", Data: "base64data"},
				{Type: "text", Text: strings.Repeat("a", 100000)},
			},
			maxBytes:        50000,
			shouldTrunc:     true,
			expectFirstType: "image",
		},
		{
			name:     "interleaved content preserves order",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: "hello"},
				{Type: "image", Data: "imgdata"},
				{Type: "text", Text: strings.Repeat("b", 100000)},
			},
			maxBytes:    50000,
			shouldTrunc: true,
		},
		{
			name:     "non-text content contributes to byte budget",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "image", Data: strings.Repeat("x", 40000)},
				{Type: "text", Text: strings.Repeat("a", 20000)},
			},
			maxBytes:    50000,
			shouldTrunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &types.CallResult{
				Content: tt.content,
			}

			result, err := truncateToolOutput(ctx, tt.toolName, response, tt.maxBytes)
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

				// Verify a truncation/drop notice is in the content
				foundNotice := false
				for _, content := range result.Content {
					if content.Type == "text" && (strings.Contains(content.Text, "Output truncated") || strings.Contains(content.Text, "dropped")) {
						foundNotice = true
					}
				}
				if !foundNotice {
					t.Error("truncation/drop notice not found in content")
				}

				if tt.expectPreviewNonEmpty {
					// The truncated text content is separate from the notice
					if len(result.Content) < 2 {
						t.Fatal("expected at least 2 content items for preview check")
					}
					// First text item should have the preview content
					if strings.TrimSpace(result.Content[0].Text) == "" {
						t.Error("expected truncated preview to contain original content, got empty preview")
					}
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

func TestTruncateInterleavedOrder(t *testing.T) {
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "truncate-order-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// [text, image, text] where total exceeds budget
	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "text", Text: "first"},
			{Type: "image", Data: "imgdata"},
			{Type: "text", Text: strings.Repeat("z", 100000)},
		},
	}

	result, err := truncateToolOutput(ctx, "order-tool", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}

	// Verify ordering: first text, then image, then truncated text, then notice
	if len(result.Content) < 4 {
		t.Fatalf("expected at least 4 content items, got %d", len(result.Content))
	}
	if result.Content[0].Type != "text" || result.Content[0].Text != "first" {
		t.Errorf("expected first item to be text 'first', got type=%s text=%q", result.Content[0].Type, result.Content[0].Text)
	}
	if result.Content[1].Type != "image" {
		t.Errorf("expected second item to be image, got %s", result.Content[1].Type)
	}
	if result.Content[2].Type != "text" {
		t.Errorf("expected third item to be truncated text, got %s", result.Content[2].Type)
	}
	if result.Content[3].Type != "text" || !strings.Contains(result.Content[3].Text, "Output truncated") {
		t.Errorf("expected fourth item to be truncation notice, got type=%s text=%q", result.Content[3].Type, result.Content[3].Text)
	}
}

func TestTruncateNonTextByteBudget(t *testing.T) {
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "truncate-budget-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Image takes most of the budget, text must be truncated
	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "image", Data: strings.Repeat("x", 40000)},
			{Type: "text", Text: strings.Repeat("a", 20000)},
		},
	}

	result, err := truncateToolOutput(ctx, "budget-tool", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}

	// Image should be kept (fits in budget), text should be truncated, then notice
	if len(result.Content) != 3 {
		t.Fatalf("expected 3 content items, got %d", len(result.Content))
	}
	if result.Content[0].Type != "image" {
		t.Errorf("expected image first, got %s", result.Content[0].Type)
	}
	if result.Content[1].Type != "text" {
		t.Errorf("expected text second, got %s", result.Content[1].Type)
	}
	// The text should be shorter than original 20000 bytes (budget remaining = 10000)
	if len(result.Content[1].Text) >= 20000 {
		t.Errorf("expected text to be truncated, got %d bytes", len(result.Content[1].Text))
	}
	if !strings.Contains(result.Content[2].Text, "Output truncated") {
		t.Errorf("expected truncation notice, got %q", result.Content[2].Text)
	}
}

func TestTruncateLargeNonTextDropped(t *testing.T) {
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "truncate-drop-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Image exceeds budget entirely, text is small
	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "image", Data: strings.Repeat("x", 60000)},
			{Type: "text", Text: "small text"},
		},
	}

	result, err := truncateToolOutput(ctx, "drop-tool", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}

	// Image should be dropped (doesn't fit), text should be kept, drop notice appended
	for _, c := range result.Content {
		if c.Type == "image" {
			t.Error("expected large image to be dropped")
		}
	}

	var droppedFilePath string
	for _, c := range result.Content {
		if strings.Contains(c.Text, "1 non-text content item(s) dropped") {
			// Extract the file path from the notice
			idx := strings.Index(c.Text, "saved to: ")
			if idx >= 0 {
				droppedFilePath = strings.TrimSuffix(strings.TrimSpace(c.Text[idx+len("saved to: "):]), ")")
			}
		}
	}
	if droppedFilePath == "" {
		t.Fatal("expected drop notice with file path for removed non-text item")
	}

	// Verify the dropped content JSON file exists and contains the image
	droppedData, err := os.ReadFile(droppedFilePath)
	if err != nil {
		t.Fatalf("failed to read dropped content file: %v", err)
	}
	var droppedContent []mcp.Content
	if err := json.Unmarshal(droppedData, &droppedContent); err != nil {
		t.Fatalf("failed to unmarshal dropped content: %v", err)
	}
	if len(droppedContent) != 1 {
		t.Fatalf("expected 1 dropped item, got %d", len(droppedContent))
	}
	if droppedContent[0].Type != "image" {
		t.Errorf("expected dropped item to be image, got %s", droppedContent[0].Type)
	}
}

func TestTruncateTwoImagesExceedBudget(t *testing.T) {
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "truncate-two-img-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Two images that individually fit but together exceed budget
	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "image", Data: strings.Repeat("a", 30000)},
			{Type: "image", Data: strings.Repeat("b", 30000)},
		},
	}

	result, err := truncateToolOutput(ctx, "two-img-tool", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}

	// First image fits, second should be dropped
	imageCount := 0
	for _, c := range result.Content {
		if c.Type == "image" {
			imageCount++
		}
	}
	if imageCount != 1 {
		t.Errorf("expected 1 image kept, got %d", imageCount)
	}

	// Should have a drop notice with file path
	var droppedFilePath string
	for _, c := range result.Content {
		if c.Type == "text" && strings.Contains(c.Text, "1 non-text content item(s) dropped") {
			idx := strings.Index(c.Text, "saved to: ")
			if idx >= 0 {
				droppedFilePath = strings.TrimSuffix(strings.TrimSpace(c.Text[idx+len("saved to: "):]), ")")
			}
		}
	}
	if droppedFilePath == "" {
		t.Fatal("expected drop notice with file path for second image")
	}

	// Verify the dropped content JSON file contains the second image
	droppedData, err := os.ReadFile(droppedFilePath)
	if err != nil {
		t.Fatalf("failed to read dropped content file: %v", err)
	}
	var droppedContent []mcp.Content
	if err := json.Unmarshal(droppedData, &droppedContent); err != nil {
		t.Fatalf("failed to unmarshal dropped content: %v", err)
	}
	if len(droppedContent) != 1 {
		t.Fatalf("expected 1 dropped item, got %d", len(droppedContent))
	}
	if droppedContent[0].Type != "image" {
		t.Errorf("expected dropped item to be image, got %s", droppedContent[0].Type)
	}
	if droppedContent[0].Data != strings.Repeat("b", 30000) {
		t.Error("expected dropped item to be the second image")
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
