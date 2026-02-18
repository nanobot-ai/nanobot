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

	// Verify description, URI format, and _meta from frontmatter
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
	if testWf.Meta == nil {
		t.Fatal("test-workflow Meta should not be nil")
	}
	if testWf.Meta["name"] != "Test Workflow" {
		t.Errorf("test-workflow Meta[name] = %q, want 'Test Workflow'", testWf.Meta["name"])
	}
	if testWf.Meta["description"] != "This is a test workflow for unit testing purposes." {
		t.Errorf("test-workflow Meta[description] = %q", testWf.Meta["description"])
	}
	if testWf.Meta["createdAt"] != "2026-01-15T09:00:00Z" {
		t.Errorf("test-workflow Meta[createdAt] = %q", testWf.Meta["createdAt"])
	}

	anotherWf := resourceMap["another"]
	if anotherWf.Description != "Another workflow for testing multiple workflow listing." {
		t.Errorf("another description = %q, want 'Another workflow for testing multiple workflow listing.'", anotherWf.Description)
	}
	if anotherWf.Meta == nil {
		t.Fatal("another Meta should not be nil")
	}
	if anotherWf.Meta["name"] != "Another Workflow" {
		t.Errorf("another Meta[name] = %q, want 'Another Workflow'", anotherWf.Meta["name"])
	}

	// no-description should have empty description since it doesn't have "# Workflow:" header
	noDescWf := resourceMap["no-description"]
	if noDescWf.Description != "" {
		t.Errorf("no-description should have empty description, got %q", noDescWf.Description)
	}
	if noDescWf.Meta != nil {
		t.Errorf("no-description Meta should be nil, got %v", noDescWf.Meta)
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
		name             string
		uri              string
		expectError      bool
		shouldContain    string
		shouldNotContain string
		expectName       string
		expectMeta       map[string]string
	}{
		{
			name:             "read workflow with standard URI",
			uri:              "workflow:///test-workflow",
			expectError:      false,
			shouldContain:    "## Inputs",
			shouldNotContain: "---\nname:",
			expectName:       "test-workflow",
			expectMeta: map[string]string{
				"name":        "Test Workflow",
				"description": "This is a test workflow for unit testing purposes.",
				"createdAt":   "2026-01-15T09:00:00Z",
			},
		},
		{
			name:             "read another workflow",
			uri:              "workflow:///another",
			expectError:      false,
			shouldContain:    "## Steps",
			shouldNotContain: "---\nname:",
			expectName:       "another",
			expectMeta: map[string]string{
				"name":        "Another Workflow",
				"description": "Another workflow for testing multiple workflow listing.",
				"createdAt":   "2026-01-16T10:30:00Z",
			},
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
				if tt.shouldNotContain != "" && strings.Contains(*content.Text, tt.shouldNotContain) {
					t.Errorf("content should not contain %q (frontmatter should be stripped)", tt.shouldNotContain)
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
				if tt.expectMeta != nil {
					if content.Meta == nil {
						t.Fatal("expected _meta to be set, got nil")
					}
					for key, want := range tt.expectMeta {
						got, ok := content.Meta[key]
						if !ok {
							t.Errorf("_meta missing key %q", key)
						} else if got != want {
							t.Errorf("_meta[%q] = %q, want %q", key, got, want)
						}
					}
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

func TestParseWorkflowFrontmatter(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		expectMeta workflowMeta
		expectBody string
	}{
		{
			name:    "full frontmatter",
			content: "---\nname: Code Review\ndescription: Review code.\ncreatedAt: 2026-01-15T09:00:00Z\n---\n\n## Steps\n\nBody here.",
			expectMeta: workflowMeta{
				Name:        "Code Review",
				Description: "Review code.",
				CreatedAt:   "2026-01-15T09:00:00Z",
			},
			expectBody: "## Steps\n\nBody here.",
		},
		{
			name:       "no frontmatter",
			content:    "# Workflow: test\n\nJust a body.",
			expectMeta: workflowMeta{},
			expectBody: "# Workflow: test\n\nJust a body.",
		},
		{
			name:    "partial fields",
			content: "---\nname: My Workflow\n---\n\n## Steps",
			expectMeta: workflowMeta{
				Name: "My Workflow",
			},
			expectBody: "## Steps",
		},
		{
			name:       "malformed frontmatter - no closing delimiter",
			content:    "---\nname: Broken\n# Workflow: test",
			expectMeta: workflowMeta{},
			expectBody: "---\nname: Broken\n# Workflow: test",
		},
		{
			name:       "malformed frontmatter - invalid yaml",
			content:    "---\n: :\n---\n\n# Workflow: test",
			expectMeta: workflowMeta{},
			expectBody: "---\n: :\n---\n\n# Workflow: test",
		},
		{
			name:       "empty content",
			content:    "",
			expectMeta: workflowMeta{},
			expectBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, body := parseWorkflowFrontmatter(tt.content)
			if meta != tt.expectMeta {
				t.Errorf("meta = %+v, want %+v", meta, tt.expectMeta)
			}
			if body != tt.expectBody {
				t.Errorf("body = %q, want %q", body, tt.expectBody)
			}
		})
	}
}
