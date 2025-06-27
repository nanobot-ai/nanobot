package mcp

import (
	"context"
	"errors"

	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

var _ wire = (*serverWire)(nil)

func NewServerSession(ctx context.Context, handler MessageHandler) (*ServerSession, error) {
	return NewExistingServerSession(ctx,
		SessionState{
			ID: uuid.String(),
		}, handler)
}

func NewExistingServerSession(ctx context.Context, state SessionState, handler MessageHandler) (*ServerSession, error) {
	s := &serverWire{
		read:      make(chan Message),
		sessionID: state.ID,
	}
	session, err := newSession(ctx, s, handler, &state, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range state.Attributes {
		session.Set(k, v)
	}
	return &ServerSession{
		session: session,
		wire:    s,
	}, nil
}

type ServerSession struct {
	session *Session
	wire    *serverWire
}

func (s *ServerSession) ID() string {
	id := s.session.ID()
	if id == "" {
		return s.wire.SessionID()
	}
	return id
}

var ErrNoResponse = errors.New("no response")

func (s *ServerSession) GetSession() *Session {
	return s.session
}

func (s *ServerSession) Exchange(ctx context.Context, msg Message) (Message, error) {
	isInit, err := s.session.preInit(&msg)
	if err != nil {
		return Message{}, err
	}
	resp, err := s.wire.exchange(ctx, msg)
	if err != nil {
		return Message{}, err
	}
	if isInit {
		if err := s.session.postInit(&resp); err != nil {
			return Message{}, err
		}
	}
	return resp, nil
}

func (s *ServerSession) Read(ctx context.Context) (Message, bool) {
	select {
	case msg, ok := <-s.wire.read:
		if !ok {
			return Message{}, false
		}
		return msg, true
	case <-ctx.Done():
		return Message{}, false
	}
}

func (s *ServerSession) Send(ctx context.Context, req Message) error {
	return s.wire.Send(ctx, req)
}

func (s *ServerSession) Close() {
	s.session.Close()
}

type serverWire struct {
	ctx       context.Context
	cancel    context.CancelFunc
	pending   PendingRequests
	read      chan Message
	handler   wireHandler
	sessionID string
}

func (s *serverWire) SessionID() string {
	return s.sessionID
}

func (s *serverWire) exchange(ctx context.Context, msg Message) (Message, error) {
	ch := s.pending.WaitFor(msg.ID)
	defer s.pending.Done(msg.ID)

	go func() {
		s.handler(msg)
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return Message{}, ctx.Err()
	case <-s.ctx.Done():
		return Message{}, s.ctx.Err()
	case m, ok := <-ch:
		if !ok {
			return Message{}, ErrNoResponse
		}
		return m, nil
	}
}

func (s *serverWire) Close() {
	s.cancel()
}

func (s *serverWire) Wait() {
	<-s.ctx.Done()
}

func (s *serverWire) Start(ctx context.Context, handler wireHandler) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.handler = handler
	return nil
}

func (s *serverWire) Send(ctx context.Context, req Message) error {
	if s.pending.Notify(req) {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return s.ctx.Err()
	case s.read <- req:
		return nil
	}
}
