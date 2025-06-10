package config

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/obot-platform/nanobot/pkg/types"
)

func Load(ctx context.Context, path string, profiles ...string) (cfg *types.Config, cwd string, err error) {
	defer func() {
		if err != nil {
			if _, fErr := os.Stat(path); fErr == nil && !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, ".") {
				err = fmt.Errorf("failed to load %q, did you mean ./%s? local files must start with . or /: %w", path, path, err)
			}
		}
	}()
	configResource, err := resolve(path)
	if err != nil {
		return nil, "", fmt.Errorf("error resolving config path %s: %w", path, err)
	}

	return loadResource(ctx, configResource, profiles...)
}

func loadResource(ctx context.Context, configResource *resource, profiles ...string) (*types.Config, string, error) {
	targetCwd, err := configResource.Cwd()
	if err != nil {
		return nil, "", fmt.Errorf("error determining working directory: %w", err)
	}

	last, err := configResource.Load(ctx)
	if err != nil {
		return nil, "", err
	}

	if last.Extends != "" {
		parentResource, err := configResource.Rel(last.Extends)
		if err != nil {
			return nil, "", fmt.Errorf("error resolving extends %s: %w", last.Extends, err)
		}

		parent, err := parentResource.Load(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("error loading parent config %s: %w", parentResource.url, err)
		}

		last, err = merge(parent, last)
		if err != nil {
			return nil, "", fmt.Errorf("error merging %s: %w", last.Extends, err)
		}
	}

	for _, profile := range profiles {
		profileName, _, optional := strings.Cut(profile, "?")
		profileConfig, found := last.Profiles[profileName]
		if !found && !optional {
			return nil, "", fmt.Errorf("profile %s not found", profileName)
		} else if !found {
			continue
		}
		var err error
		last, err = merge(last, profileConfig)
		if err != nil {
			return nil, "", fmt.Errorf("error merging profile %s: %w", profileName, err)
		}
	}

	last, err = rewriteSourceReferences(last, configResource)
	if err != nil {
		return nil, "", fmt.Errorf("error rewriting source references: %w", err)
	}

	return &last, targetCwd, last.Validate(configResource.resourceType == "path")
}

func rewriteSourceReferences(cfg types.Config, resource *resource) (types.Config, error) {
	for name, mcpServer := range cfg.MCPServers {
		var err error
		mcpServer.Source, err = resource.SourceRel(mcpServer.Source)
		if err != nil {
			return types.Config{}, fmt.Errorf("error resolving source for MCP server %s: %w", name, err)
		}
		cfg.MCPServers[name] = mcpServer
	}
	return cfg, nil
}

func toMap(cfg types.Config) (map[string]any, error) {
	result := map[string]any{}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return result, json.Unmarshal(data, &result)
}

func mergeObject(base, overlay any) any {
	if baseMap, ok := base.(map[string]any); ok {
		if overlayMap, ok := overlay.(map[string]any); ok {
			newMap := maps.Clone(baseMap)
			for k, v := range overlayMap {
				newMap[k] = mergeObject(baseMap[k], v)
			}
			return newMap
		}
	}
	return overlay
}

func merge(base, overlay types.Config) (types.Config, error) {
	baseMap, err := toMap(base)
	if err != nil {
		return types.Config{}, err
	}
	overlayMap, err := toMap(overlay)
	if err != nil {
		return types.Config{}, err
	}

	merged := mergeObject(baseMap, overlayMap)
	mergedData, err := json.Marshal(merged)
	if err != nil {
		return types.Config{}, err
	}

	var result types.Config
	return result, json.Unmarshal(mergedData, &result)
}
