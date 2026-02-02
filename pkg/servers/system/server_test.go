package system

import (
	"os"
	"testing"
)

func TestGetWorkdir_WithExplicitWorkdir(t *testing.T) {
	expectedWorkdir := "/some/explicit/path"
	server := NewServer("", expectedWorkdir)

	workdir := server.getWorkdir()
	if workdir != expectedWorkdir {
		t.Errorf("Expected workdir %q, got %q", expectedWorkdir, workdir)
	}
}

func TestGetWorkdir_WithEmptyWorkdir(t *testing.T) {
	server := NewServer("", "")

	workdir := server.getWorkdir()

	// Should return current working directory
	expectedWorkdir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	if workdir != expectedWorkdir {
		t.Errorf("Expected workdir to be current directory %q, got %q", expectedWorkdir, workdir)
	}
}

func TestGetWorkdir_FallbackToDot(t *testing.T) {
	// This test verifies the fallback behavior, though it's hard to trigger
	// the os.Getwd() error in normal circumstances
	server := NewServer("", "")

	// Even if we can't force an error, the workdir should be valid
	workdir := server.getWorkdir()
	if workdir == "" {
		t.Error("Expected workdir to not be empty")
	}
}
