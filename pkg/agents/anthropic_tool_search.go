package agents

import (
	"context"
	"maps"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const (
	anthropicToolSearchToolName = "tool_search_tool_bm25"
	anthropicToolSearchToolType = "tool_search_tool_bm25_20251119"
)

func supportsAnthropicToolSearch(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "claude-opus-4") || strings.HasPrefix(model, "claude-sonnet-4")
}

func resolveAnthropicToolSearchModel(ctx context.Context, model string) string {
	model = strings.TrimSpace(model)
	if model != "" && model != "default" && model != "mini" {
		return model
	}

	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return model
	}

	envMap := session.GetEnvMap()
	switch model {
	case "mini":
		if value := strings.TrimSpace(envMap["NANOBOT_DEFAULT_MINI_MODEL"]); value != "" {
			return value
		}
	case "", "default":
		if value := strings.TrimSpace(envMap["NANOBOT_DEFAULT_MODEL"]); value != "" {
			return value
		}
	}

	return model
}

func shouldDeferAnthropicTool(mapping types.TargetMapping[types.TargetTool]) bool {
	if mapping.MCPServer == "" {
		return false
	}

	if strings.HasPrefix(mapping.MCPServer, "nanobot.") || mapping.MCPServer == "mcp-server-search" {
		return false
	}

	return true
}

func withAnthropicDeferredLoading(attributes map[string]any) map[string]any {
	if attributes == nil {
		attributes = map[string]any{}
	} else {
		attributes = maps.Clone(attributes)
	}
	attributes["defer_loading"] = true
	return attributes
}

func ensureAnthropicToolSearchTool(req *types.CompletionRequest) {
	for _, tool := range req.Tools {
		if tool.Name == anthropicToolSearchToolName {
			return
		}
		if tool.Attributes["type"] == anthropicToolSearchToolType {
			return
		}
	}

	req.Tools = append(req.Tools, types.ToolUseDefinition{
		Name: anthropicToolSearchToolName,
		Attributes: map[string]any{
			"type": anthropicToolSearchToolType,
		},
	})
}
