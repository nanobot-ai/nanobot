package anthropic

import (
	"encoding/json"
	"fmt"

	"github.com/obot-platform/nanobot/pkg/printer"
)

func PrintProgress(msg json.RawMessage) bool {
	var delta DeltaEvent
	if err := json.Unmarshal(msg, &struct {
		Data *DeltaEvent
	}{
		Data: &delta,
	}); err != nil {
		return false
	}

	switch delta.Type {
	case "message_start":
	case "content_block_start":
		if delta.ContentBlock.Type == "tool_use" {
			printer.Prefix("<-(llm)", fmt.Sprintf("Preparing to call (%s) with args: ", delta.ContentBlock.Name))
		}
	case "content_block_delta":
		switch delta.Delta.Type {
		case "text_delta":
			printer.Prefix("<-(llm)", delta.Delta.Text)
		case "input_json_delta":
			printer.Prefix("<-(llm)", delta.Delta.PartialJSON)
		}
	case "content_block_stop":
		printer.Prefix("<-(llm)", "\n")
	case "message_delta":
	case "message_stop":
		printer.Prefix("<-(llm)", "\n")
	default:
		return false
	}

	return true
}
