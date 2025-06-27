package mcp

import (
	"net/http"
	"sync"
)

type SessionStore interface {
	Store(*http.Request, string, *ServerSession) error
	Load(*http.Request, string) (*ServerSession, bool, error)
	LoadAndDelete(*http.Request, string) (*ServerSession, bool, error)
}

type inMemory struct {
	sessions sync.Map
}

func NewInMemorySessionStore() SessionStore {
	return &inMemory{}
}

func (s *inMemory) Store(_ *http.Request, sessionID string, session *ServerSession) error {
	s.sessions.Store(sessionID, session)
	return nil
}

func (s *inMemory) Load(_ *http.Request, sessionID string) (*ServerSession, bool, error) {
	if v, ok := s.sessions.Load(sessionID); ok {
		return v.(*ServerSession), true, nil
	}
	return nil, false, nil
}

func (s *inMemory) LoadAndDelete(_ *http.Request, sessionID string) (*ServerSession, bool, error) {
	if v, ok := s.sessions.LoadAndDelete(sessionID); ok {
		return v.(*ServerSession), true, nil
	}
	return nil, false, nil
}
