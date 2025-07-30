package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/agentbuilder"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"gorm.io/gorm"
)

func NewManager(server mcp.MessageHandler, dsn string, config types.Config) (*Manager, error) {
	store, err := NewStoreFromDSN(dsn)
	if err != nil {
		return nil, err
	}

	agentsStore, err := agentbuilder.NewStoreFromDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create agents store: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:         ctx,
		close:       cancel,
		server:      server,
		store:       store,
		agentsStore: agentsStore,
		config:      config,
		root:        &Session{},
		inMemory:    mcp.NewInMemorySessionStore(),
	}, nil
}

type Manager struct {
	ctx         context.Context
	close       context.CancelFunc
	server      mcp.MessageHandler
	store       *Store
	agentsStore *agentbuilder.Store
	root        *Session
	config      types.Config
	inMemory    mcp.SessionStore
}

func (m *Manager) newRecord(parent *Session, id string) *Session {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	return &Session{
		Type:      "thread",
		SessionID: id,
		ParentID:  parent.SessionID,
		Cwd:       cwd,
	}
}

func (m *Manager) setupAgentUUID(req *http.Request, stored *Session) string {
	agentID := m.getAgent(req)
	if agentID == "" {
		return ""
	}

	agent, err := m.agentsStore.GetByUUID(m.ctx, agentID)
	if err != nil {
		// ignore all errors related to agent not found
		return ""
	}

	if agent.AccountID != stored.AccountID && !agent.IsPublic {
		return ""
	}

	return agentID
}

func (m *Manager) newSession(req *http.Request, stored *Session) {
	stored.AccountID = m.getAccount(req)
	stored.AgentUUID = m.setupAgentUUID(req, stored)
	stored.Config = ConfigWrapper(m.config)
}

func (m *Manager) loadAttributesFromRecord(stored *Session, session *mcp.ServerSession) {
	session.GetSession().Set(types.DescriptionSessionKey, stored.Description)
	session.GetSession().Set(types.PublicSessionKey, stored.IsPublic)
	session.GetSession().Set(types.AccountIDSessionKey, stored.AccountID)
	session.GetSession().Set(types.AgentUUIDSessionKey, stored.AgentUUID)

	var (
		agentConfig types.CustomAgent
		config      = m.config
	)

	if stored.AgentUUID != "" {
		agent, err := m.agentsStore.GetByUUID(m.ctx, stored.AgentUUID)
		if err == nil && agent != nil {
			if err := json.Unmarshal([]byte(agent.Config), &agentConfig); err == nil {
				agentConfig.ID = agent.UUID
				agentConfig.Name = agent.Name
				agentConfig.Description = agent.Description
				agentConfig.IsPublic = agent.IsPublic
				session.GetSession().Set(types.CustomAgentConfigSessionKey, &agentConfig)
			}
		}
	}

	session.GetSession().Set(types.ConfigSessionKey, config)
}

func (m *Manager) saveAttributesToRecord(ctx context.Context, stored *Session, session *mcp.ServerSession) error {
	session.GetSession().Get(types.DescriptionSessionKey, &stored.Description)
	session.GetSession().Get(types.PublicSessionKey, &stored.IsPublic)
	stored.Config = ConfigWrapper(m.config)

	if updated := false; session.GetSession().Get(types.CustomAgentModifiedSessionKey, &updated) && updated {
		var agentConfig types.CustomAgent
		if session.GetSession().Get(types.CustomAgentConfigSessionKey, &agentConfig) && agentConfig.ID != "" {
			agentRecord, err := m.agentsStore.GetByUUID(ctx, agentConfig.ID)
			if err != nil {
				return fmt.Errorf("failed to get agent by UUID: %w", err)
			}

			agentRecord.Name = agentConfig.Name
			agentRecord.Description = agentConfig.Description
			agentRecord.IsPublic = agentConfig.IsPublic

			// zero out
			agentConfig.CustomAgentMeta = types.CustomAgentMeta{}
			configData, err := json.Marshal(agentConfig)
			if err != nil {
				return fmt.Errorf("failed to marshal agent config: %w", err)
			}
			agentRecord.Config = string(configData)
			if err := m.agentsStore.UpdateConfig(ctx, agentRecord); err != nil {
				return fmt.Errorf("failed to update agent config: %w", err)
			}
		}
	}

	return nil
}

func (m *Manager) getAccount(req *http.Request) string {
	return req.Header.Get("X-Nanobot-Account-Id")
}

