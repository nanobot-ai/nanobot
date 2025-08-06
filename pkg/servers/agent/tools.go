package agent

import (
	"encoding/json"

	"github.com/nanobot-ai/nanobot/pkg/types"
)

type ChatData struct {
	ID           string            `json:"id"`
	Ext          ChatDataExtension `json:"ai.nanobot/ext,omitzero"`
	CurrentAgent string            `json:"currentAgent,omitempty"`
	Messages     []types.Message   `json:"messages"`
}

func (c ChatData) MarshalJSON() ([]byte, error) {
	if c.Messages == nil {
		c.Messages = []types.Message{}
	}
	// We want to omit the empty fields in the extension
	type Alias ChatData
	return json.Marshal(Alias(c))
}

type ChatDataExtension struct {
	CustomAgent *types.CustomAgent   `json:"customAgent,omitempty"`
	Tools       types.ToolMappings   `json:"tools,omitempty"`
	Prompts     types.PromptMappings `json:"prompts,omitempty"`
	Agents      types.Agents         `json:"agents,omitempty"`
}
