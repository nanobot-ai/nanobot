package agents

import (
	"encoding/json"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/types"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

// estimateTokens estimates the total token count for a set of messages, a system prompt, and tool definitions.
// It uses the cl100k_base encoding (reasonable for both OpenAI and Anthropic models).
// Falls back to len(text)/4 heuristic if tiktoken encoding fails.
func estimateTokens(messages []types.Message, systemPrompt string, tools []types.ToolUseDefinition) int {
	var sb strings.Builder

	if systemPrompt != "" {
		sb.WriteString(systemPrompt)
		sb.WriteString("\n")
	}

	for _, msg := range messages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		for _, item := range msg.Items {
			if item.Content != nil {
				sb.WriteString(item.Content.Text)
				sb.WriteString(" ")
			}
			if item.ToolCall != nil {
				sb.WriteString(item.ToolCall.Name)
				sb.WriteString(" ")
				sb.WriteString(item.ToolCall.Arguments)
				sb.WriteString(" ")
			}
			if item.ToolCallResult != nil {
				for _, c := range item.ToolCallResult.Output.Content {
					switch c.Type {
					case "text", "":
						sb.WriteString(c.Text)
					case "image", "audio":
						sb.WriteString(c.Data)
					case "resource":
						if c.Resource != nil {
							sb.WriteString(c.Resource.Text)
							sb.WriteString(c.Resource.Blob)
						}
					default:
						if data, err := json.Marshal(c); err == nil {
							sb.Write(data)
						}
					}
					sb.WriteString(" ")
				}
			}
			if item.Reasoning != nil {
				for _, s := range item.Reasoning.Summary {
					sb.WriteString(s.Text)
					sb.WriteString(" ")
				}
			}
		}
		sb.WriteString("\n")
	}

	for _, tool := range tools {
		sb.WriteString(tool.Name)
		sb.WriteString(" ")
		sb.WriteString(tool.Description)
		sb.WriteString(" ")
		if len(tool.Parameters) > 0 {
			sb.Write(tool.Parameters)
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}

	return countTokens(sb.String())
}

// countTokens counts the tokens in the given text using tiktoken's cl100k_base encoding.
// Falls back to len(text)/4 if encoding fails.
func countTokens(text string) int {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return len(text) / 4
	}
	return len(enc.Encode(text, nil, nil))
}
