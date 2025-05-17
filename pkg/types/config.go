package types

import (
	"encoding/json"
	"strings"
)

type Config struct {
	Publish    Publish              `json:"publish,omitempty"`
	Agents     map[string]Agent     `json:"agents,omitempty"`
	MCPServers map[string]MCPServer `json:"mcpServers,omitempty"`
}

type Publish struct {
	Name         string     `json:"name,omitempty"`
	Description  string     `json:"description,omitempty"`
	Version      string     `json:"version,omitempty"`
	Instructions string     `json:"instructions,omitempty"`
	Tools        StringList `json:"tools,omitzero"`
}

type ToolRef struct {
	Server string
	Tool   string
	As     string
}

func ParseToolRef(ref string) ToolRef {
	name, as, _ := strings.Cut(ref, ":")
	server, tool, _ := strings.Cut(name, "/")
	return ToolRef{
		Server: server,
		Tool:   tool,
		As:     as,
	}
}

type ToolMappings map[string]ToolMapping

type ToolMapping struct {
	Server   string         `json:"server,omitempty"`
	ToolName string         `json:"toolName,omitempty"`
	Tool     ToolDefinition `json:"tool,omitempty"`
}

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema,omitzero"`
}

type StringList []string

func (s *StringList) UnmarshalJSON(data []byte) error {
	if data[0] == '[' && data[len(data)-1] == ']' {
		var raw []string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		*s = raw
	} else {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		*s = StringList{raw}
	}
	return nil
}

type MCPServer struct {
	Env     map[string]string `json:"env,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	BaseURL string            `json:"baseURL,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type Agent struct {
	Description    string                    `json:"description,omitempty"`
	Instructions   string                    `json:"instructions,omitempty"`
	Model          string                    `json:"model,omitempty"`
	Tools          []string                  `json:"tools,omitempty"`
	Stateful       bool                      `json:"stateful,omitempty"`
	ToolExtensions map[string]map[string]any `json:"toolExtensions,omitempty"`
	ToolChoice     string                    `json:"toolChoice,omitempty"`
	Temperature    *json.Number              `json:"temperature,omitempty"`
	TopP           *json.Number              `json:"topP,omitempty"`
	Output         *OutputSchema             `json:"output,omitempty"`
	Truncation     *string                   `json:"truncation,omitempty"`
	MaxTokens      int                       `json:"maxTokens,omitempty"`

	// Selection criteria fields

	Aliases      []string `json:"aliases,omitempty"`
	Cost         float64  `json:"cost,omitempty"`
	Speed        float64  `json:"speed,omitempty"`
	Intelligence float64  `json:"intelligence,omitempty"`
}

type OutputSchema struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema,omitzero"`
	Strict      bool            `json:"strict,omitempty"`
}
