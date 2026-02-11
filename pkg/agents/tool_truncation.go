package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

const (
	toolOutputPreviewLimit = 32 * 1024
)

func (a *Agents) truncateToolResult(callID string, result *types.CallResult) *types.CallResult {
	if result == nil {
		return nil
	}

	fullText := strings.TrimSpace(flattenCallResult(*result))
	if fullText == "" {
		return result
	}

	fullSize := len([]byte(fullText))
	if fullSize <= toolOutputPreviewLimit {
		return result
	}

	archivePath, displayPath, err := writeToolArchive(callID, result)
	if err != nil {
		return result
	}

	snippet := truncatePreview(fullText, toolOutputPreviewLimit)

	previewText := fmt.Sprintf("Output truncated to %d bytes (original %d bytes). Full output saved to %s. Use Read or Grep on that path to inspect details.\n\n%s",
		toolOutputPreviewLimit, fullSize, displayPath, snippet)

	preview := *result
	preview.Content = []mcp.Content{
		{
			Type: "text",
			Text: previewText,
		},
	}
	preview.StructuredContent = map[string]any{
		"nanobot": map[string]any{
			"toolOutputArchive": archivePath,
			"toolOutputBytes":   fullSize,
			"previewBytes":      toolOutputPreviewLimit,
		},
	}

	return &preview
}

func writeToolArchive(callID string, result *types.CallResult) (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve home dir: %w", err)
	}

	dir := filepath.Join(home, ".nanobot", "tool-output")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("failed to create tool output dir: %w", err)
	}

	archiveID := sanitizeArchiveID(callID)
	archivePath := filepath.Join(dir, archiveID+".json")

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal tool output: %w", err)
	}

	if err := os.WriteFile(archivePath, data, 0o600); err != nil {
		return "", "", fmt.Errorf("failed to write tool archive: %w", err)
	}

	displayPath := archivePath
	if strings.HasPrefix(archivePath, home) {
		displayPath = filepath.Join("~", strings.TrimPrefix(archivePath, home+string(os.PathSeparator)))
	}

	return archivePath, displayPath, nil
}

func sanitizeArchiveID(id string) string {
	if id == "" {
		return uuid.String()
	}

	var builder strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('-')
	}

	res := strings.Trim(builder.String(), "-")
	if res == "" {
		return uuid.String()
	}
	return res + "-" + uuid.String()
}

func truncatePreview(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "…"
}
