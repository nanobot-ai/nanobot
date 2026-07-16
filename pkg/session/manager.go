package session

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
	"github.com/obot-platform/nanobot/pkg/uuid"
	"gorm.io/gorm"
)

const defaultGarbageCollectionInterval = time.Hour

const (
	DefaultLiveSessionIdleTTL = 10 * time.Second
	DefaultMaxLiveSessions    = 4
)

// ManagerOptions configures persisted-session garbage collection and the
// bounded set of live sessions that may own downstream resources.
type ManagerOptions struct {
	DatabaseGarbageCollectionPeriod time.Duration
	LiveSessionIdleTTL              time.Duration
	MaxLiveSessions                 int
}

func (o ManagerOptions) Complete() ManagerOptions {
	if o.LiveSessionIdleTTL <= 0 {
		o.LiveSessionIdleTTL = DefaultLiveSessionIdleTTL
	}
	if o.MaxLiveSessions <= 0 {
		o.MaxLiveSessions = DefaultMaxLiveSessions
	}
	return o
}

func NewManager(ctx context.Context, store *Store, gcPeriod time.Duration) *Manager {
	return NewManagerWithOptions(ctx, store, ManagerOptions{
		DatabaseGarbageCollectionPeriod: gcPeriod,
	})
}

func NewManagerWithOptions(ctx context.Context, store *Store, opts ManagerOptions) *Manager {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	opts = opts.Complete()

	m := &Manager{
		ctx:                ctx,
		cancel:             cancel,
		done:               make(chan struct{}),
		DB:                 store,
		root:               &Session{},
		liveSessions:       make(map[string]liveSession),
		liveSessionIdleTTL: opts.LiveSessionIdleTTL,
		maxLiveSessions:    opts.MaxLiveSessions,
		newServerSession: func(ctx context.Context, state mcp.SessionState, server mcp.MessageHandler) (*mcp.ServerSession, error) {
			return mcp.NewExistingServerSession(ctx, state, server)
		},
	}

	m.startGarbageCollector(opts.DatabaseGarbageCollectionPeriod)
	go m.closeOnShutdown()

	return m
}

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	DB     *Store
	root   *Session

	liveSessionsLock   sync.Mutex
	liveSessions       map[string]liveSession
	pendingSessions    int
	liveSessionIdleTTL time.Duration
	maxLiveSessions    int
	newServerSession   func(context.Context, mcp.SessionState, mcp.MessageHandler) (*mcp.ServerSession, error)
}

type liveSession struct {
	session  *mcp.ServerSession
	count    int
	cancel   context.CancelFunc
	lastUsed time.Time
}

func (m *Manager) newRecord(id, accountID string) *Session {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	return &Session{
		SessionID: id,
		AccountID: accountID,
		Cwd:       cwd,
	}
}

func (m *Manager) loadAttributesFromRecord(stored *Session, session *mcp.ServerSession) {
	session.GetSession().Set(types.DescriptionSessionKey, stored.Description)
	session.GetSession().Set(types.AccountIDSessionKey, stored.AccountID)
	session.GetSession().Set(types.TaskURISessionKey, stored.TaskURI)
}

func (m *Manager) saveAttributesToRecord(stored *Session, session *mcp.ServerSession) error {
	var (
		config  types.Config
		taskURI string
	)

	session.GetSession().Get(types.DescriptionSessionKey, &stored.Description)
	session.GetSession().Get(types.ConfigSessionKey, &config)
	session.GetSession().Get(types.TaskURISessionKey, &taskURI)

	stored.Config = ConfigWrapper(config)
	stored.TaskURI = taskURI
	return nil
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
		if err := m.addLiveSession(session); err != nil {
			if deleteErr := m.DB.Delete(ctx, id); deleteErr != nil && !errors.Is(deleteErr, gorm.ErrRecordNotFound) {
				slog.Error("failed to remove rejected session record", "session_id", id, "error", deleteErr)
			}
			return err
		}
	} else {
		if err := m.DB.Update(ctx, stored); err != nil {
			return err
		}
	}

	m.loadAttributesFromRecord(stored, session)
	return nil
}

