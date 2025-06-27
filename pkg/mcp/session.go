package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"sync"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

type MessageHandler interface {
	OnMessage(ctx context.Context, msg Message)
}

type MessageHandlerFunc func(ctx context.Context, msg Message)

func (f MessageHandlerFunc) OnMessage(ctx context.Context, msg Message) {
	f(ctx, msg)
}

type wire interface {
	Close()
	Wait()
	Start(ctx context.Context, handler wireHandler) error
	Send(ctx context.Context, req Message) error
	SessionID() string
}

type wireHandler func(msg Message)

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
	ctx               context.Context
	cancel            context.CancelFunc
	wire              wire
	handler           MessageHandler
	pendingRequest    PendingRequests
	InitializeResult  InitializeResult
	InitializeRequest InitializeRequest
	recorder          *recorder
	Parent            *Session
	attributes        map[string]any
	lock              sync.Mutex
}

const SessionEnvMapKey = "env"

func (s *Session) ID() string {
	if s == nil || s.wire == nil {
		return ""
	}
	return s.wire.SessionID()
}

func (s *Session) State() *SessionState {
	if s == nil {
		return nil
	}
	return &SessionState{
		ID:                s.wire.SessionID(),
		InitializeResult:  s.InitializeResult,
		InitializeRequest: s.InitializeRequest,
		Attributes:        s.Attributes(),
	}
}
func (s *Session) EnvMap() map[string]string {
	if s == nil {
		return map[string]string{}
	}
	if s.attributes == nil {
		s.attributes = make(map[string]any)
	}

	env, ok := s.attributes[SessionEnvMapKey].(map[string]string)
	if !ok {
		env = make(map[string]string)
		s.attributes[SessionEnvMapKey] = env
	}
	return env
}

func (s *Session) Set(key string, value any) {
	if s == nil {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.attributes == nil {
		s.attributes = make(map[string]any)
	}
	s.attributes[key] = value
}

func (s *Session) Get(key string, out any) bool {
	if s == nil {
		return false
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	v, ok := s.attributes[key]
	if !ok {
		return false
	}
	if out == nil {
		return true
	}

	dstVal := reflect.ValueOf(out)
	srcVal := reflect.ValueOf(v)
	if srcVal.Type().AssignableTo(dstVal.Type()) {
		reflect.Indirect(dstVal).Set(reflect.Indirect(srcVal))
		return true
	}

	switch v := v.(type) {
	case string:
		if outStr, ok := out.(*string); ok {
			*outStr = v
			return true
		}
		if err := json.Unmarshal([]byte(v), out); err != nil {
			delete(s.attributes, key)
			return false
		}
		s.attributes[key] = out
	default:
		data, err := json.Marshal(v)
		if err != nil {
			delete(s.attributes, key)
			return false
		}
		if err := json.Unmarshal(data, out); err != nil {
			delete(s.attributes, key)
			return false
		}
		s.attributes[key] = out
	}

	return true
}

func (s *Session) Attributes() map[string]any {
	if s == nil || len(s.attributes) == 0 {
		return nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	return maps.Clone(s.attributes)
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

func (s *Session) normalizeProgress(progress *NotificationProgressRequest) {
	var (
		progressKey               = fmt.Sprintf("progress-token:%v", progress.ProgressToken)
		lastProgress, newProgress float64
	)

	if ok := s.Get(progressKey, &lastProgress); !ok {
		lastProgress = 0
	}

	if progress.Progress != "" {
		newF, err := progress.Progress.Float64()
		if err == nil {
			newProgress = newF
		}
	}

	if newProgress <= lastProgress {
		if progress.Total == nil {
			newProgress = lastProgress + 1
		} else {
			// If total is set then something is probably trying to make the progress pretty
			// so we don't want to just increment by 1 and mess that up.
			newProgress = lastProgress + 0.01
		}
	}
	progress.Progress = json.Number(fmt.Sprintf("%f", newProgress))
	s.Set(progressKey, newProgress)
}

func (s *Session) SendPayload(ctx context.Context, method string, payload any) error {
	if progress, ok := payload.(NotificationProgressRequest); ok {
		s.normalizeProgress(&progress)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return s.Send(ctx, Message{
		Method: method,
		Params: data,
	})
}

func (s *Session) Send(ctx context.Context, req Message) error {
	if s.wire == nil {
		return fmt.Errorf("empty session: wire is not initialized")
	}

	req.JSONRPC = "2.0"
	s.recorder.save(ctx, s.wire.SessionID(), true, req)
	return s.wire.Send(ctx, req)
}

type ExchangeOption struct {
	ProgressToken any
}

func (e ExchangeOption) Merge(other ExchangeOption) (result ExchangeOption) {
	result.ProgressToken = complete.Last(e.ProgressToken, other.ProgressToken)
	return
}

func (s *Session) preInit(msg *Message) (bool, error) {
	if msg.Method == "initialize" {
		var init InitializeRequest
		if err := json.Unmarshal(msg.Params, &init); err != nil {
			return false, fmt.Errorf("failed to unmarshal initialize request: %w", err)
		}
		s.InitializeRequest = init
		return true, nil
	}

	return false, nil
}

func (s *Session) postInit(msg *Message) error {
	var init InitializeResult
	if err := json.Unmarshal(msg.Result, &init); err != nil {
		return fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}
	s.InitializeResult = init
	return nil
}

func (s *Session) Exchange(ctx context.Context, method string, in, out any, opts ...ExchangeOption) error {
	opt := complete.Complete(opts...)
	req, ok := in.(*Message)
	if !ok {
		data, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		req = &Message{
			Method: method,
			Params: data,
		}
	}

	if req.ID == nil || req.ID == "" {
		req.ID = uuid.String()
	}
	if opt.ProgressToken != nil {
		if err := req.SetProgressToken(opt.ProgressToken); err != nil {
			return fmt.Errorf("failed to set progress token: %w", err)
		}
	}

	ch := s.pendingRequest.WaitFor(req.ID)
	defer s.pendingRequest.Done(req.ID)

	isInit, err := s.preInit(req)
	if err != nil {
		return err
	}

	if err := s.Send(ctx, *req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case m := <-ch:
		if mOut, ok := out.(*Message); ok {
			*mOut = m

			if isInit {
				return s.postInit(mOut)
			}
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
	s.recorder.save(s.ctx, s.wire.SessionID(), false, message)
	message.Session = s
	if s.pendingRequest.Notify(message) {
		return
	}
	s.handler.OnMessage(s.ctx, message)
}

func NewEmptySession(ctx context.Context) *Session {
	s := &Session{}
	s.ctx, s.cancel = context.WithCancel(WithSession(ctx, s))
	return s
}

func newSession(ctx context.Context, wire wire, handler MessageHandler, session *SessionState, r *recorder) (*Session, error) {
	s := &Session{
		wire:     wire,
		handler:  handler,
		recorder: r,
	}
	if session != nil {
		s.InitializeRequest = session.InitializeRequest
		s.InitializeResult = session.InitializeResult
	}
	s.ctx, s.cancel = context.WithCancel(WithSession(ctx, s))
	return s, wire.Start(s.ctx, s.onWire)
}

type recorder struct {
}

func (r *recorder) save(ctx context.Context, sessionID string, send bool, msg Message) {
}
