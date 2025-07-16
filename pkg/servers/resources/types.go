package resources

import (
	"gorm.io/gorm"
)

// Resource represents a MCP Resource
type Resource struct {
	gorm.Model
	// UID is a unique identifier for the artifact, typically a UUID
	UUID string `json:"uid" gorm:"uniqueIndex;not null"`
	// SessionID is the ID of the session that created this artifact
	SessionID string `json:"sessionID"`
	// AccountID is the ID of the account that owns this artifact
	AccountID string `json:"accountID" gorm:"index;not null"`
	// Blob is the binary content of the artifact
	Blob string `json:"blob"`
	// MimeType the mime type of the content
	MimeType    string `json:"mimeType,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description"`
}
