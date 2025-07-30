package agentbuilder

import (
	"gorm.io/gorm"
)

// Agent is a the configuration of a custom agent
type Agent struct {
	gorm.Model
	// UID is a unique identifier for the artifact, typically a UUID
	UUID string `json:"uid" gorm:"uniqueIndex;not null"`
	// SessionID is the ID of the session that created this artifact
	SessionID string `json:"sessionID"`
	// AccountID is the ID of the account that owns this artifact
	AccountID string `json:"accountID" gorm:"index;not null"`
	// Blob is the binary content of the artifact
	Config string `json:"config"`
	// Is an agent that can be shared with others
	IsPublic bool `json:"isPublic"`
	// Name is the name of the agent
	Name string `json:"name"`
	// Description is a human-readable description of the agent
	Description string `json:"description,omitempty"`
}
