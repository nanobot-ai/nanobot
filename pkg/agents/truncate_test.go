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

			result, err := truncateToolOutput(ctx, tt.toolName, "call-123", response, tt.maxBytes)
			if err != nil {
				t.Fatalf("truncateToolOutput failed: %v", err)
			}

			if tt.shouldTrunc {
				if !result.Truncated {
					t.Error("expected truncated=true")
				}

				// Verify JSON file was created
				if result.FilePath == "" {
					t.Error("expected file path to be set when truncated")
				}
				if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
					t.Errorf("truncation file not created at %s", result.FilePath)
				}

				// Verify the JSON file contains the full original content
				fileData, err := os.ReadFile(result.FilePath)
				if err != nil {
					t.Fatalf("failed to read truncation file: %v", err)
				}
				var savedContent []mcp.Content
				if err := json.Unmarshal(fileData, &savedContent); err != nil {
					t.Fatalf("failed to unmarshal saved content: %v", err)
				}
				if len(savedContent) != len(tt.content) {
					t.Errorf("saved content has %d items, expected %d", len(savedContent), len(tt.content))
				}

				// Verify a truncation notice is in the content
				foundNotice := false
				for _, content := range result.Content {
					if content.Type == "text" && strings.Contains(content.Text, "Tool output truncated") {
						foundNotice = true
					}
				}
				if !foundNotice {
					t.Error("truncation notice not found in content")
				}

				if tt.expectPreviewNonEmpty {
					if len(result.Content) < 2 {
						t.Fatal("expected at least 2 content items for preview check")
					}
					if strings.TrimSpace(result.Content[0].Text) == "" {
						t.Error("expected truncated preview to contain original content, got empty preview")
					}
				}
			} else {
				if result.Truncated {
					t.Error("expected truncated=false for no truncation")
				}
			}

			if tt.expectFirstType != "" && result.Truncated {
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

	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "text", Text: "first"},
			{Type: "image", Data: "imgdata"},
			{Type: "text", Text: strings.Repeat("z", 100000)},
		},
	}

	result, err := truncateToolOutput(ctx, "order-tool", "call-456", response, 50000)
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
	if result.Content[3].Type != "text" || !strings.Contains(result.Content[3].Text, "Tool output truncated") {
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

	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "image", Data: strings.Repeat("x", 40000)},
			{Type: "text", Text: strings.Repeat("a", 20000)},
		},
	}

	result, err := truncateToolOutput(ctx, "budget-tool", "call-789", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}

	// Image kept (fits), text truncated, then notice
	if len(result.Content) != 3 {
		t.Fatalf("expected 3 content items, got %d", len(result.Content))
	}
	if result.Content[0].Type != "image" {
		t.Errorf("expected image first, got %s", result.Content[0].Type)
	}
	if result.Content[1].Type != "text" {
		t.Errorf("expected text second, got %s", result.Content[1].Type)
	}
	if len(result.Content[1].Text) >= 20000 {
		t.Errorf("expected text to be truncated, got %d bytes", len(result.Content[1].Text))
	}
	if !strings.Contains(result.Content[2].Text, "Tool output truncated") {
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

	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "image", Data: strings.Repeat("x", 60000)},
			{Type: "text", Text: "small text"},
		},
	}

	result, err := truncateToolOutput(ctx, "drop-tool", "call-drop", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}

	// Image dropped, text kept, drop notice appended
	for _, c := range result.Content {
		if c.Type == "image" {
			t.Error("expected large image to be dropped")
		}
	}

	foundDropNotice := false
	for _, c := range result.Content {
		if strings.Contains(c.Text, "Tool output truncated") {
			foundDropNotice = true
		}
	}
	if !foundDropNotice {
		t.Fatal("expected drop notice for removed non-text item")
	}

	// Verify JSON file has the full original content including the image
	fileData, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	var savedContent []mcp.Content
	if err := json.Unmarshal(fileData, &savedContent); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(savedContent) != 2 {
		t.Fatalf("expected 2 saved items, got %d", len(savedContent))
	}
	if savedContent[0].Type != "image" {
		t.Errorf("expected saved first item to be image, got %s", savedContent[0].Type)
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

	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "image", Data: strings.Repeat("a", 30000)},
			{Type: "image", Data: strings.Repeat("b", 30000)},
		},
	}

	result, err := truncateToolOutput(ctx, "two-img-tool", "call-2img", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}

	// First image fits, second dropped
	imageCount := 0
	for _, c := range result.Content {
		if c.Type == "image" {
			imageCount++
		}
	}
	if imageCount != 1 {
		t.Errorf("expected 1 image kept, got %d", imageCount)
	}

	// Drop notice present
	foundDropNotice := false
	for _, c := range result.Content {
		if c.Type == "text" && strings.Contains(c.Text, "Tool output truncated") {
			foundDropNotice = true
		}
	}
	if !foundDropNotice {
		t.Fatal("expected drop notice for second image")
	}

	// JSON file has both images
	fileData, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	var savedContent []mcp.Content
	if err := json.Unmarshal(fileData, &savedContent); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(savedContent) != 2 {
		t.Fatalf("expected 2 saved items, got %d", len(savedContent))
	}
}

func TestTruncateCallIDInFilename(t *testing.T) {
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "truncate-callid-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "text", Text: strings.Repeat("x", 100000)},
		},
	}

	result, err := truncateToolOutput(ctx, "my-tool", "call_abc123", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}
	if !strings.Contains(result.FilePath, "call_abc123") {
		t.Errorf("expected callID in file path, got %s", result.FilePath)
	}
	if !strings.HasSuffix(result.FilePath, ".json") {
		t.Errorf("expected .json extension, got %s", result.FilePath)
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
