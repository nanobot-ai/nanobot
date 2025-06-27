package session

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/server"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"gorm.io/gorm"
)

func NewManager(server *server.Server, dsn string) (*Manager, error) {
	store, err := NewStoreFromDSN(dsn)
	if err != nil {
		return nil, err
	}

	return &Manager{
		server: server,
		store:  store,
		root: &Session{
			Config: ConfigWrapper(server.DefaultRuntime.GetConfig()),
		},
		inMemory: mcp.NewInMemorySessionStore(),
	}, nil
}

type Manager struct {
	server   *server.Server
	store    *Store
	root     *Session
	inMemory mcp.SessionStore
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
		Config:    parent.Config,
		Cwd:       cwd,
	}
}

func (m *Manager) Store(_ *http.Request, id string, session *mcp.ServerSession) error {
	var setRuntime bool

	if id == "" {
		return nil
	}

	stored, err := m.store.Get(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setRuntime = true
		stored = m.newRecord(m.root, id)
		err = m.store.Create(stored)
		if err != nil {
			return err
		}
	}

	state := (*State)(session.GetSession().State())
	stored.State = *state
	stored.Env, _ = state.Attributes[mcp.SessionEnvMapKey].(map[string]string)

	delete(stored.State.Attributes, mcp.SessionEnvMapKey)
	delete(stored.State.Attributes, runtime.SessionKey)

	if err := m.store.Update(stored); err != nil {
		return err
	}

	if setRuntime {
		newRuntime := m.newRuntime(m.server.DefaultRuntime.GetConfig())
		session.GetSession().Set(runtime.SessionKey, newRuntime)
	}

	return m.inMemory.Store(nil, id, session)
}

func (m *Manager) newRuntime(cfg types.Config) *runtime.Runtime {
	newRuntime := *m.server.DefaultRuntime
	newRuntime.Reload(cfg)
	return &newRuntime
}

func (m *Manager) Load(req *http.Request, id string) (*mcp.ServerSession, bool, error) {
	session, ok, err := m.inMemory.Load(req, id)
	if err != nil {
		return nil, false, err
	} else if ok {
		return session, true, nil
	}

	storedSession, err := m.store.Get(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	if storedSession.State.Attributes == nil {
		storedSession.State.Attributes = make(map[string]any)
	}
	storedSession.State.Attributes[mcp.SessionEnvMapKey] = (map[string]string)(storedSession.Env)

	serverSession, err := mcp.NewExistingServerSession(context.Background(),
		mcp.SessionState(storedSession.State), m.server)
	if err != nil {
		return nil, false, err
	}

	serverSession.GetSession().Set(runtime.SessionKey, m.newRuntime(types.Config(storedSession.Config)))
	err = m.inMemory.Store(nil, id, serverSession)
	return serverSession, true, err
}

func (m *Manager) LoadAndDelete(request *http.Request, id string) (*mcp.ServerSession, bool, error) {
	return nil, true, nil
}
