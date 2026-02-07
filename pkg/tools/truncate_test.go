package tools

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
		name        string
		toolName    string
		content     []mcp.Content
		maxBytes    int
		shouldTrunc bool
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
			maxBytes:    50000,
			shouldTrunc: true,
		},
		{
			name:     "mixed text and non-text exceeds budget",
			toolName: "test-tool",
			content: []mcp.Content{
				{Type: "text", Text: strings.Repeat("a", 100000)},
				{Type: "image", Data: "base64data"},
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

			result, err := TruncateToolOutput(ctx, tt.toolName, "call-123", response, tt.maxBytes)
			if err != nil {
				t.Fatalf("TruncateToolOutput failed: %v", err)
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
			} else {
				if result.Truncated {
					t.Error("expected truncated=false for no truncation")
				}
			}
		})
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

	result, err := TruncateToolOutput(ctx, "drop-tool", "call-drop", response, 50000)
	if err != nil {
		t.Fatalf("TruncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
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

func TestTruncateCallIDInFilename(t *testing.T) {
	ctx := context.Background()
	setupTruncateTestDir(t)

	response := &types.CallResult{
		Content: []mcp.Content{
			{Type: "text", Text: strings.Repeat("x", 100000)},
		},
	}

	result, err := TruncateToolOutput(ctx, "my-tool", "call_abc123", response, 50000)
	if err != nil {
		t.Fatalf("TruncateToolOutput failed: %v", err)
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

	result, err := TruncateToolOutput(ctx, "resource-tool", "call-res", response, 50000)
	if err != nil {
		t.Fatalf("TruncateToolOutput failed: %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncation")
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
			size := ContentByteSize(tt.content)
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
			result := SanitizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
