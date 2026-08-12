package mcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	filtercontract "github.com/obot-platform/nanobot/pkg/filter"
)

const mutationDisallowedReason = "mutation not allowed by hook configuration, implicit rejection"

func newMCPFilterRequest(message *Message, direction, method, identifier string, context filtercontract.Context, capabilities filtercontract.Capabilities) (filtercontract.Request, error) {
	payload, err := json.Marshal(message)
	if err != nil {
		return filtercontract.Request{}, fmt.Errorf("failed to marshal MCP Filter payload: %w", err)
	}

	phase := filtercontract.Phase(direction)
	if direction == "response" && message.Error != nil {
		phase = filtercontract.PhaseFailure
	}
	request := filtercontract.Request{
		APIVersion: filtercontract.APIVersionV1,
		Source:     filtercontract.SourceMCP,
		Event: filtercontract.Event{
			Type:       filtercontract.EventTypeMCPMessage,
			Phase:      phase,
			Method:     method,
			Identifier: identifier,
		},
		Context:      context,
		Capabilities: capabilities,
		Payload:      payload,
	}
	if err := request.Validate(); err != nil {
		return filtercontract.Request{}, fmt.Errorf("invalid MCP Filter request: %w", err)
	}
	return request, nil
}

func normalizeV1FilterResponse(response filtercontract.Response, current *Message, identifier string, capabilities filtercontract.Capabilities) (SessionMessageHook, error) {
	validationCapabilities := capabilities
	if response.Decision == filtercontract.DecisionMutate {
		// Validate the replacement payload independently from authorization so a
		// malformed mutation remains an invalid response, not a policy rejection.
		validationCapabilities.CanMutate = true
	}
	if err := response.Validate(validationCapabilities); err != nil {
		return SessionMessageHook{}, fmt.Errorf("invalid v1 Filter response: %w", err)
	}

	switch response.Decision {
	case filtercontract.DecisionAccept:
		return SessionMessageHook{Accept: true, Message: current, Reason: response.Reason, NormalizedDecision: filtercontract.DecisionAccept}, nil
	case filtercontract.DecisionReject:
		return SessionMessageHook{Message: current, Reason: response.Reason, NormalizedDecision: filtercontract.DecisionReject}, nil
	case filtercontract.DecisionMutate:
		trimmed := bytes.TrimSpace(response.Payload)
		if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || trimmed[0] != '{' {
			return SessionMessageHook{}, errors.New("invalid v1 Filter response: mutation payload must be an MCP message object")
		}
		var message Message
		if err := json.Unmarshal(response.Payload, &message); err != nil {
			return SessionMessageHook{}, fmt.Errorf("invalid v1 Filter response: mutation payload must be an MCP message: %w", err)
		}
		if message.JSONRPC != current.JSONRPC || message.Method != current.Method || !reflect.DeepEqual(message.ID, current.ID) || getMessageName(&message) != identifier {
			return SessionMessageHook{}, errors.New("invalid v1 Filter response: mutation changed immutable MCP message identity")
		}
		if !capabilities.CanMutate {
			return SessionMessageHook{
				Accept:             false,
				Message:            current,
				Reason:             appendReason(response.Reason, mutationDisallowedReason),
				NormalizedDecision: filtercontract.DecisionReject,
			}, nil
		}
		return SessionMessageHook{
			Accept:             true,
			Mutated:            true,
			Message:            &message,
			Reason:             response.Reason,
			NormalizedDecision: filtercontract.DecisionMutate,
		}, nil
	default:
		panic("validated Filter response has an unknown decision")
	}
}

func normalizeLegacyFilterResponse(response SessionMessageHook, current *Message, mutateDisallowed bool) (SessionMessageHook, error) {
	if !response.Accept {
		response.Mutated = false
		response.Message = current
		response.NormalizedDecision = filtercontract.DecisionReject
		return response, nil
	}
	if !response.Mutated {
		response.Message = current
		response.NormalizedDecision = filtercontract.DecisionAccept
		return response, nil
	}
	if response.Message == nil {
		return SessionMessageHook{}, errors.New("invalid legacy Filter response: mutation requires a message")
	}
	if mutateDisallowed {
		return SessionMessageHook{
			Accept:             false,
			Message:            current,
			Reason:             appendReason(response.Reason, mutationDisallowedReason),
			NormalizedDecision: filtercontract.DecisionReject,
		}, nil
	}
	response.NormalizedDecision = filtercontract.DecisionMutate
	return response, nil
}

func appendReason(reason, addition string) string {
	if reason == "" {
		return addition
	}
	return reason + "; " + addition
}
