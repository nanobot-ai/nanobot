package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/obot-platform/nanobot/pkg/complete"
	"github.com/obot-platform/nanobot/pkg/mcp"
)

type Config struct {
	Extends    string                `json:"extends,omitempty"`
	Env        map[string]EnvDef     `json:"env,omitempty"`
	Publish    Publish               `json:"publish,omitempty"`
	Agents     map[string]Agent      `json:"agents,omitempty"`
	MCPServers map[string]mcp.Server `json:"mcpServers,omitempty"`
	Flows      map[string]Flow       `json:"flows,omitempty"`
	Profiles   map[string]Config     `json:"profiles,omitempty"`
}

func (c Config) Validate(allowLocal bool) error {
	var (
		errs      []error
		seenNames = map[string]string{}
	)
	if strings.HasPrefix(strings.TrimSpace(c.Extends), "/") {
		errs = append(errs, fmt.Errorf("extends cannot be an absolute path: %s", c.Extends))
	}

	for agentName, agent := range c.Agents {
		if err := checkDup(seenNames, "agents", agentName); err != nil {
			errs = append(errs, err)
		}
		if err := agent.validate(agentName, c); err != nil {
			errs = append(errs, err)
		}
	}

	for mcpServerName, mcpServer := range c.MCPServers {
		if err := checkDup(seenNames, "mcpServers", mcpServerName); err != nil {
			errs = append(errs, err)
		}
		if err := validateMCPServer(mcpServerName, mcpServer, allowLocal); err != nil {
			errs = append(errs, err)
		}
	}

	for flowName, flow := range c.Flows {
		if err := checkDup(seenNames, "flows", flowName); err != nil {
			errs = append(errs, err)
		}
		if err := flow.validate(flowName, c); err != nil {
			errs = append(errs, fmt.Errorf("error validating flow %q: %w", flowName, err))
		}
	}

	return errors.Join(errs...)
}

func validateMCPServer(mcpServerName string, mcpServer mcp.Server, allowLocal bool) error {
	if allowLocal && strings.HasPrefix(mcpServer.Source.Repo, "/") {
		return nil
	}

	if mcpServer.Source.Repo != "" {
		if !strings.HasPrefix(mcpServer.Source.Repo, "https://") &&
			!strings.HasPrefix(mcpServer.Source.Repo, "http://") &&
			!strings.HasPrefix(mcpServer.Source.Repo, "git@") &&
			!strings.HasPrefix(mcpServer.Source.Repo, "ssh://") {
			return fmt.Errorf("mcpServer %q has invalid repo URL %q, must start with http://, https://, git@, or ssh://", mcpServerName, mcpServer.Source.Repo)
		}
	}
	return nil
}

type EnvDef struct {
	Default        string     `json:"default,omitempty"`
	Description    string     `json:"description,omitempty"`
	Options        StringList `json:"options,omitempty"`
	Optional       bool       `json:"optional,omitempty"`
	Sensitive      *bool      `json:"sensitive,omitempty"`
	UseBearerToken bool       `json:"useBearerToken,omitempty"`
}

func (e *EnvDef) UnmarshalJSON(data []byte) error {
	if data[0] == '"' && data[len(data)-1] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		e.Description = raw
		return nil
	}
	type Alias EnvDef
	return json.Unmarshal(data, (*Alias)(e))
}

type Flow struct {
	Description string      `json:"description,omitempty"`
	Input       InputSchema `json:"input,omitempty"`
	OutputRole  string      `json:"outputRole,omitempty"`
	Steps       []Step      `json:"steps,omitzero"`
}

func (f Flow) validate(flowName string, c Config) error {
	var errs []error
	for i, step := range f.Steps {
		if err := step.validate(c); err != nil {
			errs = append(errs, fmt.Errorf("error validating step %d in flow %q: %w", i, flowName, err))
		}
	}
	return errors.Join(errs...)
}