func (m *Manager) getAgent(req *http.Request) string {
	agent := req.Header.Get("X-Nanobot-Agent-Id")
	if agent != "" {
		return agent
	}

	parts := strings.Split(req.URL.Path, "/")
	for i, part := range parts {
		if len(strings.Split(part, "-")) == 5 && i > 0 && parts[i-1] == "agents" {
			return part
		}
	}

	return agent
}

func (m *Manager) Store(req *http.Request, id string, session *mcp.ServerSession) error {
	if id == "" {
		return nil
	}

	var create bool
	stored, err := m.store.Get(req.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		stored = m.newRecord(m.root, id)
		create = true
	} else if err != nil {
		return err
	}

	if create {
		m.newSession(req, stored)
	}

	if stored.AccountID != m.getAccount(req) {
		return fmt.Errorf("session %s not found for account %s", id, m.getAccount(req))
	}

	if err := m.saveAttributesToRecord(req.Context(), stored, session); err != nil {
		return fmt.Errorf("failed to save attributes to session record: %w", err)
	}

	state, err := session.GetSession().State()
	if err != nil {
		return fmt.Errorf("failed to get session state: %w", err)
	}
	stored.State = *(*State)(state)

	if create {
		if err := m.store.Create(req.Context(), stored); err != nil {
			return fmt.Errorf("failed to create session record: %w", err)
		}
	} else {
		if err := m.store.Update(req.Context(), stored); err != nil {
			return err
		}
	}

	m.loadAttributesFromRecord(stored, session)
	return m.inMemory.Store(nil, id, session)
}

func (m *Manager) ExtractID(req *http.Request) string {
	id := req.Header.Get("Mcp-Session-Id")
	if id != "" {
		return id
	}
	id = req.Header.Get("X-Nanobot-Session-Id")
	if id != "" {
		return id
	}
	parts := strings.Split(req.URL.Path, "/")
	for i, part := range parts {
		if i > 0 && parts[i-1] == "agents" {
			continue
		}
		if len(strings.Split(part, "-")) == 5 {
			return part
		}
	}
	return ""
}

func (m *Manager) Load(req *http.Request, id string) (ret *mcp.ServerSession, found bool, retErr error) {
	defer func() {
		if found && ret != nil {
			var account string
			ret.GetSession().Get(types.AccountIDSessionKey, &account)
			if account != m.getAccount(req) {
				var isPublic bool
				ret.GetSession().Get(types.PublicSessionKey, &isPublic)
				if isPublic {
					ret, found, retErr = m.loadSessionFromDatabase(req, id)
					if found && retErr == nil {
						ret.GetSession().Set(types.AccountIDSessionKey, m.getAccount(req))
					}
				} else {
					found = false
					ret = nil
				}
			}
		}

		if found && ret != nil {
			ret.GetSession().Set(StoreSessionKey, m.store)
		}
	}()

	if id == "new" {
		s, err := mcp.NewServerSession(m.ctx, m.server)
		if err != nil {
			return nil, false, fmt.Errorf("failed to create new server session: %w", err)
		}
		if err := m.Store(req, s.ID(), s); err != nil {
			return nil, false, fmt.Errorf("failed to store new server session: %w", err)
		}
		return s, true, nil
	}

	session, ok, err := m.inMemory.Load(req, id)
	if err != nil {
		return nil, false, err
	} else if ok {
		// Check if closed
		select {
		case <-session.GetSession().Context().Done():
		default:
			return session, true, nil
		}
	}

	serverSession, ok, err := m.loadSessionFromDatabase(req, id)
	if err != nil || !ok {
		return nil, false, err
	}

	err = m.inMemory.Store(nil, id, serverSession)
	return serverSession, true, err
}

func (m *Manager) loadSessionFromDatabase(req *http.Request, id string) (*mcp.ServerSession, bool, error) {
	storedSession, err := m.store.Get(req.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	if storedSession.State.Attributes == nil {
		storedSession.State.Attributes = make(map[string]any)
	}

	serverSession, err := mcp.NewExistingServerSession(m.ctx,
		mcp.SessionState(storedSession.State), m.server)
	if err != nil {
		return nil, false, err
	}

	m.loadAttributesFromRecord(storedSession, serverSession)
	return serverSession, true, nil
}

func (m *Manager) LoadAndDelete(request *http.Request, id string) (*mcp.ServerSession, bool, error) {
	session, found, err := m.Load(request, id)
	if !found || err != nil {
		return session, found, err
	}
	_, _, _ = m.inMemory.LoadAndDelete(request, id)
	err = m.store.Delete(request.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("failed to delete session: %w", err)
	}
	return session, true, nil
}
