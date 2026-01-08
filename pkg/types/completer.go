package types

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

type Completer interface {
	Complete(ctx context.Context, req CompletionRequest, opts ...CompletionOptions) (*CompletionResponse, error)
}

type CompletionOptions struct {
	ProgressToken      any
	Chat               *bool
	ToolChoice         *mcp.ToolChoice
	Tools              []mcp.Tool
	ToolIncludeContext string
	ToolSource         string
}

func (c CompletionOptions) Merge(other CompletionOptions) (result CompletionOptions) {
	result.ProgressToken = complete.Last(c.ProgressToken, other.ProgressToken)
	result.Chat = complete.Last(c.Chat, other.Chat)
	result.ToolChoice = complete.Last(c.ToolChoice, other.ToolChoice)
	result.Tools = append(c.Tools, other.Tools...)
	result.ToolIncludeContext = complete.Last(c.ToolIncludeContext, other.ToolIncludeContext)
	result.ToolSource = complete.Last(c.ToolSource, other.ToolSource)
	return
}

type CompletionRequest struct {
	Model            string               `json:"model,omitempty"`
	Agent            string               `json:"agent,omitempty"`
	ThreadName       string               `json:"threadName,omitempty"`
	NewThread        bool                 `json:"newThread,omitempty"`
	Input            []Message            `json:"input,omitzero"`
	ModelPreferences mcp.ModelPreferences `json:"modelPreferences,omitzero"`
	SystemPrompt     string               `json:"systemPrompt,omitzero"`
	MaxTokens        int                  `json:"maxTokens,omitempty"`
	ToolChoice       string               `json:"toolChoice,omitempty"`
	OutputSchema     *OutputSchema        `json:"outputSchema,omitempty"`
	Temperature      *json.Number         `json:"temperature,omitempty"`
	Truncation       string               `json:"truncation,omitempty"`
	TopP             *json.Number         `json:"topP,omitempty"`
	Metadata         map[string]any       `json:"metadata,omitempty"`
	Tools            []ToolUseDefinition  `json:"tools,omitzero"`
	Reasoning        *AgentReasoning      `json:"reasoning,omitempty"`
}

func (r CompletionRequest) GetAgent() string {
	if r.Agent != "" {
		return r.Agent
	}
	return r.Model
}

func (r CompletionRequest) Reset() CompletionRequest {
	r.Input = nil
	r.NewThread = false
	return r
}

