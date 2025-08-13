package types

import (
	"encoding/json"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

const (
	AgentTool            = "chat"
	AgentToolDescription = "Chat with the current agent"
)

var ChatInputSchema = []byte(`{
  "type": "object",
  "required": ["prompt"],
  "properties": {
    "prompt": {
  	  "description": "The input prompt",
  	  "type": "string"
    },
    "attachments": {
	  "type": "array",
	  "items": {
	    "description": "An attachment to the prompt (optional)",
	    "type": "object",
	    "required": ["url"],
	    "properties": {
	      "url": {
	        "description": "The URL of the attachment or data URI",
	        "type": "string"
	      },
	      "mimeType": {
	        "description": "The mime type of the content reference by the URL",
	        "type": "string"
	      }
	    }
	  }
    }
  }
}`)

type SampleCallRequest struct {
	Prompt      string       `json:"prompt"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type SampleConfirmRequest struct {
	ID     string `json:"id"`
	Accept bool   `json:"accept"`
}

type Attachment struct {
	URL      string `json:"url"`
	MimeType string `json:"mimeType,omitempty"`
}

func (a *Attachment) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var url string
		if err := json.Unmarshal(data, &url); err != nil {
			return err
		}
		a.URL = url
		return nil
	}
	type Alias Attachment
	return json.Unmarshal(data, (*Alias)(a))
}

type ChatData struct {
	ID           string            `json:"id"`
	Ext          ChatDataExtension `json:"ai.nanobot/ext,omitzero"`
	CurrentAgent string            `json:"currentAgent,omitempty"`
	Messages     []Message         `json:"messages"`
}

func (c ChatData) MarshalJSON() ([]byte, error) {
	if c.Messages == nil {
		c.Messages = []Message{}
	}
	// We want to omit the empty fields in the extension
	type Alias ChatData
	return json.Marshal(Alias(c))
}

type ChatDataExtension struct {
	Tools   ToolMappings   `json:"tools,omitempty"`
	Prompts PromptMappings `json:"prompts,omitempty"`
	Agents  Agents         `json:"agents,omitempty"`
}

type ChatList struct {
	Chats []Chat `json:"chats"`
}

type Chat struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Created    time.Time `json:"created"`
	ReadOnly   bool      `json:"readonly,omitempty"`
	Visibility string    `json:"visibility,omitempty"`
}

type ProjectConfig struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Instructions string         `json:"instructions"`
	DefaultAgent string         `json:"defaultAgent"`
	Agents       []AgentDisplay `json:"agents"`
}

func (p *ProjectConfig) Deserialize(v any) (any, error) {
	return &p, mcp.JSONCoerce(v, &p)
}

func (p *ProjectConfig) Serialize() (any, error) {
	return p, nil
}
