package agents

import (
	"bytes"
	"context"
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
	// Default truncation limits
	DefaultMaxLines = 2000
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
func truncateToolOutput(ctx context.Context, toolName string, response *types.CallResult, maxLines, maxBytes int) (*TruncationResult, error) {
	if maxLines <= 0 {
		maxLines = DefaultMaxLines
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}

	// Combine all text content
	var fullText strings.Builder
	for _, content := range response.Content {
		if content.Type == "text" {
			fullText.WriteString(content.Text)
		}
	}

	text := fullText.String()
	if text == "" {
		// No text to truncate
		return &TruncationResult{
			Content:   response.Content,
			Truncated: false,
		}, nil
	}

	// Check if truncation is needed
	lines := strings.Split(text, "\n")
	textBytes := len(text)

	needsTruncation := len(lines) > maxLines || textBytes > maxBytes

	if !needsTruncation {
		return &TruncationResult{
			Content:   response.Content,
			Truncated: false,
		}, nil
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

	fileName := fmt.Sprintf("%s-%s.txt", timestamp, safeToolName)
	filePath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(filePath, []byte(text), 0644); err != nil {
		return nil, fmt.Errorf("failed to write full output to file: %w", err)
	}

	// Truncate the content
	truncatedText, removedCount, unit := truncateText(text, lines, maxLines, maxBytes)

	// Add truncation message
	truncationMsg := fmt.Sprintf("\n\n(Output truncated: removed %d %s. Full output saved to: %s)",
		removedCount, unit, filePath)
	truncatedText += truncationMsg

	// Build result while preserving original non-text ordering
	result := make([]mcp.Content, 0, len(response.Content))
	insertedTruncation := false
	for _, content := range response.Content {
		if content.Type == "text" {
			if !insertedTruncation {
				result = append(result, mcp.Content{Type: "text", Text: truncatedText})
				insertedTruncation = true
			}
			continue
		}
		result = append(result, content)
	}
	if !insertedTruncation {
		// Safety fallback; should not occur because empty text content short-circuits earlier
		result = append(result, mcp.Content{Type: "text", Text: truncatedText})
	}

	return &TruncationResult{
		Content:   result,
		Truncated: true,
		FilePath:  filePath,
	}, nil
}

// truncateText performs the actual truncation, returning the truncated text,
// the count of items removed, and the unit (lines/bytes)
func truncateText(fullText string, lines []string, maxLines, maxBytes int) (string, int, string) {
	// Determine what caused the truncation
	hitBytes := len(fullText) > maxBytes

	if hitBytes {
		// Truncate by bytes
		var buf bytes.Buffer
		for _, line := range lines {
			lineWithNewline := line + "\n"
			if buf.Len()+len(lineWithNewline) > maxBytes {
				remaining := maxBytes - buf.Len()
				for _, r := range lineWithNewline {
					runeSize := utf8.RuneLen(r)
					if runeSize < 0 {
						runeSize = 1
					}
					if remaining < runeSize {
						break
					}
					buf.WriteRune(r)
					remaining -= runeSize
					if remaining == 0 {
						break
					}
				}
				break
			}
			buf.WriteString(lineWithNewline)
		}
		truncated := strings.TrimRight(buf.String(), "\n")
		removed := len(fullText) - len(truncated)
		return truncated, removed, "bytes"
	}

	// Truncate by lines
	truncatedLines := lines[:maxLines]
	truncated := strings.Join(truncatedLines, "\n")
	removed := len(lines) - maxLines
	return truncated, removed, "lines"
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
