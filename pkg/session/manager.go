package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"gorm.io/gorm"
)

func NewManager(server mcp.MessageHandler, dsn string, config types.Config) (*Manager, error) {
	store, err := NewStoreFromDSN(dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:      ctx,
		close:    cancel,
		server:   server,
		store:    store,
		config:   config,
		root:     &Session{},
		inMemory: mcp.NewInMemorySessionStore(),
	}, nil
}

type Manager struct {
	ctx      context.Context
	close    context.CancelFunc
	server   mcp.MessageHandler
	store    *Store
	root     *Session
	config   types.Config
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
		Cwd:       cwd,
	}
}

func (m *Manager) Store(req *http.Request, id string, session *mcp.ServerSession) error {
	if id == "" {
		return nil
	}

	stored, err := m.store.Get(req.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		stored = m.newRecord(m.root, id)
		err = m.store.Create(req.Context(), stored)
		if err != nil {
			return err
		}
	}

	if !session.GetSession().Get(types.ConfigSessionKey, &types.Config{}) {
		session.GetSession().Set(types.ConfigSessionKey, setupConfig(m.config))
	}

	state, err := session.GetSession().State()
	if err != nil {
		return fmt.Errorf("failed to get session state: %w", err)
	}
	stored.State = *(*State)(state)

	if err := m.store.Update(req.Context(), stored); err != nil {
		return err
	}

	return m.inMemory.Store(nil, id, session)
}

func setupConfig(config types.Config) *types.Config {
	if _, hasMain := config.Agents["main"]; !slices.Contains(config.Publish.MCPServers, "__meta") && (config.Publish.Entrypoint != "" || hasMain) {
		config.Publish.MCPServers = append(config.Publish.MCPServers, "__meta")
		if config.MCPServers == nil {
			config.MCPServers = make(map[string]mcp.Server)
		}
		config.MCPServers["__meta"] = mcp.Server{}
	}
	return &config
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
	for _, part := range strings.Split(req.URL.Path, "/") {
		parts := strings.Split(part, "-")
		if len(parts) == 5 {
			return part
		}
	}
	return ""
}

func (m *Manager) Load(req *http.Request, id string) (ret *mcp.ServerSession, found bool, _ error) {
	defer func() {
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

	err = m.inMemory.Store(nil, id, serverSession)
	return serverSession, true, err
}

func (m *Manager) LoadAndDelete(request *http.Request, id string) (*mcp.ServerSession, bool, error) {
	return nil, true, nil
}
