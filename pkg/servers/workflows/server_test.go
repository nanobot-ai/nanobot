package workflows

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

// testdataDir returns the absolute path to the testdata directory
func testdataDir(t *testing.T, subdir string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get caller info")
	}
	return filepath.Join(filepath.Dir(filename), "testdata", subdir)
}

// withWorkingDir temporarily changes to a directory and restores it after the test
func withWorkingDir(t *testing.T, dir string) func() {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to directory %s: %v", dir, err)
	}
	return func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}
}

func TestResourcesList(t *testing.T) {
	restore := withWorkingDir(t, testdataDir(t, "with-workflows"))
	defer restore()

	server := NewServer()
	ctx := context.Background()

	result, err := server.resourcesList(ctx, mcp.Message{}, mcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("resourcesList() failed: %v", err)
	}
	if result == nil {
		t.Fatal("resourcesList() returned nil result")
	}

	// We should have 3 workflows in the test directory
	if len(result.Resources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(result.Resources))
	}

	// Verify resources are present with correct names and URIs
	resourceMap := make(map[string]mcp.Resource)
	for _, res := range result.Resources {
		resourceMap[res.Name] = res
	}

	if _, ok := resourceMap["test-workflow"]; !ok {
		t.Error("should have test-workflow")
	}
	if _, ok := resourceMap["another"]; !ok {
		t.Error("should have another workflow")
	}
	if _, ok := resourceMap["no-description"]; !ok {
		t.Error("should have no-description workflow")
	}

	// Verify description and URI format
	testWf := resourceMap["test-workflow"]
	if testWf.Description != "This is a test workflow for unit testing purposes." {
		t.Errorf("test-workflow description = %q, want 'This is a test workflow for unit testing purposes.'", testWf.Description)
	}
	if testWf.URI != "workflow:///test-workflow" {
		t.Errorf("test-workflow URI = %q, want 'workflow:///test-workflow'", testWf.URI)
	}
	if testWf.MimeType != "text/markdown" {
		t.Errorf("test-workflow MimeType = %q, want 'text/markdown'", testWf.MimeType)
	}

	anotherWf := resourceMap["another"]
	if anotherWf.Description != "Another workflow for testing multiple workflow listing." {
		t.Errorf("another description = %q, want 'Another workflow for testing multiple workflow listing.'", anotherWf.Description)
	}

	// no-description should have empty description since it doesn't have "# Workflow:" header
	noDescWf := resourceMap["no-description"]
	if noDescWf.Description != "" {
		t.Errorf("no-description should have empty description, got %q", noDescWf.Description)
	}
}

func TestResourcesListMissingDirectory(t *testing.T) {
	// Create a temp directory without a workflows subdirectory
	tempDir := t.TempDir()
	restore := withWorkingDir(t, tempDir)
	defer restore()

	server := NewServer()
	ctx := context.Background()

	result, err := server.resourcesList(ctx, mcp.Message{}, mcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("resourcesList() should not error on missing directory: %v", err)
	}
	if result == nil {
		t.Fatal("resourcesList() returned nil result")
	}

	// Should return empty list
	if len(result.Resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(result.Resources))
	}
}

func TestResourcesListEmptyDirectory(t *testing.T) {
	restore := withWorkingDir(t, testdataDir(t, "empty"))
	defer restore()

	server := NewServer()
	ctx := context.Background()

	result, err := server.resourcesList(ctx, mcp.Message{}, mcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("resourcesList() failed: %v", err)
	}
	if result == nil {
		t.Fatal("resourcesList() returned nil result")
	}

	// Should return empty list (only .gitkeep exists)
	if len(result.Resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(result.Resources))
	}
}

func TestResourcesRead(t *testing.T) {
	restore := withWorkingDir(t, testdataDir(t, "with-workflows"))
	defer restore()

	server := NewServer()
	ctx := context.Background()

	tests := []struct {
		name          string
		uri           string
		expectError   bool
		shouldContain string
		expectName    string
	}{
		{
			name:          "read workflow with standard URI",
			uri:           "workflow:///test-workflow",
			expectError:   false,
			shouldContain: "# Workflow: test-workflow",
			expectName:    "test-workflow",
		},
		{
			name:          "read another workflow",
			uri:           "workflow:///another",
			expectError:   false,
			shouldContain: "# Workflow: another",
			expectName:    "another",
		},
		{
			name:        "nonexistent workflow",
			uri:         "workflow:///nonexistent-workflow",
			expectError: true,
		},
		{
			name:        "invalid URI format",
			uri:         "invalid://workflow",
			expectError: true,
		},
		{
			name:        "empty workflow name",
			uri:         "workflow:///",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.resourcesRead(ctx, mcp.Message{}, mcp.ReadResourceRequest{URI: tt.uri})

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("resourcesRead() failed: %v", err)
				}
				if result == nil || len(result.Contents) == 0 {
					t.Fatal("expected non-empty contents")
				}
				content := result.Contents[0]
				if content.Text == nil || *content.Text == "" {
					t.Error("expected non-empty text content")
				}
				if tt.shouldContain != "" && !strings.Contains(*content.Text, tt.shouldContain) {
					t.Errorf("content should contain %q", tt.shouldContain)
				}
				if content.Name != tt.expectName {
					t.Errorf("content name = %q, want %q", content.Name, tt.expectName)
				}
				if content.MIMEType != "text/markdown" {
					t.Errorf("content MIMEType = %q, want 'text/markdown'", content.MIMEType)
				}
				if content.URI != tt.uri {
					t.Errorf("content URI = %q, want %q", content.URI, tt.uri)
				}
			}
		})
	}
}

func TestParseWorkflowDescription(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "standard workflow header",
			content: `# Workflow: my-workflow

This is the description.

## Steps`,
			expected: "This is the description.",
		},
		{
			name: "multi-line description",
			content: `# Workflow: my-workflow

This is the first line.
This is the second line.

## Steps`,
			expected: "This is the first line. This is the second line.",
		},
		{
			name: "no workflow header",
			content: `# Something Else

Description here.`,
			expected: "",
		},
		{
			name: "workflow header with no description",
			content: `# Workflow: my-workflow

## Steps`,
			expected: "",
		},
		{
			name:     "workflow header at end of file",
			content:  `# Workflow: my-workflow`,
			expected: "",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name: "description with multiple paragraphs",
			content: `# Workflow: my-workflow

First paragraph only.

Second paragraph is ignored.

## Steps`,
			expected: "First paragraph only.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseWorkflowDescription(tt.content)
			if result != tt.expected {
				t.Errorf("parseWorkflowDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}
