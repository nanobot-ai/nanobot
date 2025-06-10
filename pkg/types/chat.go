package types

import "encoding/json"

var AgentTool = "__agent"

var ChatInputSchema = []byte(`{
  "type": "object",
  "required": [
    "prompt"
  ],
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
	      }
	    }
	  }
    }
  }
}`)

func Marshal[T any](in any, out *T) error {
	switch s := any(out).(type) {
	case *string:
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		*s = string(data)
		return nil
	}

	if v, ok := in.(T); ok {
		*out = v
		return nil
	}
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

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
