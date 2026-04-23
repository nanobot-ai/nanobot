package obotmcp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
)

const (
	configDirName  = "mcp-cli"
	configFileName = "config.json"
)

var (
	nameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	configLocks   sync.Map
)

type ConnectedServer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitempty"`
	Description string `json:"description,omitempty"`
	ConnectURL  string `json:"connect_url"`
}

type inventoryEntry struct {
	Name   string
	Server ConnectedServer
}

type config struct {
	MCPServers map[string]configServer `json:"mcpServers"`
}

type configServer struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func mcpCLIConfigPath(configDir string) string {
	return filepath.Join(storageRoot(configDir), configDirName, configFileName)
}

func storageRoot(configDir string) string {
	base := configDir
	if base == "" {
		if cwd, err := os.Getwd(); err == nil {
			return filepath.Join(cwd, ".nanobot")
		}
		return ".nanobot"
	}

	if !filepath.IsAbs(base) {
		if cwd, err := os.Getwd(); err == nil {
			base = filepath.Join(cwd, base)
		}
	}

	return base
}

func buildConfig(servers []ConnectedServer) config {
	cfg := config{
		MCPServers: map[string]configServer{},
	}

	for _, entry := range buildInventoryEntries(servers) {
		cfg.MCPServers[entry.Name] = configServer{
			URL: entry.Server.ConnectURL,
			Headers: map[string]string{
				"Authorization": "Bearer ${MCP_API_KEY}",
			},
		}
	}

	return cfg
}

func PrepareMCPCLIConfig(ctx context.Context, configDir string, force bool) (string, error) {
	return prepareMCPCLIConfig(ctx, configDir, force, obotConnectedServerLister{})
}

func prepareMCPCLIConfig(ctx context.Context, configDir string, force bool, lister connectedServerLister) (string, error) {
	configPath := mcpCLIConfigPath(configDir)

	if force {
		globalConnectedServersCache.invalidate()
	}

	servers, err := lister.ConnectedMCPServers(ctx)
	if err != nil {
		if !errors.Is(err, ErrSearchNotConfigured) {
			return "", err
		}
		servers = nil
	}

	if err := writeMCPCLIConfig(configPath, buildConfig(servers)); err != nil {
		return "", err
	}

	return configPath, nil
}

func buildInventoryEntries(servers []ConnectedServer) []inventoryEntry {
	if len(servers) == 0 {
		return nil
	}

	sortedServers := slices.Clone(servers)
	slices.SortFunc(sortedServers, func(a, b ConnectedServer) int {
		if cmp := strings.Compare(a.Name, b.Name); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.ID, b.ID)
	})

	entries := make([]inventoryEntry, 0, len(sortedServers))
	seenURLs := map[string]struct{}{}
	usedNames := map[string]struct{}{}
	for _, server := range sortedServers {
		if server.ConnectURL == "" {
			continue
		}
		if _, seen := seenURLs[server.ConnectURL]; seen {
			continue
		}
		seenURLs[server.ConnectURL] = struct{}{}

		entries = append(entries, inventoryEntry{
			Name:   configServerName(server, usedNames),
			Server: server,
		})
	}

	return entries
}

func writeMCPCLIConfig(path string, cfg config) error {
	lock := configLock(path)
	lock.Lock()
	defer lock.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create mcp-cli config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mcp-cli config: %w", err)
	}
	data = append(data, '\n')

	if existing, err := os.ReadFile(path); err == nil {
		if bytes.Equal(existing, data) {
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read existing mcp-cli config: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp mcp-cli config: %w", err)
	}

	tmpPath := tmpFile.Name()
	cleanup := true
	defer func() {
		_ = tmpFile.Close()
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("write temp mcp-cli config: %w", err)
	}
	if err := tmpFile.Chmod(0o644); err != nil {
		return fmt.Errorf("chmod temp mcp-cli config: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp mcp-cli config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("install mcp-cli config: %w", err)
	}

	cleanup = false
	return nil
}

func configServerName(server ConnectedServer, usedNames map[string]struct{}) string {
	base := server.Alias
	if base == "" {
		base = server.Name
	}
	if base == "" {
		base = server.ID
	}

	name := sanitizeName(base)
	if name == "" {
		name = "server"
	}
	if _, exists := usedNames[name]; !exists {
		usedNames[name] = struct{}{}
		return name
	}

	suffixSource := server.ID
	if suffixSource == "" {
		suffixSource = server.ConnectURL
	}
	name = fmt.Sprintf("%s-%s", name, shortHash(suffixSource))
	if _, exists := usedNames[name]; !exists {
		usedNames[name] = struct{}{}
		return name
	}

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", name, i)
		if _, exists := usedNames[candidate]; !exists {
			usedNames[candidate] = struct{}{}
			return candidate
		}
	}
}

func sanitizeName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = nameSanitizer.ReplaceAllString(value, "-")
	return strings.Trim(value, "-._")
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:8]
}

func configLock(path string) *sync.Mutex {
	lock, _ := configLocks.LoadOrStore(path, &sync.Mutex{})
	return lock.(*sync.Mutex)
}
