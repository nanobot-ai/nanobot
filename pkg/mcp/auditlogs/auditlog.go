package auditlogs

import (
	"encoding/json"
	"time"
)

// MCPAuditLog represents an audit log entry for MCP API calls
type MCPAuditLog struct {
	// Metadata is additional information about this server that a user can provide for audit log tracking purposes.
	// For example Obot uses this to track catalog information.
	Metadata         map[string]string  `json:"metadata,omitempty"`
	CreatedAt        time.Time          `json:"createdAt"`
	Subject          string             `json:"subject"`
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
