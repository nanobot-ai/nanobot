package session

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"gorm.io/gorm"
)

func NewManager(dsn string) (*Manager, error) {
	store, err := NewStoreFromDSN(dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:          ctx,
		close:        cancel,
		DB:           store,
		root:         &Session{},
		liveSessions: make(map[string]liveSession),
	}, nil
}

type Manager struct {
	ctx   context.Context
	close context.CancelFunc
	DB    *Store
	root  *Session

	liveSessionsLock sync.Mutex
	liveSessions     map[string]liveSession

	eventSubscribersLock sync.RWMutex
	eventSubscribers     map[uint64]func(SessionEvent)
	nextSubscriberID     uint64
}

type SessionEventType string

const (
	SessionEventCreated SessionEventType = "created"
	SessionEventDeleted SessionEventType = "deleted"
	SessionEventUpdated SessionEventType = "updated"
)

// SessionEvent describes a persisted thread lifecycle change emitted by Manager.
type SessionEvent struct {
	Type        SessionEventType
	SessionType string
	SessionID   string
	AccountID   string
}

type liveSession struct {
	session *mcp.ServerSession
	count   int
	cancel  context.CancelFunc
}

func defaultSessionCwd(sessionID string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, "sessions", types.SanitizeSessionDirectoryName(sessionID))
}

func (m *Manager) newRecord(id, accountID string) *Session {
	return &Session{
		SessionID: id,
		AccountID: accountID,
		Cwd:       defaultSessionCwd(id),
	}
}

func (m *Manager) loadAttributesFromRecord(stored *Session, session *mcp.ServerSession) {
	if stored.Cwd == "" {
		stored.Cwd = defaultSessionCwd(session.ID())
	}
	session.GetSession().Set(types.DescriptionSessionKey, stored.Description)
	session.GetSession().Set(types.AccountIDSessionKey, stored.AccountID)
	session.GetSession().Set(types.CwdSessionKey, stored.Cwd)
	session.GetSession().Set(types.WorkflowURIsSessionKey, stored.WorkflowURIs)
}

func (m *Manager) saveAttributesToRecord(stored *Session, session *mcp.ServerSession) error {
	var (
		config types.Config
		cwd    string
	)

	session.GetSession().Get(types.DescriptionSessionKey, &stored.Description)
	session.GetSession().Get(types.ConfigSessionKey, &config)
	session.GetSession().Get(types.CwdSessionKey, &cwd)
	session.GetSession().Get(types.WorkflowURIsSessionKey, &stored.WorkflowURIs)
	if cwd == "" {
		cwd = defaultSessionCwd(session.ID())
	}
	stored.Cwd = cwd

	stored.Config = ConfigWrapper(config)
	return nil
}

// SubscribeEvents registers a callback for session create/update/delete events.
// The returned function unregisters the handler.
func (m *Manager) SubscribeEvents(handler func(SessionEvent)) (unsubscribe func()) {
	if handler == nil {
		return func() {}
	}

	id := atomic.AddUint64(&m.nextSubscriberID, 1)

	m.eventSubscribersLock.Lock()
	if m.eventSubscribers == nil {
		m.eventSubscribers = make(map[uint64]func(SessionEvent))
	}
	m.eventSubscribers[id] = handler
	m.eventSubscribersLock.Unlock()

	return func() {
		m.eventSubscribersLock.Lock()
		delete(m.eventSubscribers, id)
		m.eventSubscribersLock.Unlock()
	}
}

func (m *Manager) emitEvent(event SessionEvent) {
	m.eventSubscribersLock.RLock()
	if len(m.eventSubscribers) == 0 {
		m.eventSubscribersLock.RUnlock()
		return
	}

	handlers := make([]func(SessionEvent), 0, len(m.eventSubscribers))
	for _, handler := range m.eventSubscribers {
		handlers = append(handlers, handler)
	}
	m.eventSubscribersLock.RUnlock()

	for _, handler := range handlers {
		func() {
			defer func() {
				if recoverErr := recover(); recoverErr != nil {
					log.Errorf(m.ctx, "panic while handling session event: %v", recoverErr)
				}
			}()
			handler(event)
		}()
	}
}

func (m *Manager) Store(ctx context.Context, id string, session *mcp.ServerSession) error {
	if id == "" {
		return nil
	}

	var accountID string
	session.GetSession().Get(types.AccountIDSessionKey, &accountID)

	var create bool
	stored, err := m.DB.Get(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		stored = m.newRecord(id, accountID)
		create = true
	} else if err != nil {
		return err
	}

	if stored.AccountID != accountID {
		return fmt.Errorf("session %s not found for account %s", id, accountID)
	}

	previousDescription := stored.Description

	if err := m.saveAttributesToRecord(stored, session); err != nil {
		return fmt.Errorf("failed to save attributes to session record: %w", err)
	}

	state, err := session.GetSession().State()
	if err != nil {
		return fmt.Errorf("failed to get session state: %w", err)
	}
	stored.State = *(*State)(state)

	if create {
		if err := m.DB.Create(ctx, stored); err != nil {
			return fmt.Errorf("failed to create session record: %w", err)
		}
		m.emitEvent(SessionEvent{
			Type:        SessionEventCreated,
			SessionType: stored.Type,
			SessionID:   stored.SessionID,
			AccountID:   stored.AccountID,
		})

		m.liveSessionsLock.Lock()
		live, ok := m.liveSessions[id]
		if ok {
			if live.session != nil {
				live.session.Close(false)
			}
			live.count++
			live.session = session

			m.liveSessions[id] = live
		} else {
			m.liveSessions[id] = liveSession{
				session: session,
				count:   1,
			}
		}
		m.liveSessionsLock.Unlock()
	} else {
		if err := m.DB.Update(ctx, stored); err != nil {
			return err
		}

		if previousDescription != stored.Description {
			sessionType := stored.Type
			if sessionType == "" {
				sessionType = "thread"
			}

			// Description changes are surfaced as resource updates for thread UIs.
			m.emitEvent(SessionEvent{
				Type:        SessionEventUpdated,
				SessionType: sessionType,
				SessionID:   stored.SessionID,
				AccountID:   stored.AccountID,
			})
		}
	}

	m.loadAttributesFromRecord(stored, session)
	return nil
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

		if uuid.ValidUUID(part) {
			return part
		}
	}
	return ""
}

