// Package filter defines the versioned wire contract used by Obot Filters.
//
// The package deliberately contains only contract types and validation helpers
// so other services can consume it without importing nanobot application code.
package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const APIVersionV1 = "obot.obot.ai/filter/v1"

type ContractVersion string

const (
	ContractVersionLegacyMCP ContractVersion = ""
	ContractVersionV1        ContractVersion = APIVersionV1
)

type Source string

const (
	SourceMCP        Source = "mcp"
	SourceLocalAgent Source = "local_agent"
)

type EventType string

const (
	EventTypeMCPMessage EventType = "mcp_message"
	EventTypeUserPrompt EventType = "user_prompt"
	EventTypeToolCall   EventType = "tool_call"
)

type Phase string

const (
	PhaseRequest  Phase = "request"
	PhaseResponse Phase = "response"
	PhaseFailure  Phase = "failure"
)

type Surface string

const (
	SurfaceUserPrompt    Surface = "user_prompt"
	SurfaceToolArguments Surface = "tool_arguments"
	SurfaceToolResponse  Surface = "tool_response"
)

type Decision string

const (
	DecisionAccept Decision = "accept"
	DecisionReject Decision = "reject"
	DecisionMutate Decision = "mutate"
)

// ToolRequest is the MCP tool input for a v1 Filter. The wrapper maps directly
// to an ordinary MCP tool parameter named "request" in SDKs such as FastMCP.
type ToolRequest struct {
	Request Request `json:"request"`
}

// Request is the common v1 Filter request envelope. Payload is the complete
// value being evaluated; it is never a partial update.
type Request struct {
	APIVersion   string          `json:"apiVersion"`
	Source       Source          `json:"source"`
	Event        Event           `json:"event"`
	Context      Context         `json:"context"`
	Capabilities Capabilities    `json:"capabilities"`
	Payload      json.RawMessage `json:"payload"`
}

type Event struct {
	Type       EventType `json:"type"`
	Phase      Phase     `json:"phase"`
	Surface    Surface   `json:"surface,omitempty"`
	Method     string    `json:"method,omitempty"`
	Identifier string    `json:"identifier,omitempty"`
}

type Context struct {
	Trace       *TraceContext       `json:"trace,omitempty"`
	MCP         *MCPContext         `json:"mcp,omitempty"`
	LocalAgent  *LocalAgentContext  `json:"localAgent,omitempty"`
	Device      *DeviceContext      `json:"device,omitempty"`
	Environment *EnvironmentContext `json:"environment,omitempty"`
}

type TraceContext struct {
	EventID   string `json:"eventId,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
	TurnID    string `json:"turnId,omitempty"`
	ToolUseID string `json:"toolUseId,omitempty"`
}

type MCPContext struct {
	ServerName      string `json:"serverName,omitempty"`
	ServerShortName string `json:"serverShortName,omitempty"`
}

type LocalAgentContext struct {
	Provider     string `json:"provider"`
	AgentVersion string `json:"agentVersion,omitempty"`
	Model        string `json:"model,omitempty"`
	ModelID      string `json:"modelId,omitempty"`
	ToolName     string `json:"toolName,omitempty"`
	ToolKind     string `json:"toolKind,omitempty"`
}

type DeviceContext struct {
	ID           string `json:"id"`
	DeploymentID string `json:"deploymentId"`
}

type EnvironmentContext struct {
	OperatingSystem  string `json:"operatingSystem,omitempty"`
	Architecture     string `json:"architecture,omitempty"`
	WorkingDirectory string `json:"workingDirectory,omitempty"`
}

type Capabilities struct {
	CanReject bool `json:"canReject"`
	CanMutate bool `json:"canMutate"`
}

// Response is the common v1 Filter response. Payload must be present only for
// mutate and is the complete replacement value.
type Response struct {
	Decision Decision        `json:"decision"`
	Reason   string          `json:"reason,omitempty"`
	Payload  json.RawMessage `json:"payload,omitempty"`
}

func (r Request) Validate() error {
	if r.APIVersion != APIVersionV1 {
		return fmt.Errorf("unsupported filter API version %q", r.APIVersion)
	}
	if len(r.Payload) == 0 || !json.Valid(r.Payload) {
		return errors.New("filter payload must be valid JSON")
	}

	switch r.Source {
	case SourceMCP:
		if r.Event.Type != EventTypeMCPMessage {
			return errors.New("MCP source requires an mcp_message event")
		}
		if r.Event.Phase != PhaseRequest && r.Event.Phase != PhaseResponse && r.Event.Phase != PhaseFailure {
			return errors.New("MCP event requires a request, response, or failure phase")
		}
		if r.Event.Surface != "" {
			return errors.New("MCP events must not declare a local-agent surface")
		}
	case SourceLocalAgent:
		if err := r.Event.validateLocalAgent(); err != nil {
			return err
		}
		if r.Context.LocalAgent == nil || r.Context.LocalAgent.Provider == "" {
			return errors.New("local-agent events require an agent provider")
		}
		if r.Context.Device == nil || r.Context.Device.ID == "" || r.Context.Device.DeploymentID == "" {
			return errors.New("local-agent events require authenticated device and deployment identity")
		}
		if r.Event.Surface == SurfaceUserPrompt && r.Capabilities.CanMutate {
			return errors.New("user prompts cannot advertise mutation capability")
		}
	default:
		return fmt.Errorf("unknown filter source %q", r.Source)
	}

	return nil
}

func (e Event) validateLocalAgent() error {
	switch e.Surface {
	case SurfaceUserPrompt:
		if e.Type != EventTypeUserPrompt || e.Phase != PhaseRequest {
			return errors.New("user_prompt surface requires a user_prompt request event")
		}
	case SurfaceToolArguments:
		if e.Type != EventTypeToolCall || e.Phase != PhaseRequest {
			return errors.New("tool_arguments surface requires a tool_call request event")
		}
	case SurfaceToolResponse:
		if e.Type != EventTypeToolCall || (e.Phase != PhaseResponse && e.Phase != PhaseFailure) {
			return errors.New("tool_response surface requires a tool_call response or failure event")
		}
	default:
		return fmt.Errorf("unknown local-agent surface %q", e.Surface)
	}
	return nil
}

func (r Response) Validate(capabilities Capabilities) error {
	hasPayload := len(bytes.TrimSpace(r.Payload)) != 0
	switch r.Decision {
	case DecisionAccept:
		if hasPayload {
			return fmt.Errorf("%s response must not contain a payload", r.Decision)
		}
	case DecisionReject:
		if !capabilities.CanReject {
			return errors.New("rejection not allowed by filter capabilities")
		}
		if hasPayload {
			return fmt.Errorf("%s response must not contain a payload", r.Decision)
		}
	case DecisionMutate:
		if !capabilities.CanMutate {
			return errors.New("mutation not allowed by filter capabilities")
		}
		if !hasPayload || !json.Valid(r.Payload) {
			return errors.New("mutate response requires a valid JSON payload")
		}
	default:
		return fmt.Errorf("unknown filter decision %q", r.Decision)
	}
	return nil
}