func (m *Manager) Reserve(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var evicted *mcp.ServerSession
	m.liveSessionsLock.Lock()
	if err := m.ctx.Err(); err != nil {
		m.liveSessionsLock.Unlock()
		return nil, mcp.ErrSessionStoreClosed
	}

	if len(m.liveSessions)+m.pendingSessions >= m.maxLiveSessions {
		evictionID, candidate := m.leastRecentlyUsedIdleSessionLocked()
		if candidate == nil {
			err := &mcp.SessionCapacityError{
				MaxSessions:     m.maxLiveSessions,
				ActiveSessions:  len(m.liveSessions),
				PendingSessions: m.pendingSessions,
			}
			m.liveSessionsLock.Unlock()
			return nil, err
		}
		delete(m.liveSessions, evictionID)
		if candidate.cancel != nil {
			candidate.cancel()
		}
		evicted = candidate.session
		slog.Info("evicting idle live session to reserve capacity",
			"session_id", evictionID,
			"max_live_sessions", m.maxLiveSessions)
	}

	m.pendingSessions++
	m.liveSessionsLock.Unlock()

	if evicted != nil {
		evicted.Close(true)
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			m.liveSessionsLock.Lock()
			if m.pendingSessions > 0 {
				m.pendingSessions--
			}
			m.liveSessionsLock.Unlock()
		})
	}, nil
}

func (m *Manager) addLiveSession(session *mcp.ServerSession) error {
	var evicted *mcp.ServerSession

	m.liveSessionsLock.Lock()
	if err := m.ctx.Err(); err != nil {
		m.liveSessionsLock.Unlock()
		return mcp.ErrSessionStoreClosed
	}
	if _, exists := m.liveSessions[session.ID()]; exists {
		m.liveSessionsLock.Unlock()
		return fmt.Errorf("live session %q already exists", session.ID())
	}
	if len(m.liveSessions) >= m.maxLiveSessions {
		evictionID, candidate := m.leastRecentlyUsedIdleSessionLocked()
		if candidate == nil {
			err := &mcp.SessionCapacityError{
				MaxSessions:     m.maxLiveSessions,
				ActiveSessions:  len(m.liveSessions),
				PendingSessions: m.pendingSessions,
			}
			m.liveSessionsLock.Unlock()
			return err
		}
		delete(m.liveSessions, evictionID)
		if candidate.cancel != nil {
			candidate.cancel()
		}
		evicted = candidate.session
		slog.Info("evicting idle live session at capacity",
			"session_id", evictionID,
			"max_live_sessions", m.maxLiveSessions)
	}

	m.liveSessions[session.ID()] = liveSession{
		session:  session,
		count:    1,
		lastUsed: time.Now(),
	}
	m.liveSessionsLock.Unlock()

	if evicted != nil {
		evicted.Close(true)
	}
	return nil
}

func (m *Manager) leastRecentlyUsedIdleSessionLocked() (string, *liveSession) {
	var (
		evictionID string
		eviction   *liveSession
	)
	for id, candidate := range m.liveSessions {
		if candidate.count != 0 {
			continue
		}
		if eviction == nil || candidate.lastUsed.Before(eviction.lastUsed) {
			copy := candidate
			evictionID = id
			eviction = &copy
		}
	}
	return evictionID, eviction
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
			delete(m.liveSessions, id)
			if live.cancel != nil {
				live.cancel()
			}
			m.liveSessionsLock.Unlock()
			return nil, false, nil
		default:
		}

		if !checkAccount(ctx, live.session) {
			m.liveSessionsLock.Unlock()
			return nil, false, nil
		}

		if live.cancel != nil {
			live.cancel()
			live.cancel = nil
		}

		live.count++
		live.lastUsed = time.Now()
		m.liveSessions[id] = live
		m.liveSessionsLock.Unlock()
		return live.session, true, nil
	}
	m.liveSessionsLock.Unlock()

	releaseReservation, err := m.Reserve(ctx)
	if err != nil {
		return nil, false, err
	}
	defer releaseReservation()

	serverSession, ok, err := m.loadSessionFromDatabase(ctx, server, id)
	if err != nil || !ok {
		return nil, false, err
	}

	if !checkAccount(ctx, serverSession) {
		serverSession.Close(true)
		return nil, false, nil
	}

	var evicted *mcp.ServerSession
	m.liveSessionsLock.Lock()
	live, ok = m.liveSessions[id]
	if ok {
		serverSession.Close(true)
		if !checkAccount(ctx, live.session) {
			m.liveSessionsLock.Unlock()
			return nil, false, nil
		}
		if live.cancel != nil {
			live.cancel()
			live.cancel = nil
		}
		live.count++
		live.lastUsed = time.Now()
		m.liveSessions[id] = live
		m.liveSessionsLock.Unlock()
		return live.session, true, nil
	}
	if len(m.liveSessions) >= m.maxLiveSessions {
		evictionID, candidate := m.leastRecentlyUsedIdleSessionLocked()
		if candidate == nil {
			err := &mcp.SessionCapacityError{
				MaxSessions:     m.maxLiveSessions,
				ActiveSessions:  len(m.liveSessions),
				PendingSessions: m.pendingSessions,
			}
			m.liveSessionsLock.Unlock()
			serverSession.Close(true)
			return nil, false, err
		}
		delete(m.liveSessions, evictionID)
		if candidate.cancel != nil {
			candidate.cancel()
		}
		evicted = candidate.session
	}
	m.liveSessions[id] = liveSession{
		session:  serverSession,
		count:    1,
		lastUsed: time.Now(),
	}
	m.liveSessionsLock.Unlock()

	if evicted != nil {
		evicted.Close(true)
	}
	return serverSession, true, err
}

