package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/obot-platform/nanobot/pkg/log"
)

type MessageHandler interface {
	OnMessage(ctx context.Context, msg Message) error
}

type MessageHandlerFunc func(ctx context.Context, msg Message) error

func (f MessageHandlerFunc) OnMessage(ctx context.Context, msg Message) error {
	return f(ctx, msg)
}

type wire interface {
	Close()
	Wait()
	Start(ctx context.Context, handler wireHandler) error
	Send(ctx context.Context, req Message) error
}

type wireHandler func(msg Message)

type pendingRequest struct {
	lock sync.Mutex
	ids  map[any]chan Message
}

func (p *pendingRequest) waitFor(id any) chan Message {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.ids == nil {
		p.ids = make(map[any]chan Message)
	}
	ch := make(chan Message, 1)
	p.ids[id] = ch
	return ch
}

func (p *pendingRequest) notify(msg Message) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	ch, ok := p.ids[msg.ID]
	if ok {
		select {
		case ch <- msg:
			return true
			// don't let it block, we are holding the lock
		default:
		}
		delete(p.ids, msg.ID)
	}
	return false
}

func (p *pendingRequest) done(id any) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.ids, id)
}

var sessionKey = struct{}{}

func SessionFromContext(ctx context.Context) *Session {
	if ctx == nil {
		return nil
	}
	s, ok := ctx.Value(sessionKey).(*Session)
	if !ok {
		return nil
	}
	return s
}

func WithSession(ctx context.Context, s *Session) context.Context {
	if s == nil {
		return ctx
	}
	return context.WithValue(ctx, sessionKey, s)
}

type Session struct {
	ctx                context.Context
	cancel             context.CancelFunc
	wire               wire
	handler            MessageHandler
	pendingRequest     pendingRequest
	ClientCapabilities *ClientCapabilities
	recorder           *recorder
	sessionID          string
	attributes         map[string]any
}

func (s *Session) Set(key string, value any) {
	if s == nil {
		return
	}
	if s.attributes == nil {
		s.attributes = make(map[string]any)
	}
	s.attributes[key] = value
}

func (s *Session) Get(key string) any {
	if s == nil {
		return nil
	}
	return s.attributes[key]
}

func (s *Session) ID() string {
	if s == nil {
		return ""
	}
	return s.sessionID
}

func (s *Session) Close() {
	if s.wire != nil {
		s.wire.Close()
	}
	s.cancel()
}

func (s *Session) Wait() {
	if s.wire == nil {
		<-s.ctx.Done()
		return
	}
	s.wire.Wait()
}

func (s *Session) Send(ctx context.Context, req Message) error {
	if s.wire == nil {
		return fmt.Errorf("empty session: wire is not initialized")
	}

	req.JSONRPC = "2.0"
	if req.Method == "initialize" {
		var init InitializeRequest
		if err := json.Unmarshal(req.Params, &init); err != nil {
			return fmt.Errorf("failed to unmarshal initialize request: %w", err)
		}
		s.ClientCapabilities = &init.Capabilities
	}
	s.recorder.save(ctx, s.sessionID, true, req)
	return s.wire.Send(ctx, req)
}

type ExchangeOption struct {
	ID            string
	ProgressToken any
}

func completeExchangeOption(opts ...ExchangeOption) ExchangeOption {
	var opt ExchangeOption
	for _, o := range opts {
		if o.ID != "" {
			opt.ID = o.ID
		}
		if o.ProgressToken != nil {
			opt.ProgressToken = o.ProgressToken
		}
	}
	return opt
}

func (s *Session) Exchange(ctx context.Context, method string, in, out any, opts ...ExchangeOption) error {
	opt := completeExchangeOption(opts...)
	req, ok := in.(*Message)
	if !ok {
		data, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		req = &Message{
			ID:     opt.ID,
			Method: method,
			Params: data,
		}
	}

	if req.ID == nil || req.ID == "" {
		req.ID = uuid.New().String()
	}
	if opt.ProgressToken != nil {
		if err := req.SetProgressToken(opt.ProgressToken); err != nil {
			return fmt.Errorf("failed to set progress token: %w", err)
		}
	}

	ch := s.pendingRequest.waitFor(req.ID)
	defer s.pendingRequest.done(req.ID)

	if err := s.Send(ctx, *req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case m := <-ch:
		if mOut, ok := out.(*Message); ok {
			*mOut = m
			return nil
		}
		if m.Error != nil {
			return fmt.Errorf("error from server: %s", m.Error.Message)
		}
		if m.Result == nil {
			return fmt.Errorf("no result in response")
		}
		if err := json.Unmarshal(m.Result, out); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
		return nil
	}
}

func (s *Session) onWire(message Message) {
	s.recorder.save(s.ctx, s.sessionID, false, message)
	message.Session = s
	if s.pendingRequest.notify(message) {
		return
	}
	if err := s.handler.OnMessage(s.ctx, message); err != nil {
		if message.IsRequest() {
			message.SendError(s.ctx, 500, err.Error(), nil)
		} else {
			log.Errorf(s.ctx, "failed to handle message %s: %v", message.ID, err)
		}
	}
	return
}

func NewEmptySession(ctx context.Context, sessionID string) *Session {
	s := &Session{
		sessionID: sessionID,
	}
	s.ctx, s.cancel = context.WithCancel(WithSession(ctx, s))
	return s
}

func newSession(ctx context.Context, wire wire, handler MessageHandler, sessionID string, r *recorder) (*Session, error) {
	s := &Session{
		wire:      wire,
		handler:   handler,
		sessionID: sessionID,
		recorder:  r,
	}
	s.ctx, s.cancel = context.WithCancel(WithSession(ctx, s))
	return s, wire.Start(s.ctx, s.onWire)
}

type recorder struct {
}

func (r *recorder) save(ctx context.Context, sessionID string, send bool, msg Message) {
}
