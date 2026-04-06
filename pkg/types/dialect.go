package types

import "strings"

type Dialect string

const (
	DialectAnthropicMessages Dialect = "AnthropicMessages"
	DialectOpenResponses     Dialect = "OpenResponses"
	DialectDefault                   = DialectOpenResponses
)

// DialectFromString returns the Dialect matching s (case-insensitive).
// Returns an empty Dialect if s is empty or unrecognized.
func DialectFromString(s string) Dialect {
	switch strings.ToLower(s) {
	case strings.ToLower(string(DialectAnthropicMessages)):
		return DialectAnthropicMessages
	case strings.ToLower(string(DialectOpenResponses)):
		return DialectOpenResponses
	default:
		return Dialect(s)
	}
}
