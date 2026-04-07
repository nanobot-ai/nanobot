package types

import "strings"

type Dialect string

const (
	DialectAnthropicMessages Dialect = "AnthropicMessages"
	DialectOpenAIResponses   Dialect = "OpenAIResponses"
	DialectOpenResponses     Dialect = "OpenResponses"
	DialectDefault                   = DialectOpenAIResponses
)

// DialectFromString returns the Dialect matching s (case-insensitive).
// Returns an empty Dialect if s is empty or unrecognized.
func DialectFromString(s string) Dialect {
	switch strings.ToLower(s) {
	case strings.ToLower(string(DialectAnthropicMessages)):
		return DialectAnthropicMessages
	case strings.ToLower(string(DialectOpenAIResponses)):
		return DialectOpenAIResponses
	case strings.ToLower(string(DialectOpenResponses)):
		return DialectOpenResponses
	default:
		return Dialect(s)
	}
}
