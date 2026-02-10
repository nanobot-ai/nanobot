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

// truncateToolOutput checks if a tool output needs truncation and truncates it if necessary.
// It saves the full output to disk and returns a message with the truncated content.
func truncateToolOutput(ctx context.Context, toolName string, response *types.CallResult, maxBytes int) (*TruncationResult, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}

	// Calculate total size across all content items
	totalBytes := 0
	for _, c := range response.Content {
		if c.Type == "text" {
			totalBytes += len(c.Text)
		} else {
			totalBytes += len(c.Data)
		}
	}

	if totalBytes <= maxBytes {
		return &TruncationResult{
			Content:   response.Content,
			Truncated: false,
		}, nil
	}

	// Collect full text for saving to disk
	var fullText strings.Builder
	for _, c := range response.Content {
		if c.Type == "text" {
			fullText.WriteString(c.Text)
		}
	}

	// Save full output to disk
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	if sessionID == "" {
		sessionID = "default"
	}

	timestamp := time.Now().Format("20060102-150405")
	safeToolName := sanitizeFileName(toolName)
	outputDir := filepath.Join(".nanobot", sessionID, "tool-outputs")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tool output directory: %w", err)
	}

	fileName := fmt.Sprintf("%s-%s-full-text.txt", timestamp, safeToolName)
	filePath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(filePath, []byte(fullText.String()), 0644); err != nil {
		return nil, fmt.Errorf("failed to write full output to file: %w", err)
	}

	// Walk content items in original order with a remaining byte budget
	remaining := maxBytes
	result := make([]mcp.Content, 0, len(response.Content))
	var droppedItems []mcp.Content
	textTruncated := false

	for _, c := range response.Content {
		if c.Type != "text" {
			// Non-text: keep if it fits entirely, otherwise drop
			if len(c.Data) <= remaining {
				result = append(result, c)
				remaining -= len(c.Data)
			} else {
				droppedItems = append(droppedItems, c)
			}
			continue
		}

		// Text item: keep if it fits entirely
		if len(c.Text) <= remaining {
			result = append(result, c)
			remaining -= len(c.Text)
			continue
		}

		// Partial text truncation: truncate on a clean UTF-8 boundary
		truncated := truncateUTF8(c.Text, remaining)
		result = append(result, mcp.Content{Type: "text", Text: truncated})
		textTruncated = true
		// Stop processing further items after text truncation
		break
	}

	// Save dropped non-text items to disk
	var droppedFilePath string
	if len(droppedItems) > 0 {
		droppedFileName := fmt.Sprintf("%s-%s-dropped-items.json", timestamp, safeToolName)
		droppedFilePath = filepath.Join(outputDir, droppedFileName)
		droppedJSON, err := json.Marshal(droppedItems)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dropped content items: %w", err)
		}
		if err := os.WriteFile(droppedFilePath, droppedJSON, 0644); err != nil {
			return nil, fmt.Errorf("failed to write dropped content to file: %w", err)
		}
	}

	// Build truncation notice
	var notice strings.Builder
	if textTruncated {
		fmt.Fprintf(&notice, "Output truncated. Full output saved to: %s", filePath)
	}
	if len(droppedItems) > 0 {
		if notice.Len() > 0 {
			notice.WriteString(". ")
		}
		fmt.Fprintf(&notice, "%d non-text content item(s) dropped due to size, saved to: %s", len(droppedItems), droppedFilePath)
	}
	if notice.Len() > 0 {
		result = append(result, mcp.Content{Type: "text", Text: fmt.Sprintf("\n\n(%s)", notice.String())})
	}

	return &TruncationResult{
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
		" ", "-",
	)
	return replacer.Replace(name)
}