type Step struct {
	ID         string    `json:"id,omitempty"`
	Agent      AgentCall `json:"agent,omitempty"`
	Tool       string    `json:"tool,omitempty"`
	Flow       string    `json:"flow,omitempty"`
	If         string    `json:"if,omitempty"`
	ForEach    any       `json:"forEach,omitempty"`
	ForEachVar string    `json:"forEachVar,omitempty"`
	Input      any       `json:"input,omitempty"`
	Steps      []Step    `json:"steps,omitzero"`
}

func ignoreEmptyStringList(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

func (s Step) validate(c Config) error {
	_, _, errs := validateReferences(c, ignoreEmptyStringList(s.Tool), ignoreEmptyStringList(s.Agent.Name), ignoreEmptyStringList(s.Flow))
	for i, step := range s.Steps {
		if err := step.validate(c); err != nil {
			errs = append(errs, fmt.Errorf("error validating nested step %d: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

type AgentCall struct {
	Name        string        `json:"name,omitempty"`
	Output      *OutputSchema `json:"output,omitempty"`
	ChatHistory *bool         `json:"chatHistory,omitempty"`
	ToolChoice  string        `json:"toolChoice,omitempty"`
	Temperature *json.Number  `json:"temperature,omitempty"`
	TopP        *json.Number  `json:"topP,omitempty"`
	// NOTE: DON'T ADD A NEW FIELD HERE WITHOUT UPDATING MarshalJSON/UnmarshalJSON/Merge
}

func (a AgentCall) Merge(other AgentCall) (result AgentCall) {
	result.Name = complete.Last(a.Name, other.Name)
	result.Output = complete.Last(a.Output, other.Output)
	result.ChatHistory = complete.Last(a.ChatHistory, other.ChatHistory)
	result.ToolChoice = complete.Last(a.ToolChoice, other.ToolChoice)
	result.Temperature = complete.Last(a.Temperature, other.Temperature)
	result.TopP = complete.Last(a.TopP, other.TopP)
	return
}

func (a AgentCall) MarshalJSON() ([]byte, error) {
	if a.Output == nil && a.ChatHistory == nil && a.ToolChoice == "" && a.Temperature == nil && a.TopP == nil {
		return json.Marshal(a.Name)
	}
	type Alias AgentCall
	return json.Marshal(Alias(a))
}

func (a *AgentCall) UnmarshalJSON(data []byte) error {
	if data[0] == '"' && data[len(data)-1] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		a.Name = raw
		return nil
	}
	type Alias AgentCall
	return json.Unmarshal(data, (*Alias)(a))
}

type Publish struct {
	Name              string              `json:"name,omitempty"`
	Introduction      DynamicInstructions `json:"introduction,omitempty"`
	Version           string              `json:"version,omitempty"`
	Instructions      string              `json:"instructions,omitempty"`
	Tools             StringList          `json:"tools,omitzero"`
	Prompts           StringList          `json:"prompts,omitzero"`
	Resources         StringList          `json:"resources,omitzero"`
	ResourceTemplates StringList          `json:"resourceTemplates,omitzero"`
	MCPServers        StringList          `json:"mcpServers,omitzero"`
	Entrypoint        string              `json:"entrypoint,omitempty"`
}

type ToolRef struct {
	Server string
	Tool   string
	As     string
}

func (t ToolRef) PublishedName(name string) string {
	if t.As != "" {
		return t.As
	}
	if t.Tool != "" {
		return t.Tool
	}
	return name
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

type ResourceMappings map[string]TargetMapping

type ResourceTemplateMappings map[string]TargetMapping

type PromptMappings map[string]TargetMapping

type TargetMapping struct {
	MCPServer  string `json:"mcpServer,omitempty"`
	TargetName string `json:"targetName,omitempty"`
	Target     any    `json:"target,omitempty"`
}

type ToolMappings map[string]TargetMapping

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

type Agent struct {
	Description    string                    `json:"description,omitempty"`
	Instructions   DynamicInstructions       `json:"instructions,omitempty"`
	Model          string                    `json:"model,omitempty"`
	Tools          StringList                `json:"tools,omitempty"`
	Agents         StringList                `json:"agents,omitempty"`
	Flows          StringList                `json:"flows,omitempty"`
	ChatHistory    *bool                     `json:"chatHistory,omitempty"`
	ToolExtensions map[string]map[string]any `json:"toolExtensions,omitempty"`
	ToolChoice     string                    `json:"toolChoice,omitempty"`
	Temperature    *json.Number              `json:"temperature,omitempty"`
	TopP           *json.Number              `json:"topP,omitempty"`
	Output         *OutputSchema             `json:"output,omitempty"`
	Truncation     string                    `json:"truncation,omitempty"`
	MaxTokens      int                       `json:"maxTokens,omitempty"`

	// Selection criteria fields

	Aliases      []string `json:"aliases,omitempty"`
	Cost         float64  `json:"cost,omitempty"`
	Speed        float64  `json:"speed,omitempty"`
	Intelligence float64  `json:"intelligence,omitempty"`
}

const mcpServerName = "MCP Server"

func validateReference[T any](ref string, targetType string, targets map[string]T) (string, error) {
	if targetType != mcpServerName && strings.Contains(ref, "/") {
		return "", fmt.Errorf("invalid %s reference %q: slashes are not allowed", targetType, ref)
	}

	toolRef := ParseToolRef(ref)
	if _, ok := targets[toolRef.Server]; !ok {
		return "", fmt.Errorf("can not find %s %q, missing in config", targetType, ref)
	}

	if targetType == mcpServerName {
		return toolRef.PublishedName(""), nil
	}

	return toolRef.PublishedName(toolRef.Server), nil
}

func validateReferences(c Config, tools, agents, flows StringList) (bool, map[string]struct{}, []error) {
	var (
		errs              []error
		unknownNames      bool
		resolvedToolNames = make(map[string]struct{})
	)

	for _, ref := range tools {
		targetName, err := validateReference(ref, mcpServerName, c.MCPServers)
		if err != nil {
			errs = append(errs, fmt.Errorf("error validating tool reference %q: %w", ref, err))
		}
		if targetName == "" {
			unknownNames = true
		} else {
			resolvedToolNames[targetName] = struct{}{}
		}
	}

	for _, ref := range agents {
		targetName, err := validateReference(ref, "agent", c.Agents)
		if err != nil {
			errs = append(errs, fmt.Errorf("error validating agent reference %q: %w", ref, err))
		}
		resolvedToolNames[targetName] = struct{}{}
	}

	for _, ref := range flows {
		targetName, err := validateReference(ref, "flow", c.Flows)
		if err != nil {
			errs = append(errs, fmt.Errorf("error validating flow reference %q: %w", ref, err))
		}
		resolvedToolNames[targetName] = struct{}{}
	}

	return unknownNames, resolvedToolNames, errs
}

func (a Agent) validate(agentName string, c Config) error {
	unknownNames, resolvedToolNames, errs := validateReferences(c, a.Tools, a.Agents, a.Flows)

	if a.Instructions.IsSet() && a.Instructions.IsPrompt() {
		_, ok := c.MCPServers[a.Instructions.MCPServer]
		if !ok {
			errs = append(errs, fmt.Errorf("agent %q has instructions with MCP server %q that is not defined in config", agentName, a.Instructions.MCPServer))
		}
	}

	if !unknownNames && a.ToolChoice != "" && a.ToolChoice != "none" && a.ToolChoice != "auto" {
		if _, ok := resolvedToolNames[a.ToolChoice]; !ok {
			errs = append(errs, fmt.Errorf("agent %q has tool choice %q that is not defined in tools", agentName, a.ToolChoice))
		}
	}

	return errors.Join(errs...)
}

type DynamicInstructions struct {
	Instructions string            `json:"-"`
	MCPServer    string            `json:"mcpServer"`
	Prompt       string            `json:"prompt"`
	Args         map[string]string `json:"args"`
}

func (a DynamicInstructions) IsPrompt() bool {
	return a.MCPServer != "" && a.Prompt != ""
}

func (a DynamicInstructions) IsSet() bool {
	return a.IsPrompt() || a.Instructions != ""
}

func (a *DynamicInstructions) UnmarshalJSON(data []byte) error {
	if data[0] == '"' && data[len(data)-1] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		a.Instructions = raw
		return nil
	}
	type Alias DynamicInstructions
	return json.Unmarshal(data, (*Alias)(a))
}

func (a DynamicInstructions) MarshalJSON() ([]byte, error) {
	if a.Instructions != "" {
		return json.Marshal(a.Instructions)
	}
	type Alias DynamicInstructions
	return json.Marshal(Alias(a))
}

type OutputSchema struct {
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	Schema      json.RawMessage  `json:"schema,omitzero"`
	Strict      bool             `json:"strict,omitempty"`
	Fields      map[string]Field `json:"fields,omitempty"`
}

type Field struct {
	Description string           `json:"description,omitempty"`
	Fields      map[string]Field `json:"fields,omitempty"`
}

func (f *Field) UnmarshalJSON(data []byte) error {
	if data[0] == '"' && data[len(data)-1] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		f.Description = raw
		f.Fields = nil
		return nil
	}
	type Alias Field
	return json.Unmarshal(data, (*Alias)(f))
}

func (f Field) MarshalJSON() ([]byte, error) {
	if len(f.Fields) > 0 {
		type Alias Field
		return json.Marshal(Alias(f))
	}
	return json.Marshal(f.Description)
}

func (o OutputSchema) ToSchema() json.RawMessage {
	if len(o.Fields) > 0 {
		data, _ := json.Marshal(BuildSimpleSchema(o.Name, o.Description, o.Fields))
		return data
	}
	return o.Schema
}

type InputSchema struct {
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	Schema      json.RawMessage  `json:"schema,omitzero"`
	Fields      map[string]Field `json:"fields,omitempty"`
}

func (i InputSchema) ToSchema() json.RawMessage {
	if len(i.Fields) > 0 {
		data, _ := json.Marshal(BuildSimpleSchema(i.Name, i.Description, i.Fields))
		return data
	}
	return i.Schema
}

func BuildSimpleSchema(name, description string, args map[string]Field) map[string]any {
	required := make([]string, 0)
	jsonschema := map[string]any{
		"type":                 "object",
		"properties":           map[string]any{},
		"additionalProperties": false,
	}

	if name != "" {
		jsonschema["title"] = name
	}

	if description != "" {
		jsonschema["description"] = description
	}

	for name, field := range args {
		if strings.HasSuffix(name, "[]") {
			name = strings.TrimSuffix(name, "[]")
			jsonschema["properties"].(map[string]any)[name] = map[string]any{
				"type":        "array",
				"description": field.Description,
				"items": map[string]any{
					"type": "string",
				},
			}
			if len(field.Fields) > 0 {
				jsonschema["properties"].(map[string]any)[name].(map[string]any)["items"] =
					BuildSimpleSchema("", field.Description, field.Fields)
			}
		} else if strings.HasSuffix(name, "(int)") || strings.HasSuffix(name, "(integer)") {
			name = strings.Split(name, "(")[0]
			jsonschema["properties"].(map[string]any)[name] = map[string]any{
				"type":        "integer",
				"description": field.Description,
			}
		} else if strings.HasSuffix(name, "(float)") || strings.HasSuffix(name, "(number)") {
			name = strings.Split(name, "(")[0]
			jsonschema["properties"].(map[string]any)[name] = map[string]any{
				"type":        "number",
				"description": field.Description,
			}
		} else if strings.HasSuffix(name, "(bool)") || strings.HasSuffix(name, "(boolean)") {
			name = strings.Split(name, "(")[0]
			jsonschema["properties"].(map[string]any)[name] = map[string]any{
				"type":        "boolean",
				"description": field.Description,
			}
		} else if len(field.Fields) > 0 {
			jsonschema["properties"].(map[string]any)[name] = BuildSimpleSchema("", field.Description, field.Fields)
		} else {
			jsonschema["properties"].(map[string]any)[name] = map[string]any{
				"type":        "string",
				"description": field.Description,
			}
		}

		required = append(required, name)
	}

	jsonschema["required"] = required
	return jsonschema
}
