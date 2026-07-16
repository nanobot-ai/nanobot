package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmdpkg "github.com/obot-platform/nanobot/pkg/cmd"
	"github.com/obot-platform/nanobot/pkg/config"
	"github.com/obot-platform/nanobot/pkg/session"
)

func TestNanobotConfigPathsDefault(t *testing.T) {
	n := &Nanobot{}

	paths := n.ConfigPaths()
	if len(paths) != 1 || paths[0] != config.DefaultConfigPath {
		t.Fatalf("expected default config path [.nanobot/], got %v", paths)
	}
}

func TestRuntimeConfigDirDefaultsWhenNoPathsExist(t *testing.T) {
	configDir := runtimeConfigDir([]string{"./missing", "https://example.com/nanobot.yaml"})
	if configDir != config.DefaultConfigPath {
		t.Fatalf("expected default config dir %q, got %q", config.DefaultConfigPath, configDir)
	}
}

func TestRuntimeConfigDirKeepsDefaultForExistingDirectory(t *testing.T) {
	dir := t.TempDir()

	configDir := runtimeConfigDir([]string{"./missing", dir})
	if configDir != config.DefaultConfigPath {
		t.Fatalf("expected default config dir %q, got %q", config.DefaultConfigPath, configDir)
	}
}

func TestRuntimeConfigDirReturnsFileParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nanobot.yaml")
	if err := os.WriteFile(path, []byte("agents: {}\n"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	configDir := runtimeConfigDir([]string{path})
	if configDir != dir {
		t.Fatalf("expected config dir %q, got %q", dir, configDir)
	}
}

func TestRunSessionLifecycleFlagsExposeBoundedDefaultsAndEnvironmentVariables(t *testing.T) {
	run := NewRun(&Nanobot{})
	command := cmdpkg.Command(run)

	tests := []struct {
		name         string
		defaultValue string
		env          string
	}{
		{name: "max-live-sessions", defaultValue: "4", env: "NANOBOT_MAX_LIVE_SESSIONS"},
		{name: "live-session-idle-ttl", defaultValue: "10s", env: "NANOBOT_LIVE_SESSION_IDLE_TTL"},
		{name: "event-stream-max-lifetime", defaultValue: "5m", env: "NANOBOT_EVENT_STREAM_MAX_LIFETIME"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := command.PersistentFlags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag --%s was not registered", tt.name)
			}
			if flag.DefValue != tt.defaultValue {
				t.Fatalf("--%s default = %q, want %q", tt.name, flag.DefValue, tt.defaultValue)
			}
			if !strings.Contains(flag.Usage, "$"+tt.env) {
				t.Fatalf("--%s help does not mention $%s: %q", tt.name, tt.env, flag.Usage)
			}
		})
	}
}

func TestRunSessionLifecycleOptions(t *testing.T) {
	run := &Run{
		SessionGCPeriod:        "168h",
		LiveSessionIdleTTL:     "10s",
		MaxLiveSessions:        4,
		EventStreamMaxLifetime: "5m",
	}

	managerOptions, streamLifetime, err := run.sessionLifecycleOptions()
	if err != nil {
		t.Fatal(err)
	}
	if managerOptions.DatabaseGarbageCollectionPeriod != 168*time.Hour {
		t.Fatalf("database GC period = %s", managerOptions.DatabaseGarbageCollectionPeriod)
	}
	if managerOptions.LiveSessionIdleTTL != session.DefaultLiveSessionIdleTTL {
		t.Fatalf("live session idle TTL = %s", managerOptions.LiveSessionIdleTTL)
	}
	if managerOptions.MaxLiveSessions != session.DefaultMaxLiveSessions {
		t.Fatalf("max live sessions = %d", managerOptions.MaxLiveSessions)
	}
	if streamLifetime != 5*time.Minute {
		t.Fatalf("event stream max lifetime = %s", streamLifetime)
	}
}

func TestRunSessionLifecycleOptionsRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		run  Run
	}{
		{
			name: "non-positive max sessions",
			run: Run{
				SessionGCPeriod:        "168h",
				LiveSessionIdleTTL:     "10s",
				MaxLiveSessions:        0,
				EventStreamMaxLifetime: "5m",
			},
		},
		{
			name: "non-positive idle TTL",
			run: Run{
				SessionGCPeriod:        "168h",
				LiveSessionIdleTTL:     "0s",
				MaxLiveSessions:        4,
				EventStreamMaxLifetime: "5m",
			},
		},
		{
			name: "non-positive stream lifetime",
			run: Run{
				SessionGCPeriod:        "168h",
				LiveSessionIdleTTL:     "10s",
				MaxLiveSessions:        4,
				EventStreamMaxLifetime: "0s",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := tt.run.sessionLifecycleOptions(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
