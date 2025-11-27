package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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

func (h Hooks) CallAllHooks(ctx context.Context, r HookRunner, msg *Message, name, direction string, err error) error {
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
	)
	for def, targets := range h {
		if def.Matches(hookDef) {
			for _, target := range targets {
				if accepted, reason, m, err = r.RunHook(ctx, string(message), target); err != nil {
					return fmt.Errorf("failed to run hook %s: %w", def, err)
				} else if !accepted {
					return fmt.Errorf("hook %s rejected message: %s", def, reason)
				} else if m != nil {
					*msg = *m
				}
			}
		}
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
	if direction != "" && direction != "in" && direction != "out" {
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
