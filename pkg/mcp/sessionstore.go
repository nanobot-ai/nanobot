package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	// DefaultInMemorySessionIdleTTL is how long an unreferenced session remains
	// available before the in-memory store closes it.
	DefaultInMemorySessionIdleTTL = 30 * time.Minute
	// DefaultInMemorySessionReapInterval controls how frequently the in-memory
	// store looks for expired sessions.
	DefaultInMemorySessionReapInterval = time.Minute
	// DefaultInMemorySessionMaxSessions bounds the number of sessions retained
	// by a default in-memory store.
	DefaultInMemorySessionMaxSessions = 32
)

var (
	// ErrSessionCapacity indicates that a bounded session store has no idle
	// session available for eviction.
	ErrSessionCapacity = errors.New("session store capacity reached")
	// ErrSessionStoreClosed indicates that the store has begun shutting down.
	ErrSessionStoreClosed = errors.New("session store is closed")
)

// SessionCapacityError describes a failed attempt to store a session while
// every retained session is acquired.
type SessionCapacityError struct {
	MaxSessions     int
	ActiveSessions  int
	PendingSessions int
}

func (e *SessionCapacityError) Error() string {
	return fmt.Sprintf("%v: %d active and %d pending sessions (limit %d)",
		ErrSessionCapacity, e.ActiveSessions, e.PendingSessions, e.MaxSessions)
}

func (e *SessionCapacityError) Unwrap() error {
	return ErrSessionCapacity
}

type SessionStore interface {
	ExtractID(*http.Request) string
	Store(context.Context, string, *ServerSession) error
	Acquire(context.Context, MessageHandler, string) (*ServerSession, bool, error)
	Release(*ServerSession)
	LoadAndDelete(context.Context, MessageHandler, string) (*ServerSession, bool, error)
}

// SessionStoreReserver allows an HTTP server to reserve capacity before it
// initializes a session and creates any session-scoped downstream resources.
// The returned release function must be called exactly once on every path.
type SessionStoreReserver interface {
	Reserve(context.Context) (release func(), err error)
}

// InMemorySessionStoreOptions configures in-memory session lifetime and
// capacity. All zero values receive bounded production defaults.
type InMemorySessionStoreOptions struct {
	BaseContext  context.Context
	IdleTTL      time.Duration
	ReapInterval time.Duration
	MaxSessions  int
}

func (o InMemorySessionStoreOptions) Complete() InMemorySessionStoreOptions {
	if o.BaseContext == nil {
		o.BaseContext = context.Background()
	}
	if o.IdleTTL <= 0 {
		o.IdleTTL = DefaultInMemorySessionIdleTTL
	}
	if o.ReapInterval <= 0 {
		o.ReapInterval = DefaultInMemorySessionReapInterval
	}
	if o.MaxSessions <= 0 {
		o.MaxSessions = DefaultInMemorySessionMaxSessions
	}
	return o
}

func (o InMemorySessionStoreOptions) Merge(other InMemorySessionStoreOptions) InMemorySessionStoreOptions {
	if other.BaseContext != nil {
		o.BaseContext = other.BaseContext
	}
	if other.IdleTTL != 0 {
		o.IdleTTL = other.IdleTTL
	}
	if other.ReapInterval != 0 {
		o.ReapInterval = other.ReapInterval
	}
	if other.MaxSessions != 0 {
		o.MaxSessions = other.MaxSessions
	}
	return o
}

type inMemorySession struct {
	session    *ServerSession
	references int
	lastUsed   time.Time
}

type InMemorySessionStore struct {
	cancel context.CancelFunc
	done   chan struct{}

	mu          sync.Mutex
	closed      bool
	sessions    map[string]*inMemorySession
	pending     int
	idleTTL     time.Duration
	maxSessions int
}

// NewInMemorySessionStore preserves the original constructor while applying
// bounded defaults. Use NewInMemorySessionStoreWithOptions to tune them.
func NewInMemorySessionStore() SessionStore {
	return NewInMemorySessionStoreWithOptions(InMemorySessionStoreOptions{})
}

// NewInMemorySessionStoreWithOptions creates an in-memory session store and
// starts its idle-session reaper. Close the returned store, or cancel
// BaseContext, to stop the reaper and close all retained sessions.
func NewInMemorySessionStoreWithOptions(opts InMemorySessionStoreOptions) *InMemorySessionStore {
	o := opts.Complete()

	ctx, cancel := context.WithCancel(o.BaseContext)
	store := &InMemorySessionStore{
		cancel:      cancel,
		done:        make(chan struct{}),
		sessions:    make(map[string]*inMemorySession),
		idleTTL:     o.IdleTTL,
		maxSessions: o.MaxSessions,
	}
	go store.runReaper(ctx, o.ReapInterval)
	return store
}

func (s *InMemorySessionStore) ExtractID(req *http.Request) string {
	return req.Header.Get("Mcp-Session-Id")
}

