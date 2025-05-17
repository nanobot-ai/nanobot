package mcp

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var _ wire = (*serverWire)(nil)

func newServerSession(ctx context.Context, handler MessageHandler) (*serverSession, error) {
	s := &serverWire{
		read: make(chan Message),
	}
	id := uuid.New().String()
	session, err := newSession(ctx, s, handler, id, nil)
	if err != nil {
		return nil, err
	}
	return &serverSession{
		session: session,
		wire:    s,
	}, nil
}

type serverSession struct {
	session *Session
	wire    *serverWire
}

var ErrNoResponse = errors.New("no response")

func (s *serverSession) Exchange(ctx context.Context, msg Message) (Message, error) {
	return s.wire.exchange(ctx, msg)
}

func (s *serverSession) Read(ctx context.Context) (Message, bool) {
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

func (s *serverSession) Send(ctx context.Context, req Message) error {
	return s.wire.Send(ctx, req)
}

type serverWire struct {
	ctx     context.Context
	cancel  context.CancelFunc
	pending pendingRequest
	read    chan Message
	handler wireHandler
}

func (s *serverWire) exchange(ctx context.Context, msg Message) (Message, error) {
	ch := s.pending.waitFor(msg.ID)
	defer s.pending.done(msg.ID)

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
	if s.pending.notify(req) {
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
