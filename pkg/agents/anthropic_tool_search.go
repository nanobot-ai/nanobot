package agents

import (
	"maps"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/types"
)

const (
	anthropicToolSearchToolName = "anthropic_tool_search"
	anthropicToolSearchToolType = "tool_search_20251119"
)

func supportsAnthropicToolSearch(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "claude-opus-4") || strings.HasPrefix(model, "claude-sonnet-4")
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
			"type":             anthropicToolSearchToolType,
			"tool_search_type": "bm25_search",
		},
	})
}
