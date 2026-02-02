package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCloneGitHubRepoIfConfigured_EmptyCloneURL(t *testing.T) {
	// Save original env vars
	originalCloneURL := os.Getenv("GITHUB_CLONE_URL")
	originalRepo := os.Getenv("GITHUB_REPO")
	defer func() {
		os.Setenv("GITHUB_CLONE_URL", originalCloneURL)
		os.Setenv("GITHUB_REPO", originalRepo)
	}()

	// Test with empty GITHUB_CLONE_URL
	os.Setenv("GITHUB_CLONE_URL", "")
	os.Setenv("GITHUB_REPO", "")

	err := cloneGitHubRepoIfConfigured()
	if err != nil {
		t.Errorf("Expected no error with empty GITHUB_CLONE_URL, got: %v", err)
	}
}

func TestCloneGitHubRepoIfConfigured_MissingGitHubRepo(t *testing.T) {
	// Save original env vars
	originalCloneURL := os.Getenv("GITHUB_CLONE_URL")
	originalRepo := os.Getenv("GITHUB_REPO")
	defer func() {
		os.Setenv("GITHUB_CLONE_URL", originalCloneURL)
		os.Setenv("GITHUB_REPO", originalRepo)
	}()

	// Test with GITHUB_CLONE_URL set but GITHUB_REPO empty
	os.Setenv("GITHUB_CLONE_URL", "https://github.com/example/repo.git")
	os.Setenv("GITHUB_REPO", "")

	err := cloneGitHubRepoIfConfigured()
	if err == nil {
		t.Error("Expected error when GITHUB_CLONE_URL is set but GITHUB_REPO is not")
	}
	expectedMsg := "GITHUB_CLONE_URL is set but GITHUB_REPO is not"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestCloneGitHubRepoIfConfigured_InvalidRepoFormat(t *testing.T) {
	// Save original env vars
	originalCloneURL := os.Getenv("GITHUB_CLONE_URL")
	originalRepo := os.Getenv("GITHUB_REPO")
	defer func() {
		os.Setenv("GITHUB_CLONE_URL", originalCloneURL)
		os.Setenv("GITHUB_REPO", originalRepo)
	}()

	testCases := []struct {
		name     string
		repoName string
	}{
		{
			name:     "no slash",
			repoName: "repo",
		},
		{
			name:     "too many slashes",
			repoName: "owner/repo/extra",
		},
		{
			name:     "empty string",
			repoName: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("GITHUB_CLONE_URL", "https://github.com/example/repo.git")
			os.Setenv("GITHUB_REPO", tc.repoName)

			err := cloneGitHubRepoIfConfigured()
			if err == nil {
				t.Errorf("Expected error for invalid repo format %q", tc.repoName)
			}
		})
	}
}

func TestCloneGitHubRepoIfConfigured_DirectoryAlreadyExists(t *testing.T) {
	// Save original env vars and working directory
	originalCloneURL := os.Getenv("GITHUB_CLONE_URL")
	originalRepo := os.Getenv("GITHUB_REPO")
	originalWd, _ := os.Getwd()
	defer func() {
		os.Setenv("GITHUB_CLONE_URL", originalCloneURL)
		os.Setenv("GITHUB_REPO", originalRepo)
		os.Chdir(originalWd)
	}()

	// Create a temporary directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create a directory that matches the repo name
	repoName := "test-repo"
	repoDir := filepath.Join(tmpDir, repoName)
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Set env vars
	os.Setenv("GITHUB_CLONE_URL", "https://github.com/owner/test-repo.git")
	os.Setenv("GITHUB_REPO", "owner/test-repo")

	// Should not error and should not attempt to clone
	err := cloneGitHubRepoIfConfigured()
	if err != nil {
		t.Errorf("Expected no error when directory already exists, got: %v", err)
	}

	// Verify we're now in the repo directory
	// Use EvalSymlinks to resolve any symlinks (e.g., /var -> /private/var on macOS)
	currentWd, _ := os.Getwd()
	expectedPath, _ := filepath.EvalSymlinks(repoDir)
	actualPath, _ := filepath.EvalSymlinks(currentWd)
	if actualPath != expectedPath {
		t.Errorf("Expected working directory to be %q, got %q", expectedPath, actualPath)
	}
}
