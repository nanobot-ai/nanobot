package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const (
	DefaultMaxBytes = 50 * 1024 // 50KB
)

// TruncationResult contains the truncated content and metadata
type TruncationResult struct {
	Content   []mcp.Content
	Truncated bool
	FilePath  string
}

// contentByteSize returns the byte size of a content item's payload.
func contentByteSize(c mcp.Content) int {
	switch c.Type {
	case "text":
		return len(c.Text)
	case "resource":
		if c.Resource != nil {
			return len(c.Resource.Text) + len(c.Resource.Blob)
		}
		return 0
	default:
		// image, audio, and other types use the Data field
		return len(c.Data)
	}
}

// truncateToolOutput checks if a tool output needs truncation and truncates it if necessary.
// It saves the full original content as JSON to disk and returns truncated content.
func truncateToolOutput(ctx context.Context, toolName, callID string, response *types.CallResult, maxBytes int) (TruncationResult, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}

	// Calculate total size across all content items
	totalBytes := 0
	for _, c := range response.Content {
		totalBytes += contentByteSize(c)
	}

	// No truncation needed
	if totalBytes <= maxBytes {
		return TruncationResult{}, nil
	}

	// Save full original content to disk as JSON
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	if sessionID == "" {
		sessionID = "default"
	}

	timestamp := time.Now().Format("20060102-150405")
	safeSessionID := sanitizeFileName(sessionID)
	safeToolName := sanitizeFileName(toolName)
	safeCallID := sanitizeFileName(callID)
	outputDir := filepath.Join(".nanobot", safeSessionID, "truncated-tool-outputs")

	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return TruncationResult{}, fmt.Errorf("failed to create tool output directory: %w", err)
	}

	filePath := filepath.Join(outputDir, fmt.Sprintf("%s-%s-%s.json", timestamp, safeToolName, safeCallID))

	fullJSON, err := json.Marshal(response.Content)
	if err != nil {
		return TruncationResult{}, fmt.Errorf("failed to marshal full content: %w", err)
	}
	if err := os.WriteFile(filePath, fullJSON, 0600); err != nil {
		return TruncationResult{}, fmt.Errorf("failed to write full output to file: %w", err)
	}

	// Walk content items in original order with a remaining byte budget
	remaining := maxBytes
	result := make([]mcp.Content, 0, len(response.Content))

	for _, c := range response.Content {
		size := contentByteSize(c)

		if c.Type != "text" {
			// Non-text: keep if it fits entirely, otherwise drop
			if size <= remaining {
				result = append(result, c)
				remaining -= size
			}
			continue
		}

		// Text item: keep if it fits entirely
		if size <= remaining {
			result = append(result, c)
			remaining -= size
			continue
		}

		// Partial text truncation: truncate on a clean UTF-8 boundary
		truncated := truncateUTF8(c.Text, remaining)
		result = append(result, mcp.Content{Type: "text", Text: truncated})
		// Stop processing further items after text truncation
		break
	}

	result = append(result, mcp.Content{
		Type: "text",
		Text: fmt.Sprintf("\n\n(Tool output truncated because it is too large. Full output saved to: %s)", filePath),
	})

	return TruncationResult{
		Content:   result,
		Truncated: true,
		FilePath:  filePath,
	}, nil
}

// truncateUTF8 truncates s to at most maxBytes while preserving valid UTF-8.
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Walk backwards from maxBytes to find a valid UTF-8 boundary
	for maxBytes > 0 && !utf8.RuneStart(s[maxBytes]) {
		maxBytes--
	}
	return s[:maxBytes]
}

// sanitizeFileName removes characters that aren't safe for filenames
func sanitizeFileName(name string) string {
	// Replace unsafe characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		".", "_",
		" ", "-",
	)
	return replacer.Replace(name)
}
