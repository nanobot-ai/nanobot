package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNanobotConfigPathsDefault(t *testing.T) {
	n := &Nanobot{}

	paths := n.ConfigPaths()
	if len(paths) != 1 || paths[0] != ".nanobot/" {
		t.Fatalf("expected default config path [.nanobot/], got %v", paths)
	}
}

func TestNanobotRuntimeConfigDirDirectory(t *testing.T) {
	dir := t.TempDir()
	n := &Nanobot{ConfigPath: []string{dir, "./overlay.yaml"}}

	configDir, err := n.RuntimeConfigDir()
	if err != nil {
		t.Fatalf("expected directory config path to succeed, got %v", err)
	}
	if configDir != dir {
		t.Fatalf("expected config dir %q, got %q", dir, configDir)
	}
}

func TestNanobotRuntimeConfigDirFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nanobot.yaml")
	if err := os.WriteFile(path, []byte("agents: {}\n"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	n := &Nanobot{ConfigPath: []string{path}}
	_, err := n.RuntimeConfigDir()
	if err == nil || !strings.Contains(err.Error(), "must be a directory, not a file") {
		t.Fatalf("expected file config path error, got %v", err)
	}
}

func TestNanobotRuntimeConfigDirURL(t *testing.T) {
	n := &Nanobot{ConfigPath: []string{"https://example.com/nanobot.yaml"}}
	_, err := n.RuntimeConfigDir()
	if err == nil || !strings.Contains(err.Error(), "must be a local directory, not a URL") {
		t.Fatalf("expected URL config path error, got %v", err)
	}
}

func TestNanobotRuntimeConfigDirGitHubRepo(t *testing.T) {
	n := &Nanobot{ConfigPath: []string{"owner/repo"}}
	_, err := n.RuntimeConfigDir()
	if err == nil || !strings.Contains(err.Error(), "must be a local directory, not a GitHub repository") {
		t.Fatalf("expected GitHub repo config path error, got %v", err)
	}
}
