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

// setupTruncateTestDir creates a temp directory and changes to it, returning a cleanup function.
func setupTruncateTestDir(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "truncate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory %q: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory to %q: %v", origDir, err)
		}
		os.RemoveAll(tmpDir)
	})
}

func TestTruncateToolOutput(t *testing.T) {
	ctx := context.Background()
	setupTruncateTestDir(t)

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
						t.Fatal("expected at least 2 content items (notice + preview)")
					}
					// Index 0 is the truncation notice, index 1 is the first original content item
					if strings.TrimSpace(result.Content[1].Text) == "" {
						t.Error("expected truncated preview to contain original content, got empty preview")
					}
				}
			} else {
				if result.Truncated {
					t.Error("expected truncated=false for no truncation")
				}
			}

			if tt.expectFirstType != "" && result.Truncated {
				// Index 0 is the truncation notice, index 1 is the first original content item
				if len(result.Content) < 2 || result.Content[1].Type != tt.expectFirstType {
					t.Errorf("expected first content type after notice to be %s, got %v", tt.expectFirstType, result.Content)
				}
			}
		})
	}
}

func TestTruncateInterleavedOrder(t *testing.T) {
	ctx := context.Background()
	setupTruncateTestDir(t)

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

	// Verify ordering: notice first, then text, image, truncated text
	if len(result.Content) < 4 {
		t.Fatalf("expected at least 4 content items, got %d", len(result.Content))
	}
	if result.Content[0].Type != "text" || !strings.Contains(result.Content[0].Text, "Tool output truncated") {
		t.Errorf("expected first item to be truncation notice, got type=%s text=%q", result.Content[0].Type, result.Content[0].Text)
	}
	if result.Content[1].Type != "text" || result.Content[1].Text != "first" {
		t.Errorf("expected second item to be text 'first', got type=%s text=%q", result.Content[1].Type, result.Content[1].Text)
	}
	if result.Content[2].Type != "image" {
		t.Errorf("expected third item to be image, got %s", result.Content[2].Type)
	}
	if result.Content[3].Type != "text" {
		t.Errorf("expected fourth item to be truncated text, got %s", result.Content[3].Type)
	}
}

func TestTruncateNonTextByteBudget(t *testing.T) {
	ctx := context.Background()
	setupTruncateTestDir(t)

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

	// Notice first, then image kept (fits), then text truncated
	if len(result.Content) != 3 {
		t.Fatalf("expected 3 content items, got %d", len(result.Content))
	}
	if !strings.Contains(result.Content[0].Text, "Tool output truncated") {
		t.Errorf("expected truncation notice first, got %q", result.Content[0].Text)
	}
	if result.Content[1].Type != "image" {
		t.Errorf("expected image second, got %s", result.Content[1].Type)
	}
	if result.Content[2].Type != "text" {
		t.Errorf("expected text third, got %s", result.Content[2].Type)
	}
	if len(result.Content[2].Text) >= 20000 {
		t.Errorf("expected text to be truncated, got %d bytes", len(result.Content[2].Text))
	}
}

func TestTruncateLargeNonTextDropped(t *testing.T) {
	ctx := context.Background()
	setupTruncateTestDir(t)

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
	setupTruncateTestDir(t)

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
	setupTruncateTestDir(t)

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

func TestTruncateEmbeddedResource(t *testing.T) {
	ctx := context.Background()
	setupTruncateTestDir(t)

	response := &types.CallResult{
		Content: []mcp.Content{
			{
				Type: "resource",
				Resource: &mcp.EmbeddedResource{
					URI:      "file:///large.txt",
					MIMEType: "text/plain",
					Text:     strings.Repeat("r", 60000),
				},
			},
			{Type: "text", Text: "small text"},
		},
	}

	result, err := truncateToolOutput(ctx, "resource-tool", "call-res", response, 50000)
	if err != nil {
		t.Fatalf("truncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
	}

	// Large resource dropped, text kept
	for _, c := range result.Content {
		if c.Type == "resource" {
			t.Error("expected large resource to be dropped")
		}
	}

	foundNotice := false
	for _, c := range result.Content {
		if strings.Contains(c.Text, "Tool output truncated") {
			foundNotice = true
		}
	}
	if !foundNotice {
		t.Fatal("expected truncation notice")
	}
}

func TestContentByteSize(t *testing.T) {
	tests := []struct {
		name     string
		content  mcp.Content
		expected int
	}{
		{
			name:     "text content",
			content:  mcp.Content{Type: "text", Text: "hello"},
			expected: 5,
		},
		{
			name:     "image content",
			content:  mcp.Content{Type: "image", Data: "base64data"},
			expected: 10,
		},
		{
			name:     "audio content",
			content:  mcp.Content{Type: "audio", Data: "audiodata"},
			expected: 9,
		},
		{
			name: "resource with text",
			content: mcp.Content{
				Type:     "resource",
				Resource: &mcp.EmbeddedResource{Text: "resource text"},
			},
			expected: 13,
		},
		{
			name: "resource with blob",
			content: mcp.Content{
				Type:     "resource",
				Resource: &mcp.EmbeddedResource{Blob: "blobdata"},
			},
			expected: 8,
		},
		{
			name:     "resource with nil resource",
			content:  mcp.Content{Type: "resource"},
			expected: 0,
		},
		{
			name:     "resource_link",
			content:  mcp.Content{Type: "resource_link", URI: "file:///test"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := contentByteSize(tt.content)
			if size != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, size)
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
		{"with.dot", "with_dot"},
		{"..", "__"},
		{"../../../etc", "_________etc"},
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
