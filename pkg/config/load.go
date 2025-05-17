package config

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/obot-platform/nanobot/pkg/types"
	"sigs.k8s.io/yaml"
)

func Load(path string) (*types.Config, string, error) {
	var (
		last types.Config
		data []byte
		dir  = "."
	)

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}
	} else {
		s, err := os.Stat(path)
		if err != nil {
			return nil, "", fmt.Errorf("error reading %s: %w", path, err)
		}

		if s.IsDir() {
			path = filepath.Join(path, "nanobot.yaml")
		}

		dir = filepath.Dir(path)
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, "", err
		}
	}

	if err := yaml.Unmarshal(data, &last); err != nil {
		return nil, "", err
	}

	seen := map[string]string{}
	if err := checkDup(seen, "mcpServer", slices.Collect(maps.Keys(last.MCPServers))); err != nil {
		return nil, "", err
	}
	if err := checkDup(seen, "agent", slices.Collect(maps.Keys(last.Agents))); err != nil {
		return nil, "", err
	}

	return &last, dir, nil
}

func checkDup(seen map[string]string, category string, keys []string) error {
	for _, k := range keys {
		if oldCategory, ok := seen[k]; ok {
			return fmt.Errorf("duplicate name [%s] used for both [%s] and [%s]", k, oldCategory, category)
		}
		seen[k] = category
	}
	return nil
}
