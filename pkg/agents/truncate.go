package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const maxToolResultSize = 50 * 1024 // 50 KiB

var sanitizeRe = regexp.MustCompile(`[^a-zA-Z0-9_\-.]`)

func truncateToolResult(ctx context.Context, toolName, callID string, msg *types.Message) *types.Message {
	fmt.Printf("[DEBUG truncate] truncateToolResult called: toolName=%q callID=%q\n", toolName, callID)

	if msg == nil || len(msg.Items) == 0 {
		fmt.Printf("[DEBUG truncate] msg is nil or has no items, returning unchanged\n")
		return msg
	}

	result := msg.Items[0].ToolCallResult
	if result == nil {
		fmt.Printf("[DEBUG truncate] no ToolCallResult in first item, returning unchanged\n")
		return msg
	}

	if result.Output.IsError {
		fmt.Printf("[DEBUG truncate] tool result is an error, skipping truncation\n")
		return msg
	}

	content := result.Output.Content
	if len(content) == 0 {
		fmt.Printf("[DEBUG truncate] tool result has no content, returning unchanged\n")
		return msg
	}

	size := contentSize(content)
	fmt.Printf("[DEBUG truncate] tool=%q contentSize=%d maxToolResultSize=%d contentParts=%d\n", toolName, size, maxToolResultSize, len(content))

	if size <= maxToolResultSize {
		fmt.Printf("[DEBUG truncate] size %d <= limit %d, no truncation needed\n", size, maxToolResultSize)
		return msg
	}

	fmt.Printf("[DEBUG truncate] TRUNCATING tool=%q: size %d exceeds limit %d (overflow=%d bytes)\n", toolName, size, maxToolResultSize, size-maxToolResultSize)

	// Determine file extension
	ext := ".txt"
	for _, c := range content {
		if c.Type != "" && c.Type != "text" {
			ext = ".json"
			break
		}
	}
	fmt.Printf("[DEBUG truncate] file extension=%q\n", ext)

	// Build file path
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	if sessionID == "" {
		sessionID = "default"
	}

	filePath := filepath.Join(".nanobot", sanitizePathComponent(sessionID),
		"truncated-outputs",
		sanitizePathComponent(toolName)+"-"+sanitizePathComponent(callID)+ext)

	fmt.Printf("[DEBUG truncate] writing full result to %s\n", filePath)

	if err := writeFullResult(content, filePath); err != nil {
		fmt.Printf("[DEBUG truncate] ERROR writing full result: %v\n", err)
		log.Errorf(ctx, "failed to write truncated tool result to %s: %v", filePath, err)
		return msg
	}

	truncated := buildTruncatedContent(content, maxToolResultSize, filePath)
	fmt.Printf("[DEBUG truncate] built truncated content: %d parts, total truncated size=%d\n", len(truncated), contentSize(truncated))

	return &types.Message{
		ID:   msg.ID,
		Role: msg.Role,
		Items: []types.CompletionItem{
			{
				ID: msg.Items[0].ID,
				ToolCallResult: &types.ToolCallResult{
					CallID: result.CallID,
					Output: types.CallResult{
						Content: truncated,
					},
				},
			},
		},
	}
}

func contentSize(content []mcp.Content) int {
	total := 0
	for i, c := range content {
		partSize := 0
		switch c.Type {
		case "text", "":
			partSize = len(c.Text)
		case "image", "audio":
			partSize = len(c.Data)
		case "resource":
			if c.Resource != nil {
				partSize = len(c.Resource.Text) + len(c.Resource.Blob)
			}
		default:
			data, err := json.Marshal(c)
			if err == nil {
				partSize = len(data)
			}
		}
		total += partSize
		fmt.Printf("[DEBUG truncate] contentSize part[%d] type=%q size=%d runningTotal=%d\n", i, c.Type, partSize, total)
	}
	return total
}

func writeFullResult(content []mcp.Content, filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	allText := true
	for _, c := range content {
		if c.Type != "" && c.Type != "text" {
			allText = false
			break
		}
	}

	if allText {
		var sb strings.Builder
		for i, c := range content {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(c.Text)
		}
		return os.WriteFile(filePath, []byte(sb.String()), 0644)
	}

	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func buildTruncatedContent(content []mcp.Content, budget int, filePath string) []mcp.Content {
	suffix := fmt.Sprintf("\n\n[Truncated: full output available at %s]", filePath)
	remaining := budget - len(suffix)
	if remaining < 0 {
		remaining = 0
	}

	fmt.Printf("[DEBUG truncate] buildTruncatedContent: budget=%d suffixLen=%d remaining=%d inputParts=%d\n", budget, len(suffix), remaining, len(content))

	var result []mcp.Content

	for i, c := range content {
		if remaining <= 0 {
			fmt.Printf("[DEBUG truncate] buildTruncatedContent: budget exhausted at part[%d], skipping remaining %d parts\n", i, len(content)-i)
			break
		}

		switch c.Type {
		case "text", "":
			text := c.Text
			truncated := false
			if len(text) > remaining {
				truncated = true
				text = text[:remaining]
			}
			result = append(result, mcp.Content{
				Type: "text",
				Text: text,
			})
			if truncated {
				fmt.Printf("[DEBUG truncate] buildTruncatedContent: part[%d] text truncated from %d to %d bytes\n", i, len(c.Text), len(text))
			} else {
				fmt.Printf("[DEBUG truncate] buildTruncatedContent: part[%d] text kept in full (%d bytes)\n", i, len(text))
			}
			remaining -= len(text)
		default:
			note := fmt.Sprintf("[%s content written to %s]", c.Type, filePath)
			result = append(result, mcp.Content{
				Type: "text",
				Text: note,
			})
			fmt.Printf("[DEBUG truncate] buildTruncatedContent: part[%d] non-text type=%q replaced with note (%d bytes)\n", i, c.Type, len(note))
			remaining -= len(note)
		}
	}

	result = append(result, mcp.Content{
		Type: "text",
		Text: suffix,
	})

	fmt.Printf("[DEBUG truncate] buildTruncatedContent: final result has %d parts\n", len(result))

	return result
}

func sanitizePathComponent(s string) string {
	s = sanitizeRe.ReplaceAllString(s, "_")
	s = strings.TrimLeft(s, ".")
	if len(s) > 100 {
		s = s[:100]
	}
	if s == "" {
		s = "unnamed"
	}
	return s
}