type ToolUseDefinition struct {
	Name        string          `json:"name,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
	Description string          `json:"description,omitempty"`
	Attributes  map[string]any  `json:"-"`
}

type CompletionProgress struct {
	Model     string         `json:"model,omitempty"`
	Agent     string         `json:"agent,omitempty"`
	MessageID string         `json:"messageID,omitempty"`
	Role      string         `json:"role,omitempty"`
	Item      CompletionItem `json:"item,omitempty"`
}

const CompletionProgressMetaKey = "ai.nanobot.progress/completion"

type Message struct {
	ID      string           `json:"id,omitempty"`
	Created *time.Time       `json:"created,omitempty"`
	Role    string           `json:"role,omitempty"`
	Items   []CompletionItem `json:"items,omitempty"`
	HasMore bool             `json:"hasMore,omitempty"`
}

type CompletionItem struct {
	ID string `json:"id,omitempty"`
	// Partial indicates that the content may not be complete and later content may be appended
	// If false the content is complete and further content
	Partial        bool            `json:"partial,omitempty"`
	Content        *mcp.Content    `json:"content,omitempty"`
	ToolCall       *ToolCall       `json:"toolCall,omitempty"`
	ToolCallResult *ToolCallResult `json:"toolCallResult,omitempty"`
	Reasoning      *Reasoning      `json:"reasoning,omitempty"`
}

func (c *CompletionItem) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.ID = ""
	if id, ok := raw["id"]; ok {
		if err := json.Unmarshal(id, &c.ID); err != nil {
			return err
		}
	}

	if partial, ok := raw["partial"]; ok {
		if err := json.Unmarshal(partial, &c.Partial); err != nil {
			return err
		}
	}

	var typeField string
	if t, ok := raw["type"]; ok {
		if err := json.Unmarshal(t, &typeField); err != nil {
			return err
		}
	}

	switch typeField {
	case "text", "image", "audio", "resource":
		c.Content = &mcp.Content{}
		if err := json.Unmarshal(data, c.Content); err != nil {
			return err
		}
		err := json.Unmarshal(data, &struct {
			ID *string `json:"id,omitempty"`
		}{
			ID: &c.ID,
		})
		if err != nil {
			return err
		}
	case "tool":
		if _, ok := raw["name"]; ok {
			c.ToolCall = &ToolCall{}
			if err := json.Unmarshal(data, c.ToolCall); err != nil {
				return err
			}
		} else if _, ok := raw["arguments"]; ok {
			// Handle partial tool call with only arguments
			c.ToolCall = &ToolCall{}
			if err := json.Unmarshal(data, c.ToolCall); err != nil {
				return err
			}
		}
		if _, ok := raw["output"]; ok {
			c.ToolCallResult = &ToolCallResult{}
			if err := json.Unmarshal(data, c.ToolCallResult); err != nil {
				return err
			}
		}
	case "reasoning":
		c.Reasoning = &Reasoning{}
		return json.Unmarshal(data, c.Reasoning)
	}

	return nil
}

// AppendProgress appends Messages to the InternalMessages field matching the Message IDs and Item IDs, appending
// content as necessary. The Agent and Model fields are also updated if the progress has non-zero values.
// This function does not modify the input resp; it returns a new CompletionResponse with updated data.
func AppendProgress(resp CompletionResponse, progress CompletionProgress) CompletionResponse {
	// Create a new response to avoid modifying the input
	result := CompletionResponse{
		Output:        resp.Output,
		Agent:         resp.Agent,
		Model:         resp.Model,
		HasMore:       resp.HasMore,
		Error:         resp.Error,
		ProgressToken: resp.ProgressToken,
	}

	// Update Agent and Model if progress has non-zero values
	if progress.Agent != "" {
		result.Agent = progress.Agent
	}
	if progress.Model != "" {
		result.Model = progress.Model
	}

	// Copy the internal messages slice to avoid modifying the original
	result.InternalMessages = make([]Message, len(resp.InternalMessages))
	copy(result.InternalMessages, resp.InternalMessages)

	// If no message ID in progress, return the copy
	if progress.MessageID == "" {
		return result
	}

	// Find or create the message with the matching ID
	messageIdx := -1
	for i, msg := range result.InternalMessages {
		if msg.ID == progress.MessageID {
			messageIdx = i
			break
		}
	}

	// If message not found, create a new one
	if messageIdx == -1 {
		now := time.Now()
		role := progress.Role
		if role == "" {
			role = "assistant"
		}
		newMsg := Message{
			ID:      progress.MessageID,
			Created: &now,
			Role:    role,
			Items:   []CompletionItem{progress.Item},
		}
		result.InternalMessages = append(result.InternalMessages, newMsg)
		return result
	}

	// Message exists, copy it to avoid modifying the original
	msg := result.InternalMessages[messageIdx]

	// Update role if provided in progress
	if progress.Role != "" {
		msg.Role = progress.Role
	}

	// Copy items slice to avoid modifying the original
	msg.Items = make([]CompletionItem, len(msg.Items))
	copy(msg.Items, result.InternalMessages[messageIdx].Items)

	// If the item has an ID, try to find and merge it
	if progress.Item.ID != "" {
		for i, item := range msg.Items {
			if item.ID == progress.Item.ID {
				// Found matching item, merge content
				msg.Items[i] = mergeCompletionItems(item, progress.Item)
				result.InternalMessages[messageIdx] = msg
				return result
			}
		}
	}

	// Item not found or no ID, append as new item
	msg.Items = append(msg.Items, progress.Item)
	result.InternalMessages[messageIdx] = msg
	return result
}

// mergeCompletionItems merges two completion items, appending content when both items are partial.
// This function creates copies to avoid side effects on the input objects.
func mergeCompletionItems(existing, new CompletionItem) CompletionItem {
	// If the new item is not partial, it replaces the existing one
	if !new.Partial {
		return new
	}

	// Create a copy to avoid modifying the existing item
	result := CompletionItem{
		ID:      existing.ID,
		Partial: new.Partial,
	}

	// Merge based on content type
	switch {
	case new.Content != nil:
		result.Content = mergeContent(existing.Content, new.Content)
		return result
	case new.ToolCall != nil:
		result.ToolCall = mergeToolCall(existing.ToolCall, new.ToolCall)
		return result
	case new.Reasoning != nil:
		result.Reasoning = mergeReasoning(existing.Reasoning, new.Reasoning)
		return result
	case new.ToolCallResult != nil:
		result.ToolCallResult = new.ToolCallResult
		return result
	}

	// Default: if new has no content fields, keep existing unchanged
	return existing
}

// mergeContent merges text content, or returns new content if types differ or existing is nil
func mergeContent(existing, new *mcp.Content) *mcp.Content {
	if existing != nil && existing.Type == "text" && new.Type == "text" {
		return &mcp.Content{
			Type: "text",
			Text: existing.Text + new.Text,
		}
	}
	return new
}

// mergeToolCall merges tool call arguments and updates fields, or returns new if existing is nil
func mergeToolCall(existing, new *ToolCall) *ToolCall {
	if existing == nil {
		return new
	}

	merged := existing.Clone()
	merged.Arguments += new.Arguments
	if new.Name != "" {
		merged.Name = new.Name
	}
	if new.CallID != "" {
		merged.CallID = new.CallID
	}
	return merged
}

// mergeReasoning merges reasoning content and summaries, or returns new if existing is nil
func mergeReasoning(existing, new *Reasoning) *Reasoning {
	if existing == nil {
		return new
	}

	merged := existing.Clone()
	merged.EncryptedContent += new.EncryptedContent
	merged.Summary = append(merged.Summary, new.Summary...)
	return merged
}

func (c CompletionItem) MarshalJSON() ([]byte, error) {
	if c.ID == "" {
		c.ID = uuid.String()
	}

	if c.Content != nil {
		// mcp.Content has a custom MarshalJSON method that messes up things, so this is
		// a workaround to ensure we get the correct JSON structure.
		content, err := json.Marshal(c.Content)
		if err != nil {
			return nil, err
		}

		header, err := json.Marshal(map[string]any{
			"id":      c.ID,
			"partial": c.Partial,
		})

		// length 2 means it is an empty object
		if len(header) == 2 {
			return content, nil
		} else if len(content) == 2 {
			return header, nil
		}

		return slices.Concat(header[:len(header)-1], []byte(","), content[1:]), nil
	} else if c.ToolCallResult != nil || c.ToolCall != nil {
		var (
			tc     ToolCall
			output CallResult
		)
		if c.ToolCall != nil {
			tc = *c.ToolCall
		} else {
			tc = ToolCall{
				CallID: c.ToolCallResult.CallID,
			}
		}
		if c.ToolCallResult != nil {
			output = c.ToolCallResult.Output
		}
		return json.Marshal(struct {
			ID      string     `json:"id,omitempty"`
			Partial bool       `json:"partial,omitempty"`
			Type    string     `json:"type,omitempty"`
			Output  CallResult `json:"output,omitzero"`
			ToolCall
		}{
			ID:       c.ID,
			Type:     "tool",
			Partial:  c.Partial,
			ToolCall: tc,
			Output:   output,
		})
	} else if c.Reasoning != nil {
		return json.Marshal(struct {
			ID      string `json:"id,omitempty"`
			Type    string `json:"type,omitempty"`
			Partial bool   `json:"partial,omitempty"`
			*Reasoning
		}{
			ID:        c.ID,
			Type:      "reasoning",
			Partial:   c.Partial,
			Reasoning: c.Reasoning,
		})
	}
	type Alias CompletionItem
	return json.Marshal(Alias(c))
}

type Reasoning struct {
	EncryptedContent string        `json:"encryptedContent,omitempty"`
	Summary          []SummaryText `json:"summary,omitempty"`
}

func (r *Reasoning) Clone() *Reasoning {
	if r == nil {
		return nil
	}
	clone := &Reasoning{
		EncryptedContent: r.EncryptedContent,
		Summary:          append([]SummaryText(nil), r.Summary...),
	}
	return clone
}

type SummaryText struct {
	Text string `json:"text,omitempty"`
}

type CompletionResponse struct {
	Output           Message   `json:"output,omitempty"`
	InternalMessages []Message `json:"internalMessages,omitempty"`
	Agent            string    `json:"agent,omitempty"`
	Model            string    `json:"model,omitempty"`
	HasMore          bool      `json:"hasMore,omitempty"`
	Error            string    `json:"error,omitempty"`
	ProgressToken    any       `json:"progressToken,omitempty"`
}

func (c *CompletionResponse) Serialize() (any, error) {
	return c, nil
}

func (c *CompletionResponse) Deserialize(data any) (any, error) {
	return c, mcp.JSONCoerce(data, &c)
}

type ToolCallResult struct {
	CallID string     `json:"callID,omitempty"`
	Output CallResult `json:"output,omitzero"`
	// NOTE: If you add fields here, make sure to update the CompletionItem.MarshalJSON method, it
	//has special handling for ToolCallResult.
}

type ToolCall struct {
	Arguments string `json:"arguments,omitempty"`
	CallID    string `json:"callID,omitempty"`
	Name      string `json:"name,omitempty"`
}

func (t *ToolCall) Clone() *ToolCall {
	if t == nil {
		return nil
	}
	return &ToolCall{
		Arguments: t.Arguments,
		CallID:    t.CallID,
		Name:      t.Name,
	}
}

type CallResult struct {
	Content           []mcp.Content `json:"content,omitempty"`
	IsError           bool          `json:"isError,omitempty"`
	Agent             string        `json:"agent,omitempty"`
	Model             string        `json:"model,omitempty"`
	StopReason        string        `json:"stopReason,omitempty"`
	StructuredContent any           `json:"structuredContent,omitempty"`
}

type AsyncCallResult struct {
	IsError       bool          `json:"isError"`
	Content       []mcp.Content `json:"content,omitzero"`
	InProgress    bool          `json:"inProgress,omitempty"`
	ToolName      string        `json:"toolName,omitempty"`
	ProgressToken any           `json:"progressToken,omitempty"`
}
