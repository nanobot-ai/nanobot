package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp/auditlogs"
	"github.com/tidwall/gjson"
)

type HookRunner interface {
	RunHook(ctx context.Context, msg, target string) (bool, string, *Message, error)
}

type Hooks map[HookDefinition][]string

func (h *Hooks) UnmarshalJSON(data []byte) error {
	var m map[string][]string
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	*h = make(Hooks, len(m))
	for key, value := range m {
		def, err := parseHookDefinition(key)
		if err != nil {
			return fmt.Errorf("failed to parse hook definition %s: %w", key, err)
		}

		(*h)[def] = value
	}

	return nil
}

func (h *Hooks) MarshalJSON() ([]byte, error) {
	if h == nil {
		return nil, nil
	}

	m := make(map[string][]string, len(*h))
	for def, targets := range *h {
		m[def.String()] = targets
	}

	return json.Marshal(m)
}

func (h *Hooks) CallAllHooks(ctx context.Context, r HookRunner, msg *Message, direction string, err error) error {
	if h == nil {
		h = MCPServerConfigFromContext(ctx).Hooks
		if h == nil {
			return nil
		}

		ctx = WithMCPServerConfig(ctx, Server{})
	}

	var name string
	switch msg.Method {
	case "resources/read", "resources/subscribe", "resources/unsubscribe":
		name = gjson.GetBytes(msg.Params, "uri").String()
	case "tools/call", "prompts/get":
		name = gjson.GetBytes(msg.Params, "name").String()
	default:
	}

	hookDef := HookDefinition{
		Type:        "message",
		Method:      msg.Method,
		Name:        name,
		Direction:   direction,
		CallOnError: err != nil || msg.Error != nil,
	}

	message, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	var (
		accepted bool
		reason   string
		m        *Message
		errs     []error
		statuses []auditlogs.MCPWebhookStatus
	)
	for def, targets := range *h {
		if def.Matches(hookDef) {
			for _, target := range targets {
				if accepted, reason, m, err = r.RunHook(ctx, string(message), target); err != nil {
					errs = append(errs, fmt.Errorf("failed to run hook %s: %w", def, err))
				} else if !accepted {
					errs = append(errs, fmt.Errorf("hook %s rejected message: %s", def, reason))
				} else if m != nil {
					*msg = *m
				}

				status := "ok"
				if !accepted {
					status = "rejected"
				}
				statuses = append(statuses, auditlogs.MCPWebhookStatus{
					Type:    direction,
					Method:  msg.Method,
					Name:    target,
					Status:  status,
					Message: reason,
				})
			}
		}
	}

	if auditLog := AuditLogFromContext(ctx); auditLog != nil {
		auditLog.WebhookStatuses = append(auditLog.WebhookStatuses, statuses...)
		if len(errs) > 0 {
			auditLog.ResponseStatus = http.StatusFailedDependency
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to run hooks: %v", errs)
	}

	return nil
}

type HookDefinition struct {
	Type        string
	Method      string
	Name        string
	Direction   string
	CallOnError bool
}

func parseHookDefinition(data string) (HookDefinition, error) {
	if !strings.HasPrefix(data, "message:") {
		return HookDefinition{}, fmt.Errorf("invalid hook definition")
	}

	parts := strings.Split(data, "?")
	if len(parts) > 2 {
		return HookDefinition{}, fmt.Errorf("invalid hook definition, too many '?'")
	}

	var values url.Values
	if len(parts) > 1 {
		var err error
		values, err = url.ParseQuery(parts[1])
		if err != nil {
			return HookDefinition{}, fmt.Errorf("failed to parse hook parameters: %w", err)
		}
	}

	parts = strings.Split(parts[0], ":")
	var method string
	if len(parts) > 1 {
		method = parts[1]
	}

	name := values.Get("name")
	direction := values.Get("direction")
	if direction != "" && direction != "request" && direction != "response" {
		return HookDefinition{}, fmt.Errorf("invalid direction value: %s", direction)
	}

	var callOnError bool
	if onError := values.Get("onError"); onError != "false" && onError != "no" {
		callOnError = true
	}

	return HookDefinition{
		Type:        "message",
		Method:      method,
		Name:        name,
		Direction:   direction,
		CallOnError: callOnError,
	}, nil
}

// Matches will indicate if the hook definition applies to the given hook definition.
// It is assumed that the `other` hook definition is fully defined.
func (h HookDefinition) Matches(other HookDefinition) bool {
	if h.Type != "" && h.Type != other.Type {
		return false
	}
	if h.Method != "" && h.Method != "*" && h.Method != other.Method {
		return false
	}
	if h.Direction != "" && h.Direction != other.Direction {
		return false
	}
	if h.Name != "" && h.Name != other.Name {
		return false
	}
	if !h.CallOnError && other.CallOnError {
		return false
	}
	return true
}

func (h HookDefinition) String() string {
	s := strings.Builder{}
	s.WriteString(h.Type)
	s.WriteString(":")
	s.WriteString(h.Method)
	s.WriteString("?")

	if h.Name != "" {
		s.WriteString("name=")
		s.WriteString(h.Name)
	}

	if h.Direction != "" {
		s.WriteString("&direction=")
		s.WriteString(h.Direction)
	}

	if !h.CallOnError {
		s.WriteString("&onError=false")
	}

	return strings.TrimSuffix(s.String(), "?")
}

type HookResponse struct {
	Accepted   bool     `json:"accepted"`
	Message    string   `json:"message"`
	Modified   bool     `json:"modified"`
	NewMessage *Message `json:"newMessage,omitempty"`
}
