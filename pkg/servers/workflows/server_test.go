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

	// We should have 3 workflows + 1 supporting file (from in-progress dir without SKILL.md)
	if len(result.Resources) != 4 {
		t.Errorf("expected 4 resources, got %d", len(result.Resources))
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
	if testWf.Meta["name"] != "test-workflow" {
		t.Errorf("test-workflow Meta[name] = %q, want 'test-workflow'", testWf.Meta["name"])
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
	if anotherWf.Meta["name"] != "another" {
		t.Errorf("another Meta[name] = %q, want 'another'", anotherWf.Meta["name"])
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
		name          string
		uri           string
		expectError   bool
		shouldContain string
		expectName    string
		expectMeta    map[string]string
	}{
		{
			name:          "read workflow with standard URI",
			uri:           "workflow:///test-workflow",
			expectError:   false,
			shouldContain: "## Inputs",
			expectName:    "test-workflow",
			expectMeta: map[string]string{
				"name":      "test-workflow",
				"createdAt": "2026-01-15T09:00:00Z",
			},
		},
		{
			name:          "read another workflow",
			uri:           "workflow:///another",
			expectError:   false,
			shouldContain: "## Steps",
			expectName:    "another",
			expectMeta: map[string]string{
				"name":      "another",
				"createdAt": "2026-01-16T10:30:00Z",
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

func TestResourcesListSupportingFilesWithoutSkillMD(t *testing.T) {
	restore := withWorkingDir(t, testdataDir(t, "with-workflows"))
	defer restore()

	server := NewServer()
	ctx := context.Background()

	result, err := server.resourcesList(ctx, mcp.Message{}, mcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("resourcesList() failed: %v", err)
	}

	// The in-progress directory has script.py but no SKILL.md.
	// Supporting files should still be listed so clients can read them.
	var found bool
	for _, res := range result.Resources {
		if res.Name == "script.py" && strings.Contains(res.URI, "in-progress") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected supporting file script.py from in-progress workflow (no SKILL.md) to be listed")
	}

	// The in-progress directory should NOT produce a workflow:/// resource
	for _, res := range result.Resources {
		if res.URI == "workflow:///in-progress" {
			t.Error("should not list workflow:/// resource when SKILL.md is missing")
		}
	}
}

func TestResourcesReadSupportingFileWithoutSkillMD(t *testing.T) {
	restore := withWorkingDir(t, testdataDir(t, "with-workflows"))
	defer restore()

	server := NewServer()
	ctx := context.Background()

	// Read the supporting file directly via file:/// URI
	uri := "file:///workflows/in-progress/script.py"
	result, err := server.resourcesRead(ctx, mcp.Message{}, mcp.ReadResourceRequest{URI: uri})
	if err != nil {
		t.Fatalf("resourcesRead() should succeed for supporting file without SKILL.md: %v", err)
	}
	if result == nil || len(result.Contents) == 0 {
		t.Fatal("expected non-empty contents")
	}
	if result.Contents[0].Text == nil || *result.Contents[0].Text == "" {
		t.Error("expected non-empty text content")
	}
}
