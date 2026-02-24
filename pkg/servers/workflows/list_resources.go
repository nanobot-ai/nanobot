package workflows

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

// ListWorkflowResources returns workflow resources found in workflowsPath.
// If the directory cannot be read, it returns an empty list.
func ListWorkflowResources(ctx context.Context, workflowsPath string) []mcp.Resource {
	entries, err := os.ReadDir(workflowsPath)
	if err != nil {
		return []mcp.Resource{}
	}

	result := make([]mcp.Resource, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")

		contentBytes, err := os.ReadFile(filepath.Join(workflowsPath, entry.Name()))
		if err != nil {
			continue
		}

		meta, err := parseWorkflowFrontmatter(string(contentBytes))
		if err != nil {
			log.Debugf(ctx, "failed to parse frontmatter for workflow %s: %v", entry.Name(), err)
		}

		resourceMeta := make(map[string]any)
		if meta.Name != "" {
			resourceMeta["name"] = meta.Name
		}
		if meta.CreatedAt != "" {
			resourceMeta["createdAt"] = meta.CreatedAt
		}

		res := mcp.Resource{
			URI:         fmt.Sprintf("workflow:///%s", name),
			Name:        name,
			Description: meta.Description,
			MimeType:    "text/markdown",
		}
		if len(resourceMeta) > 0 {
			res.Meta = resourceMeta
		}

		result = append(result, res)
	}
	return result
}
