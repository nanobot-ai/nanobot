package types

import (
	"encoding/json"
)

var AgentTool = "chat"

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