func (m *Manager) Release(session *mcp.ServerSession) {
	m.liveSessionsLock.Lock()

	live, ok := m.liveSessions[session.ID()]
	if ok && live.session == session {
		if live.count == 0 {
			m.liveSessionsLock.Unlock()
			return
		}
		live.count--
		live.lastUsed = time.Now()
		if live.count == 0 {
			ctx, cancel := context.WithCancel(m.ctx)
			live.cancel = cancel

			go func(ctx context.Context, sessionID string) {
				defer cancel()
				timer := time.NewTimer(m.liveSessionIdleTTL)
				defer timer.Stop()
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
				}

				m.liveSessionsLock.Lock()
				live, ok := m.liveSessions[sessionID]
				if ok && live.count == 0 {
					delete(m.liveSessions, sessionID)
					m.liveSessionsLock.Unlock()
					live.session.Close(true)
					return
				}
				m.liveSessionsLock.Unlock()
			}(ctx, session.ID())
		} else if live.cancel != nil {
			live.cancel()
			live.cancel = nil
		}

		m.liveSessions[session.ID()] = live
		m.liveSessionsLock.Unlock()
	} else {
		m.liveSessionsLock.Unlock()
		session.Close(true)
	}
}

func (m *Manager) liveSessionIDs() []string {
	m.liveSessionsLock.Lock()
	defer m.liveSessionsLock.Unlock()

	return slices.Collect(maps.Keys(m.liveSessions))
}

func (m *Manager) garbageCollect(maxIdle time.Duration) (int64, error) {
	if maxIdle <= 0 {
		return 0, nil
	}

	cutoff := time.Now().Add(-maxIdle)
	return m.DB.DeleteSessionsUpdatedBefore(m.ctx, cutoff, m.liveSessionIDs()...)
}

func (m *Manager) startGarbageCollector(maxIdle time.Duration) {
	if maxIdle <= 0 {
		return
	}

	go func() {
		if deleted, err := m.garbageCollect(maxIdle); err != nil {
			slog.Error("failed to garbage collect sessions", "max_idle", maxIdle, "error", err)
		} else if deleted > 0 {
			slog.Info("garbage collected sessions", "max_idle", maxIdle, "deleted", deleted)
		}

		ticker := time.NewTicker(defaultGarbageCollectionInterval)
		defer ticker.Stop()
		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
			}

			if deleted, err := m.garbageCollect(maxIdle); err != nil {
				slog.Error("failed to garbage collect sessions", "max_idle", maxIdle, "error", err)
			} else if deleted > 0 {
				slog.Info("garbage collected sessions", "max_idle", maxIdle, "deleted", deleted)
			}
		}
	}()
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

	serverSession, err := m.newServerSession(m.ctx, mcp.SessionState(storedSession.State), server)
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

	m.liveSessionsLock.Lock()
	if live, ok := m.liveSessions[id]; ok && live.session == session {
		delete(m.liveSessions, id)
		if live.cancel != nil {
			live.cancel()
		}
	}
	m.liveSessionsLock.Unlock()

	err = m.DB.Delete(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		session.Close(true)
		return nil, false, nil
	} else if err != nil {
		session.Close(true)
		return nil, false, fmt.Errorf("failed to delete session: %w", err)
	}
	return session, true, nil
}

func (m *Manager) Close() error {
	m.cancel()
	<-m.done
	return nil
}

func (m *Manager) closeOnShutdown() {
	defer close(m.done)
	<-m.ctx.Done()

	var sessions []*mcp.ServerSession
	m.liveSessionsLock.Lock()
	sessions = make([]*mcp.ServerSession, 0, len(m.liveSessions))
	for id, live := range m.liveSessions {
		delete(m.liveSessions, id)
		if live.cancel != nil {
			live.cancel()
		}
		sessions = append(sessions, live.session)
	}
	m.liveSessionsLock.Unlock()

	for _, session := range sessions {
		session.Close(true)
	}
}
