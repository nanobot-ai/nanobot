package session

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"gorm.io/gorm"
)

type ConfigWrapper types.Config

func (c ConfigWrapper) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfigWrapper) Scan(value interface{}) error {
	return scan(value, c)
}

type Env map[string]string

func (e Env) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *Env) Scan(value interface{}) error {
	return scan(value, e)
}

type State mcp.SessionState

func (m State) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *State) Scan(value interface{}) error {
	return scan(value, m)
}

func scan(value interface{}, obj any) error {
	if value == nil {
		return nil
	}
	if data, ok := value.([]byte); ok {
		return json.Unmarshal(data, obj)
	}
	if data, ok := value.(string); ok {
		return json.Unmarshal([]byte(data), obj)
	}
	return fmt.Errorf("cannot scan %T into %T", value, obj)
}

type Session struct {
	gorm.Model
	Type        string        `json:"type,omitempty"`
	SessionID   string        `json:"sessionID" gorm:"uniqueIndex;not null"`
	Description string        `json:"description,omitempty"`
	AccountID   string        `json:"accountID,omitempty"`
	State       State         `json:"state" gorm:"type:json"`
	Config      ConfigWrapper `json:"config,omitempty" gorm:"type:json"`
	Cwd         string        `json:"cwd,omitempty"`
	IsPublic    bool          `json:"isPublic"`
}

type Token struct {
	gorm.Model
	AccountID string `json:"accountID,omitempty"`
	URL       string `json:"url,omitempty"`
	Data      string `json:"data,omitempty"`
}

func (s *Session) Clone(accountID string) *Session {
	newSession := *s
	newSession.SessionID = uuid.String()
	newSession.AccountID = accountID
	newSession.IsPublic = false
	newSession.Model = gorm.Model{}
	newSession.State.ID = newSession.SessionID
	newSession.State.Attributes = make(map[string]any, len(s.State.Attributes))
	for k, v := range s.State.Attributes {
		if k == types.CurrentAgentSessionKey || strings.HasPrefix(k, "thread") ||
			k == types.ConfigSessionKey {
			newSession.State.Attributes[k] = v
		}
	}
	return &newSession
}
