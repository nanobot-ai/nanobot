package agents

import (
	"fmt"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/types"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

// estimateTokens estimates the total token count for a set of messages, a system prompt, and tool definitions.
// It uses the cl100k_base encoding (reasonable for both OpenAI and Anthropic models).
// Falls back to len(text)/4 heuristic if tiktoken encoding fails.
func estimateTokens(messages []types.Message, systemPrompt string, tools []types.ToolUseDefinition) int {
	fmt.Printf("[DEBUG tokencount] estimateTokens: messages=%d systemPromptLen=%d tools=%d\n",
		len(messages), len(systemPrompt), len(tools))

	var sb strings.Builder

	if systemPrompt != "" {
		sb.WriteString(systemPrompt)
		sb.WriteString("\n")
	}

	totalContentBytes := 0
	totalToolCallBytes := 0
	totalToolResultBytes := 0
	totalReasoningBytes := 0

	for _, msg := range messages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		for _, item := range msg.Items {
			if item.Content != nil {
				sb.WriteString(item.Content.Text)
				sb.WriteString(" ")
				totalContentBytes += len(item.Content.Text)
			}
			if item.ToolCall != nil {
				sb.WriteString(item.ToolCall.Name)
				sb.WriteString(" ")
				sb.WriteString(item.ToolCall.Arguments)
				sb.WriteString(" ")
				totalToolCallBytes += len(item.ToolCall.Name) + len(item.ToolCall.Arguments)
			}
			if item.ToolCallResult != nil {
				for _, c := range item.ToolCallResult.Output.Content {
					sb.WriteString(c.Text)
					sb.WriteString(" ")
					totalToolResultBytes += len(c.Text)
				}
			}
			if item.Reasoning != nil {
				for _, s := range item.Reasoning.Summary {
					sb.WriteString(s.Text)
					sb.WriteString(" ")
					totalReasoningBytes += len(s.Text)
				}
			}
		}
		sb.WriteString("\n")
	}

	totalToolDefBytes := 0
	for _, tool := range tools {
		sb.WriteString(tool.Name)
		sb.WriteString(" ")
		sb.WriteString(tool.Description)
		sb.WriteString(" ")
		toolBytes := len(tool.Name) + len(tool.Description)
		if len(tool.Parameters) > 0 {
			sb.Write(tool.Parameters)
			sb.WriteString(" ")
			toolBytes += len(tool.Parameters)
		}
		sb.WriteString("\n")
		totalToolDefBytes += toolBytes
	}

	text := sb.String()
	tokens := countTokens(text)

	fmt.Printf("[DEBUG tokencount] estimateTokens breakdown: systemPrompt=%d bytes, content=%d bytes, toolCalls=%d bytes, toolResults=%d bytes, reasoning=%d bytes, toolDefs=%d bytes\n",
		len(systemPrompt), totalContentBytes, totalToolCallBytes, totalToolResultBytes, totalReasoningBytes, totalToolDefBytes)
	fmt.Printf("[DEBUG tokencount] estimateTokens: totalTextLen=%d bytes, estimatedTokens=%d\n", len(text), tokens)

	return tokens
}

// countTokens counts the tokens in the given text using tiktoken's cl100k_base encoding.
// Falls back to len(text)/4 if encoding fails.
func countTokens(text string) int {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		fallback := len(text) / 4
		fmt.Printf("[DEBUG tokencount] countTokens: tiktoken encoding failed (%v), using fallback len/4=%d\n", err, fallback)
		return fallback
	}
	tokens := enc.Encode(text, nil, nil)
	fmt.Printf("[DEBUG tokencount] countTokens: tiktoken encoded %d bytes -> %d tokens\n", len(text), len(tokens))
	return len(tokens)
}
