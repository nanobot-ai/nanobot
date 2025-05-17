package agents

import "github.com/obot-platform/nanobot/pkg/types"

type run struct {
	Request          types.Request       `json:"request,omitempty"`
	Done             bool                `json:"done,omitempty"`
	PopulatedRequest *types.Request      `json:"populatedRequest,omitempty"`
	ToolToMCPServer  types.ToolMappings  `json:"toolToMCPServer,omitempty"`
	Response         *types.Response     `json:"response,omitempty"`
	ToolOutputs      map[string]toolCall `json:"toolOutputs,omitempty"`
}

type toolCall struct {
	Output []types.InputItem `json:"output,omitempty"`
	Done   bool              `json:"done,omitempty"`
}