func checkAccount(ctx context.Context, serverSession *mcp.ServerSession) bool {
	var (
		account        string
		nanobotContext = types.NanobotContext(ctx)
	)
	serverSession.GetSession().Get(types.AccountIDSessionKey, &account)
	return account == nanobotContext.User.ID
}

func (m *Manager) Acquire(ctx context.Context, server mcp.MessageHandler, id string) (ret *mcp.ServerSession, found bool, retErr error) {
	m.liveSessionsLock.Lock()
	live, ok := m.liveSessions[id]
	if ok {
		select {
		case <-live.session.GetSession().Context().Done():
			m.liveSessionsLock.Unlock()
			return nil, false, nil
		default:
		}

		if !checkAccount(ctx, live.session) {
			m.liveSessionsLock.Unlock()
			return nil, false, nil
		}

		live.count++
		m.liveSessions[id] = live
		m.liveSessionsLock.Unlock()
		return live.session, true, nil
	}
	m.liveSessionsLock.Unlock()

	serverSession, ok, err := m.loadSessionFromDatabase(ctx, server, id)
	if err != nil || !ok {
		return nil, false, err
	}

	if !checkAccount(ctx, serverSession) {
		return nil, false, nil
	}

	m.liveSessionsLock.Lock()
	live, ok = m.liveSessions[id]
	if ok {
		serverSession.Close(false)
		live.count++
		m.liveSessions[id] = live
		m.liveSessionsLock.Unlock()
		return live.session, true, nil
	}
	m.liveSessions[id] = liveSession{
		session: serverSession,
		count:   1,
	}
	m.liveSessionsLock.Unlock()

	return serverSession, true, err
}

func (m *Manager) Release(session *mcp.ServerSession) {
	m.liveSessionsLock.Lock()
	defer m.liveSessionsLock.Unlock()

	live, ok := m.liveSessions[session.ID()]
	if ok {
		live.count--
		if live.count == 0 {
			ctx, cancel := context.WithCancel(context.Background())
			live.cancel = cancel

			go func(ctx context.Context, sessionID string) {
				defer cancel()
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Second):
				}

				m.liveSessionsLock.Lock()
				defer m.liveSessionsLock.Unlock()

				live, ok := m.liveSessions[sessionID]
				if ok && live.count == 0 {
					delete(m.liveSessions, sessionID)
					live.session.Close(false)
				}
			}(ctx, session.ID())
		} else if live.cancel != nil {
			live.cancel()
			live.cancel = nil
		}

		m.liveSessions[session.ID()] = live
	} else {
		session.Close(false)
	}
}

func (m *Manager) loadSessionFromDatabase(ctx context.Context, server mcp.MessageHandler, id string) (*mcp.ServerSession, bool, error) {
	storedSession, err := m.DB.Get(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	if storedSession.State.Attributes == nil {
		storedSession.State.Attributes = make(map[string]any)
	} else {
		storedSession.State.Attributes[".keys"] = slices.Collect(maps.Keys(storedSession.State.Attributes))
	}

	serverSession, err := mcp.NewExistingServerSession(m.ctx,
		mcp.SessionState(storedSession.State), server)
	if err != nil {
		return nil, false, err
	}

	m.loadAttributesFromRecord(storedSession, serverSession)
	return serverSession, true, nil
}

func (m *Manager) LoadAndDelete(ctx context.Context, server mcp.MessageHandler, id string) (*mcp.ServerSession, bool, error) {
	session, found, err := m.Acquire(ctx, server, id)
	if !found || err != nil {
		return session, found, err
	}
	defer m.Release(session)

	stored, err := m.DB.Get(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("failed to load session before delete: %w", err)
	}

	err = m.DB.Delete(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("failed to delete session: %w", err)
	}

	m.emitEvent(SessionEvent{
		Type:        SessionEventDeleted,
		SessionType: stored.Type,
		SessionID:   stored.SessionID,
		AccountID:   stored.AccountID,
	})
	return session, true, nil
}

func (m *Manager) UpdateThreadDescription(ctx context.Context, id, accountID, description string) (*Session, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("session ID cannot be empty")
	}
	if accountID == "" {
		return nil, false, fmt.Errorf("account ID cannot be empty")
	}

	stored, err := m.DB.GetByIDByAccountID(ctx, id, accountID)
	if err != nil {
		return nil, false, err
	}
	if stored.Type != "" && stored.Type != "thread" {
		return nil, false, fmt.Errorf("session %s is not a thread", id)
	}
	if stored.Description == description {
		// No-op updates should not write to DB or trigger notifications.
		return stored, false, nil
	}

	stored.Description = description
	if err := m.DB.Update(ctx, stored); err != nil {
		return nil, false, err
	}

	m.liveSessionsLock.Lock()
	if live, ok := m.liveSessions[id]; ok && live.session != nil {
		live.session.GetSession().Set(types.DescriptionSessionKey, description)
	}
	m.liveSessionsLock.Unlock()

	m.emitEvent(SessionEvent{
		Type:        SessionEventUpdated,
		SessionType: "thread",
		SessionID:   stored.SessionID,
		AccountID:   stored.AccountID,
	})

	return stored, true, nil
}
