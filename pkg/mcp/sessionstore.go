package mcp

import (
	"sync"
)

type SessionStore interface {
	Store(string, *ServerSession)
	Load(string) (*ServerSession, bool)
	LoadAndDelete(string) (*ServerSession, bool)
}

type inMemory struct {
	sessions sync.Map
}

func NewInMemorySessionStore() SessionStore {
	return &inMemory{}
}

func (s *inMemory) Store(sessionID string, session *ServerSession) {
	s.sessions.Store(sessionID, session)
}

func (s *inMemory) Load(sessionID string) (*ServerSession, bool) {
	if v, ok := s.sessions.Load(sessionID); ok {
		return v.(*ServerSession), true
	}
	return nil, false
}

func (s *inMemory) LoadAndDelete(sessionID string) (*ServerSession, bool) {
	if v, ok := s.sessions.LoadAndDelete(sessionID); ok {
		return v.(*ServerSession), true
	}
	return nil, false
}