func (s *InMemorySessionStore) Reserve(_ context.Context) (func(), error) {
	var evicted *ServerSession

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, ErrSessionStoreClosed
	}

	if len(s.sessions)+s.pending >= s.maxSessions {
		var (
			evictionID string
			eviction   *inMemorySession
		)
		for id, candidate := range s.sessions {
			if candidate.references != 0 {
				continue
			}
			if eviction == nil || candidate.lastUsed.Before(eviction.lastUsed) {
				evictionID = id
				eviction = candidate
			}
		}
		if eviction == nil {
			err := &SessionCapacityError{
				MaxSessions:     s.maxSessions,
				ActiveSessions:  len(s.sessions),
				PendingSessions: s.pending,
			}
			s.mu.Unlock()
			return nil, err
		}
		delete(s.sessions, evictionID)
		evicted = eviction.session
		slog.Info("evicting idle MCP session to reserve in-memory store capacity",
			"session_id", evictionID,
			"max_sessions", s.maxSessions)
	}

	s.pending++
	s.mu.Unlock()

	if evicted != nil {
		evicted.Close(true)
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			s.mu.Lock()
			if s.pending > 0 {
				s.pending--
			}
			s.mu.Unlock()
		})
	}, nil
}

func (s *InMemorySessionStore) Store(_ context.Context, sessionID string, session *ServerSession) error {
	if session == nil {
		return errors.New("session is nil")
	}
	if sessionID == "" {
		return errors.New("session ID is empty")
	}
	if session.ID() != sessionID {
		return fmt.Errorf("session ID %q does not match stored session ID %q", sessionID, session.ID())
	}

	now := time.Now()
	var toClose []*ServerSession

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSessionStoreClosed
	}

	if existing, ok := s.sessions[sessionID]; ok {
		if existing.session == session {
			existing.lastUsed = now
			s.mu.Unlock()
			return nil
		}
		if existing.references > 0 {
			s.mu.Unlock()
			return fmt.Errorf("session %q is already acquired by another instance", sessionID)
		}
		delete(s.sessions, sessionID)
		toClose = append(toClose, existing.session)
	}

	if len(s.sessions) >= s.maxSessions {
		var (
			evictionID string
			eviction   *inMemorySession
		)
		for id, candidate := range s.sessions {
			if candidate.references != 0 {
				continue
			}
			if eviction == nil || candidate.lastUsed.Before(eviction.lastUsed) {
				evictionID = id
				eviction = candidate
			}
		}
		if eviction == nil {
			activeSessions := len(s.sessions)
			s.mu.Unlock()
			for _, replaced := range toClose {
				replaced.Close(true)
			}
			return &SessionCapacityError{
				MaxSessions:     s.maxSessions,
				ActiveSessions:  activeSessions,
				PendingSessions: s.pending,
			}
		}
		delete(s.sessions, evictionID)
		toClose = append(toClose, eviction.session)
		slog.Info("evicting idle MCP session at in-memory store capacity",
			"session_id", evictionID,
			"max_sessions", s.maxSessions)
	}

	s.sessions[sessionID] = &inMemorySession{
		session:    session,
		references: 1,
		lastUsed:   now,
	}
	s.mu.Unlock()

	for _, replacedOrEvicted := range toClose {
		replacedOrEvicted.Close(true)
	}
	return nil
}

func (s *InMemorySessionStore) Acquire(_ context.Context, _ MessageHandler, sessionID string) (*ServerSession, bool, error) {
	now := time.Now()
	var toClose *ServerSession

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, false, ErrSessionStoreClosed
	}

	entry, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return nil, false, nil
	}

	if !entry.session.GetSession().IsActive() ||
		(entry.references == 0 && now.Sub(entry.lastUsed) >= s.idleTTL) {
		delete(s.sessions, sessionID)
		toClose = entry.session
		s.mu.Unlock()
		toClose.Close(true)
		return nil, false, nil
	}

	entry.references++
	entry.lastUsed = now
	session := entry.session
	s.mu.Unlock()
	return session, true, nil
}

func (s *InMemorySessionStore) Release(session *ServerSession) {
	if session == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.sessions[session.ID()]
	if !ok || entry.session != session || entry.references == 0 {
		return
	}

	entry.references--
	if entry.references == 0 {
		entry.lastUsed = time.Now()
	}
}

func (s *InMemorySessionStore) LoadAndDelete(_ context.Context, _ MessageHandler, sessionID string) (*ServerSession, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, false, ErrSessionStoreClosed
	}

	entry, ok := s.sessions[sessionID]
	if !ok {
		return nil, false, nil
	}
	delete(s.sessions, sessionID)
	return entry.session, true, nil
}

func (s *InMemorySessionStore) Close() error {
	s.cancel()
	<-s.done
	return nil
}

func (s *InMemorySessionStore) runReaper(ctx context.Context, interval time.Duration) {
	defer close(s.done)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.closeAll()
			return
		case now := <-ticker.C:
			s.reap(now)
		}
	}
}

func (s *InMemorySessionStore) reap(now time.Time) {
	var expired []*ServerSession

	s.mu.Lock()
	if !s.closed {
		for id, entry := range s.sessions {
			if entry.references == 0 && now.Sub(entry.lastUsed) >= s.idleTTL {
				delete(s.sessions, id)
				expired = append(expired, entry.session)
			}
		}
	}
	s.mu.Unlock()

	if len(expired) > 0 {
		slog.Info("reaping idle MCP sessions", "count", len(expired), "idle_ttl", s.idleTTL)
	}
	for _, session := range expired {
		session.Close(true)
	}
}

func (s *InMemorySessionStore) closeAll() {
	var sessions []*ServerSession

	s.mu.Lock()
	if !s.closed {
		s.closed = true
		sessions = make([]*ServerSession, 0, len(s.sessions))
		for id, entry := range s.sessions {
			delete(s.sessions, id)
			sessions = append(sessions, entry.session)
		}
	}
	s.mu.Unlock()

	for _, session := range sessions {
		session.Close(true)
	}
}
