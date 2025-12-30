package auditlogs

import (
	"encoding/json"
	"strings"
	"time"
)

// MCPAuditLog represents an audit log entry for MCP API calls
type MCPAuditLog struct {
	// Metadata is additional information about this server that a user can provide for audit log tracking purposes.
	// For example Obot uses this to track catalog information.
	Metadata         map[string]string  `json:"metadata,omitempty"`
	CreatedAt        time.Time          `json:"createdAt"`
	Subject          string             `json:"subject"`
	APIKey           string             `json:"apiKey,omitempty"`
	ClientName       string             `json:"clientName"`
	ClientVersion    string             `json:"clientVersion"`
	ClientIP         string             `json:"clientIP"`
	CallType         string             `json:"callType"`
	CallIdentifier   string             `json:"callIdentifier,omitempty"`
	RequestBody      json.RawMessage    `json:"requestBody,omitempty"`
	ResponseBody     json.RawMessage    `json:"responseBody,omitempty"`
	ResponseStatus   int                `json:"responseStatus"`
	Error            string             `json:"error,omitempty"`
	ProcessingTimeMs int64              `json:"processingTimeMs"`
	SessionID        string             `json:"sessionID,omitempty"`
	WebhookStatuses  []MCPWebhookStatus `json:"webhookStatuses,omitempty"`

	// Additional metadata
	RequestID       string          `json:"requestID,omitempty"`
	UserAgent       string          `json:"userAgent,omitempty"`
	RequestHeaders  json.RawMessage `json:"requestHeaders,omitempty"`
	ResponseHeaders json.RawMessage `json:"responseHeaders,omitempty"`
}

type MCPWebhookStatus struct {
	Type    string `json:"type,omitempty"`
	Method  string `json:"method,omitempty"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// RedactAPIKey redacts an API key in the format ok1-<user ID>-<key ID>-<secret>
// to ok1-<user ID>-<key ID>-*****
func RedactAPIKey(apiKey string) string {
	// API keys have the format: ok1-<user ID>-<key ID>-<secret>
	if !strings.HasPrefix(apiKey, "ok1-") {
		return ""
	}

	parts := strings.SplitN(apiKey, "-", 4)
	if len(parts) != 4 {
		return ""
	}

	// Return redacted version: ok1-<user ID>-<key ID>-*****
	return parts[0] + "-" + parts[1] + "-" + parts[2] + "-*****"
}
